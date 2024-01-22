package eventemitter

import (
	"context"
	"sync"

	"github.com/go-zoox/logger"
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
	sync.RWMutex

	opt *Option

	ctx context.Context

	eventChans  map[string]chan *evt
	subscribers map[string][]Handler
}

type evt struct {
	event   string
	payload any
}

// Option is the option for the event emitter.
type Option struct {
	Context context.Context
	//
	BufferSize int
}

// New creates a new EventEmitter.
func New(opts ...func(opt *Option)) EventEmitter {
	opt := &Option{
		Context:    context.Background(),
		BufferSize: 10,
	}
	for _, o := range opts {
		o(opt)
	}

	e := &eventemitter{
		opt: opt,
		//
		ctx: opt.Context,
		//
		subscribers: make(map[string][]Handler),
		//
		eventChans: make(map[string]chan *evt),
	}

	return e
}

// On registers a handler for the given event type.
func (e *eventemitter) On(event string, subscriber Handler) {
	e.Lock()
	defer e.Unlock()

	if _, ok := e.eventChans[event]; !ok {
		eventChan := make(chan *evt, e.opt.BufferSize)
		e.eventChans[event] = eventChan

		go e.handleEvents(eventChan)
	}

	e.subscribers[event] = append(e.subscribers[event], subscriber)
}

// handleEvents 处理特定类型事件的goroutine。
func (e *eventemitter) handleEvents(eventChan chan *evt) {
	for {
		select {
		case <-e.ctx.Done():
			return
		case evt := <-eventChan:
			e.RLock()
			subscribers, ok := e.subscribers[evt.event]
			e.RUnlock()
			if !ok {
				break
			}

			for _, subscriber := range subscribers {
				err := safe.Do(func() error {
					subscriber.Serve(evt.payload)
					return nil
				})
				if err != nil {
					logger.Errorf("[eventemitter] failed to handle event(%s): %s", evt.event, err)
				}
			}
		}
	}
}

// Emit emits an event.
func (e *eventemitter) Emit(event string, payload any) {
	e.RLock()
	defer e.RUnlock()

	eventChan, ok := e.eventChans[event]
	if !ok {
		return // 没有对应的事件channel，无需操作
	}

	eventChan <- &evt{
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

	for i, h := range e.subscribers[typ] {
		if h.ID() == handler.ID() {
			e.subscribers[typ] = append(e.subscribers[typ][:i], e.subscribers[typ][i+1:]...)
			break
		}
	}
}
