package eventemitter

import (
	"github.com/go-zoox/uuid"
)

// Handle is a function that can be registered to an event.
type Handle interface {
	ID() string
	Serve(payload any)
}

// HandleFunc creates a Handle from a function.
func HandleFunc(handler func(payload any)) Handle {
	return &handleFuncCreator{
		id: uuid.V4(),
		fn: handler,
	}
}

type handleFuncCreator struct {
	id string
	fn func(payload any)
}

// Serve calls the function.
func (h *handleFuncCreator) Serve(payload any) {
	h.fn(payload)
}

// ID returns the id of the handle.
func (h *handleFuncCreator) ID() string {
	return h.id
}
