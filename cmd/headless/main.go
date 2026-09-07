// Command headless runs a seeded world with no player and prints what the
// settlement does over time. It is the first renderer and the tuning tool:
// a sandbox only works if a city grows on its own.
package main

import (
	"flag"
	"fmt"
	"strings"

	"lreat/core/event"
	"lreat/core/observe"
	"lreat/core/system"
	"lreat/core/world"
	"lreat/ui/ascii"
)

var names = []string{
	"Ada", "Bo", "Cai", "Dag", "Eli", "Fen", "Gus", "Hal", "Ivo", "Jun",
	"Kai", "Lin", "Mo", "Nia", "Odd", "Pim", "Quin", "Rui", "Sol", "Tam",
	"Uma", "Vic", "Wen", "Xin", "Yara", "Zed",
}

func main() {
	seed := flag.Uint64("seed", 1, "world seed")
	ticks := flag.Int("ticks", 5000, "ticks to simulate")
	agents := flag.Int("agents", 20, "starting population")
	every := flag.Int("every", 250, "report interval in ticks")
	showMap := flag.Bool("map", false, "print the map at each report")
	value := flag.Bool("value", false, "agents choose by expected value, the original rule, instead of by recognition")
	temp := flag.Float64("temp", world.DefaultRules().Temperature, "base temperature of recognition; 0 always takes the best fit")
	flag.Parse()

	w := world.New(*seed)
	w.Rules.Fit = !*value
	w.Rules.Temperature = *temp
	for i := 0; i < *agents; i++ {
		w.Spawn(fmt.Sprintf("%s%d", names[i%len(names)], i/len(names)), w.RandomPersonality())
	}

	fmt.Printf("%6s %4s %4s | %5s %5s %5s %5s %5s | %5s | %5s %5s %6s | %4s %4s | %4s %4s | %5s %5s %5s | %s\n",
		"tick", "pop", "died", "phys", "safe", "belng", "estm", "actl", "hlth", "gini", "price", "knowl", "hous", "fild", "frnd", "feud", "reach", "sprd", "open", "doing")
	lastReported := 0
	for w.Tick < *ticks {
		system.Step(w)
		if w.Tick%*every == 0 || w.Tick == *ticks {
			s := observe.Take(w)
			report(s)
			for _, e := range w.Log.Since(lastReported) {
				if e.Kind == event.Discovered || e.Kind == event.Died {
					fmt.Printf("       %-4s   %s\n", "", e.Text)
				}
			}
			lastReported = w.Tick + 1
			if *showMap {
				fmt.Println(strings.Join(ascii.Lines(s.Map), "\n"))
			}
		}
	}
	if !*showMap {
		fmt.Println()
		fmt.Println(strings.Join(ascii.Lines(observe.Take(w).Map), "\n"))
	}
	fmt.Printf("\ntechs: %v\nevents: %d retained, %d dropped\n", w.Techs(), w.Log.Len(), w.Log.Dropped())
}

func report(s observe.Snapshot) {
	var doing []string
	for i, a := range s.Activity {
		if i == 3 {
			break
		}
		doing = append(doing, fmt.Sprintf("%s:%d", a.Action, a.Agents))
	}
	n := s.MeanNeeds
	fmt.Printf("%6d %4d %4d | %5.2f %5.2f %5.2f %5.2f %5.2f | %5.2f | %5.2f %5.2f %6.1f | %4d %4d | %4d %4d | %5.2f %5.2f %5.2f | %s\n",
		s.Tick, s.Population, s.Deaths, n[0], n[1], n[2], n[3], n[4], s.MeanHealth,
		s.WealthGini, s.FoodPrice, s.Knowledge, s.Houses, s.Fields, s.Friendships, s.Feuds,
		s.GatedReach, s.HabitSpread, s.ChoiceEntropy, strings.Join(doing, " "))
}
