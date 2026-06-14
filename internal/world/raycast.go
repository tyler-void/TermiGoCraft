package world

import (
	"math"

	"github.com/tyler-void/TermiGoCraft/internal/geom"
)

// Face identifies which side of a block a ray entered through. The value is the
// face whose outward normal points back toward the ray, i.e. the face you'd
// place a new block against.
type Face uint8

const (
	FaceNone Face = iota
	FacePosX
	FaceNegX
	FacePosY
	FaceNegY
	FacePosZ
	FaceNegZ
)

// Normal returns the unit offset of the face's outward normal.
func (f Face) Normal() (int, int, int) {
	switch f {
	case FacePosX:
		return 1, 0, 0
	case FaceNegX:
		return -1, 0, 0
	case FacePosY:
		return 0, 1, 0
	case FaceNegY:
		return 0, -1, 0
	case FacePosZ:
		return 0, 0, 1
	case FaceNegZ:
		return 0, 0, -1
	}
	return 0, 0, 0
}

// Hit is the result of a ray cast into the world.
type Hit struct {
	X, Y, Z int     // the solid cell that was hit
	Face    Face    // face entered through (for placement)
	Dist    float64 // Euclidean distance along the (normalized) ray
	Block   Block   // the block type hit
	OK      bool    // false when nothing solid was found within maxDist
}

// Cast walks the voxel grid from origin along dir using the Amanatides & Woo
// "Fast Voxel Traversal" algorithm and returns the first solid cell within
// maxDist. dir need not be normalized. The returned Dist is Euclidean distance;
// callers that need perpendicular (camera-Z) depth multiply by the cosine of
// the angle between the ray and the camera forward axis.
func (w *World) Cast(origin, dir geom.Vec3, maxDist float64) Hit {
	d := dir.Normalize()

	x := int(math.Floor(origin.X))
	y := int(math.Floor(origin.Y))
	z := int(math.Floor(origin.Z))

	// Starting already inside a solid cell counts as an immediate hit.
	if w.Solid(x, y, z) {
		return Hit{X: x, Y: y, Z: z, Face: FaceNone, Dist: 0, Block: w.At(x, y, z), OK: true}
	}

	stepX, tMaxX, tDeltaX := initAxis(origin.X, d.X)
	stepY, tMaxY, tDeltaY := initAxis(origin.Y, d.Y)
	stepZ, tMaxZ, tDeltaZ := initAxis(origin.Z, d.Z)

	var face Face
	t := 0.0
	for t <= maxDist {
		switch {
		case tMaxX < tMaxY && tMaxX < tMaxZ:
			x += stepX
			t = tMaxX
			tMaxX += tDeltaX
			if stepX > 0 {
				face = FaceNegX
			} else {
				face = FacePosX
			}
		case tMaxY < tMaxZ:
			y += stepY
			t = tMaxY
			tMaxY += tDeltaY
			if stepY > 0 {
				face = FaceNegY
			} else {
				face = FacePosY
			}
		default:
			z += stepZ
			t = tMaxZ
			tMaxZ += tDeltaZ
			if stepZ > 0 {
				face = FaceNegZ
			} else {
				face = FacePosZ
			}
		}

		if t > maxDist {
			break
		}
		// Once we have exited the grid on an axis and are still moving away on
		// that axis, no solid cell can ever be reached again — stop early.
		if (x < 0 && stepX < 0) || (x >= w.SX && stepX > 0) ||
			(y < 0 && stepY < 0) || (y >= w.SY && stepY > 0) ||
			(z < 0 && stepZ < 0) || (z >= w.SZ && stepZ > 0) {
			break
		}
		if w.Solid(x, y, z) {
			return Hit{X: x, Y: y, Z: z, Face: face, Dist: t, Block: w.At(x, y, z), OK: true}
		}
	}
	return Hit{OK: false}
}

// initAxis returns the step direction, the distance to the first grid boundary,
// and the per-cell distance increment for one axis.
func initAxis(start, dir float64) (step int, tMax, tDelta float64) {
	switch {
	case dir > 0:
		step = 1
		tDelta = 1 / dir
		tMax = (math.Floor(start) + 1 - start) / dir
	case dir < 0:
		step = -1
		tDelta = -1 / dir
		tMax = (start - math.Floor(start)) / -dir
	default:
		step = 0
		tDelta = math.Inf(1)
		tMax = math.Inf(1)
	}
	return
}
