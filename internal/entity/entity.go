// Package entity holds non-block scene objects (and the held item) that are
// drawn by the triangle rasterizer rather than the voxel raycaster.
package entity

import "github.com/tyler-void/TermiGoCraft/internal/geom"

// Entity is a positioned, colored triangle mesh in the world.
type Entity struct {
	Pos   geom.Vec3
	Mesh  []geom.Triangle
	Color geom.RGB
}

// Cube returns the 12 triangles of an axis-aligned cube of the given side
// length, centered on the origin, with outward-facing winding.
func Cube(size float64) []geom.Triangle {
	s := size / 2
	// 8 corners
	p := [8]geom.Vec3{
		{X: -s, Y: -s, Z: -s}, // 0
		{X: s, Y: -s, Z: -s},  // 1
		{X: s, Y: s, Z: -s},   // 2
		{X: -s, Y: s, Z: -s},  // 3
		{X: -s, Y: -s, Z: s},  // 4
		{X: s, Y: -s, Z: s},   // 5
		{X: s, Y: s, Z: s},    // 6
		{X: -s, Y: s, Z: s},   // 7
	}
	quad := func(a, b, c, d int) []geom.Triangle {
		return []geom.Triangle{
			{A: p[a], B: p[b], C: p[c]},
			{A: p[a], B: p[c], C: p[d]},
		}
	}
	var tris []geom.Triangle
	tris = append(tris, quad(4, 5, 6, 7)...) // +Z front
	tris = append(tris, quad(1, 0, 3, 2)...) // -Z back
	tris = append(tris, quad(0, 4, 7, 3)...) // -X left
	tris = append(tris, quad(5, 1, 2, 6)...) // +X right
	tris = append(tris, quad(3, 7, 6, 2)...) // +Y top
	tris = append(tris, quad(0, 1, 5, 4)...) // -Y bottom
	return tris
}
