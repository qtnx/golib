package pubsub

import (
	"context"
	"reflect"

	"github.com/golibs-starter/golib/web/event"
)

var _bus EventBus = NewDefaultEventBus()
var _publisher Publisher = NewDefaultPublisher(_bus)

func GetEventBus() EventBus {
	return _bus
}

func GetPublisher() Publisher {
	return _publisher
}

func Register(subscribers ...Subscriber) {
	_bus.Register(subscribers...)
}

func Run() {
	_bus.Run()
}

func Publish(event Event) {
	_publisher.Publish(event)
}

// PublishEvent is a helper function to publish a message as an event directly.
func PublishEvent[T any](ctx context.Context, msg T) {
	Publish(MessageEvent[T]{
		AbstractEvent: event.NewAbstractEvent(ctx, reflect.TypeOf(msg).Name()),
		PayloadData:   msg,
	})
}

// PublishEventWithAbstractEvent is a helper function to publish a message as an event directly.
func PublishEventWithAbstractEvent[T any](ctx context.Context, abstractEvent *event.AbstractEvent, msg T) {
	Publish(MessageEvent[T]{
		AbstractEvent: abstractEvent,
		PayloadData:   msg,
	})
}

func ReplaceGlobal(bus EventBus, publisher Publisher) {
	_bus = bus
	_publisher = publisher
}
