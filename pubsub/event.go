package pubsub

type Event interface {

	// Name returns event name of current event
	Name() string

	// Payload returns event payload of current event
	Payload() interface{}

	// String convert event data to string
	String() string
}
