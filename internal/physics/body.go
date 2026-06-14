// Package physics gives the player a physical body: an axis-aligned bounding box
// integrated under gravity with per-axis voxel collision, plus a free-fly mode.
// It depends only on geom and world (no renderer / tcell), so it is fully
// testable headlessly, and Step performs no heap allocation.
package physics

import "github.com/tyler-void/TermiGoCraft/internal/geom"

// Mode is the movement model.
type Mode uint8

const (
	Walk Mode = iota // gravity + collision
	Fly              // free movement, no gravity/collision
)

// Body is the player's AABB. Feet is the bottom-center of the box.
type Body struct {
	Feet      geom.Vec3
	Vel       geom.Vec3
	Half      float64 // half width/depth
	Height    float64 // total height
	EyeHeight float64 // eye offset above feet
	Grounded  bool
	Mode      Mode
}

// Intent is the per-tick movement request built from input.
type Intent struct {
	Forward, Strafe float64 // -1..1, camera-relative (yaw)
	Up, Down, Jump  bool
}

// NewBody returns a player-sized body at the given feet position, in Walk mode.
func NewBody(feet geom.Vec3) *Body {
	return &Body{Feet: feet, Half: 0.3, Height: 1.8, EyeHeight: 1.6, Mode: Walk}
}

// Eye is the camera position derived from the body each tick.
func (b *Body) Eye() geom.Vec3 {
	return geom.Vec3{X: b.Feet.X, Y: b.Feet.Y + b.EyeHeight, Z: b.Feet.Z}
}
