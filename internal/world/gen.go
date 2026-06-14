package world

import "math"

// Generate fills w with a simple landscape: a rolling ground surface (grass over
// dirt over stone), occasional sand patches, and a few trees for parallax. It
// returns a reasonable spawn height (top of the ground at the world centre) so
// the camera can be placed above the surface.
func Generate(w *World) int {
	base := w.SY / 3
	for z := 0; z < w.SZ; z++ {
		for x := 0; x < w.SX; x++ {
			h := base + int(2*math.Sin(float64(x)*0.25)+2*math.Cos(float64(z)*0.25))
			if h < 1 {
				h = 1
			}
			if h >= w.SY {
				h = w.SY - 1
			}
			top := Grass
			// low-lying cells become sandy beaches
			if h <= base-1 {
				top = Sand
			}
			for y := 0; y <= h; y++ {
				switch {
				case y == h:
					w.Set(x, y, z, top)
				case y >= h-2:
					w.Set(x, y, z, Dirt)
				default:
					w.Set(x, y, z, Stone)
				}
			}
		}
	}

	// Scatter a few trees on a fixed lattice (deterministic — no RNG needed).
	for z := 6; z < w.SZ-6; z += 11 {
		for x := 5; x < w.SX-5; x += 13 {
			plantTree(w, x, z)
		}
	}

	// Spawn height: ground top at the world centre, plus eye height.
	cx, cz := w.SX/2, w.SZ/2
	gy := groundTop(w, cx, cz)
	return gy + 2
}

func groundTop(w *World, x, z int) int {
	for y := w.SY - 1; y >= 0; y-- {
		if w.Solid(x, y, z) {
			return y
		}
	}
	return 0
}

func plantTree(w *World, x, z int) {
	gy := groundTop(w, x, z)
	if gy <= 0 || gy >= w.SY-5 {
		return
	}
	trunk := gy + 3
	for y := gy + 1; y <= trunk; y++ {
		w.Set(x, y, z, Wood)
	}
	// leaf canopy
	for dy := 0; dy <= 1; dy++ {
		r := 2 - dy
		for dz := -r; dz <= r; dz++ {
			for dx := -r; dx <= r; dx++ {
				if dx == 0 && dz == 0 && dy == 0 {
					continue
				}
				w.Set(x+dx, trunk+dy, z+dz, Leaves)
			}
		}
	}
	w.Set(x, trunk+2, z, Leaves)
}
