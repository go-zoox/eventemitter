package eventemitter

import (
	"sync"
)

// EventEmitter is a simple event emitter.
type EventEmitter struct {
	ch       chan *action
	handlers map[string][]Handle
}

type action struct {
	Type    string
	Payload any
}

// New creates a new EventEmitter.
func New() *EventEmitter {
	return &EventEmitter{
		handlers: make(map[string][]Handle),
	}
}

// On registers a handler for the given event type.
func (e *EventEmitter) On(typ string, handler Handle) {
	e.handlers[typ] = append(e.handlers[typ], handler)
}

// Emit emits an event.
func (e *EventEmitter) Emit(typ string, payload any) {
	if e.ch == nil {
		panic("event worker is not started or stopped	")
	}

	e.ch <- &action{
		Type:    typ,
		Payload: payload,
	}
}

// Once performs exactly one action.
func (e *EventEmitter) Once(typ string, handler Handle) {
	var once sync.Once
	e.On(typ, HandleFunc(func(payload any) {
		once.Do(func() {
			handler.Serve(payload)
		})
	}))
}

// Off removes specify the given event type.
func (e *EventEmitter) Off(typ string, handler Handle) {
	for i, h := range e.handlers[typ] {
		if h.ID() == handler.ID() {
			e.handlers[typ] = append(e.handlers[typ][:i], e.handlers[typ][i+1:]...)
			break
		}
	}
}

// Start starts the event worker.
func (e *EventEmitter) Start() {
	e.ch = make(chan *action)

	go func() {
		for {
			select {
			case action := <-e.ch:
				for _, handler := range e.handlers[action.Type] {
					handler.Serve(action.Payload)
				}
			}
		}
	}()
}

// Stop stops the event worker.
func (e *EventEmitter) Stop() {
	close(e.ch)
	e.ch = nil
}
