// Package camera is the first-person viewpoint: a position plus yaw/pitch, from
// which it derives an orthonormal basis and per-pixel rays. It is pure geometry
// (depends only on geom); player concerns like reach and selected block live in
// the game package.
package camera

import (
	"math"

	"github.com/tyler-void/TermiGoCraft/internal/geom"
)

// PitchLimit caps how far up/down the camera can look, avoiding a flip-over.
const PitchLimit = 1.5

// Camera is the eye into the world.
type Camera struct {
	Pos   geom.Vec3
	Yaw   float64 // radians; 0 looks toward +Z, increasing turns toward +X
	Pitch float64 // radians; clamped to ±PitchLimit
	FOV   float64 // vertical field of view in radians
}

var worldUp = geom.Vec3{X: 0, Y: 1, Z: 0}

// Direction returns the forward unit vector.
func (c Camera) Direction() geom.Vec3 {
	return geom.Vec3{
		X: math.Cos(c.Pitch) * math.Sin(c.Yaw),
		Y: math.Sin(c.Pitch),
		Z: math.Cos(c.Pitch) * math.Cos(c.Yaw),
	}
}

// Right returns the right-hand unit vector (worldUp × forward).
func (c Camera) Right() geom.Vec3 {
	return worldUp.Cross(c.Direction()).Normalize()
}

// Up returns the camera up unit vector (forward × right).
func (c Camera) Up() geom.Vec3 {
	return c.Direction().Cross(c.Right()).Normalize()
}

// ClampPitch keeps Pitch within ±PitchLimit.
func (c *Camera) ClampPitch() {
	if c.Pitch > PitchLimit {
		c.Pitch = PitchLimit
	}
	if c.Pitch < -PitchLimit {
		c.Pitch = -PitchLimit
	}
}

// ScreenRay returns the (normalized) ray direction through pixel (px,py) of a
// w×h pixel viewport. Used for single-ray aiming; the world render pass inlines
// the same math for speed.
func (c Camera) ScreenRay(px, py, w, h int) geom.Vec3 {
	fwd := c.Direction()
	right := c.Right()
	up := c.Up()
	aspect := float64(w) / float64(h)
	tanF := math.Tan(c.FOV / 2)
	ndcX := 2*(float64(px)+0.5)/float64(w) - 1
	ndcY := 1 - 2*(float64(py)+0.5)/float64(h)
	dir := fwd.
		Add(right.Scale(ndcX * tanF * aspect)).
		Add(up.Scale(ndcY * tanF))
	return dir.Normalize()
}
