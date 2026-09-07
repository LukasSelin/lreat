// Command tune runs a batch of seeded settlements and prints how they came
// out, for judging one set of priors against another. A single seed is a
// coin toss; this is the tool that says whether a change helped.
package main

import (
	"flag"
	"fmt"
	"sort"
	"strings"
	"sync"

	"lreat/core/action"
	"lreat/core/clock"
	"lreat/core/entity"
	"lreat/core/event"
	"lreat/core/need"
	"lreat/core/observe"
	"lreat/core/system"
	"lreat/core/world"
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
	quiet := flag.Bool("quiet", false, "summary only")
	flag.Parse()
	action.BornNoise = *born
	action.InheritNoise = *inherit
	system.Workers = 1 // the seeds are the parallelism here

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
			w := world.New(uint64(*offset + i + 1))
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
					if e.Kind == event.Acted {
						if k := strings.Index(e.Text, " finished "); k >= 0 {
							acts[e.Text[k+len(" finished "):]]++
						}
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
	if !*quiet {
		fmt.Printf("%4s %4s %5s %6s %6s %6s | %5s %5s %5s %5s %5s | %5s\n",
			"seed", "pop", "died", "births", "houses", "fields", "phys", "safe", "belng", "estm", "actl", "order")
	}
	for _, r := range rows {
		if !*quiet {
			fmt.Printf("%4d %4d %5d %6d %6d %6d | %5.2f %5.2f %5.2f %5.2f %5.2f | %5.2f\n",
				r.seed, r.pop, r.deaths, r.births, r.houses, r.fields,
				r.needs[0], r.needs[1], r.needs[2], r.needs[3], r.needs[4], r.order)
		}
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
	if !*quiet {
		total := 0
		names := make([]string, 0, len(tally))
		for k, v := range tally {
			total += v
			names = append(names, k)
		}
		sort.Slice(names, func(i, j int) bool { return tally[names[i]] > tally[names[j]] })
		fmt.Println()
		for _, n := range names {
			fmt.Printf("%-16s %6d %5.1f%%\n", n, tally[n], 100*float64(tally[n])/float64(total))
		}
		fmt.Printf("\nborn %.2f inherit %.2f temp %.2f\n", *born, *inherit, *temp)
	}
	if gateRuns > 0 {
		for k := range gates {
			gates[k] /= gateRuns
		}
	}
	fmt.Printf("gates: fed %.2f safe %.2f held %.2f all %.3f food %.2f hungry-with-food %.2f | ", gates[0], gates[1], gates[2], gates[3], gates[4], gates[5])
	fmt.Printf("lasted %d/%d extinct %d mean %.1f median %d | phys %.2f safe %.2f belng %.2f estm %.2f\n",
		lasted, *seeds, gone, float64(sum)/float64(*seeds), pops[len(pops)/2],
		needs[0], needs[1], needs[2], needs[3])
}
