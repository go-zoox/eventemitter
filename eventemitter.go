package eventemitter

import (
	"fmt"
	"sync"

	"github.com/go-zoox/logger"
	"github.com/go-zoox/safe"
)

// EventEmitter is a simple event emitter.
type EventEmitter struct {
	actionCh chan *action
	quitCh   chan struct{}

	handlers map[string][]Handle
	m        sync.Mutex
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
	if e.actionCh == nil {
		panic("event worker is not started or stopped	")
	}

	e.actionCh <- &action{
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
func (e *EventEmitter) Start() error {
	if e.actionCh != nil {
		return fmt.Errorf("event emitter is already started")
	}

	e.actionCh = make(chan *action)
	e.quitCh = make(chan struct{})

	go func() {
		for {
			select {
			case action := <-e.actionCh:
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
			case <-e.quitCh:
				return
			}
		}
	}()

	return nil
}

// Stop stops the event worker.
func (e *EventEmitter) Stop() error {
	return safe.Do(func() error {
		e.quitCh <- struct{}{}

		close(e.actionCh)
		close(e.quitCh)
		return nil
	})
}
