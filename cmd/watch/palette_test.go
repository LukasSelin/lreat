package main

import (
	"testing"

	"github.com/gdamore/tcell/v2"

	"lreat/core/observe"
	"lreat/core/world"
	"lreat/ui/ascii"
)

// rgb unpacks one of the 256 terminal colours into its channels, which is the
// only way to ask a question about what a colour looks like rather than about
// which number it is.
func rgb(c tcell.Color) (r, g, b int) {
	i := int(c &^ tcell.ColorValid)
	if i >= 232 {
		v := 8 + (i-232)*10
		return v, v, v
	}
	i -= 16
	level := []int{0, 95, 135, 175, 215, 255}
	return level[i/36], level[(i/6)%6], level[i%6]
}

// vegetation reports whether a colour reads as something growing: the green
// channel clearly ahead of both the others.
func vegetation(c tcell.Color) bool {
	r, g, b := rgb(c)
	return g > r+30 && g > b+30
}

// A colour has to be one tcell will actually draw. tcell.Color is a bitfield,
// not a palette index: the entry for a palette colour carries ColorValid, and
// a bare tcell.Color(22) is not colour 22 at all but an invalid value that
// draws in the terminal's default. That is white on a dark terminal, and it
// is silent - nothing fails, the map simply comes out blank.
//
// Every open-ground colour on this map was written that way from the day the
// palette was first set down, so the grass had always drawn white; the mistake
// only became visible when the same form was used for the woods, which until
// then had been tcell.ColorGreen and had therefore worked. Use
// tcell.PaletteColor, or the named tcell.Color22.
func TestEveryColourIsOneTcellWillDraw(t *testing.T) {
	for c, style := range palette {
		fg, _, _ := style.Decompose()
		if fg == tcell.ColorDefault {
			continue // deliberately the terminal's own colour
		}
		if fg&tcell.ColorValid == 0 {
			t.Errorf("colour %d has foreground %d, which is not a valid tcell colour "+
				"and will draw as the terminal default", c, uint64(fg))
		}
	}
}

// Every colour the renderer can emit needs a style. A colour with no entry
// draws in the terminal's default, which is how a wood turns white: nothing
// fails, nothing is logged, and the map is simply wrong. The ramps are a
// contiguous block of constants, so anything inserted into the middle of the
// enum without a style lands here rather than on screen.
func TestEveryColourHasAStyle(t *testing.T) {
	for c := ascii.Color(0); c <= ascii.Wood5; c++ {
		if _, ok := palette[c]; !ok {
			t.Errorf("colour %d has no style, so it would draw as the terminal default", c)
		}
	}
}

// A wood is green wherever it stands. The colour ramps carry height, and the
// first pair of them carried it by draining the colour out of the map: the
// woods went sage and khaki up the ramp, and two thirds of the open ground
// was a grey-green or a khaki, because a settlement lives in the valley and
// the valley is most of what is on screen. Height is worth showing and it is
// not worth the whole picture.
func TestTheLandLooksLikeLand(t *testing.T) {
	for _, seed := range []uint64{1, 2, 3} {
		w := world.New(seed)
		m := observe.Take(w).Map
		var land, green, woods int
		for _, row := range ascii.Render(m) {
			for _, c := range row {
				wood := c.Color >= ascii.Wood0 && c.Color <= ascii.Wood5
				ground := c.Color >= ascii.Ground0 && c.Color <= ascii.Ground5
				if !wood && !ground {
					continue
				}
				style, ok := palette[c.Color]
				if !ok {
					t.Fatalf("seed %d: colour %d has no style", seed, c.Color)
				}
				fg, _, _ := style.Decompose()
				land++
				if vegetation(fg) {
					green++
				}
				if wood {
					woods++
					if !vegetation(fg) {
						t.Errorf("seed %d: a wood at band %d is not drawn green", seed, c.Color-ascii.Wood0)
					}
				}
			}
		}
		if woods == 0 {
			t.Fatalf("seed %d has no woods on it at all", seed)
		}
		// Most of a map is the lowland a settlement lives on, and the lowland
		// is green. Four fifths is comfortably clear of the three ramp steps
		// above the valley, which are meant to be bare.
		if share := float64(green) / float64(land); share < 0.8 {
			t.Errorf("seed %d: only %.0f%% of the land reads as green, want at least 80%%", seed, 100*share)
		}
	}
}
