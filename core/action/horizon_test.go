package action

import (
	"testing"

	"lreat/core/entity"
)

// A wider horizon looks further for ground worth having. It is the one
// measure of a mind that is a distance rather than a rate, and it is read
// here because the ordinary reach is a constant of this package.
func TestAWiderHorizonLooksFurther(t *testing.T) {
	var near, far entity.Agent
	near.Body, near.Mind = entity.Ordinary()
	far.Body, far.Mind = entity.Ordinary()
	near.Mind.Horizon, far.Mind.Horizon = 0.7, 1.3
	if ranging(&far) <= ranging(&near) {
		t.Fatalf("the wider horizon reaches %d and the narrower %d", ranging(&far), ranging(&near))
	}
	var undrawn entity.Agent
	if got := ranging(&undrawn); got != searchRadius {
		t.Fatalf("an agent with no mind drawn reaches %d, want the ordinary %d", got, searchRadius)
	}
}
