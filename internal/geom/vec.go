// Package geom holds the small value types shared across the renderer and
// game logic: a 3D vector, an RGB color, and a triangle. Everything here is
// allocation-free (plain value types) so it never contributes GC pressure.
package geom

import "math"

// Vec3 is a point or direction in 3D space.
type Vec3 struct {
	X, Y, Z float64
}

// Add returns v+o.
func (v Vec3) Add(o Vec3) Vec3 { return Vec3{v.X + o.X, v.Y + o.Y, v.Z + o.Z} }

// Sub returns v-o.
func (v Vec3) Sub(o Vec3) Vec3 { return Vec3{v.X - o.X, v.Y - o.Y, v.Z - o.Z} }

// Scale returns v scaled by s.
func (v Vec3) Scale(s float64) Vec3 { return Vec3{v.X * s, v.Y * s, v.Z * s} }

// Dot returns the dot product v·o.
func (v Vec3) Dot(o Vec3) float64 { return v.X*o.X + v.Y*o.Y + v.Z*o.Z }

// Cross returns the cross product v×o.
func (v Vec3) Cross(o Vec3) Vec3 {
	return Vec3{
		X: v.Y*o.Z - v.Z*o.Y,
		Y: v.Z*o.X - v.X*o.Z,
		Z: v.X*o.Y - v.Y*o.X,
	}
}

// Length returns the Euclidean length of v.
func (v Vec3) Length() float64 { return math.Sqrt(v.Dot(v)) }

// Normalize returns a unit vector in the same direction, or the zero vector
// if v has zero length.
func (v Vec3) Normalize() Vec3 {
	l := v.Length()
	if l == 0 {
		return Vec3{}
	}
	return v.Scale(1 / l)
}

// RGB is a 24-bit truecolor value.
type RGB struct {
	R, G, B uint8
}

// Triangle is three world-space corners. Color is supplied per-mesh at draw
// time, so triangles stay pure geometry.
type Triangle struct {
	A, B, C Vec3
}
