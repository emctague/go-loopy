package main

import (
	"github.com/faiface/pixel/pixelgl"
)

type Player struct{}

type ePlayer struct {
	*Transform
	*Physics
	*Player
	*Interactor
}

var PlayerSystem = func(e *ECS, win *pixelgl.Window) {
	BehaviorSystem(func(e *ECS, ev EventContainer, delta float64, entityID uint64, player ePlayer) {
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
			ev.Next <- ApplyVelocityEvent{entityID, velX, velY}
		}
	})(e)
}
