package ecs

import "reflect"

// UpdateBeginEvent is triggered when an update begins
type UpdateBeginEvent struct {
	Delta float64
}

// UpdateEndEvent is triggered after UpdateBeginEvent.
// This is used to ensure certain tasks are completed before moving on to the next frame.
type UpdateEndEvent struct {
	Delta float64
}

// SetupEvent is triggered once the ECS is ready to run - it is safe to add required entities here.
type SetupEvent struct{}

// EntityAddedEvent is triggered when an entity is added. Stores the entity ID and a map of all components.
type EntityAddedEvent struct {
	ID         uint64
	Components map[reflect.Type]interface{}
}

// EntityRemovedEvent is triggered when an entity is removed.
type EntityRemovedEvent struct {
	ID uint64
}
