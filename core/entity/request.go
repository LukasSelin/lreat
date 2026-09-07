package entity

// RequestID identifies a request. Zero means "no request".
type RequestID int

// RequestKind is what is being asked for.
type RequestKind int

const (
	// Deliver asks someone to hand over goods they have or can get.
	Deliver RequestKind = iota
	// Serve asks someone to do skilled work for the requester.
	Serve
)

// Request is one agent asking another to do something for payment.
//
// Nothing authors these. They appear when an agent wants something, believes
// it cannot or should not get it itself, and believes paying someone else is
// worth more than going without. The player receives them through exactly the
// same mechanism as everybody else.
type Request struct {
	ID        RequestID
	Requester ID
	// Directed is who was asked by name, chosen on believed competence.
	// Zero means the request is open to anyone who hears of it.
	Directed ID
	Kind     RequestKind
	Good     Good
	Amount   float64
	Skill    Skill
	Reward   float64
	Posted   int
	Deadline int
	// Greed marks a request posted out of surplus rather than need: the
	// requester could have done it, but would rather pay than spend the time.
	Greed bool
}

// Wants reports whether id is an acceptable doer for this request.
func (r *Request) Wants(id ID) bool {
	return r.Requester != id && (r.Directed == 0 || r.Directed == id)
}
