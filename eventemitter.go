package eventemitter

import (
	"sync"

	"github.com/go-zoox/safe"
)

// EventEmitter is a simple event emitter.
type EventEmitter struct {
	sync.Mutex

	handlers map[string][]Handler
}

// Option is the option for the event emitter.
type Option struct {
}

// New creates a new EventEmitter.
func New(opts ...func(opt *Option)) *EventEmitter {
	opt := &Option{}
	for _, o := range opts {
		o(opt)
	}

	e := &EventEmitter{
		handlers: make(map[string][]Handler),
	}

	return e
}

// On registers a handler for the given event type.
func (e *EventEmitter) On(event string, handler Handler) {
	e.Lock()
	defer e.Unlock()

	e.handlers[event] = append(e.handlers[event], handler)
}

// Emit emits an event.
func (e *EventEmitter) Emit(event string, payload any) {
	e.Lock()
	handlers, ok := e.handlers[event]
	e.Unlock()

	if !ok {
		return
	}

	wg := &sync.WaitGroup{}
	for _, handler := range handlers {
		wg.Add(1)
		go func(handler Handler) {
			defer wg.Done()
			safe.Do(func() error {
				handler.Serve(payload)
				return nil
			})
		}(handler)
	}
	wg.Wait()
}

// Once performs exactly one action.
func (e *EventEmitter) Once(typ string, handler Handler) {
	var once sync.Once
	e.On(typ, HandleFunc(func(payload any) {
		once.Do(func() {
			handler.Serve(payload)
		})
	}))
}

// Off removes specify the given event type.
func (e *EventEmitter) Off(typ string, handler Handler) {
	e.Lock()
	defer e.Unlock()

	for i, h := range e.handlers[typ] {
		if h.ID() == handler.ID() {
			e.handlers[typ] = append(e.handlers[typ][:i], e.handlers[typ][i+1:]...)
			break
		}
	}
}
