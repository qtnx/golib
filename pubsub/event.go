package pubsub

import (
	"context"

	baseEvent "github.com/golibs-starter/golib/web/event"
)

type Event interface {
	// Identifier returns the ID of event
	Identifier() string

	// Name returns event name of current event
	Name() string

	// Context returns the current context of event
	Context() context.Context

	// Payload returns event payload of current event
	Payload() interface{}

	// String convert event data to string
	String() string
}

type MessageEvent[T any] struct {
	*baseEvent.AbstractEvent
	PayloadData T `json:"payload"`
}

// Payload return payload of event
func (e MessageEvent[T]) Payload() interface{} {
	return e.PayloadData
}

// String() convert event to string
func (e MessageEvent[T]) String() string {
	return e.ToString(e)
}
