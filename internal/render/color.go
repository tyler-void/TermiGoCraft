// Package render is the hybrid renderer: a per-pixel voxel raycaster for the
// world plus a filled-triangle rasterizer for entities, both writing into one
// shared color+depth framebuffer so occlusion between them is automatically
// correct. The steady-state frame performs no heap allocation.
package render

import (
	"github.com/tyler-void/TermiGoCraft/internal/geom"
	"github.com/tyler-void/TermiGoCraft/internal/world"
)

// Sky is the background color used to clear each frame.
var Sky = geom.RGB{R: 135, G: 206, B: 235}

// palette maps each block type to its base color. Per-face and per-distance
// shading are derived at draw time, so only the base color is stored.
var palette = [...]geom.RGB{
	world.Air:    {R: 0, G: 0, B: 0},
	world.Grass:  {R: 86, G: 170, B: 64},
	world.Dirt:   {R: 134, G: 96, B: 67},
	world.Stone:  {R: 128, G: 128, B: 128},
	world.Wood:   {R: 120, G: 85, B: 50},
	world.Leaves: {R: 60, G: 140, B: 55},
	world.Sand:   {R: 214, G: 198, B: 140},
}

// BlockColor returns the base color of a block type.
func BlockColor(b world.Block) geom.RGB {
	if int(b) >= len(palette) {
		return geom.RGB{}
	}
	return palette[b]
}

// shade multiplies a color by a brightness factor in [0,1].
func shade(c geom.RGB, f float64) geom.RGB {
	if f < 0 {
		f = 0
	}
	if f > 1 {
		f = 1
	}
	return geom.RGB{
		R: uint8(float64(c.R) * f),
		G: uint8(float64(c.G) * f),
		B: uint8(float64(c.B) * f),
	}
}

// faceFactor gives a flat directional-light factor per face so adjacent faces
// of the same block read distinctly (top brightest, bottom darkest).
func faceFactor(f world.Face) float64 {
	switch f {
	case world.FacePosY:
		return 1.0
	case world.FaceNegY:
		return 0.5
	case world.FacePosX, world.FaceNegX:
		return 0.8
	case world.FacePosZ, world.FaceNegZ:
		return 0.65
	}
	return 0.9
}

// distFactor darkens distant surfaces for depth cueing, never below 0.4.
func distFactor(dist, maxDist float64) float64 {
	f := 1 - (dist/maxDist)*0.6
	if f < 0.4 {
		f = 0.4
	}
	return f
}
