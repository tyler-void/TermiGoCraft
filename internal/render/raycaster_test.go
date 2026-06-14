package render

import (
	"math"
	"testing"

	"github.com/tyler-void/TermiGoCraft/internal/camera"
	"github.com/tyler-void/TermiGoCraft/internal/geom"
	"github.com/tyler-void/TermiGoCraft/internal/world"
)

// buildWall returns a world with a solid wall plane at z=12 and a camera in the
// middle looking down +Z toward it.
func buildWall() (*world.World, *camera.Camera) {
	w := world.New(40, 40, 40)
	for y := 0; y < 40; y++ {
		for x := 0; x < 40; x++ {
			w.Set(x, y, 12, world.Stone)
		}
	}
	cam := &camera.Camera{Pos: geom.Vec3{X: 20, Y: 20, Z: 1}, Yaw: 0, Pitch: 0, FOV: 1.2}
	return w, cam
}

func TestRenderWorldFillsAndMissesSky(t *testing.T) {
	w, cam := buildWall()
	b := NewBuffer(60, 30)
	b.Clear(Sky)
	RenderWorld(b, w, cam)

	// Center pixel should hit the wall (depth set, color not sky).
	ci := (b.H/2)*b.W + b.W/2
	if math.IsInf(float64(b.Depth[ci]), 1) {
		t.Fatal("center pixel should have hit the wall (finite depth)")
	}
	if b.Color[ci] == Sky {
		t.Fatal("center pixel should not be sky")
	}
}

func TestRenderWorldNoFisheye(t *testing.T) {
	// A flat wall viewed head-on should have near-constant perpendicular depth
	// across the row (no fisheye): edge depth ≈ center depth.
	w, cam := buildWall()
	b := NewBuffer(80, 40)
	b.Clear(Sky)
	RenderWorld(b, w, cam)

	row := b.H / 2
	center := b.Depth[row*b.W+b.W/2]
	edge := b.Depth[row*b.W+b.W-3]
	if math.IsInf(float64(center), 1) || math.IsInf(float64(edge), 1) {
		t.Skip("wall not covering sampled pixels at this FOV")
	}
	if d := math.Abs(float64(center - edge)); d > 1.0 {
		t.Fatalf("perpendicular depth varies too much across a flat wall (%.3f): fisheye?", d)
	}
}

func BenchmarkFrame(b *testing.B) {
	w, cam := buildWall()
	buf := NewBuffer(120, 40) // 240x80 px, the default target
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Clear(Sky)
		RenderWorld(buf, w, cam)
	}
}
