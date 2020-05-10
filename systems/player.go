package systems

import (
	"github.com/emctague/go-loopy/ecs"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"math"
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
var PlayerSystem = func(e *ecs.ECS, win *pixelgl.Window, pic *pixel.Picture) {
	ecs.BehaviorSystem(func(e *ecs.ECS, ev ecs.EventContainer, delta float64, entityID uint64, player ePlayer) {
		// Don't deal with movement in menus.
		if player.Menu != nil {
			return
		}

		mousePos := win.MousePosition()
		playerPos := pixel.V(player.X, player.Y)
		diff := mousePos.To(playerPos).Unit().Rotated(math.Pi)
		player.Rotation = diff.Angle() - math.Pi/2

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

		if win.JustPressed(pixelgl.MouseButtonLeft) {
			e.AddEntity(&Transform{X: player.X, Y: player.Y, Rotation: player.Rotation}, &Physics{VelX: diff.X * 200, VelY: diff.Y * 200, DragFactor: 1}, &Renderable{Sprite: pixel.NewSprite(*pic, pixel.R(69, 28, 69+8, 28+8))}, &Projectile{}, &Bullet{})
		}

		// Store the new velocity.
		if velX != 0 || velY != 0 {
			ev.Next <- ApplyVelocityEvent{EntityID: entityID, VelX: velX, VelY: velY}
		}
	})(e)
}
