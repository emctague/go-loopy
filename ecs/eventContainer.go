package ecs

import "sync"

// EventContainer stores an event and related utilities.
type EventContainer struct {
	Wg    *sync.WaitGroup  // This wait group is used to wait for all systems to handle the event.
	Event interface{}      // The event struct.
	Next  chan interface{} // Channel used to submit new events to occur immediately after the current one.
}

// Done should be called by all systems after handling a given event.
func (e EventContainer) Done() {
	e.Wg.Done()
}
