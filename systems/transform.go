package systems

import (
	"github.com/emctague/go-loopy/ecs"
	"log"
)

// Transform is a component which represents the position of some entity.
type Transform struct {
	X        float64
	Y        float64
	Rotation float64
	Width    float64
	Height   float64
	ParentID uint64 // This transform will follow all the same movements as its parent. Set to 0 for 'no parent'.
}

// TransformEvent represents a change in the position of an entity.
// This may then fire transform events for any transforms which use the given entity as a parent.
type TransformEvent struct {
	EntityID uint64 // The entity to transform.
	OffsetX  float64
	OffsetY  float64
	Absolute bool // True if offsets are actually absolute screen coordinates.
}

// SetTransformParentEvent changes which entity a transform is parented to.
// This does not change the current position of the entity.
type SetTransformParentEvent struct {
	EntityID uint64 // The entity whose parent should be changed.
	ParentID uint64 // The new parent for the entity.
}

type eTransform struct{ *Transform }
type eTransformParent struct{ EntityID uint64 }

// TransformSystem keeps track of the transformation of entities and parenting of entity positions to those of other
// entities.
func TransformSystem(e *ecs.ECS) {
	events := e.Subscribe()
	entities := make(map[uint64]eTransform)
	parents := make(map[uint64][]eTransformParent)

	go func() {

		for ev := range events {
			switch event := ev.Event.(type) {
			case ecs.EntityAddedEvent:
				addedSet := ecs.UnpackEntity(event, &entities)
				if addedSet == nil {
					break
				}
				addedCSet := addedSet.(*eTransform)

				if addedCSet.ParentID != 0 {
					tempParentID := addedCSet.ParentID
					addedCSet.ParentID = 0
					setParent(&entities, &parents, event.ID, tempParentID)
				}

			case ecs.EntityRemovedEvent:
				// Change the parent to no-parent so that the entity is removed from any child lists.
				if entity, ok := entities[event.ID]; ok && entity.ParentID != 0 {
					setParent(&entities, &parents, event.ID, 0)
				}

				// Remove from parent list if appropriate
				if _, ok := parents[event.ID]; ok {
					delete(parents, event.ID)
				}

				ecs.RemoveEntity(event.ID, &entities)

			case SetTransformParentEvent:
				setParent(&entities, &parents, event.EntityID, event.ParentID)

			case TransformEvent:
				transformedEntity, ok := entities[event.EntityID]
				if !ok {
					log.Fatal("Transform event on entity without a transform!")
				}

				// Turn absolute values into relative ones.
				if event.Absolute {
					event.OffsetX = event.OffsetX - transformedEntity.X
					event.OffsetY = event.OffsetY - transformedEntity.Y
				}

				// Change the position
				transformedEntity.X += event.OffsetX
				transformedEntity.Y += event.OffsetY

				// Propagate transformation to children.
				children, ok := parents[event.EntityID]
				if ok {
					for _, child := range children {
						ev.Next <- TransformEvent{child.EntityID, event.OffsetX, event.OffsetY, false}
					}
				}
			}

			ev.Wg.Done()
		}
	}()
}

// Change the parent of the given entity to the given parent entity, updating the appropriate structures.
func setParent(entities *map[uint64]eTransform, parents *map[uint64][]eTransformParent, childID uint64, newParentID uint64) {
	comSet, ok := (*entities)[childID]
	if !ok {
		log.Fatal("Cannot set parent on nonexistent component")
	}

	// Exit if we're trying to set the same exact parent ID
	if comSet.Transform.ParentID == newParentID {
		return
	}

	// Remove an entry from the old parent's list if it isn't no parent (0)
	if comSet.Transform.ParentID != 0 {
		oldParentList, ok := (*parents)[comSet.Transform.ParentID]
		if !ok {
			log.Fatal("Entity's old parent is invalid!?")
		}

		if len(oldParentList) == 0 {
			delete(*parents, comSet.Transform.ParentID)
		} else {
			// Remove the item from the list
			for i, oldChild := range oldParentList {
				if oldChild.EntityID == childID {
					oldParentList[i] = oldParentList[len(oldParentList)-1]
					(*parents)[comSet.Transform.ParentID] = oldParentList[:len(oldParentList)-1]
					break
				}
			}
		}
	}

	comSet.Transform.ParentID = newParentID

	if newParentID == 0 {
		return
	}

	// Add an entry to the new parent's list if it isn't no parent (0)
	_, ok = (*parents)[newParentID]
	if !ok {
		(*parents)[newParentID] = []eTransformParent{}
	}
	(*parents)[newParentID] = append((*parents)[newParentID], eTransformParent{childID})

}
