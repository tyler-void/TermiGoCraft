package physics

import (
	"math"

	"github.com/tyler-void/TermiGoCraft/internal/geom"
	"github.com/tyler-void/TermiGoCraft/internal/world"
)

// Tunables (cells, seconds).
const (
	Gravity    = -30.0 // downward acceleration
	JumpSpeed  = 9.5   // initial jump velocity (~1.5 block jump)
	WalkSpeed  = 5.0
	FlySpeed   = 8.0
	MaxFall    = -45.0 // terminal velocity
	horizDecay = 0.5   // per-tick horizontal velocity decay when no input
	subStep    = 0.05  // collision sweep granularity (< 1 cell ⇒ no tunneling)
)

// Step advances the body by one fixed timestep dt. yaw is the camera heading
// used to orient horizontal movement. It is allocation-free and deterministic.
func Step(w *world.World, b *Body, in Intent, yaw, dt float64) {
	if b.Mode == Fly {
		stepFly(w, b, in, yaw, dt)
		return
	}
	stepWalk(w, b, in, yaw, dt)
}

// horizDir converts camera-relative intent into a world-space horizontal
// direction (flattened to the ground plane), capped to unit length.
func horizDir(in Intent, yaw float64) geom.Vec3 {
	sinY, cosY := math.Sin(yaw), math.Cos(yaw)
	fwd := geom.Vec3{X: sinY, Y: 0, Z: cosY}
	right := geom.Vec3{X: cosY, Y: 0, Z: -sinY}
	h := fwd.Scale(in.Forward).Add(right.Scale(in.Strafe))
	h.Y = 0
	if h.Length() > 1 {
		h = h.Normalize()
	}
	return h
}

func stepWalk(w *world.World, b *Body, in Intent, yaw, dt float64) {
	// If we begin overlapping solids (spawn/edit), lift to free space.
	for i := 0; i < w.SY && Overlaps(w, b.Feet, b.Half, b.Height); i++ {
		b.Feet.Y += 1.0
	}

	// Horizontal velocity from intent (instant, with decay on release).
	h := horizDir(in, yaw)
	if h.X != 0 || h.Z != 0 {
		b.Vel.X = h.X * WalkSpeed
		b.Vel.Z = h.Z * WalkSpeed
	} else {
		b.Vel.X *= horizDecay
		b.Vel.Z *= horizDecay
	}

	// Gravity.
	b.Vel.Y += Gravity * dt
	if b.Vel.Y < MaxFall {
		b.Vel.Y = MaxFall
	}

	// Jump uses the grounded state from the previous tick.
	if in.Jump && b.Grounded {
		b.Vel.Y = JumpSpeed
		b.Grounded = false
	}

	// Integrate, resolving each axis independently (slide along walls).
	b.Grounded = false
	resolveAxis(w, b, 0, b.Vel.X*dt)
	resolveAxis(w, b, 1, b.Vel.Y*dt)
	resolveAxis(w, b, 2, b.Vel.Z*dt)
	clampBounds(w, b)
}

func stepFly(w *world.World, b *Body, in Intent, yaw, dt float64) {
	h := horizDir(in, yaw)
	b.Vel.X = h.X * FlySpeed
	b.Vel.Z = h.Z * FlySpeed
	vy := 0.0
	if in.Up {
		vy += FlySpeed
	}
	if in.Down {
		vy -= FlySpeed
	}
	b.Vel.Y = vy
	b.Feet = b.Feet.Add(b.Vel.Scale(dt))
	b.Grounded = false
	clampBounds(w, b)
}

// resolveAxis sweeps the body along one axis in small steps, stopping flush
// against the first solid it would enter and zeroing velocity on that axis.
func resolveAxis(w *world.World, b *Body, axis int, delta float64) {
	if delta == 0 {
		return
	}
	dir := 1.0
	if delta < 0 {
		dir = -1.0
	}
	remaining := math.Abs(delta)
	for remaining > 0 {
		s := subStep
		if remaining < s {
			s = remaining
		}
		prev := axisGet(b, axis)
		axisSet(b, axis, prev+dir*s)
		if Overlaps(w, b.Feet, b.Half, b.Height) {
			axisSet(b, axis, prev) // back off to last free position
			switch axis {
			case 0:
				b.Vel.X = 0
			case 1:
				if dir < 0 {
					b.Grounded = true
				}
				b.Vel.Y = 0
			case 2:
				b.Vel.Z = 0
			}
			return
		}
		remaining -= s
	}
}

func axisGet(b *Body, axis int) float64 {
	switch axis {
	case 0:
		return b.Feet.X
	case 1:
		return b.Feet.Y
	default:
		return b.Feet.Z
	}
}

func axisSet(b *Body, axis int, v float64) {
	switch axis {
	case 0:
		b.Feet.X = v
	case 1:
		b.Feet.Y = v
	default:
		b.Feet.Z = v
	}
}

// clampBounds keeps the body inside the world (no walking off the edge or
// falling out the bottom).
func clampBounds(w *world.World, b *Body) {
	if b.Feet.X < b.Half {
		b.Feet.X = b.Half
	}
	if b.Feet.X > float64(w.SX)-b.Half {
		b.Feet.X = float64(w.SX) - b.Half
	}
	if b.Feet.Z < b.Half {
		b.Feet.Z = b.Half
	}
	if b.Feet.Z > float64(w.SZ)-b.Half {
		b.Feet.Z = float64(w.SZ) - b.Half
	}
	if b.Feet.Y < 0 {
		b.Feet.Y = 0
		if b.Vel.Y < 0 {
			b.Vel.Y = 0
			b.Grounded = true
		}
	}
	if b.Feet.Y > float64(w.SY)-b.Height {
		b.Feet.Y = float64(w.SY) - b.Height
	}
}
