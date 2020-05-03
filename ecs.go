package main

import (
	"reflect"
	"sync"
	"time"
)

type ECS struct {
	EventReceivers  []chan EventContainer
	CurrentEvents   chan interface{}
	NextFrameEvents chan interface{}
	EIDCounter      uint64
	Running         bool
	LastFrame       time.Time
}

// Stores an event and related utilities.
type EventContainer struct {
	Wg    *sync.WaitGroup  // This wait group is used to wait for all systems to handle the event.
	Event interface{}      // The event struct.
	Next  chan interface{} // Channel used to submit new events to occur immediately after the current one.
}

// Should be called by all systems after handling a given event.
func (e EventContainer) Done() {
	e.Wg.Done()
}

// Triggered when an update begins
type UpdateBeginEvent struct {
	Delta float64
}

// Triggered after update begin. This is used to ensure certain tasks are completed before moving on to the next frame.
type UpdateEndEvent struct {
	Delta float64
}

// This event is triggered once the ECS is ready to run - it is safe to add required entities here.
type SetupEvent struct{}

// Triggered when an entity is added. Stores the entity ID and a map of all components.
type EntityAddedEvent struct {
	ID         uint64
	Components map[reflect.Type]interface{}
}

// Triggered when an entity is removed.
type EntityRemovedEvent struct {
	ID uint64
}

// Initialize a new ECS
func NewECS() ECS {
	return ECS{[]chan EventContainer{}, make(chan interface{}, 50),
		make(chan interface{}, 50), 1, false, time.Now()}
}

// Subscribe to events. Returns a channel which can be used to receive events.
// The passed event container's `Done` method should be called after each event is fully processed.
func (e *ECS) Subscribe() chan EventContainer {
	echan := make(chan EventContainer, 10)
	e.EventReceivers = append(e.EventReceivers, echan)
	return echan
}

// Publish an event to occur NOW. Using the event container's `Next` channel is the preferred way to do this from
// systems.
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

// Publish an event to be handled on the next frame.
func (e *ECS) PublishNextFrame(event interface{}) {
	e.NextFrameEvents <- event
}

// Publish an event to be handled later on the current frame. Note that this is unreliable if used during update end.
func (e *ECS) PublishThisFrame(event interface{}) {
	e.CurrentEvents <- event
}

// Run an update from the loop.
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

// Start an ECS loop.
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

// Stop the ECS loop.
func (e *ECS) Stop() {
	e.Running = false
}

// Add an entity with the given components. These should be pointers to structs. Returns the ID of the new entity.
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

// Remove the given entity from the ECS on the next frame.
func (e *ECS) RemoveEntity(id uint64) {
	e.PublishNextFrame(EntityRemovedEvent{id})
}

// Given an EntityAddedEvent and a map which maps uint64 (entity IDs) to some structure containing pointers to
// components, add an entry to that map and return the same new entry *if* the new entity contains all the required
// components.
func UnpackEntity(event EntityAddedEvent, entityMap interface{}) interface{} {
	structType := reflect.TypeOf(entityMap).Elem().Elem()

	newEntry := reflect.New(structType)

	// Populate fields on the new entry. Abort if one of the required components does not exist on it.
	for i := 0; i < structType.NumField(); i++ {
		com, ok := event.Components[structType.Field(i).Type]
		if !ok {
			return nil
		}

		newEntry.Elem().Field(i).Set(reflect.ValueOf(com))
	}

	// Store the value in the map and return it
	reflect.ValueOf(entityMap).Elem().SetMapIndex(reflect.ValueOf(event.ID), newEntry.Elem())
	return newEntry.Interface()
}

// Remove an entity from the given map of entity IDs (uint64) if it exists. Essentially undoes the work of UnpackEntity.
func RemoveEntity(eid uint64, entityMap interface{}) {
	eMap := reflect.ValueOf(entityMap).Elem()
	eMap.SetMapIndex(reflect.ValueOf(eid), reflect.Value{})
}
