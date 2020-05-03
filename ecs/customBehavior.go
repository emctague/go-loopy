package ecs

import (
	"reflect"
)

// BehaviorSystem is a shorthand for simple systems that only keep track of one type of entity, and only perform
// operations in Update.
//
// updater: func(e *ECS, ev EventContainer, delta float64, entityID uint64, componentSet interface{})
func BehaviorSystem(updater interface{}) func(e *ECS) {
	return func(e *ECS) {
		componentSetType := reflect.TypeOf(updater).In(4)
		updaterVal := reflect.ValueOf(updater)

		entitiesType := reflect.MapOf(reflect.TypeOf(uint64(0)), componentSetType)
		entities := reflect.New(entitiesType)
		entities.Elem().Set(reflect.MakeMap(entitiesType))

		events := e.Subscribe()

		go func() {
			for ev := range events {
				switch event := ev.Event.(type) {
				case EntityAddedEvent:
					UnpackEntity(event, entities.Interface())

				case EntityRemovedEvent:
					RemoveEntity(event.ID, entities.Interface())

				case UpdateBeginEvent:
					iter := entities.Elem().MapRange()
					for iter.Next() {
						updaterVal.Call([]reflect.Value{reflect.ValueOf(e), reflect.ValueOf(ev), reflect.ValueOf(event.Delta), iter.Key(), iter.Value()})
					}
				}

				ev.Wg.Done()
			}
		}()
	}
}
