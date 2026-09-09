// Command headless runs a seeded world with no player and prints what the
// settlement does over time. It is the first renderer and the tuning tool:
// a sandbox only works if a city grows on its own.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"runtime/pprof"
	"strings"
	"time"

	"lreat/core/clock"
	"lreat/core/event"
	"lreat/core/observe"
	"lreat/core/system"
	"lreat/core/world"
	"lreat/report"
	"lreat/ui/ascii"
)

var names = []string{
	"Ada", "Bo", "Cai", "Dag", "Eli", "Fen", "Gus", "Hal", "Ivo", "Jun",
	"Kai", "Lin", "Mo", "Nia", "Odd", "Pim", "Quin", "Rui", "Sol", "Tam",
	"Uma", "Vic", "Wen", "Xin", "Yara", "Zed",
}

func main() {
	seed := flag.Uint64("seed", 1, "world seed")
	width := flag.Int("width", world.DefaultWidth, "map width")
	height := flag.Int("height", world.DefaultHeight, "map height")
	wrap := flag.Bool("wrap", false, "join the east edge to the west: a globe drawn as a cylinder rather than a valley")
	ticks := flag.Int("ticks", 50*clock.Year, "days to simulate")
	agents := flag.Int("agents", 20, "starting population")
	every := flag.Int("every", 5*clock.Year, "report interval in days")
	showMap := flag.Bool("map", false, "print the map at each report")
	value := flag.Bool("value", false, "agents choose by expected value, the original rule, instead of by recognition")
	temp := flag.Float64("temp", world.DefaultRules().Temperature, "base temperature of recognition; 0 always takes the best fit")
	pave := flag.Int("pave", 0, "lay streets through the settlement every N days (0 never)")
	workers := flag.Int("workers", system.Workers, "goroutines to decide over (1 decides one agent at a time)")
	timing := flag.Bool("timing", false, "print what each report interval spent on each phase of the day, per tick")
	profile := flag.String("cpuprofile", "", "write a CPU profile of the run to this file")
	flag.Parse()
	system.Workers = *workers
	if *profile != "" {
		f, err := os.Create(*profile)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		defer pprof.StopCPUProfile()
	}

	// The run is kept as well as printed. Everything below goes through out,
	// so the file is what was on the terminal rather than a second account
	// of it. See package report.
	rep := report.Open("headless")
	out := rep.Out(os.Stdout)
	defer func() { fmt.Print(rep.Close()) }()

	w := world.NewWith(*seed, world.Config{Width: *width, Height: *height, Wrap: *wrap})
	w.Rules.Fit = !*value
	w.Rules.Temperature = *temp
	for i := 0; i < *agents; i++ {
		w.Spawn(fmt.Sprintf("%s%d", names[i%len(names)], i/len(names)), w.RandomPersonality())
	}

	fmt.Fprintf(out, "%6s %4s %4s | %5s %5s %5s %5s %5s | %5s %5s %4s | %5s %5s %6s | %4s %4s %4s %4s | %4s %4s | %5s %5s %5s | %-18s %5s | %s\n",
		"day", "pop", "died", "phys", "safe", "belng", "estm", "actl", "hlth", "age", "eld", "gini", "price", "knowl", "hous", "road", "fild", "wood", "frnd", "feud", "reach", "sprd", "open", "date", "deg", "doing")
	lastReported := 0
	spent := newTimer()
	for w.Tick < *ticks {
		if *timing {
			system.StepWith(w, spent.around)
		} else {
			system.Step(w)
		}
		// The settlement paves for itself; this lays the whole network at
		// once, for comparing a built-out network against what agents get
		// round to on their own.
		if *pave > 0 && w.Tick%*pave == 0 {
			w.PaveStreets()
		}
		if w.Tick%*every == 0 || w.Tick == *ticks {
			s := observe.Take(w)
			row(out, s)
			for _, e := range w.Log.Since(lastReported) {
				if e.Kind == event.Discovered || e.Kind == event.Died {
					fmt.Fprintf(out, "       %-4s   %s\n", "", e.Text)
				}
			}
			lastReported = w.Tick + 1
			if *timing {
				awake := 0
				for i := range w.Grid.Chunks {
					if w.Grid.Awake(i) {
						awake++
					}
				}
				spent.awake = fmt.Sprintf("%d/%d", awake, len(w.Grid.Chunks))
				fmt.Fprintln(out, spent.line())
			}
			if *showMap {
				fmt.Fprintln(out, strings.Join(ascii.Lines(s.Map), "\n"))
			}
		}
	}
	if !*showMap {
		fmt.Fprintln(out)
		fmt.Fprintln(out, strings.Join(ascii.Lines(observe.Take(w).Map), "\n"))
	}
	fmt.Fprintf(out, "\ntechs: %v\nevents: %d retained, %d dropped\n", w.Techs(), w.Log.Len(), w.Log.Dropped())
	score(rep, observe.Take(w))
}

// score picks the few numbers off the last tick worth holding against
// another run's. They go into the folder's index as well as the report, so
// that a run can be found again by how it came out rather than only by when
// it was taken.
func score(rep *report.Run, s observe.Snapshot) {
	rep.Score("pop", float64(s.Population))
	rep.Score("died", float64(s.Deaths))
	rep.Score("phys", s.MeanNeeds[0])
	rep.Score("safe", s.MeanNeeds[1])
	rep.Score("belng", s.MeanNeeds[2])
	rep.Score("estm", s.MeanNeeds[3])
	rep.Score("actl", s.MeanNeeds[4])
	rep.Score("health", s.MeanHealth)
	rep.Score("houses", float64(s.Houses))
	rep.Score("fields", float64(s.Fields))
	rep.Score("forest", float64(s.Forest))
	rep.Score("knowledge", s.Knowledge)
	rep.Score("order", s.Safety)
}

func row(out io.Writer, s observe.Snapshot) {
	var doing []string
	for i, a := range s.Activity {
		if i == 3 {
			break
		}
		doing = append(doing, fmt.Sprintf("%s:%d", a.Action, a.Agents))
	}
	n := s.MeanNeeds
	fmt.Fprintf(out, "%6d %4d %4d | %5.2f %5.2f %5.2f %5.2f %5.2f | %5.2f %5d %4d | %5.2f %5.2f %6.1f | %4d %4d %4d %4d | %4d %4d | %5.2f %5.2f %5.2f | %6s %5.1f | %s\n",
		s.Tick, s.Population, s.Deaths, n[0], n[1], n[2], n[3], n[4], s.MeanHealth, s.MeanAge, s.Elders,
		s.WealthGini, s.FoodPrice, s.Knowledge, s.Houses, s.Roads, s.Fields, s.Forest, s.Friendships, s.Feuds,
		s.GatedReach, s.HabitSpread, s.ChoiceEntropy, s.Date, s.Temp, strings.Join(doing, " "))
}

// timer is what a run spent on each phase of the day since it last said.
// The wall clock is read here and nowhere in the core.
type timer struct {
	spent []time.Duration
	ticks int
	index map[string]int
	// awake is how much of the ground was awake at the last report, as
	// chunks of chunks: what the passes over the ground are paying for.
	awake string
}

func newTimer() *timer {
	t := &timer{spent: make([]time.Duration, len(system.Phases)), index: map[string]int{}}
	for i, p := range system.Phases {
		t.index[p.Name] = i
	}
	return t
}

func (t *timer) around(name string, run func()) {
	start := time.Now()
	run()
	i := t.index[name]
	t.spent[i] += time.Since(start)
	if i == len(t.spent)-1 {
		t.ticks++
	}
}

// line says what a tick cost, by phase and in all, and starts counting again.
func (t *timer) line() string {
	if t.ticks == 0 {
		return "       timing: nothing ticked"
	}
	var parts []string
	var total time.Duration
	for i, p := range system.Phases {
		total += t.spent[i]
		parts = append(parts, fmt.Sprintf("%s %s", p.Name, per(t.spent[i], t.ticks)))
		t.spent[i] = 0
	}
	s := fmt.Sprintf("       timing: %s per tick, %s awake | %s", per(total, t.ticks), t.awake, strings.Join(parts, " "))
	t.ticks = 0
	return s
}

func per(d time.Duration, n int) string {
	return (d / time.Duration(n)).Round(time.Microsecond).String()
}
