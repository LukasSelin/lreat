// The OWL rendering of lreat's ontology is its own module so that lreat
// itself stays what its README says it is: a simulation with no dependency
// beyond the terminal library. `go build ./...` at the repository root walks
// past this directory, because a nested module is not part of the parent.
//
// gowl declares its module path as `gowl` rather than as a fetchable one, so
// it is reached by a replace directive and is expected to sit beside lreat:
//
//	repos/
//	  gowl/
//	  lreat/
//	    owl/   <- here
module lreat/owl

go 1.27.0

require (
	gowl v0.0.0
	lreat v0.0.0
)

replace (
	gowl => ../../gowl
	lreat => ..
)
