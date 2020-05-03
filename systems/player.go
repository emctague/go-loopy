package systems

import (
	"github.com/emctague/go-loopy/ecs"
	"github.com/faiface/pixel/pixelgl"
)

// Player is a component which signifies that an entity is the player.
type Player struct{}

type ePlayer struct {
	*Transform
	*Physics
	*Player
	*Interactor
}

// PlayerSystem is a system which handles basic player controls
var PlayerSystem = func(e *ecs.ECS, win *pixelgl.Window) {
	ecs.BehaviorSystem(func(e *ecs.ECS, ev ecs.EventContainer, delta float64, entityID uint64, player ePlayer) {
		// Don't deal with movement in menus.
		if player.Menu != nil {
			return
		}

		// Apply velocity related to held arrow keys.
		var velX, velY float64
		if win.Pressed(pixelgl.KeyUp) {
			velY += 800 * delta
		}
		if win.Pressed(pixelgl.KeyDown) {
			velY -= 800 * delta
		}
		if win.Pressed(pixelgl.KeyLeft) {
			velX -= 800 * delta
		}
		if win.Pressed(pixelgl.KeyRight) {
			velX += 800 * delta
		}

		// Store the new velocity.
		if velX != 0 || velY != 0 {
			ev.Next <- ApplyVelocityEvent{EntityID: entityID, VelX: velX, VelY: velY}
		}
	})(e)
}
