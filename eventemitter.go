package eventemitter

import "sync"

// EventEmitter is a simple event emitter.
type EventEmitter struct {
	ch       chan *action
	handlers map[string][]func(payload any)
}

type action struct {
	Type    string
	Payload any
}

// New creates a new EventEmitter.
func New() *EventEmitter {
	return &EventEmitter{
		handlers: make(map[string][]func(payload any)),
	}
}

// On registers a handler for the given event type.
func (e *EventEmitter) On(typ string, handler func(payload any)) {
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
func (e *EventEmitter) Once(typ string, handler func(payload any)) {
	var once sync.Once
	e.On(typ, func(payload any) {
		once.Do(func() {
			handler(payload)
		})
	})
}

// Start starts the event worker.
func (e *EventEmitter) Start() {
	e.ch = make(chan *action)

	go func() {
		for {
			select {
			case action := <-e.ch:
				for _, handler := range e.handlers[action.Type] {
					handler(action.Payload)
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
