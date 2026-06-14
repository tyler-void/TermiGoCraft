package physics

import (
	"math"

	"github.com/tyler-void/TermiGoCraft/internal/geom"
	"github.com/tyler-void/TermiGoCraft/internal/world"
)

// Overlaps reports whether any solid block intersects the AABB defined by the
// feet position, half-width, and height. Allocation-free.
func Overlaps(w *world.World, feet geom.Vec3, half, height float64) bool {
	minX := int(math.Floor(feet.X - half))
	maxX := int(math.Floor(feet.X + half))
	minY := int(math.Floor(feet.Y))
	maxY := int(math.Floor(feet.Y + height))
	minZ := int(math.Floor(feet.Z - half))
	maxZ := int(math.Floor(feet.Z + half))
	for y := minY; y <= maxY; y++ {
		for z := minZ; z <= maxZ; z++ {
			for x := minX; x <= maxX; x++ {
				if w.Solid(x, y, z) {
					return true
				}
			}
		}
	}
	return false
}

// OccupiesCell reports whether cell (x,y,z) overlaps the body's AABB. Used by the
// block-placement guard so a block can't be placed inside the player (FR-009).
func (b *Body) OccupiesCell(x, y, z int) bool {
	minX := int(math.Floor(b.Feet.X - b.Half))
	maxX := int(math.Floor(b.Feet.X + b.Half))
	minY := int(math.Floor(b.Feet.Y))
	maxY := int(math.Floor(b.Feet.Y + b.Height))
	minZ := int(math.Floor(b.Feet.Z - b.Half))
	maxZ := int(math.Floor(b.Feet.Z + b.Half))
	return x >= minX && x <= maxX && y >= minY && y <= maxY && z >= minZ && z <= maxZ
}
