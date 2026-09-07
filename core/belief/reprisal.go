package belief

// Severity is how transgressive an act is, independent of who is judging: the
// combined weight of the norms it violates. Unlike Conscience, it does not
// consult anybody's values, because a person can know an act is the sort of
// thing that gets punished here without personally thinking it wrong.
func Severity(v Valence) float64 {
	var s float64
	for _, x := range v {
		if x < 0 {
			s -= x
		}
	}
	return s
}

const (
	// ReprisalScale converts an expectation of punishment into the same units
	// as need satisfaction.
	ReprisalScale = 0.6
	// CautionLearned is how much witnessing a reprisal teaches an agent that
	// transgression is answered in this place.
	CautionLearned = 0.2
	// CautionSuffered is the much sharper lesson of being on the receiving end.
	CautionSuffered = 0.5
	// CautionDecay is how fast the lesson fades when nothing reinforces it.
	// Deterrence has to be maintained; a watch that stops watching stops
	// deterring within a few hundred ticks.
	CautionDecay = 0.003
)

// Reprisal is what an agent expects doing this act to cost it in trouble.
// It is the product of how punishable the act is and how much punishment this
// agent has learned to expect, so the same theft is cheap in a lawless
// settlement and expensive in a policed one.
func Reprisal(v Valence, caution float64) float64 {
	return Severity(v) * caution * ReprisalScale
}
