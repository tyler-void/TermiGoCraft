package render

import (
	"testing"

	"github.com/tyler-void/TermiGoCraft/internal/camera"
	"github.com/tyler-void/TermiGoCraft/internal/entity"
	"github.com/tyler-void/TermiGoCraft/internal/geom"
	"github.com/tyler-void/TermiGoCraft/internal/world"
)

// countNonSky returns how many pixels differ from the sky color.
func countNonSky(b *Buffer) int {
	n := 0
	for _, c := range b.Color {
		if c != Sky {
			n++
		}
	}
	return n
}

// TestEntityOcclusion is the proof that the hybrid shares one depth buffer:
// an entity behind a wall must not be drawn; the same entity in front must be.
func TestEntityOcclusion(t *testing.T) {
	cam := &camera.Camera{Pos: geom.Vec3{X: 0, Y: 0, Z: 0}, Yaw: 0, Pitch: 0, FOV: 1.2}
	cube := entity.Cube(1.5)
	red := geom.RGB{R: 255, G: 0, B: 0}

	// Wall directly ahead at z=10.
	w := world.New(40, 40, 40)
	// center the world around the camera line by offsetting via cam? Simpler:
	// put a wall plane the ray actually crosses. Camera at origin looks +Z, so
	// world cells with negative coords are out of bounds. Move camera into the
	// grid instead.
	cam.Pos = geom.Vec3{X: 20, Y: 20, Z: 2}
	for y := 0; y < 40; y++ {
		for x := 0; x < 40; x++ {
			w.Set(x, y, 10, world.Stone)
		}
	}

	// Case 1: entity BEHIND the wall (z=20). Render world then entity.
	behind := NewBuffer(60, 30)
	behind.Clear(Sky)
	RenderWorld(behind, w, cam)
	before := countNonSky(behind)
	RenderMesh(behind, cube, geom.Vec3{X: 20, Y: 20, Z: 20}, red, cam)
	afterBehind := countNonSky(behind)
	redBehind := countColor(behind, red)

	if redBehind != 0 {
		t.Fatalf("entity behind the wall leaked %d red pixels (should be fully occluded)", redBehind)
	}
	_ = before
	_ = afterBehind

	// Case 2: entity IN FRONT of the wall (z=5). Should draw red pixels.
	front := NewBuffer(60, 30)
	front.Clear(Sky)
	RenderWorld(front, w, cam)
	RenderMesh(front, cube, geom.Vec3{X: 20, Y: 20, Z: 5}, red, cam)
	if countColor(front, red) == 0 {
		t.Fatal("entity in front of the wall was not drawn")
	}
}

func countColor(b *Buffer, c geom.RGB) int {
	n := 0
	for _, p := range b.Color {
		if p == c {
			n++
		}
	}
	return n
}

func TestRenderMeshDrawsSomething(t *testing.T) {
	cam := &camera.Camera{Pos: geom.Vec3{X: 0, Y: 0, Z: 0}, Yaw: 0, Pitch: 0, FOV: 1.2}
	b := NewBuffer(60, 30)
	b.Clear(Sky)
	RenderMesh(b, entity.Cube(2), geom.Vec3{X: 0, Y: 0, Z: 6}, geom.RGB{R: 10, G: 200, B: 10}, cam)
	if countNonSky(b) == 0 {
		t.Fatal("a cube in front of the camera should rasterize visible pixels")
	}
}
