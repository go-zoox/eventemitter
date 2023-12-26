package eventemitter

import (
	"sync"

	"github.com/go-zoox/logger"
)

// EventEmitter is a simple event emitter.
type EventEmitter struct {
	ch       chan *action
	handlers map[string][]Handle
	m        sync.Mutex

	//
	isStarted bool
}

type action struct {
	Type    string
	Payload any
}

type Option struct {
	DisableAutoStart bool `json:"disable-auto-start"`
}

// New creates a new EventEmitter.
func New(opts ...func(opt *Option)) *EventEmitter {
	opt := &Option{}
	for _, o := range opts {
		o(opt)
	}

	e := &EventEmitter{
		handlers: make(map[string][]Handle),
	}

	if !opt.DisableAutoStart {
		e.Start()
	}

	return e
}

// On registers a handler for the given event type.
func (e *EventEmitter) On(typ string, handler Handle) {
	e.m.Lock()
	e.handlers[typ] = append(e.handlers[typ], handler)
	e.m.Unlock()
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
	e.m.Lock()
	handlers := e.handlers[typ]
	e.m.Unlock()

	for i, h := range handlers {
		if h.ID() == handler.ID() {
			e.handlers[typ] = append(e.handlers[typ][:i], e.handlers[typ][i+1:]...)
			break
		}
	}
}

// Start starts the event worker.
func (e *EventEmitter) Start() {
	if e.ch != nil {
		return
	}

	e.ch = make(chan *action)

	go func() {
		for {
			select {
			case action := <-e.ch:
				e.m.Lock()
				handlers := e.handlers[action.Type]
				e.m.Unlock()

				for _, handler := range handlers {
					go func(handler Handle) {
						if err := handler.Serve(action.Payload); err != nil {
							logger.Errorf("event handler error: %v (type: %s)", err, action.Type)
						}
					}(handler)
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
