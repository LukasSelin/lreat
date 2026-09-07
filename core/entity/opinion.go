package entity

import "lreat/core/belief"

// MaxBonds caps how many people an agent keeps an opinion about. Beyond it
// the least recently seen is forgotten, so nobody carries the whole city in
// their head and reputation stays local.
const MaxBonds = 24

// bond returns the existing record for id, creating one if there is room and
// evicting the stalest if there is not.
func (a *Agent) bond(id ID, tick int) *Bond {
	for i := range a.Bonds {
		if a.Bonds[i].To == id {
			a.Bonds[i].LastSeen = tick
			return &a.Bonds[i]
		}
	}
	if len(a.Bonds) >= MaxBonds {
		stale := 0
		for i := range a.Bonds {
			if a.Bonds[i].LastSeen < a.Bonds[stale].LastSeen {
				stale = i
			}
		}
		a.Bonds[stale] = Bond{To: id, LastSeen: tick}
		return &a.Bonds[stale]
	}
	a.Bonds = append(a.Bonds, Bond{To: id, LastSeen: tick})
	return &a.Bonds[len(a.Bonds)-1]
}

// Look returns the opinion record for id, or nil if this agent has none.
func (a *Agent) Look(id ID) *Bond {
	for i := range a.Bonds {
		if a.Bonds[i].To == id {
			return &a.Bonds[i]
		}
	}
	return nil
}

// Judge moves this agent's moral regard for another, clamped to [-1,1].
func (a *Agent) Judge(id ID, delta float64, tick int) {
	b := a.bond(id, tick)
	b.Regard += delta
	if b.Regard > 1 {
		b.Regard = 1
	}
	if b.Regard < -1 {
		b.Regard = -1
	}
}

// Regard is how well this agent thinks of another. Zero for strangers.
func (a *Agent) Regard(id ID) float64 {
	if b := a.Look(id); b != nil {
		return b.Regard
	}
	return 0
}

// Rate moves this agent's belief about another's skill toward evidence.
func (a *Agent) Rate(id ID, s Skill, evidence float64, tick int) {
	b := a.bond(id, tick)
	b.Competence[s] = belief.Clamp(belief.Update(b.Competence[s], evidence, belief.LearnRate))
}

// Competence is what this agent believes another can do. Strangers are
// assumed unremarkable rather than useless, so newcomers get a first chance.
func (a *Agent) Competence(id ID, s Skill) float64 {
	if b := a.Look(id); b != nil {
		return b.Competence[s]
	}
	return 0.15
}

// Believes is what the agent thinks it can do itself.
func (a *Agent) Believes(s Skill) float64 { return a.Efficacy[s] }
