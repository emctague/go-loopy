package systems

import (
	"github.com/emctague/go-loopy/ecs"
	"github.com/faiface/pixel/pixelgl"
	"log"
)

// Physics is a component which specifies that an entity should be affected by the physics system.
type Physics struct {
	VelX       float64
	VelY       float64
	DragFactor float64
}

// ApplyVelocityEvent is used to add instantaneous velocity to an entity.
type ApplyVelocityEvent struct {
	EntityID uint64
	VelX     float64
	VelY     float64
}

// PhysicsSystem handles object physics (velocity, etc.)
func PhysicsSystem(e *ecs.ECS, win *pixelgl.Window) {
	type ComponentSet struct {
		*Transform
		*Physics
	}
	entities := make(map[uint64]ComponentSet)
	events := e.Subscribe()

	go func() {
		for ev := range events {
			switch event := ev.Event.(type) {
			case ecs.EntityAddedEvent:
				ecs.UnpackEntity(event, &entities)

			case ecs.EntityRemovedEvent:
				ecs.RemoveEntity(event.ID, &entities)

			case ApplyVelocityEvent:
				ent, ok := entities[event.EntityID]
				if !ok {
					log.Fatal("Cannot apply velocity to entity without physics")
				}
				ent.VelX += event.VelX
				ent.VelY += event.VelY

			case ecs.UpdateBeginEvent:
				for eid, entity := range entities {

					entity.VelX *= entity.DragFactor
					entity.VelY *= entity.DragFactor

					ev.Next <- TransformEvent{eid, entity.VelX * event.Delta, entity.VelY * event.Delta, false}
				}

			}

			ev.Wg.Done()
		}
	}()
}
