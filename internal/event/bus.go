package event

import (
	"context"
	"sync"

	log "github.com/sirupsen/logrus"
)

type Handler func(ctx context.Context, e Event)

type Bus struct {
	mu       sync.RWMutex
	handlers map[string][]Handler
	ch       chan Event
}

func NewBus(bufferSize int) *Bus {
	return &Bus{
		handlers: make(map[string][]Handler),
		ch:       make(chan Event, bufferSize),
	}
}

func (b *Bus) Subscribe(eventName string, h Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[eventName] = append(b.handlers[eventName], h)
}

func (b *Bus) Publish(e Event) {
	b.ch <- e
}

func (b *Bus) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case e := <-b.ch:
				b.dispatch(ctx, e)
			}
		}
	}()
}

func (b *Bus) dispatch(ctx context.Context, e Event) {
	b.mu.RLock()
	handlers := b.handlers[e.EventName()]
	b.mu.RUnlock()

	for _, h := range handlers {
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Errorf("event handler panic [%s]: %v", e.EventName(), r)
				}
			}()
			h(ctx, e)
		}()
	}
}
