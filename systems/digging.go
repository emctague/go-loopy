package systems

import (
	"github.com/emctague/go-loopy/ecs"
	"github.com/faiface/pixel/pixelgl"
)

// Diggable is a component attached to objects which can be broken in a way which resembles mining in games like
// minecraft.
type Diggable struct {
	BaseDurability float64
	Durability     float64
}

type eDiggable struct {
	*Diggable
	*Transform
	*Renderable
}

// DigSystem provides the ability for the player to click over an entity and eventually break it.
var DigSystem = func(e *ecs.ECS, win *pixelgl.Window) {
	ecs.BehaviorSystem(func(e *ecs.ECS, ev ecs.EventContainer, delta float64, entityID uint64, diggable eDiggable) {

		if win.Pressed(pixelgl.MouseButtonLeft) {
			mp := win.MousePosition()

			if mp.X > diggable.X+5 || mp.X < diggable.X-5 || mp.Y > diggable.Y+5 || mp.Y < diggable.Y-5 {
				return
			}

			diggable.Durability -= delta
			if diggable.Durability <= 0 {
				ev.Next <- ecs.EntityRemovedEvent{ID: entityID}
			}
		}
	})(e)
}
