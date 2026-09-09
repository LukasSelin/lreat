package world

import (
	"testing"

	"lreat/core/entity"
)

func TestDistanceIsShortestRoundTheCylinder(t *testing.T) {
	g := NewGrid(100, 20)
	a, b := entity.Pos{X: 2, Y: 5}, entity.Pos{X: 97, Y: 5}
	if d := g.Dist(a, b); d != 95 {
		t.Fatalf("a valley 100 wide puts them %d apart, want 95", d)
	}
	g.Wrap = true
	if d := g.Dist(a, b); d != 5 {
		t.Fatalf("a globe 100 round puts them %d apart, want 5", d)
	}
	if d := g.Delta(a, b); d != (entity.Pos{X: -5, Y: 0}) {
		t.Fatalf("the short way from 2 to 97 is %v, want five steps west", d)
	}
	if p := g.Toward(entity.Pos{X: 0, Y: 5}, b); p != (entity.Pos{X: 99, Y: 5}) {
		t.Fatalf("a step from the west edge toward 97 lands on %v, want 99", p)
	}
	if p := g.Norm(entity.Pos{X: -1, Y: 3}); p != (entity.Pos{X: 99, Y: 3}) {
		t.Fatalf("west of the west edge is %v, want 99", p)
	}
	if g.In(entity.Pos{X: 500, Y: 3}) != true || g.In(entity.Pos{X: 3, Y: -1}) {
		t.Fatal("every column is on a globe; a row past the pole is not")
	}
}

func TestNearestFindsGroundAcrossTheSeam(t *testing.T) {
	g := NewGrid(40, 10)
	g.Wrap = true
	g.Turn(entity.Pos{X: 38, Y: 5}, Rock)
	p, ok := g.Nearest(entity.Pos{X: 1, Y: 5}, 10, func(_ entity.Pos, t *Tile) bool { return t.Terrain == Rock })
	if !ok || p != (entity.Pos{X: 38, Y: 5}) {
		t.Fatalf("the outcrop three tiles west across the seam was found at %v, %v", p, ok)
	}
	// A ring wider than the map reads each column once and still finds it.
	p, ok = g.Nearest(entity.Pos{X: 20, Y: 0}, 30, func(_ entity.Pos, t *Tile) bool { return t.Terrain == Rock })
	if !ok || p != (entity.Pos{X: 38, Y: 5}) {
		t.Fatalf("from the far side of the map the outcrop was found at %v, %v", p, ok)
	}
}

func TestAWrappedMapRoutesAcrossTheSeam(t *testing.T) {
	g := NewGrid(60, 12)
	g.Wrap = true
	from, to := entity.Pos{X: 1, Y: 6}, entity.Pos{X: 58, Y: 6}
	path := g.Path(from, to)
	if len(path) != 3 {
		t.Fatalf("the way from 1 to 58 round the seam is %d steps: %v", len(path), path)
	}
	for _, p := range path {
		if p.X < 0 || p.X >= g.W {
			t.Fatalf("a step off the map in the route: %v", path)
		}
	}
	if c := g.TravelCost(from, to); c != 3 {
		t.Fatalf("three tiles of grass round the seam cost %v", c)
	}
	g.Wrap = false
	if len(g.Path(from, to)) != 57 {
		t.Fatalf("the same way across a valley is %d steps", len(g.Path(from, to)))
	}
}
