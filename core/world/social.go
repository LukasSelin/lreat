package world

import "lreat/core/entity"

// AgentAt returns the agent closest to p within radius, ignoring except.
// Ties go to the earliest in the slice, which keeps runs deterministic.
func (w *World) AgentAt(p entity.Pos, radius int, except *entity.Agent) *entity.Agent {
	var best *entity.Agent
	bestD := radius + 1
	for _, o := range w.Agents {
		if o == except {
			continue
		}
		if d := entity.Dist(p, o.Pos); d < bestD {
			best, bestD = o, d
		}
	}
	return best
}
