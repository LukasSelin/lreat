package action

import (
	"lreat/core/entity"
	"lreat/core/habit"
	"lreat/core/world"
)

// The priors: for each action, the kind of moment it belongs to. These are
// not payoffs. They are the situation an agent would recognise as calling
// for the act before it has any experience of its own. Every agent starts
// from these and every agent's copy drifts with what happens to it. The
// moral coordinates never drift; they are matched against the agent's own
// values, so an act that belongs to dishonest moments is one an honest agent
// rarely recognises as fitting.
//
// Coordinates are in [-1, 1]. Unmentioned coordinates are 0, which means the
// act does not care. Fit is by direction, so a prior should name only the
// coordinates that predict its moment: every extra coordinate dilutes the
// ones that matter, and a coordinate that is the same for everyone (a
// newcomer's low belief in every skill, say) is a tax on whichever acts
// mention it. Compare with the situation table in docs/action-space.md.

// Reach at birth. Ordinary living is fully within reach, and standing guard
// is ordinary living: it is the one public good in the catalog, and a
// settlement that has to discover it first has died of disorder before it
// does. Crafts and learning begin far off and are brought closer by study,
// teaching, and discovery.
const (
	reachEveryday = 1.0
	reachGuard    = 1.0
	reachCraft    = 0.5
	reachPave     = 0.5
	reachStudy    = 0.6
	reachTeach    = 0.3
)

// uses is a Skilled that always names the same skill.
func uses(s entity.Skill) func(*entity.Agent, *world.World) (entity.Skill, bool) {
	return func(*entity.Agent, *world.World) (entity.Skill, bool) { return s, true }
}

// atTarget is a With for acts done to whoever is at the target.
func atTarget(a *entity.Agent, w *world.World, target entity.Pos) *entity.Agent {
	return intended(a, w, target)
}

func init() {
	// Rest recovers a little of the body, so its moment is a mild version
	// of eating's. It has no constant coordinate on purpose: a fallback
	// that always fits a little would beat every real but moderate match,
	// and under sampling the least bad option is fallback enough.
	seed(Rest, reachEveryday, habit.Signature{habit.Hunger: 1})
	seed(Eat, reachEveryday, habit.Signature{
		habit.Hunger: 1, habit.Near: 0.5,
	})
	// Foraging belongs to the green half of the year, which is the half in
	// which the forest puts back most of what is taken from it. A
	// settlement that forages through a winter is eating what will not be
	// replaced until spring.
	seed(Forage, reachEveryday, habit.Signature{
		habit.Hunger: 0.8, habit.Food: -0.8, habit.Chill: -0.3, habit.Near: 0.6,
	})
	// Farming is not what hunger calls for; foraging is. Farming is what an
	// industrious person with a field nearby does whether or not the larder
	// is low. That is tradition, and tradition is what carries farming
	// through its bad years: a first field feeds less than the forest, and
	// were farming judged by hunger alone the harvests would learn it away
	// before the settlement had learned to rotate its fields or anyone had
	// learned the work. Once they have, a field feeds two or three meals to
	// the forest's one, the harvests thank it, and the forest empties.
	// Farming is a season's work as well as a habit: little comes up in a
	// frost, and the tradition that carries it is a tradition of sowing in
	// spring.
	seed(Farm, reachEveryday, habit.Signature{
		habit.Food: -0.3, habit.Chill: -0.5, habit.Industry: 0.7, habit.Near: 0.5, habit.Skill: 0.3,
	})
	Farm.Skilled = uses(entity.Farming)
	// Wood is measured against the cost of a house, so "enough wood" reads
	// as +1 exactly when a shelter can be built. Gathering belongs to the
	// unsheltered moment more than to the empty-handed one; building to the
	// unsheltered moment with the wood in hand.
	//
	// What both lean into is exposure and not the weather itself. The cold
	// alone is the same news to everybody, so an act that names it is an act
	// the whole settlement turns to at once: told it is February, the people
	// who already have roofs put down their tools and go to the woods with
	// everyone else, and the settlement eats its winter stores while the
	// timber piles up in the hands of people who do not need it. Told
	// instead what each body is actually standing in, the same coordinate
	// sorts them: whoever is out in the cold goes for wood, whoever is under
	// a roof stays in the field. That is the difference between a settlement
	// that suffers a winter and one that houses itself over one.
	//
	// The weather itself stays on gathering at a mild weight, because
	// felling is winter work whoever does it - the sap is down and there is
	// least else to do - and off building altogether: a house is worth
	// raising in any weather, and a settlement that waited for the frost to
	// start would spend the winter it was building for outdoors.
	//
	// Exposure is weighted as lightly as it is because the hours it moves
	// have to come from somewhere, and where they come from is the fields.
	// Leant on harder it houses more people and feeds fewer: at 0.5 the
	// share with no house falls to 32% and the median population falls with
	// it from 42 to 29. At 0.3 the settlement gets most of the housing for
	// the smallest part of the harvest, and that is the whole of why this
	// number is small rather than round.
	seed(GatherWood, reachEveryday, habit.Signature{
		habit.Unsafe: 0.6, habit.Wood: -0.6, habit.Shelter: -0.6,
		habit.Chill: 0.3, habit.Exposure: 0.3, habit.Near: 0.5,
	})
	seed(BuildShelter, reachEveryday, habit.Signature{
		habit.Unsafe: 1, habit.Wood: 1, habit.Shelter: -1, habit.Exposure: 0.3,
	})
	BuildShelter.Skilled = uses(entity.Building)
	seed(Sell, reachEveryday, habit.Signature{
		habit.Food: 1, habit.Wealth: -0.6, habit.Near: 0.4,
	})
	seed(Buy, reachEveryday, habit.Signature{
		habit.Hunger: 0.8, habit.Food: -1, habit.Wealth: 0.6, habit.Near: 0.5,
	})
	seed(Guard, reachGuard, habit.Signature{
		habit.Unsafe: 0.6, habit.Company: 0.5, habit.Order: -1, habit.Charity: 0.4, habit.Tradition: 0.3,
	})
	Guard.Skilled = uses(entity.Guarding)
	seed(Socialize, reachEveryday, habit.Signature{
		habit.Lonely: 1, habit.Company: 0.8, habit.Near: 0.4, habit.Rapport: 0.7,
	})
	Socialize.With = atTarget
	seed(Craft, reachCraft, habit.Signature{
		habit.Hunger: -0.3, habit.Unproven: 0.8, habit.Wood: 0.7, habit.Skill: 0.5,
	})
	Craft.Skilled = uses(entity.Crafting)
	seed(Teach, reachTeach, habit.Signature{
		habit.Unproven: 0.8, habit.Company: 0.7, habit.Charity: 0.3, habit.Rapport: 0.5, habit.Skill: 1,
	})
	Teach.Skilled = func(a *entity.Agent, _ *world.World) (entity.Skill, bool) {
		s, _ := a.BestSkill()
		return s, true
	}
	Teach.With = atTarget
	// Paving belongs to the settled moment: somebody already under a roof,
	// with wood past what that roof needed, among neighbours whose comings
	// and goings have worn a way. Shelter is what separates it from gathering
	// and building, which want the opposite; charity and industry are what it
	// shares with standing guard. It says nothing of hunger and nothing of
	// skill: a newcomer believes itself unskilled at everything, and a prior
	// that mentions skill taxes exactly the acts a young settlement needs.
	seed(Pave, reachPave, habit.Signature{
		habit.Wood: 0.6, habit.Shelter: 0.7, habit.Company: 0.6,
		habit.Charity: 0.5, habit.Industry: 0.6, habit.Near: 0.5,
	})
	seed(Study, reachStudy, habit.Signature{
		habit.Hunger: -0.5, habit.Unsafe: -0.3, habit.Curious: 1, habit.Tradition: -0.4,
	})
	Study.Skilled = uses(entity.Scholarship)
	seed(Steal, reachEveryday, habit.Signature{
		habit.Hunger: 1, habit.Food: -1, habit.Order: -0.3,
		habit.Honesty: -1, habit.Caution: -0.7, habit.Near: 0.6, habit.Rapport: -0.4,
	})
	Steal.With = func(a *entity.Agent, w *world.World, _ entity.Pos) *entity.Agent {
		return nearestWith(a, w, reachRadius, hasSpareFood)
	}
	seed(Give, reachEveryday, habit.Signature{
		habit.Hunger: -0.4, habit.Lonely: 0.3, habit.Food: 0.7, habit.Charity: 1, habit.Near: 0.5, habit.Rapport: 0.5,
	})
	Give.With = func(a *entity.Agent, w *world.World, _ entity.Pos) *entity.Agent {
		return nearestWith(a, w, reachRadius, inNeed)
	}
	seed(Fulfil, reachEveryday, habit.Signature{
		habit.Unproven: 0.5, habit.Wealth: -0.5, habit.Industry: 0.5, habit.Near: 0.4, habit.Rapport: 0.4, habit.Skill: 0.5,
	})
	Fulfil.Skilled = func(a *entity.Agent, w *world.World) (entity.Skill, bool) {
		r := BestRequest(a, w)
		if r == nil {
			return 0, false
		}
		if r.Kind == entity.Serve {
			return r.Skill, true
		}
		return entity.Farming, true
	}
	Fulfil.With = func(a *entity.Agent, w *world.World, _ entity.Pos) *entity.Agent {
		if r := BestRequest(a, w); r != nil {
			return w.Find(r.Requester)
		}
		return nil
	}
	seed(Retaliate, reachEveryday, habit.Signature{
		habit.Unproven: 0.6, habit.Honesty: 0.2, habit.Charity: -0.5, habit.Caution: -0.4, habit.Near: 0.5, habit.Rapport: 1,
	})
	Retaliate.With = func(a *entity.Agent, w *world.World, _ entity.Pos) *entity.Agent {
		t, _ := Grudge(a, w)
		return t
	}
}

func seed(d *Def, reach float64, prior habit.Signature) {
	d.Prior = prior
	d.Reach0 = reach
}
