package eventemitter

import (
	"fmt"
	"sync"

	"github.com/go-zoox/logger"
	"github.com/go-zoox/safe"
)

// EventEmitter is a simple event emitter.
type EventEmitter struct {
	chs      map[string]chan *action
	handlers map[string][]Handle

	quitCh chan struct{}
	m      sync.Mutex
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

	e := &EventEmitter{}

	if !opt.DisableAutoStart {
		e.Start()
	}

	return e
}

// On registers a handler for the given event type.
func (e *EventEmitter) On(typ string, handler Handle) {
	e.m.Lock()
	defer e.m.Unlock()

	if _, ok := e.chs[typ]; !ok {
		ch := make(chan *action)

		// worker
		go func(typ string, ch chan *action) {
			for {
				select {
				case data := <-ch:
					// @TODO channel closed
					if data == nil {
						return
					}

					wg := &sync.WaitGroup{}

					for _, cb := range e.handlers[typ] {
						wg.Add(1)

						go func(cb Handle) {
							if err := cb.Serve(data.Payload); err != nil {
								// @TODO error handling
								logger.Errorf("failed to handle event: %s", err)
							}
							wg.Done()
						}(cb)
					}

					wg.Wait()
				case <-e.quitCh:
					return
				}
			}
		}(typ, ch)

		e.chs[typ] = ch
	}

	e.handlers[typ] = append(e.handlers[typ], handler)
}

// Emit emits an event.
func (e *EventEmitter) Emit(typ string, payload any) {
	if e.chs == nil {
		logger.Warnf("[emit][ignore] event worker is not started or stopped")
		return
	}

	if _, ok := e.chs[typ]; !ok {
		// panic(fmt.Errorf("event: %s is not registered", typ))
		return
	}

	e.chs[typ] <- &action{
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
	if e.chs != nil {
		return fmt.Errorf("event emitter is already started")
	}

	e.chs = make(map[string]chan *action)
	e.handlers = make(map[string][]Handle)
	e.quitCh = make(chan struct{})

	return nil
}

// Stop stops the event worker.
func (e *EventEmitter) Stop() error {
	return safe.Do(func() error {
		e.quitCh <- struct{}{}

		e.chs = nil
		e.handlers = nil

		// @TODO donot close channel, let gc do its thing
		//
		// close(e.quitCh)
		// for _, c := range e.chs {
		// 	close(c)
		// }

		return nil
	})
}
