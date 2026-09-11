// Command tune runs a batch of seeded settlements and prints how they came
// out, for judging one set of priors against another. A single seed is a
// coin toss; this is the tool that says whether a change helped.
package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"sort"

	"lreat/core/action"
	"lreat/core/clock"
	"lreat/core/entity"
	"lreat/core/event"
	"lreat/core/need"
	"lreat/core/observe"
	"lreat/core/system"
	"lreat/core/world"
	"lreat/report"
)

type row struct {
	seed                        uint64
	pop, deaths, births, houses int
	fields                      int
	needs                       [5]float64
	order                       float64
}

// terms is what every settlement in a batch is founded on. One seed differs
// from another in its seed and in nothing else, which is the whole point of
// a batch.
type terms struct {
	cfg    world.Config
	fit    bool
	temp   float64
	agents int
	ticks  int
}

// found is how one seed came out: the row it puts in the table, what its
// people spent their days on, and how many of them were fed and safe and
// held. It is what a worker hands back, and nothing in it is shared with
// any other worker.
type found struct {
	// nth is which of the batch's seeds this is, so that a finding can go
	// back to its place in the table whatever order it comes home in.
	nth   int
	row   row
	acts  map[string]int
	gate  [6]float64
	gateN float64
}

func main() {
	seeds := flag.Int("seeds", 24, "seeds to run")
	offset := flag.Int("offset", 0, "first seed minus one, for an independent batch")
	ticks := flag.Int("ticks", 60*clock.Year, "days per run")
	agents := flag.Int("agents", 20, "starting population")
	deer := flag.Int("deer", 0, "deer put down in the woods around each settlement (none: the baseline has no creatures in it)")
	value := flag.Bool("value", false, "use the value rule")
	born := flag.Float64("born", action.BornNoise, "drift on a founder's habits")
	inherit := flag.Float64("inherit", action.InheritNoise, "drift on a child's habits")
	temp := flag.Float64("temp", world.DefaultRules().Temperature, "recognition temperature")
	cap := flag.Int("cap", system.MaxPopulation, "population ceiling; the guard on the machine, not a fact about the world (0 takes it off, and a runaway seed then runs as long as the machine bears it)")
	quiet := flag.Bool("quiet", false, "summary only")
	preset := flag.String("preset", "", "the terms to found each world on: valley (the default map), globe, or ancient (a valley made out of its own history); -width, -height and -wrap override it")
	width := flag.Int("width", world.DefaultWidth, "map width")
	height := flag.Int("height", world.DefaultHeight, "map height")
	wrap := flag.Bool("wrap", false, "join the east edge to the west")
	together := flag.Int("together", runtime.NumCPU(), "how many of the seeds to run at the same time; lower it for a batch of globes, where every world in flight is a couple of hundred megabytes")
	flag.Parse()
	cfg, ok := world.Preset(*preset)
	if !ok {
		fmt.Fprintf(os.Stderr, "no such preset: %q\n", *preset)
		os.Exit(2)
	}
	flag.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "width":
			cfg.Width = *width
		case "height":
			cfg.Height = *height
		case "wrap":
			cfg.Wrap = *wrap
		}
	})
	cfg.Deer = *deer
	action.BornNoise = *born
	action.InheritNoise = *inherit
	system.MaxPopulation = *cap
	world.Workers = 1 // the seeds are the parallelism here

	// The batch is kept as well as printed; see package report. -quiet is
	// about the terminal and not about the record, so the table and the
	// tally still go into the file when the terminal is only given the
	// summary line — a batch is worth its few minutes twice over if the
	// rows can be read a week later.
	rep := report.Open("tune")
	out := rep.Out(os.Stdout)
	defer func() { fmt.Print(rep.Close()) }()
	full := out
	if *quiet {
		full = rep
	}

	// A batch is a few dozen settlements and each of them is minutes of
	// work, so the seeds are run side by side - but not all at once. A globe
	// is seventy megabytes of ground before anybody is standing on it and a
	// couple of hundred once it is running, and a goroutine for every seed
	// held every one of those worlds in memory at the same moment: eight
	// globes took 1.35GB, and the forty a real batch would want could not
	// have been asked for at all.
	//
	// Through a fixed few workers instead, what a batch costs the machine is
	// what the machine has cores for rather than what the batch has seeds.
	// Those same forty run through four at a time in 1.0GB, and -seeds can
	// now be raised as far as patience allows without asking the machine for
	// anything more.
	at := min(*together, *seeds)
	if at < 1 {
		at = 1
	}
	t := terms{cfg: cfg, fit: !*value, temp: *temp, agents: *agents, ticks: *ticks}
	jobs := make(chan int)
	done := make(chan found, at)
	for k := 0; k < at; k++ {
		go func() {
			for i := range jobs {
				done <- settle(i, uint64(*offset+i+1), t)
			}
		}()
	}
	go func() {
		for i := 0; i < *seeds; i++ {
			jobs <- i
		}
		close(jobs)
	}()

	// Everything the batch adds up is added up here, on this one goroutine,
	// and in seed order. That is not only tidier than a lock around the
	// totals: the gates are sums of floats, and a lock adds them in whatever
	// order the workers happen to finish in, so the same batch could come
	// out differing in the last bits from one run to the next. Added in seed
	// order they cannot, and a batch is reproducible to the last bit like
	// everything else here.
	all := make([]found, *seeds)
	for n := 0; n < *seeds; n++ {
		f := <-done
		all[f.nth] = f
		// A batch used to say nothing at all until it was over, which is
		// minutes. This goes to the error stream, so the report keeps what
		// the batch found and not how long it took to find it.
		fmt.Fprintf(os.Stderr, "\r%d/%d seeds", n+1, *seeds)
	}
	fmt.Fprintln(os.Stderr)

	rows := make([]row, *seeds)
	tally := map[string]int{}
	var gates [6]float64
	gateRuns := 0.0
	for i, f := range all {
		rows[i] = f.row
		for k, v := range f.acts {
			tally[k] += v
		}
		if f.gateN > 0 {
			for k := range gates {
				gates[k] += f.gate[k] / f.gateN
			}
			gateRuns++
		}
	}

	lasted, gone, sum := 0, 0, 0
	pops := make([]int, 0, len(rows))
	var needs [5]float64
	live := 0.0
	fmt.Fprintf(full, "%4s %4s %5s %6s %6s %6s | %5s %5s %5s %5s %5s | %5s\n",
		"seed", "pop", "died", "births", "houses", "fields", "phys", "safe", "belng", "estm", "actl", "order")
	for _, r := range rows {
		fmt.Fprintf(full, "%4d %4d %5d %6d %6d %6d | %5.2f %5.2f %5.2f %5.2f %5.2f | %5.2f\n",
			r.seed, r.pop, r.deaths, r.births, r.houses, r.fields,
			r.needs[0], r.needs[1], r.needs[2], r.needs[3], r.needs[4], r.order)
		sum += r.pop
		pops = append(pops, r.pop)
		if r.pop >= *agents {
			lasted++
		}
		if r.pop == 0 {
			gone++
		} else {
			live++
			for t := range needs {
				needs[t] += r.needs[t]
			}
		}
	}
	sort.Ints(pops)
	if live > 0 {
		for t := range needs {
			needs[t] /= live
		}
	}
	total := 0
	names := make([]string, 0, len(tally))
	for k, v := range tally {
		total += v
		names = append(names, k)
	}
	// Busiest first, and ties by name. The names come out of a map, so
	// without the second clause two kinds of work tallied alike came out in
	// whichever order the map was walked in, and one batch's tally could not
	// be diffed against another's.
	sort.Slice(names, func(i, j int) bool {
		if tally[names[i]] != tally[names[j]] {
			return tally[names[i]] > tally[names[j]]
		}
		return names[i] < names[j]
	})
	fmt.Fprintln(full)
	for _, n := range names {
		fmt.Fprintf(full, "%-34s %7d %5.1f%%\n", n, tally[n], 100*float64(tally[n])/float64(total))
	}
	fmt.Fprintf(full, "\nborn %.2f inherit %.2f temp %.2f\n", *born, *inherit, *temp)
	if gateRuns > 0 {
		for k := range gates {
			gates[k] /= gateRuns
		}
	}
	fmt.Fprintf(out, "gates: fed %.2f safe %.2f held %.2f all %.3f food %.2f hungry-with-food %.2f | ", gates[0], gates[1], gates[2], gates[3], gates[4], gates[5])
	fmt.Fprintf(out, "lasted %d/%d extinct %d mean %.1f median %d | phys %.2f safe %.2f belng %.2f estm %.2f\n",
		lasted, *seeds, gone, float64(sum)/float64(*seeds), pops[len(pops)/2],
		needs[0], needs[1], needs[2], needs[3])

	// The headline, number by number, into the folder's index. These are
	// what one batch is held against another by, and the thresholds they
	// are worth believing at are in docs/baseline.md.
	rep.Score("fed", gates[0])
	rep.Score("safe", gates[1])
	rep.Score("held", gates[2])
	rep.Score("gates-all", gates[3])
	rep.Score("food", gates[4])
	rep.Score("hungry-with-food", gates[5])
	rep.Score("lasted", float64(lasted))
	rep.Score("extinct", float64(gone))
	rep.Score("mean-pop", float64(sum)/float64(*seeds))
	rep.Score("median-pop", float64(pops[len(pops)/2]))
	rep.Score("phys", needs[0])
	rep.Score("safe-need", needs[1])
	rep.Score("belng", needs[2])
	rep.Score("estm", needs[3])
}

// settle runs one seeded settlement to the end of the batch's span and says
// how it came out. It reads nothing any other seed writes and writes nothing
// any other seed reads - each has a world of its own and the world is the
// whole of the state - which is what lets the seeds run side by side.
func settle(nth int, seed uint64, t terms) found {
	w := world.NewWith(seed, t.cfg)
	w.Rules.Fit = t.fit
	w.Rules.Temperature = t.temp
	for j := 0; j < t.agents; j++ {
		w.Spawn("a", w.RandomPersonality())
	}
	w.Populate()
	f := found{nth: nth, acts: map[string]int{}}
	last := 0
	for w.Tick < t.ticks {
		system.Run(w, 200)
		for _, a := range w.Agents {
			if !a.Species().Settles || !entity.Fertile(a.Age(w.Tick)) {
				continue
			}
			f.gateN++
			fed := a.Needs[need.Physiological] >= 0.7
			safe := a.Needs[need.Safety] >= 0.6
			held := a.Needs[need.Belonging] >= 0.6
			if fed {
				f.gate[0]++
			}
			if safe {
				f.gate[1]++
			}
			if held {
				f.gate[2]++
			}
			if fed && safe && held {
				f.gate[3]++
			}
			f.gate[4] += a.Inventory[entity.Food]
			if !fed && a.Inventory[entity.Food] >= 1 {
				f.gate[5]++
			}
		}
		for _, e := range w.Log.Since(last) {
			if e.Kind == event.Acted && e.Act != "" {
				f.acts[e.Act]++
			}
		}
		last = w.Tick + 1
	}
	s := observe.Take(w)
	births := 0
	for _, e := range w.Log.All() {
		if e.Kind == event.Born {
			births++
		}
	}
	f.row = row{seed: seed, pop: s.Population, deaths: s.Deaths,
		births: births, houses: s.Houses, fields: s.Fields, order: s.Safety}
	copy(f.row.needs[:], s.MeanNeeds[:])
	return f
}
