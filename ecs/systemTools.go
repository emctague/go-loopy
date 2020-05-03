package ecs

import "reflect"

// UnpackEntity takes an event and adds it to the given map if its components match the fields in the map's value struct.
// entityMap should be a pointer to a map[uint64]eCollection, where eCollection is some struct type (generally with a
// name beginning with e for entity) which represents some subset of an entity's components that are used by the system.
// This also returns a pointer to the the added structure, or nil if the entity didn't meet requirements.
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

// RemoveEntity Removes an entity from the given map of entity IDs (uint64) if it exists. Essentially undoes the work
// of UnpackEntity.
func RemoveEntity(eid uint64, entityMap interface{}) {
	eMap := reflect.ValueOf(entityMap).Elem()
	eMap.SetMapIndex(reflect.ValueOf(eid), reflect.Value{})
}
