package render

import (
	"math"

	"github.com/tyler-void/TermiGoCraft/internal/camera"
	"github.com/tyler-void/TermiGoCraft/internal/world"
)

// ViewDist is the maximum distance the world raycaster looks before giving up
// on a pixel (it stays sky).
const ViewDist = 48.0

// RenderWorld casts one ray per pixel into the voxel world (Pass A). The camera
// basis and projection constants are computed once per frame; the per-pixel
// inner loop only does value math and a DDA walk, so it allocates nothing.
func RenderWorld(b *Buffer, w *world.World, cam *camera.Camera) {
	fwd := cam.Direction()
	right := cam.Right()
	up := cam.Up()
	aspect := float64(b.W) / float64(b.H)
	tanF := math.Tan(cam.FOV / 2)
	invW := 1.0 / float64(b.W)
	invH := 1.0 / float64(b.H)

	for y := 0; y < b.H; y++ {
		ndcY := 1 - 2*(float64(y)+0.5)*invH
		oy := up.Scale(ndcY * tanF)
		for x := 0; x < b.W; x++ {
			ndcX := 2*(float64(x)+0.5)*invW - 1
			dir := fwd.Add(right.Scale(ndcX * tanF * aspect)).Add(oy).Normalize()

			hit := w.Cast(cam.Pos, dir, ViewDist)
			if !hit.OK {
				continue
			}
			// Perpendicular (camera-Z) depth: project the Euclidean hit
			// distance onto the forward axis. This is the shared depth metric
			// the rasterizer also writes, and it removes fisheye distortion.
			z := hit.Dist * dir.Dot(fwd)
			f := faceFactor(hit.Face) * distFactor(hit.Dist, ViewDist)
			b.TestSet(x, y, float32(z), shade(BlockColor(hit.Block), f))
		}
	}
}
