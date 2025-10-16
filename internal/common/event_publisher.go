package common

import (
	"context"
	"log"
)

// EventPublisher interface for publishing domain events
type EventPublisher interface {
	Publish(ctx context.Context, eventType string, data map[string]interface{}) error
}

// simpleEventPublisher is a simple in-memory event publisher
type simpleEventPublisher struct {
	handlers map[string][]EventHandler
}

// EventHandler is a function that handles events
type EventHandler func(ctx context.Context, eventType string, data map[string]interface{}) error

var globalPublisher = &simpleEventPublisher{
	handlers: make(map[string][]EventHandler),
}

// PublishEvent publishes an event to all registered handlers
func PublishEvent(ctx context.Context, eventType string, data map[string]interface{}) {
	if err := globalPublisher.Publish(ctx, eventType, data); err != nil {
		log.Printf("Failed to publish event %s: %v", eventType, err)
	}
}

// RegisterEventHandler registers a handler for a specific event type
func RegisterEventHandler(eventType string, handler EventHandler) {
	globalPublisher.handlers[eventType] = append(globalPublisher.handlers[eventType], handler)
}

// Publish publishes an event to all registered handlers
func (p *simpleEventPublisher) Publish(ctx context.Context, eventType string, data map[string]interface{}) error {
	handlers, exists := p.handlers[eventType]
	if !exists || len(handlers) == 0 {
		// No handlers registered, just log the event
		log.Printf("Event published: %s (no handlers registered)", eventType)
		return nil
	}

	// Call all registered handlers
	for _, handler := range handlers {
		if err := handler(ctx, eventType, data); err != nil {
			log.Printf("Event handler error for %s: %v", eventType, err)
			// Continue with other handlers even if one fails
		}
	}

	return nil
}

// ClearEventHandlers clears all event handlers (useful for testing)
func ClearEventHandlers() {
	globalPublisher.handlers = make(map[string][]EventHandler)
}
