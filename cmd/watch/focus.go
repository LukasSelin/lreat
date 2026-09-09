package main

import (
	"fmt"

	"github.com/gdamore/tcell/v2"

	"lreat/ui/ascii"
)

// Every page here shows a dozen measurements at once, and every one of them
// is squeezed: a measure of the world gets a single row, a kind of work gets
// whatever share of six rows it happens to hold. Squeezed that far a series
// says whether it went up and nothing about the shape it went up in — and a
// band that holds three percent of the population is not drawn at all, which
// is why the world page writes the shares out in numbers underneath.
//
// The arrows step through whatever the page has, and the one stepped onto is
// opened out: drawn on its own, over the whole of the page's largest space,
// scaled to its own high-water mark rather than to whatever it shares the
// space with. Esc drops it and the page goes back to showing everything at
// once. Nothing is hidden by stepping — the rows and the legend stay where
// they were, with the one being read picked out.

// graph is one of the things a page can open out. Most are a series of
// numbers over the run and are drawn from series; the two that are already
// pictures rather than lines — the weave of what people are doing, and the
// population curve — draw themselves through whole.
type graph struct {
	name   string
	color  ascii.Color
	digits int
	// series is the graph's history as one value per column, oldest first,
	// gathered into at most w columns. It is nil when whole is set.
	series func(v *view, w int) []float64
	// whole draws a graph that is not a single line into the space given.
	whole func(v *view, x, y, w, h int)
	// note is what is worth saying about it beside the numbers.
	note func(v *view) string
}

// graphs is what the page now up has to step through, in the order it draws
// them. The map page starts from the whole weave and then goes through the
// kinds of work in it; the world page from its measures to the same weave
// over the whole run; the vitals page from the curve to what moves it.
func (v *view) graphs() []graph {
	switch {
	case v.vitals:
		return vitalGraphs()
	case v.world:
		return worldGraphs()
	}
	return mapGraphs()
}

// step moves the focus over the page's graphs, and off the end of them in
// either direction: stepping past the last is how one stops reading a single
// measure and goes back to the page as a whole. focus counts from one, and
// is zero when nothing is opened out — a view is made with nothing opened.
func (v *view) step(d int) {
	n := len(v.graphs())
	if n == 0 {
		return
	}
	at := v.focus + d
	if v.focus == 0 && d < 0 {
		at = n // stepping back from the page itself takes the last graph
	}
	if at < 1 || at > n {
		at = 0
	}
	v.focus = at
}

// focused is the graph being read, and whether there is one. The index is
// kept rather than the graph itself because the lists are built fresh on
// every draw, out of a history that is a tick longer each time.
func (v *view) focused() (graph, bool) {
	gs := v.graphs()
	if v.focus < 1 || v.focus > len(gs) {
		return graph{}, false
	}
	return gs[v.focus-1], true
}

// drawOpen draws a graph opened out: its whole history, a column to each
// stretch of it, scaled to the highest it ever reached, filling the height
// it is given from the bottom up. A row is drawn in eighths, so a series
// that never leaves the bottom of its range still has a shape.
func (v *view) drawOpen(x, y, w, h int, g graph) {
	if h < 1 || w < 1 {
		return
	}
	if g.whole != nil {
		g.whole(v, x, y, w, h)
		return
	}
	vals := g.series(v, w)
	var high float64
	for _, f := range vals {
		high = max(high, f)
	}
	if high <= 0 {
		return
	}
	style := palette[g.color]
	for i, f := range vals {
		cx := x + w - len(vals) + i
		eighths := int(f/high*float64(h*8) + 0.5)
		for r := 0; r < h; r++ {
			// The bottom row is filled first: what is left of the column's
			// eighths once every row under this one has taken its fill.
			cell := min(eighths-(h-1-r)*8, 8)
			if cell <= 0 {
				continue
			}
			v.screen.SetContent(cx, y+r, steps[cell], nil, style)
		}
	}
}

// headline is the line over an opened graph: what it is, where it stands
// now, the high the height stands for, and whatever else is worth saying.
func (v *view) headline(g graph) string {
	if g.whole != nil {
		return g.name
	}
	vals := g.series(v, max(len(v.traces), 1))
	if len(vals) == 0 {
		return g.name
	}
	high, low, now := vals[0], vals[0], vals[len(vals)-1]
	for _, f := range vals {
		high, low = max(high, f), min(low, f)
	}
	line := fmt.Sprintf("%s   now %.*f   full height is %.*f", g.name, g.digits, now, g.digits, high)
	// The low is only worth the room when the series never came down to
	// nothing: a measure that has been at zero says so by touching the
	// floor of its own chart.
	if low > 0 {
		line += fmt.Sprintf("   low %.*f", g.digits, low)
	}
	if g.note != nil {
		if s := g.note(v); s != "" {
			line += "   " + s
		}
	}
	return line
}

// keyed is the line telling the reader that stepping is a thing the page
// does, and how to stop. It stands where the page's own keys are.
func (v *view) keyed(keys string) string {
	if _, ok := v.focused(); ok {
		return "up/down another graph   esc back to the page   " + keys
	}
	return "up/down open a graph   " + keys
}

// graphKeys is the same line for the map page, which has the panel's width
// to say it in rather than the screen's, and has to say which pair of keys:
// where the arrows are looking around, the graphs answer to pgup and pgdn
// instead. A line naming the keys of the other case is worse than no line —
// it is a reader pressing what they were told to and watching the map move.
func (v *view) graphKeys() string {
	step := "up/down"
	if v.looking() {
		step = "pgup/pgdn"
	}
	if _, ok := v.focused(); ok {
		return step + " another graph  esc back"
	}
	return step + " open a graph"
}

// mapGraphs are the band under the map: the whole weave, and then each kind
// of work in it on its own. Opened out, a kind is scaled to its own high
// rather than to the population, which is the only way the thin ones — the
// guarding, the studying — are visible at all.
func mapGraphs() []graph {
	gs := []graph{{
		name:  "everyone",
		whole: func(v *view, x, y, w, h int) { v.drawGraph(x, y, w, h) },
	}}
	for i, g := range ascii.Groups {
		gs = append(gs, graph{
			name: g.Name, color: g.Color, digits: 2,
			series: func(v *view, w int) []float64 {
				hist := v.hist
				if len(hist) > w {
					hist = hist[len(hist)-w:]
				}
				out := make([]float64, len(hist))
				for j, col := range hist {
					out[j] = col[i]
				}
				return out
			},
			note: func(v *view) string { return "share of everyone alive" },
		})
	}
	return gs
}

// worldGraphs are the measures down the world page, and under them the same
// weave over the whole run rather than the last few hundred days.
func worldGraphs() []graph {
	gs := make([]graph, 0, len(measures)+1)
	for _, m := range measures {
		gs = append(gs, measureGraph(m))
	}
	return append(gs, graph{
		name: "what they have been doing with themselves",
		whole: func(v *view, x, y, w, h int) {
			v.drawWork(x, y, w, h, v.fit(w))
		},
	})
}

// measureGraph is one row of the world page as something that can be opened
// out. It takes the high over each stretch rather than the mean, for the
// same reason the row does: a history that averages away the week the
// granary emptied is no use for finding out what happened in it.
func measureGraph(m measure) graph {
	return graph{
		name: m.name, color: m.color, digits: m.digits,
		series: func(v *view, w int) []float64 {
			cols := v.fit(w)
			out := make([]float64, len(cols))
			for i, b := range cols {
				for _, t := range v.traces[b.lo:b.hi] {
					out[i] = max(out[i], m.at(t))
				}
			}
			return out
		},
		note: func(v *view) string {
			if m.note == nil || v.snap == nil {
				return ""
			}
			return m.note(v.snap)
		},
	}
}

// vitalGraphs are the curve at the top of the vitals page and the two flows
// that make it: the settlement is only ever as many people as were born into
// it less those it has buried, and either of those can be the one that moved.
func vitalGraphs() []graph {
	return []graph{{
		name: "population",
		whole: func(v *view, x, y, w, h int) {
			v.drawPopulation(x, y, w, v.fit(w))
		},
	}, {
		name: "born", color: ascii.AgentSocial,
		series: bucketSeries(func(b bucket) float64 { return float64(b.born) }),
		note:   func(v *view) string { return "children, over each stretch of the run" },
	}, {
		name: "died", color: ascii.AgentGuard,
		series: bucketSeries(func(b bucket) float64 { return float64(b.died) }),
		note:   func(v *view) string { return "burials, over each stretch of the run" },
	}, {
		name: "fed", color: ascii.AgentFood, digits: 2,
		series: bucketSeries(func(b bucket) float64 { return b.phys }),
		note:   func(v *view) string { return "the worst anyone was fed in the stretch" },
	}}
}

// bucketSeries reads one number off each stretch of the kept history.
func bucketSeries(at func(bucket) float64) func(*view, int) []float64 {
	return func(v *view, w int) []float64 {
		cols := v.fit(w)
		out := make([]float64, len(cols))
		for i, b := range cols {
			out[i] = at(b)
		}
		return out
	}
}

// markGroup is how a kind of work is written in the legend under the map:
// picked out when it is the one opened out above, dimmed while another one
// is, and left as it was when the whole weave is up.
func (v *view) markGroup(i int) tcell.Style {
	g, ok := v.focused()
	if !ok {
		return tcell.StyleDefault
	}
	if g.name == ascii.Groups[i].Name {
		return tcell.StyleDefault.Bold(true)
	}
	return tcell.StyleDefault.Dim(true)
}
