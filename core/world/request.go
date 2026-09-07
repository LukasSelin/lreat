package world

import "lreat/core/entity"

// MaxOpenRequests bounds the board so a large settlement cannot spend all its
// time asking rather than doing.
const MaxOpenRequests = 200

// Post adds a request to the board and returns its id, or zero if the board
// is full.
func (w *World) Post(r entity.Request) entity.RequestID {
	if len(w.Requests) >= MaxOpenRequests {
		return 0
	}
	w.nextReqID++
	r.ID = w.nextReqID
	r.Posted = w.Tick
	w.Requests = append(w.Requests, &r)
	return r.ID
}

// Request returns the open request with the given id, or nil.
func (w *World) Request(id entity.RequestID) *entity.Request {
	for _, r := range w.Requests {
		if r.ID == id {
			return r
		}
	}
	return nil
}

// Close removes a request from the board.
func (w *World) Close(id entity.RequestID) {
	for i, r := range w.Requests {
		if r.ID == id {
			w.Requests = append(w.Requests[:i], w.Requests[i+1:]...)
			return
		}
	}
}

// HasRequestFrom reports whether this agent already has something on the
// board. One at a time keeps the board legible and the agents decisive.
func (w *World) HasRequestFrom(id entity.ID) bool {
	for _, r := range w.Requests {
		if r.Requester == id {
			return true
		}
	}
	return false
}
