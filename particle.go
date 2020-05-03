package main

type Particle struct {
	Lifetime float64
}

type eParticle struct {
	*Transform
	*Physics
	*Particle
	*DebugCircle
}

// Deals with a very specific type of onscreen particle - a circle whose size decreases over time and which obeys some
// gravity.
var ParticleSystem = BehaviorSystem(func(e *ECS, ev EventContainer, delta float64, entityID uint64, particle eParticle) {
	// Track particle lifetime
	particle.Lifetime -= delta
	if particle.Lifetime <= 0 {
		e.RemoveEntity(entityID)
	}

	// Reduce particle size
	particle.Radius = 5 * (particle.Lifetime / 0.25)

	// Apply velocity to the particle
	ev.Next <- ApplyVelocityEvent{entityID, 0, -1500 * delta}
})
