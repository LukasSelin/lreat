package system

import (
	"math"
	"testing"
	"time"

	"lreat/core/entity"
	"lreat/core/world"
)

// planRoute is one route a settled agent was actually walking: where it
// stood, where its plan was taking it, and what it was carrying.
type planRoute struct {
	from, to entity.Pos
	load     float64
	who      entity.ID
}

// harvest runs a world on until it has settled and gathers the routes its
// agents' plans were walking, so that routing can be measured over the
// journeys the simulation asks for rather than over made-up pairs.
func harvest(w *world.World, ticks int) []planRoute {
	Run(w, ticks)
	var routes []planRoute
	for _, a := range w.Agents {
		if a.Plan == nil || a.Plan.Target == a.Pos {
			continue
		}
		routes = append(routes, planRoute{from: a.Pos, to: a.Plan.Target, load: a.Load(), who: a.ID})
	}
	return routes
}

// benchPlanRoutes times Path over the harvested routes on a grid, and says
// how many tiles a route opened on average.
func benchPlanRoutes(b *testing.B, g *world.Grid, routes []planRoute) {
	r := g.Router()
	// What the search knows at the start against what the walk comes to,
	// over the routes with a way: 1 is a bound that already knew the
	// answer. Read before the clock starts, being a question about the
	// bound rather than the search.
	known, found := 0.0, 0
	for _, p := range routes {
		cost := r.Carrying(p.load).Holding(p.who).TravelCost(p.from, p.to)
		if math.IsInf(cost, 1) {
			continue
		}
		known += r.Carrying(p.load).Holding(p.who).AtLeast(p.from, p.to) / cost
		found++
	}
	b.ResetTimer()
	work, noWay, noWayWork, length := 0, 0, 0, 0
	for i := 0; i < b.N; i++ {
		for _, p := range routes {
			r.Reset()
			path := r.Carrying(p.load).Holding(p.who).Path(p.from, p.to)
			if len(path) == 0 {
				noWay++
				noWayWork += r.Work
			}
			length += len(path)
			work += r.Work
		}
	}
	b.StopTimer()
	b.ReportMetric(float64(work)/float64(b.N)/float64(len(routes)), "tiles/route")
	b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N)/float64(len(routes)), "ns/route")
	b.ReportMetric(float64(len(routes)), "routes")
	// A search that finds no way has opened everything it could reach,
	// and no bound on what is left can spare it that; they are counted
	// apart so that the rest can be read.
	b.ReportMetric(float64(noWay)/float64(b.N), "noway")
	if noWay > 0 {
		b.ReportMetric(float64(noWayWork)/float64(noWay), "tiles/noway")
	}
	if ways := len(routes)*b.N - noWay; ways > 0 {
		b.ReportMetric(float64(work-noWayWork)/float64(ways), "tiles/found")
		b.ReportMetric(float64(length)/float64(ways), "len/found")
	}
	if found > 0 {
		b.ReportMetric(known/float64(found), "known/cost")
	}
}

// The same routes with the landmark tables standing and without them. A
// clone of the grid has no tables, and the ground on it is the same.
func BenchmarkPlanRoutesScatterLandmarks(b *testing.B) {
	w := scatter(1, 1000)
	routes := harvest(w, settling+200)
	w.Grid.RefreshLandmarks(w.Tick)
	benchPlanRoutes(b, w.Grid, routes)
}

func BenchmarkPlanRoutesScatterPlain(b *testing.B) {
	w := scatter(1, 1000)
	routes := harvest(w, settling+200)
	benchPlanRoutes(b, w.Grid.Clone(), routes)
}

func BenchmarkPlanRoutesValleyLandmarks(b *testing.B) {
	w := world.New(1)
	for i := 0; i < 40; i++ {
		w.Spawn("a", w.RandomPersonality())
	}
	routes := harvest(w, 6000)
	w.Grid.RefreshLandmarks(w.Tick)
	benchPlanRoutes(b, w.Grid, routes)
}

func BenchmarkPlanRoutesValleyPlain(b *testing.B) {
	w := world.New(1)
	for i := 0; i < 40; i++ {
		w.Spawn("a", w.RandomPersonality())
	}
	routes := harvest(w, 6000)
	benchPlanRoutes(b, w.Grid.Clone(), routes)
}

// What taking the tables costs, on the valley and on the globe.
func BenchmarkLandmarksValley(b *testing.B) {
	w := world.New(1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.Grid.Rekind() // puts the tables out, so the refresh takes them again
		w.Grid.RefreshLandmarks(w.Tick)
	}
}

func BenchmarkLandmarksGlobe(b *testing.B) {
	w := world.NewWith(1, world.Globe())
	start := time.Now()
	w.Grid.RefreshLandmarks(w.Tick)
	b.ReportMetric(float64(time.Since(start).Milliseconds()), "first-ms")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.Grid.Rekind()
		w.Grid.RefreshLandmarks(w.Tick)
	}
}
