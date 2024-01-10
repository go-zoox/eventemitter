package eventemitter

import (
	"context"
	"sync"

	"github.com/go-zoox/safe"
)

// EventEmitter is a simple event emitter.
type EventEmitter interface {
	On(event string, handler Handler)
	Once(event string, handler Handler)
	Off(event string, handler Handler)
	Emit(event string, payload any)
}

type eventemitter struct {
	sync.Mutex

	ctx      context.Context
	handlers map[string][]Handler
	evtCh    chan *evt
}

type evt struct {
	event   string
	payload any
}

// Option is the option for the event emitter.
type Option struct {
	Context context.Context
}

// New creates a new EventEmitter.
func New(opts ...func(opt *Option)) EventEmitter {
	opt := &Option{
		Context: context.Background(),
	}
	for _, o := range opts {
		o(opt)
	}

	e := &eventemitter{
		handlers: make(map[string][]Handler),
		//
		ctx:   opt.Context,
		evtCh: make(chan *evt),
	}

	go e.work()

	return e
}

// On registers a handler for the given event type.
func (e *eventemitter) On(event string, handler Handler) {
	e.Lock()
	defer e.Unlock()

	e.handlers[event] = append(e.handlers[event], handler)
}

// Emit emits an event.
func (e *eventemitter) Emit(event string, payload any) {
	e.evtCh <- &evt{
		event:   event,
		payload: payload,
	}
}

// Once performs exactly one action.
func (e *eventemitter) Once(typ string, handler Handler) {
	var once sync.Once
	e.On(typ, HandleFunc(func(payload any) {
		once.Do(func() {
			handler.Serve(payload)
		})
	}))
}

// Off removes specify the given event type.
func (e *eventemitter) Off(typ string, handler Handler) {
	e.Lock()
	defer e.Unlock()

	for i, h := range e.handlers[typ] {
		if h.ID() == handler.ID() {
			e.handlers[typ] = append(e.handlers[typ][:i], e.handlers[typ][i+1:]...)
			break
		}
	}
}

func (e *eventemitter) work() {
	for {
		select {
		case <-e.ctx.Done():
			return
		case evt := <-e.evtCh:
			e.Lock()
			handlers, ok := e.handlers[evt.event]
			e.Unlock()

			if !ok {
				return
			}

			go func() {
				wg := &sync.WaitGroup{}
				for _, handler := range handlers {
					wg.Add(1)
					go func(handler Handler) {
						defer wg.Done()
						safe.Do(func() error {
							handler.Serve(evt.payload)
							return nil
						})
					}(handler)
				}
				wg.Wait()
			}()
		}
	}
}
