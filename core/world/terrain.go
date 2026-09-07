package world

import (
	"math"

	"lreat/core/entity"
)

// GenerateTerrain lays down a meandering river, scattered forest, a fertility
// gradient falling away from the water, and a market on the bank near the
// middle. Everything else on the map is built by agents.
func (w *World) GenerateTerrain(width, height int) {
	g := NewGrid(width, height)

	phase := w.RNG.Float64() * 2 * math.Pi
	amp := float64(width) / 6
	riverX := func(y int) int {
		return width/2 + int(amp*math.Sin(float64(y)*0.12+phase))
	}
	for y := 0; y < height; y++ {
		half := 1
		if w.RNG.Float64() < 0.2 {
			half = 2
		}
		for dx := -half; dx <= half; dx++ {
			p := entity.Pos{X: riverX(y) + dx, Y: y}
			if g.In(p) {
				g.At(p).Terrain = Water
			}
		}
	}

	blobs := width * height / 150
	for i := 0; i < blobs; i++ {
		p := entity.Pos{X: w.RNG.IntN(width), Y: w.RNG.IntN(height)}
		for s := 0; s < 45; s++ {
			if g.In(p) {
				if t := g.At(p); t.Terrain == Grass {
					t.Terrain = Forest
					t.Wood = 0.6 + 0.4*w.RNG.Float64()
				}
			}
			p.X += w.RNG.IntN(3) - 1
			p.Y += w.RNG.IntN(3) - 1
		}
	}

	// Fertility: breadth-first distance from water, eight-connected.
	dist := make([]int, len(g.Tiles))
	queue := make([]int, 0, len(g.Tiles))
	for i := range g.Tiles {
		if g.Tiles[i].Terrain == Water {
			dist[i] = 0
			queue = append(queue, i)
		} else {
			dist[i] = -1
		}
	}
	for head := 0; head < len(queue); head++ {
		i := queue[head]
		p := entity.Pos{X: i % width, Y: i / width}
		for dy := -1; dy <= 1; dy++ {
			for dx := -1; dx <= 1; dx++ {
				q := entity.Pos{X: p.X + dx, Y: p.Y + dy}
				if !g.In(q) {
					continue
				}
				j := q.Y*width + q.X
				if dist[j] == -1 {
					dist[j] = dist[i] + 1
					queue = append(queue, j)
				}
			}
		}
	}
	for i := range g.Tiles {
		if g.Tiles[i].Terrain == Water {
			continue
		}
		d := float64(dist[i])
		g.Tiles[i].Fertility = 0.15 + 0.8*math.Max(0, 1-d/10)
	}

	center := entity.Pos{X: riverX(height/2) + 4, Y: height / 2}
	mp, ok := g.Nearest(center, width+height, func(_ entity.Pos, t *Tile) bool { return t.Terrain == Grass })
	if !ok {
		mp = entity.Pos{X: width / 2, Y: height / 2}
	}
	g.At(mp).Structure = Market
	w.Grid = g
	w.MarketPos = mp
}
