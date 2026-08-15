// Package events implements the domain event bus. Modules publish domain
// events (FeePaid, ResultsPublished, ...); subscribers (notifications,
// dashboard cache invalidation) react. This is the microservice seam: today
// dispatch is in-process, tomorrow it becomes a Redis stream / message broker
// without touching publishing modules.
package events

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Event is a domain event envelope.
type Event struct {
	ID        uuid.UUID
	Type      string
	SchoolID  uuid.UUID
	UserID    uuid.UUID
	OccurredAt time.Time
	Payload   any
}

// Handler processes a single event.
type Handler func(ctx context.Context, ev Event)

// Bus fans out events to registered handlers asynchronously.
type Bus struct {
	mu       sync.RWMutex
	handlers []Handler
	log      *slog.Logger
}

// New creates a Bus.
func New(log *slog.Logger) *Bus {
	return &Bus{log: log}
}

// Subscribe registers a handler for all events.
func (b *Bus) Subscribe(h Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers = append(b.handlers, h)
}

// Publish enqueues an event for async dispatch.
func (b *Bus) Publish(ctx context.Context, ev Event) {
	b.mu.RLock()
	handlers := make([]Handler, len(b.handlers))
	copy(handlers, b.handlers)
	b.mu.RUnlock()

	for _, h := range handlers {
		go func(h Handler) {
			defer func() {
				if r := recover(); r != nil {
					b.log.Error("event handler panicked", "type", ev.Type, "panic", r)
				}
			}()
			h(ctx, ev)
		}(h)
	}
}

// NewEvent constructs an Event with defaults.
func NewEvent(evType string, schoolID, userID uuid.UUID, payload any) Event {
	return Event{
		ID:         uuid.New(),
		Type:       evType,
		SchoolID:   schoolID,
		UserID:     userID,
		OccurredAt: time.Now().UTC(),
		Payload:    payload,
	}
}
