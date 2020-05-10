package systems

import (
	"github.com/emctague/go-loopy/ecs"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"math"
)

// Projectile is a component added to entities which fly around and bounce off walls, being removed after a particular
// number of bounces.
type Projectile struct {
	Bounces int
}

type eProjectile struct {
	*Transform
	*Physics
	*Projectile
}

// ProjectileSystem handles projectile movement and rotation
var ProjectileSystem = func(e *ecs.ECS, win *pixelgl.Window) {
	ecs.BehaviorSystem(func(e *ecs.ECS, ev ecs.EventContainer, delta float64, entityID uint64, projectile eProjectile) {
		if projectile.Y+projectile.VelY*delta < 20 {
			projectile.VelY = -projectile.VelY * 0.5
			projectile.Y = 20
			projectile.Bounces++
		}

		if projectile.Y+projectile.VelY*delta > win.Bounds().Max.Y-20 {
			projectile.VelY = -projectile.VelY
			projectile.Y = win.Bounds().Max.Y - 20
			projectile.Bounces++
		}

		if projectile.X+projectile.VelX*delta > win.Bounds().Max.X-20 {
			projectile.VelX = -projectile.VelX
			projectile.X = win.Bounds().Max.X - 20
			projectile.Bounces++
		}

		if projectile.X+projectile.VelX*delta < 20 {
			projectile.VelX = -projectile.VelX
			projectile.X = 20
			projectile.Bounces++
		}

		projectile.Rotation = pixel.V(projectile.VelX, projectile.VelY).Angle() + math.Pi/2

		if projectile.Bounces > 5 {
			e.RemoveEntity(entityID)
		}
	})(e)
}
