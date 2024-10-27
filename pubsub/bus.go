package pubsub

type EventBus interface {

	// Register subscriber(s) with the bus
	Register(subscribers ...Subscriber)

	// Deliver an event
	Deliver(event Event)

	// Run the bus
	Run()

	// Stop the bus
	Stop()

	// IsRunning returns whether the bus is running
	IsRunning() bool
}
