package main

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"

	"lreat/core/observe"
	"lreat/core/world"
	"lreat/ui/ascii"
)

// The world page is the other half of the post-mortem. The vitals page says
// what became of the people; this one says what became of the ground they
// stood on and what they spent themselves on standing there — the forest
// coming down, the fields going in, the houses going up, the price of food,
// what the settlement knew and how safe it was.
//
// It is drawn as one row per measure over the whole run rather than as this
// tick's numbers, because every one of these figures is only worth anything
// against its own past. Forest 255 says nothing. Forest 255 of a wood that
// was 900 and has been falling in a straight line for four thousand ticks
// says why nobody is building any more.
//
// The band at the bottom is the same weave as under the map, over the whole
// run instead of the last few hundred ticks: what a settlement spends its
// people on is the nearest thing it has to a policy, and it changes on the
// scale of generations rather than of ticks.

// measure is one row of the page: a name, how to read it off a moment of
// history, and how to write the number down.
type measure struct {
	name   string
	at     func(trace) float64
	digits int
	color  ascii.Color
	// note is what stands to the right of the line, where the plain range
	// of the series is not the useful thing to say about it.
	note func(*observe.Snapshot) string
}

// measures are the world's own figures, in the order the settlement makes
// them: the people, then the land they take, then what they built on it,
// then the dealings between them, then what got through to them.
var measures = []measure{
	{name: "people", at: func(t trace) float64 { return float64(t.pop) }, color: ascii.AgentIdle},
	{name: "forest", at: func(t trace) float64 { return float64(t.forest) }, color: ascii.ForestRich,
		note: func(s *observe.Snapshot) string {
			if s.Forest0 == 0 {
				return ""
			}
			// Woods are cut and woods are planted, so this is a share of
			// what the world was made with and not a share of what is
			// left: a settlement can end with more trees than it found.
			return fmt.Sprintf("%.0f%% of the %d it started with", 100*float64(s.Forest)/float64(s.Forest0), s.Forest0)
		}},
	{name: "fields", at: func(t trace) float64 { return float64(t.fields) }, color: ascii.Field},
	{name: "houses", at: func(t trace) float64 { return float64(t.houses) }, color: ascii.House},
	{name: "roads", at: func(t trace) float64 { return float64(t.roads) }, color: ascii.Road},
	{name: "food kept", at: func(t trace) float64 { return t.stock }, digits: 1, color: ascii.Granary,
		note: func(s *observe.Snapshot) string { return "in the market, over everyone" }},
	{name: "food price", at: func(t trace) float64 { return t.price }, digits: 2, color: ascii.Market},
	{name: "carried", at: func(t trace) float64 { return t.carried }, digits: 2, color: ascii.AgentFood,
		note: func(s *observe.Snapshot) string { return "food on the average person" }},
	{name: "fed", at: func(t trace) float64 { return t.phys }, digits: 2, color: ascii.AgentFood},
	{name: "health", at: func(t trace) float64 { return t.health }, digits: 2, color: ascii.AgentIdle},
	{name: "safety", at: func(t trace) float64 { return t.safety }, digits: 2, color: ascii.AgentGuard},
	{name: "knowledge", at: func(t trace) float64 { return t.knowledge }, color: ascii.AgentStudy},
	{name: "gini", at: func(t trace) float64 { return t.gini }, digits: 2, color: ascii.AgentTrade,
		note: func(s *observe.Snapshot) string { return "0 is everyone alike, 1 is one owner" }},
	{name: "growth", at: func(t trace) float64 { return t.growth }, digits: 2, color: ascii.Grass,
		note: func(s *observe.Snapshot) string { return fmt.Sprintf("%s, %+.0f degrees", s.Season, s.Temp) }},
}

const (
	// The row is name, then where the measure stands now, then the line of
	// its history, then whatever is worth saying about it.
	measureName  = 11
	measureValue = 9
	measureNote  = 34
	// workHeight is the least the whole-run weave of what people have been
	// doing is drawn at. It takes whatever the page has left over above
	// that: the weave is the one thing here with no natural height, and a
	// tall terminal spent on blank rows under it is a tall terminal wasted.
	workHeight = 6
	// workBelow is what stands under the weave and has to be left room for:
	// the legend, a blank line, the technologies, and a blank line over the
	// keys along the foot.
	workBelow = 4
)

// found is a technology and the tick the settlement came to it. Discovery is
// the one thing in this world that never comes undone, so a list of them
// with their ticks is the settlement's whole technical history.
type found struct {
	tick int
	tech world.Tech
}

// discoveries notes anything the settlement has come to since the last look.
// It is kept here rather than read off the event log because the log is
// bounded and a long run drops its own beginning, which is where the first
// discoveries are.
func (v *view) discoveries(s *observe.Snapshot) {
	if len(s.Techs) == len(v.techs) {
		return
	}
	known := make(map[world.Tech]bool, len(v.techs))
	for _, f := range v.techs {
		known[f.tech] = true
	}
	for _, t := range s.Techs {
		if !known[t] {
			v.techs = append(v.techs, found{tick: s.Tick, tech: t})
		}
	}
}

// drawWorld draws the page.
func (v *view) drawWorld(sw, sh int) {
	sc := v.screen
	s := v.snap
	dim := tcell.StyleDefault.Dim(true)
	line := 0

	puts(sc, 0, line, tcell.StyleDefault.Bold(true),
		trim(fmt.Sprintf("─ the world ─ tick %d ─ %d by %d ─ %d people on it", s.Tick, s.Map.W, s.Map.H, s.Population), sw))
	line++
	puts(sc, 0, line, dim, trim(fmt.Sprintf("the whole run, %d ticks, oldest on the left", v.run()), sw))
	line += 2

	width := sw - measureName - measureValue - measureNote - 2
	cols := v.fit(max(width, 1))
	open, opened := v.focused()
	for _, m := range measures {
		// A measure is dropped rather than drawn over the weave: the page
		// gives the weave its floor and the keys their line whatever the
		// window is, and what will not fit above that simply is not shown.
		if line >= sh-1-workBelow-workHeight {
			break
		}
		v.drawMeasure(line, width, m, cols, opened && open.name == m.name)
		line++
	}
	line++

	// Under the measures stands whatever is being read closely: the weave
	// of what the settlement has spent its people on, or the one measure
	// stepped onto, opened out over the same space and scaled to its own
	// high-water mark. A measure squeezed into a row says whether it went
	// up; the same measure over twenty rows says the shape it went up in.
	title := "what they have been doing with themselves"
	if opened {
		title = v.headline(open)
	}
	puts(sc, 0, line, tcell.StyleDefault.Bold(true), trim(title, sw))
	line++
	// The space gets the screen's whole width rather than the measures'
	// narrower one, so it is squeezed on its own terms, and whatever height
	// the page has not already spent.
	weave := max(workHeight, sh-1-workBelow-line)
	if opened {
		v.drawOpen(0, line, sw, weave, open)
	} else {
		v.drawWork(0, line, sw, weave, v.fit(sw))
	}
	line += weave
	// The legend carries each kind's share of the whole run beside its
	// name. The weave is six rows deep and one kind of work usually takes
	// most of them, which leaves the rest too thin to see at all: the
	// numbers say what the bands cannot.
	if opened && open.whole == nil {
		// The legend belongs to the weave, and the weave is not up: a row
		// of shares under a chart of something else would be read as that
		// chart's own. What stands there instead is what the chart is made
		// of — how much of the run each of its columns is.
		puts(sc, 0, line, dim, trim(fmt.Sprintf("oldest on the left, a column to every %d days of the run",
			max(v.run()/max(len(cols), 1), 1)), sw))
	} else {
		cell := max(sw/len(ascii.Groups), 4)
		share := v.workShare()
		for i, g := range ascii.Groups {
			puts(sc, i*cell, line, palette[g.Color], "█")
			puts(sc, i*cell+2, line, dim, trim(fmt.Sprintf("%s %.0f%%", g.Name, 100*share[i]), cell-3))
		}
	}
	line += 2

	// Everything the settlement has ever worked out, with the tick it came
	// to it. Nothing here is ever lost again, so it is the one line of the
	// page that only grows.
	var techs []string
	for _, f := range v.techs {
		techs = append(techs, fmt.Sprintf("t%d %s", f.tick, f.tech))
	}
	if len(techs) == 0 {
		techs = []string{"nothing yet"}
	}
	puts(sc, 0, line, tcell.StyleDefault, trim("worked out: "+strings.Join(techs, "   "), sw))

	puts(sc, 0, sh-1, dim, trim(v.keyed("w back to the map   d vitals   q quit"), sw))
	sc.Show()
}

// drawMeasure draws one row: the name, where it stands now, its whole
// history as a line of glyphs scaled to its own high-water mark, and what is
// worth saying about it. Each row is scaled to itself, so the shape of every
// measure is legible whatever the others are doing; the number at the right
// is what the full height means.
func (v *view) drawMeasure(y, width int, m measure, cols []bucket, open bool) {
	sc := v.screen
	style := palette[m.color]
	// The row of the measure being read closely is picked out where it
	// always stood, so that what is opened out below the page is plainly
	// this line and the page is not rearranged around it. The mark stands
	// in the blank column between the number and its history, pointing at
	// the line it belongs to. See focus.go.
	name := tcell.StyleDefault.Dim(true)
	if open {
		name = tcell.StyleDefault.Bold(true)
		puts(sc, measureName+measureValue-1, y, tcell.StyleDefault.Bold(true), "▸")
	}
	puts(sc, 0, y, name, trim(m.name, measureName-1))

	var top float64
	for _, t := range v.traces {
		top = max(top, m.at(t))
	}
	now := 0.0
	if len(v.traces) > 0 {
		now = m.at(v.traces[len(v.traces)-1])
	}
	puts(sc, measureName, y, style, fmt.Sprintf("%*.*f", measureValue-1, m.digits, now))

	x := measureName + measureValue
	for i, b := range cols {
		sc.SetContent(x+width-len(cols)+i, y, v.glyph(m, b, top), nil, style)
	}

	note := fmt.Sprintf("high %.*f", m.digits, top)
	if m.note != nil {
		if s := m.note(v.snap); s != "" {
			note = s
		}
	}
	puts(sc, x+width+1, y, tcell.StyleDefault.Dim(true), trim(note, measureNote))
}

// workShare is each kind of work's mean share of the population over the
// whole run.
func (v *view) workShare() column {
	var mean column
	if len(v.traces) == 0 {
		return mean
	}
	for _, t := range v.traces {
		for i := range mean {
			mean[i] += t.work[i] / float64(len(v.traces))
		}
	}
	return mean
}

// steps are the eighths a measure's line is drawn in, blank for nothing.
var steps = []rune(" ▁▂▃▄▅▆▇█")

// glyph is one column of a measure's line: how high the measure got over the
// stretch of history the column stands for, against the highest it reached
// anywhere in the run. The high-water mark rather than the mean, because a
// compressed history that averages away the week the granary emptied is no
// use for finding out what happened in it.
func (v *view) glyph(m measure, b bucket, top float64) rune {
	if b.span == 0 {
		return ' '
	}
	var high float64
	for _, t := range v.traces[b.lo:b.hi] {
		high = max(high, m.at(t))
	}
	if high <= 0 {
		return ' '
	}
	h := 1
	if top > 0 {
		h = max(int(high/top*float64(len(steps)-1)), 1)
	}
	return steps[min(h, len(steps)-1)]
}

// drawWork stacks the kinds of work over the whole run, in the same order
// and colours as the band under the map.
func (v *view) drawWork(x, y, w, h int, cols []bucket) {
	for i, b := range cols {
		if b.span == 0 {
			continue
		}
		cx := x + w - len(cols) + i
		for r := 0; r < h; r++ {
			share := (float64(h-r) - 0.5) / float64(h)
			var cum float64
			for gi, g := range ascii.Groups {
				cum += b.work[gi]
				if share <= cum {
					v.screen.SetContent(cx, y+r, '█', nil, palette[g.Color])
					break
				}
			}
		}
	}
}

// worldReport is the same page in plain words, for the end of a run.
func (v *view) worldReport() string {
	if len(v.traces) < 2 {
		return ""
	}
	var b strings.Builder
	b.WriteString("\n─ the world ─\n")
	cols := v.fit(reportWidth)
	for _, m := range measures {
		var top float64
		for _, t := range v.traces {
			top = max(top, m.at(t))
		}
		line := make([]rune, len(cols))
		for i, c := range cols {
			line[i] = v.glyph(m, c, top)
		}
		now := m.at(v.traces[len(v.traces)-1])
		fmt.Fprintf(&b, "%-11s %*.*f %s  high %.*f\n",
			m.name, measureValue-1, m.digits, now, string(line), m.digits, top)
	}
	var techs []string
	for _, f := range v.techs {
		techs = append(techs, fmt.Sprintf("t%d %s", f.tick, f.tech))
	}
	if len(techs) == 0 {
		techs = []string{"nothing"}
	}
	var work []string
	share := v.workShare()
	for i, g := range ascii.Groups {
		work = append(work, fmt.Sprintf("%s %.0f%%", g.Name, 100*share[i]))
	}
	fmt.Fprintf(&b, "%-11s %s\n", "their work", strings.Join(work, ", "))
	fmt.Fprintf(&b, "%-11s %s\n", "worked out", strings.Join(techs, ", "))
	return b.String()
}
