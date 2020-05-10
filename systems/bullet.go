package systems

import (
	"github.com/emctague/go-loopy/ecs"
)

// Bullet is a component added to objects which can harm enemies upon collision.
type Bullet struct{}

// Enemy is a component added to objects which may take damage from bullets and eventually die.
type Enemy struct {
	Health int
}

type eBullet struct {
	*Transform
	*Bullet
}

type eEnemy struct {
	*Transform
	*Enemy
}

// BulletSystem handles gunshot collisions
func BulletSystem(e *ecs.ECS) {
	bullets := make(map[uint64]eBullet)
	enemies := make(map[uint64]eEnemy)
	events := e.Subscribe()

	go func() {
		for ev := range events {
			switch event := ev.Event.(type) {
			case ecs.EntityAddedEvent:
				ecs.UnpackEntity(event, &bullets)
				ecs.UnpackEntity(event, &enemies)

			case ecs.EntityRemovedEvent:
				ecs.RemoveEntity(event.ID, &bullets)
				ecs.RemoveEntity(event.ID, &enemies)

			case ecs.UpdateBeginEvent:
				for bid, bullet := range bullets {
					for eid, enemy := range enemies {
						// eww why
						if bullet.X > enemy.X-enemy.Width/2 && bullet.X < enemy.X+enemy.Width/2 && bullet.Y > enemy.Y-enemy.Height/2 && bullet.Y < enemy.Y+enemy.Height/2 {
							enemy.Health--
							e.RemoveEntity(bid)

							if enemy.Health < 1 {
								e.RemoveEntity(eid)
							}

							break
						}
					}
				}

			}

			ev.Wg.Done()
		}
	}()
}
