// Command tune runs a batch of seeded settlements and prints how they came
// out, for judging one set of priors against another. A single seed is a
// coin toss; this is the tool that says whether a change helped.
package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"sync"

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

func main() {
	seeds := flag.Int("seeds", 24, "seeds to run")
	offset := flag.Int("offset", 0, "first seed minus one, for an independent batch")
	ticks := flag.Int("ticks", 60*clock.Year, "days per run")
	agents := flag.Int("agents", 20, "starting population")
	value := flag.Bool("value", false, "use the value rule")
	born := flag.Float64("born", action.BornNoise, "drift on a founder's habits")
	inherit := flag.Float64("inherit", action.InheritNoise, "drift on a child's habits")
	temp := flag.Float64("temp", world.DefaultRules().Temperature, "recognition temperature")
	cap := flag.Int("cap", system.MaxPopulation, "population ceiling; the guard on the machine, not a fact about the world")
	quiet := flag.Bool("quiet", false, "summary only")
	preset := flag.String("preset", "", "the terms to found each world on: valley (the default map) or globe; -width, -height and -wrap override it")
	width := flag.Int("width", world.DefaultWidth, "map width")
	height := flag.Int("height", world.DefaultHeight, "map height")
	wrap := flag.Bool("wrap", false, "join the east edge to the west")
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
	action.BornNoise = *born
	action.InheritNoise = *inherit
	system.MaxPopulation = *cap
	system.Workers = 1 // the seeds are the parallelism here

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

	rows := make([]row, *seeds)
	tally := map[string]int{}
	var gates [6]float64
	gateRuns := 0.0
	var mu sync.Mutex
	var wg sync.WaitGroup
	for i := 0; i < *seeds; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			w := world.NewWith(uint64(*offset+i+1), cfg)
			w.Rules.Fit = !*value
			w.Rules.Temperature = *temp
			for j := 0; j < *agents; j++ {
				w.Spawn("a", w.RandomPersonality())
			}
			acts, last := map[string]int{}, 0
			var gate [6]float64
			var gateN float64
			for w.Tick < *ticks {
				system.Run(w, 200)
				for _, a := range w.Agents {
					if !entity.Fertile(a.Age(w.Tick)) {
						continue
					}
					gateN++
					fed := a.Needs[need.Physiological] >= 0.7
					safe := a.Needs[need.Safety] >= 0.6
					held := a.Needs[need.Belonging] >= 0.6
					if fed {
						gate[0]++
					}
					if safe {
						gate[1]++
					}
					if held {
						gate[2]++
					}
					if fed && safe && held {
						gate[3]++
					}
					gate[4] += a.Inventory[entity.Food]
					if !fed && a.Inventory[entity.Food] >= 1 {
						gate[5]++
					}
				}
				for _, e := range w.Log.Since(last) {
					if e.Kind == event.Acted && e.Act != "" {
						acts[e.Act]++
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
			r := row{seed: uint64(*offset + i + 1), pop: s.Population, deaths: s.Deaths,
				births: births, houses: s.Houses, fields: s.Fields, order: s.Safety}
			copy(r.needs[:], s.MeanNeeds[:])
			rows[i] = r
			mu.Lock()
			for k, v := range acts {
				tally[k] += v
			}
			if gateN > 0 {
				for k := range gate {
					gates[k] += gate[k] / gateN
				}
				gateRuns++
			}
			mu.Unlock()
		}(i)
	}
	wg.Wait()

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
	sort.Slice(names, func(i, j int) bool { return tally[names[i]] > tally[names[j]] })
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
