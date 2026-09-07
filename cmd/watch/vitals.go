package main

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"

	"lreat/core/event"
	"lreat/core/observe"
	"lreat/core/system"
	"lreat/core/world"
)

// The vitals page is the answer to the one question the map cannot answer:
// a settlement stops, and nothing on the map says why. The map shows a
// hundred figures walking, and then fewer, and then none, and by the time
// the headcount moves the reason for it is several hundred ticks in the past.
//
// So this page is a post-mortem rather than a dashboard. It is laid out in
// the order the question is actually asked: when did it end, what did the
// population do on the way there, what did people die of, and — the part
// that is invisible everywhere else — what stood between everyone still
// alive and a child. A settlement that starves and one that simply stopped
// bearing look identical on the map and nothing alike here.
const (
	// popHeight is the population curve under the heading, and each of its
	// columns covers graphTicks ticks, as the activity graph's do.
	popHeight = 10
	// blockWidth is one of the columns of figures below the curve, and
	// labelWidth how much of it the name of a figure gets before its value:
	// wide enough that "past prime" is not cut down to a stub.
	blockWidth = 30
	labelWidth = 13
	// traceMax is how many columns of history are kept. Beyond it the
	// history is halved and each remaining column stands for twice as long,
	// so the whole of a run is always there however long it has run for.
	// The alternative — a window of the last so many ticks — is exactly
	// what is no use here: a settlement's end is explained by the thousand
	// ticks before it, and those are the first thing a window drops.
	traceMax = 2048
	// rateSpan is the window the recent birth and death rates are read
	// over. Short enough to be about now, long enough that a settlement
	// losing one person every few hundred ticks still registers.
	rateSpan = 500
)

// trace is one column of the settlement's history: where it stood at the end
// of the window. The counts are cumulative, so any span's rates are the
// difference between its two ends, and the levels are what they were at that
// moment. Together they are enough to reconstruct a collapse afterwards
// without keeping every tick of it.
type trace struct {
	tick    int
	pop     int
	births  int
	starved int
	failed  int
	phys    float64
	health  float64
	carried float64
}

// keep takes what every tick has to say, whichever page is up. The moments
// here are each read once and never come round again: a settlement dies out
// once, and the tick it happened on is not one the view can go back for.
func (v *view) keep(s *observe.Snapshot) {
	if s.Population > v.peak {
		v.peak, v.peakAt = s.Population, s.Tick
	}
	// When somebody was last born and last died is read off the cumulative
	// counts rather than off this tick's, because a snapshot is dropped
	// whenever the screen falls behind, and the ticks dropped during a
	// collapse are exactly the ones worth having.
	if s.Vitals.Births > v.births {
		v.births, v.lastBirth = s.Vitals.Births, s.Tick
	}
	if s.Deaths > v.deaths {
		v.deaths, v.lastDeath = s.Deaths, s.Tick
	}
	if s.Population == 0 && v.gone == 0 && v.peak > 0 {
		v.gone = s.Tick
	}
}

// plot adds a column to the history, at the same cadence the activity graph
// draws its own, and thins the history rather than dropping the start of it
// when there is too much. Thinning loses detail evenly across the run; a
// window would lose the beginning entirely, and the beginning is where a
// settlement's end was decided.
func (v *view) plot(s *observe.Snapshot) {
	v.traces = append(v.traces, trace{
		tick: s.Tick, pop: s.Population,
		births: s.Vitals.Births, starved: s.Vitals.Starved, failed: s.Vitals.Failed,
		phys: s.MeanNeeds[0], health: s.MeanHealth, carried: s.MeanFood,
	})
	if len(v.traces) <= traceMax {
		return
	}
	// Keep every other column, and the last one whatever its parity, so
	// that what the settlement stands at right now is never the thing
	// thrown away.
	thin := v.traces[:0]
	for i := 0; i < len(v.traces)-1; i += 2 {
		thin = append(thin, v.traces[i])
	}
	v.traces = append(thin, v.traces[len(v.traces)-1])
}

// bucket is one drawn column of the curve: several columns of history
// squeezed into one place on the screen. The population is the highest the
// settlement reached inside it and the condition the worst, because a
// compressed history that averages away its own bad weeks is no use for
// finding out what went wrong in one of them.
type bucket struct {
	pop        int
	phys       float64
	born, died int
}

// fit squeezes the whole kept history into at most w columns. The whole of
// it, always: a post-mortem that shows the last screenful of a long run
// answers nothing, because the answer is nearly always further back.
func (v *view) fit(w int) []bucket {
	if len(v.traces) < 2 || w < 1 {
		return nil
	}
	n := min(len(v.traces), w)
	out := make([]bucket, n)
	for j := range out {
		lo := j * len(v.traces) / n
		hi := max((j+1)*len(v.traces)/n, lo+1)
		b := bucket{phys: 1}
		for i := lo; i < hi; i++ {
			t := v.traces[i]
			b.pop = max(b.pop, t.pop)
			b.phys = min(b.phys, t.phys)
			if i > 0 {
				p := v.traces[i-1]
				b.born += t.births - p.births
				b.died += (t.starved + t.failed) - (p.starved + p.failed)
			}
		}
		out[j] = b
	}
	return out
}

// span returns what the settlement did over the last n ticks of kept
// history: how many were born and how many died. It reports false when the
// history does not reach back that far, so that a young run says so instead
// of quoting a rate over three ticks.
func (v *view) span(n int) (born, died int, ok bool) {
	if len(v.traces) < 2 {
		return 0, 0, false
	}
	last, first := v.traces[len(v.traces)-1], v.traces[0]
	for i := len(v.traces) - 1; i >= 0; i-- {
		first = v.traces[i]
		if last.tick-first.tick >= n {
			break
		}
	}
	born = last.births - first.births
	died = (last.starved + last.failed) - (first.starved + first.failed)
	return born, died, last.tick-first.tick >= n
}

// drawVitals draws the whole page. It gives up the map to do it: on the tick
// a settlement ends there is nothing left on the map to look at, and this is
// what is worth the screen instead.
func (v *view) drawVitals(sw, sh int) {
	sc := v.screen
	s := v.snap
	bold := tcell.StyleDefault.Bold(true)
	dim := tcell.StyleDefault.Dim(true)
	line := 0
	put := func(style tcell.Style, format string, args ...any) {
		puts(sc, 0, line, style, trim(fmt.Sprintf(format, args...), sw))
		line++
	}

	// The heading says whether there is still a settlement here, and if not,
	// when it ended. That is the first thing anyone asks of this page.
	state, style := fmt.Sprintf("alive, %d people", s.Population), bold
	if v.gone != 0 {
		state = fmt.Sprintf("DIED OUT at tick %d, %d ticks ago", v.gone, s.Tick-v.gone)
		style = tcell.StyleDefault.Foreground(tcell.ColorRed).Bold(true)
	}
	put(style, "─ vitals ─ tick %d ─ %s", s.Tick, state)
	put(dim, "peak %d at tick %d   last birth %s   last death %s",
		v.peak, v.peakAt, ago(v.lastBirth, s.Tick), ago(v.lastDeath, s.Tick))
	line++

	// The curve, and under it a ribbon of which way the population was going
	// in each column: red where it was losing people, green where it was
	// making them. The curve shows the fall; the ribbon shows how long the
	// settlement had been failing to replace itself before it.
	cols := v.fit(sw)
	v.drawPopulation(0, line, sw, cols)
	line += popHeight
	v.drawFlow(0, line, sw, cols)
	line++
	put(dim, "the whole run, %d ticks →   full height is %d people   a column is %d ticks",
		v.run(), max(v.peak, 1), v.run()/max(len(cols), 1))
	line++

	top := line
	v.drawTurnover(0, top, sh-2, s)
	v.drawGates(blockWidth, top, sh-2, s)
	v.drawLarder(2*blockWidth, top, sh-2, s)

	// What the settlement itself said, newest first, filling whatever is
	// left of the width. In a collapse this is one name after another
	// starving with the tick each of them went, which is as near a cause of
	// death as this world states anywhere.
	x := 3 * blockWidth
	if x+20 <= sw {
		puts(sc, x, top, bold, "last words")
		if len(s.Chronicle) == 0 {
			puts(sc, x, top+1, dim, "nothing has happened yet")
		}
		row := top + 1
		for i := len(s.Chronicle) - 1; i >= 0 && row < sh-2; i-- {
			n := s.Chronicle[i]
			st := dim
			switch n.Kind {
			case event.Died:
				st = tcell.StyleDefault.Foreground(tcell.ColorRed)
			case event.Born:
				st = tcell.StyleDefault.Foreground(tcell.ColorLime)
			}
			puts(sc, x, row, st, trim(fmt.Sprintf("t%-6d %s", n.Tick, n.Text), sw-x))
			row++
		}
	}

	puts(sc, 0, sh-1, dim, "d back to the map   space pause  +/- speed  . step  q quit")
	sc.Show()
}

// drawPopulation draws the headcount as a filled curve, oldest column on the
// left, coloured by how well fed the settlement was at the time. The colour
// is the point: a population that held level for a thousand ticks while
// turning yellow was already over, and the headcount alone would not say so
// until the drop.
func (v *view) drawPopulation(x, y, w int, cols []bucket) {
	top := max(v.peak, 1)
	for i, b := range cols {
		cx := x + w - len(cols) + i
		h := b.pop * popHeight / top
		if h == 0 && b.pop > 0 {
			h = 1 // one person left is not the same as nobody
		}
		for r := 0; r < h; r++ {
			v.screen.SetContent(cx, y+popHeight-1-r, '█', nil, foodStyle(b.phys))
		}
	}
}

// drawFlow draws one row under the curve saying which way each column went:
// green where more were born than died, red where more died, a dim rule
// where neither happened. Read across, it is the demographic history of the
// run in a single line.
func (v *view) drawFlow(x, y, w int, cols []bucket) {
	for i, b := range cols {
		cx := x + w - len(cols) + i
		born, died := b.born, b.died
		switch {
		case born > died:
			v.screen.SetContent(cx, y, '▲', nil, tcell.StyleDefault.Foreground(tcell.ColorLime))
		case died > born:
			v.screen.SetContent(cx, y, '▼', nil, tcell.StyleDefault.Foreground(tcell.ColorRed))
		default:
			v.screen.SetContent(cx, y, '─', nil, tcell.StyleDefault.Dim(true))
		}
	}
}

// drawTurnover is the ledger of the population: everyone who arrived,
// everyone who left and what of, whether the two are keeping pace now rather
// than over the whole run, and the shape of the generations that are left.
func (v *view) drawTurnover(x, y, bottom int, s *observe.Snapshot) {
	row := v.block(x, y, bottom, "turnover")
	vi := s.Vitals
	row(tcell.StyleDefault, "born", "%d", vi.Births)
	row(tcell.StyleDefault, "died", "%d", vi.Starved+vi.Failed)
	row(tcell.StyleDefault.Foreground(tcell.ColorRed), "  starved", "%d", vi.Starved)
	row(tcell.StyleDefault, "  old age", "%d", vi.Failed)
	row(tcell.StyleDefault, "net", "%+d", vi.Births-(vi.Starved+vi.Failed))
	born, died, ok := v.span(rateSpan)
	label := fmt.Sprintf("last %dt", rateSpan)
	style := tcell.StyleDefault
	if died > born {
		style = tcell.StyleDefault.Foreground(tcell.ColorRed)
	}
	if !ok {
		label, style = "all of it", tcell.StyleDefault.Dim(true)
	}
	row(style, label, "+%d -%d", born, died)
	// The generations under the headcount. A settlement whose bearing years
	// have emptied is finished whatever its population says today, and this
	// is the only place that shows before the fall does.
	row(tcell.StyleDefault, "children", "%d", s.Children)
	style = tcell.StyleDefault
	if s.Bearing == 0 && s.Population > 0 {
		style = tcell.StyleDefault.Foreground(tcell.ColorRed).Bold(true)
	}
	row(style, "bearing", "%d", s.Bearing)
	row(tcell.StyleDefault, "elders", "%d", s.Elders)
	row(tcell.StyleDefault, "mean age", "%d", s.MeanAge)
}

// drawGates is the fertility funnel: every living agent counted once, under
// the first thing that stood between it and a child on the last tick. It is
// what nothing else in the view can say — a settlement with food in the
// market and nobody being born has a reason, and the reason is here.
func (v *view) drawGates(x, y, bottom int, s *observe.Snapshot) {
	row := v.block(x, y, bottom, "why no children")
	for _, g := range world.Gates() {
		n := s.Vitals.Gates[g]
		style := tcell.StyleDefault
		switch {
		case n == 0:
			style = tcell.StyleDefault.Dim(true)
		case g == world.Ready:
			style = tcell.StyleDefault.Foreground(tcell.ColorLime)
		}
		row(style, "  "+g.String(), "%d", n)
	}
	row(tcell.StyleDefault.Dim(true), "then draw", "%.1f%%", system.BirthChance*100)
}

// drawLarder is what there is to live on and what living on it has done to
// people. Starving counts everyone at the bottom of the tier right now, each
// of them on a clock; it moves a whole generation ahead of the deaths it
// turns into.
func (v *view) drawLarder(x, y, bottom int, s *observe.Snapshot) {
	row := v.block(x, y, bottom, "larder and condition")
	row(foodStyle(s.MeanNeeds[0]), "fed", "%.2f", s.MeanNeeds[0])
	row(healthStyle(s.MeanHealth), "health", "%.2f", s.MeanHealth)
	style := tcell.StyleDefault
	if s.Starving > 0 {
		style = tcell.StyleDefault.Foreground(tcell.ColorRed).Bold(true)
	}
	row(style, "starving", "%d of %d", s.Starving, s.Population)
	row(tcell.StyleDefault.Dim(true), "  dead in", "%d ticks", system.StarvationTicks)
	row(tcell.StyleDefault, "carried", "%.2f", s.MeanFood)
	row(tcell.StyleDefault, "in market", "%.1f", s.FoodStock)
	row(tcell.StyleDefault, "price", "%.2f", s.FoodPrice)
	row(tcell.StyleDefault, "shelter", "%.2f", s.MeanShelter)
	row(tcell.StyleDefault, "safety", "%.2f", s.Safety)
	row(tcell.StyleDefault, "fields", "%d", s.Fields)
	row(tcell.StyleDefault, "forest", "%d", s.Forest)
	row(tcell.StyleDefault, "houses", "%d", s.Houses)
	row(tcell.StyleDefault, "weather", "%s %+.0f", s.Season, s.Temp)
	row(tcell.StyleDefault, "growth", "%.2f", s.Growth)
}

// block returns a writer for one column of label-and-number rows, in the
// shape the settlement's figures are in on the map page: dim label on the
// left, value against a fixed column, so that a figure which has moved since
// the last glance is still in the place it was.
func (v *view) block(x, y, bottom int, title string) func(style tcell.Style, label, format string, args ...any) {
	puts(v.screen, x, y, tcell.StyleDefault.Bold(true), title)
	line := y + 1
	return func(style tcell.Style, label, format string, args ...any) {
		if line > bottom {
			return
		}
		puts(v.screen, x, line, tcell.StyleDefault.Dim(true), trim(label, labelWidth-1))
		puts(v.screen, x+labelWidth, line, style, trim(fmt.Sprintf(format, args...), blockWidth-labelWidth-1))
		line++
	}
}

// run is how many ticks of history are kept, which after the first thinning
// is the whole run.
func (v *view) run() int {
	if len(v.traces) < 2 {
		return 0
	}
	return v.traces[len(v.traces)-1].tick - v.traces[0].tick
}

// ago puts a tick into words against now, or says it never happened.
func ago(tick, now int) string {
	if tick == 0 {
		return "never"
	}
	return fmt.Sprintf("t%d, %d ago", tick, now-tick)
}

// foodStyle colours a physiological level: green while people are fed,
// yellow while they are thin, red once the tier is down where the starvation
// clock starts.
func foodStyle(phys float64) tcell.Style {
	switch {
	case phys < 0.3:
		return tcell.StyleDefault.Foreground(tcell.ColorRed)
	case phys < 0.6:
		return tcell.StyleDefault.Foreground(tcell.ColorYellow)
	}
	return tcell.StyleDefault.Foreground(tcell.ColorLime)
}

// Report is the same post-mortem as plain text, printed when the run ends.
// A terminal view is only ever the tick it is on: whoever asks afterwards
// why a settlement died — a colleague, a bug report, a note to oneself the
// next morning — has nothing to go on but what was left on the screen, and
// the screen is torn down on the way out. So the run says it in words, and
// what it says can be pasted, piped, or diffed against another seed's.
func (v *view) Report() string {
	if v.snap == nil {
		return "nothing ran\n"
	}
	s := v.snap
	var b strings.Builder
	state := fmt.Sprintf("alive, %d people", s.Population)
	if v.gone != 0 {
		state = fmt.Sprintf("DIED OUT at tick %d, %d ticks before the end", v.gone, s.Tick-v.gone)
	}
	line := func(label, format string, args ...any) {
		fmt.Fprintf(&b, "%-15s %s\n", label, fmt.Sprintf(format, args...))
	}
	fmt.Fprintf(&b, "\n─ vitals ─ tick %d ─ %s\n", s.Tick, state)
	line("peak", "%d people at tick %d", v.peak, v.peakAt)
	line("last birth", "%s", ago(v.lastBirth, s.Tick))
	line("last death", "%s", ago(v.lastDeath, s.Tick))
	vi := s.Vitals
	line("turnover", "born %d, died %d (starved %d, old age %d), net %+d",
		vi.Births, vi.Starved+vi.Failed, vi.Starved, vi.Failed, vi.Births-vi.Starved-vi.Failed)
	if born, died, ok := v.span(rateSpan); ok {
		line(fmt.Sprintf("last %d ticks", rateSpan), "+%d -%d", born, died)
	}
	line("generations", "children %d, bearing %d, elders %d, mean age %d",
		s.Children, s.Bearing, s.Elders, s.MeanAge)
	var gates []string
	for _, g := range world.Gates() {
		gates = append(gates, fmt.Sprintf("%s %d", g, s.Vitals.Gates[g]))
	}
	line("why no children", "%s", strings.Join(gates, ", "))
	line("larder", "fed %.2f, health %.2f, starving %d of %d, carried %.2f, market %.1f at %.2f",
		s.MeanNeeds[0], s.MeanHealth, s.Starving, s.Population, s.MeanFood, s.FoodStock, s.FoodPrice)
	line("land", "fields %d, forest %d, houses %d, shelter %.2f, safety %.2f",
		s.Fields, s.Forest, s.Houses, s.MeanShelter, s.Safety)
	line("population", "%s", v.sparkline(reportWidth))
	line("", "%d ticks, full height %d people", v.run(), max(v.peak, 1))
	b.WriteString("last words\n")
	for i := len(s.Chronicle) - 1; i >= 0 && i > len(s.Chronicle)-reportNotes; i-- {
		fmt.Fprintf(&b, "  t%-7d %s\n", s.Chronicle[i].Tick, s.Chronicle[i].Text)
	}
	return b.String()
}

const (
	// reportWidth is how wide the written population curve is, and
	// reportNotes how many of the settlement's last words are quoted.
	reportWidth = 60
	reportNotes = 16
)

// sparkline draws the population over the whole run in one row of glyphs.
// It is the shape of the thing — grew, held, fell — which is most of what
// there is to say about a run and all that fits in a line.
func (v *view) sparkline(w int) string {
	cols := v.fit(w)
	if len(cols) == 0 {
		return "(too short to draw)"
	}
	steps := []rune(" ▁▂▃▄▅▆▇█")
	top := max(v.peak, 1)
	out := make([]rune, len(cols))
	for i, c := range cols {
		n := c.pop * (len(steps) - 1) / top
		if n == 0 && c.pop > 0 {
			n = 1
		}
		out[i] = steps[n]
	}
	return string(out)
}
