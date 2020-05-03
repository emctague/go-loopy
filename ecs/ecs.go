package ecs

import (
	"reflect"
	"sync"
	"time"
)

// ECS Represents an entity component system.
type ECS struct {
	EventReceivers  []chan EventContainer
	CurrentEvents   chan interface{}
	NextFrameEvents chan interface{}
	EIDCounter      uint64
	Running         bool
	LastFrame       time.Time
}

// NewECS initializes and returns a new ECS instance
func NewECS() ECS {
	return ECS{[]chan EventContainer{}, make(chan interface{}, 50),
		make(chan interface{}, 50), 1, false, time.Now()}
}

// Subscribe subscribes to events.
// It returns a channel which can be used to receive events.
// The passed event container's `Done` method should be called after each event is fully processed.
func (e *ECS) Subscribe() chan EventContainer {
	echan := make(chan EventContainer, 10)
	e.EventReceivers = append(e.EventReceivers, echan)
	return echan
}

// publishNow publishes an event to occur NOW.
// Using the event container's `Next` channel is the preferred way to do this from systems.
func (e *ECS) publishNow(event interface{}) {

	next := make(chan interface{}, 50)

	var wg sync.WaitGroup
	wg.Add(len(e.EventReceivers))

	for _, receiver := range e.EventReceivers {
		receiver <- EventContainer{&wg, event, next}
	}

	wg.Wait()

	close(next)
	for ev := range next {
		e.publishNow(ev)
	}
}

// PublishNextFrame publishes an event to be handled on the next frame.
func (e *ECS) PublishNextFrame(event interface{}) {
	e.NextFrameEvents <- event
}

// PublishThisFrame publishes an event to be handled later on the current frame.
// Note that this is unreliable if used during update end.
func (e *ECS) PublishThisFrame(event interface{}) {
	e.CurrentEvents <- event
}

// update runs an update from the loop.
func (e *ECS) update() {
	thisFrame := time.Now()
	delta := thisFrame.Sub(e.LastFrame).Seconds()
	e.LastFrame = thisFrame

	// Deliver update before anything else
	e.publishNow(UpdateBeginEvent{delta})

	// Swap out the new events channel, so that events caused past this point occur after the next update
	e.CurrentEvents = e.NextFrameEvents
	e.NextFrameEvents = make(chan interface{}, 50)

	// Iterate over all events for the current frame, including new ones being added on
	for len(e.CurrentEvents) != 0 {
		event := <-e.CurrentEvents
		e.publishNow(event)
	}

	e.publishNow(UpdateEndEvent{delta})
	close(e.CurrentEvents)
}

// Run starts the ECS loop. This is a blocking operation.
func (e *ECS) Run() {
	e.publishNow(SetupEvent{})

	e.Running = true
	for e.Running {
		e.update()
	}

	close(e.NextFrameEvents)

	for _, recv := range e.EventReceivers {
		close(recv)
	}
}

// Stop stops the ECS loop.
func (e *ECS) Stop() {
	e.Running = false
}

// AddEntity adds an entity with the given components. These should be pointers to structs.
// Returns the ID of the new entity.
func (e *ECS) AddEntity(components ...interface{}) uint64 {

	tm := make(map[reflect.Type]interface{})

	for _, c := range components {
		tm[reflect.TypeOf(c)] = c
	}

	newEID := e.EIDCounter
	e.EIDCounter += 1

	e.PublishNextFrame(EntityAddedEvent{
		newEID,
		tm,
	})

	return newEID
}

// RemoveEntity removes the given entity from the ECS on the next frame.
func (e *ECS) RemoveEntity(id uint64) {
	e.PublishNextFrame(EntityRemovedEvent{id})
}
