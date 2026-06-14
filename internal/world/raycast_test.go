package world

import (
	"testing"

	"github.com/tyler-void/TermiGoCraft/internal/geom"
)

func TestCastHitsExpectedCellAndFace(t *testing.T) {
	w := New(10, 10, 10)
	w.Set(5, 0, 0, Stone)
	// Ray from x=0 traveling +X at the centre of the y=0,z=0 row.
	origin := geom.Vec3{X: 0.5, Y: 0.5, Z: 0.5}
	hit := w.Cast(origin, geom.Vec3{X: 1, Y: 0, Z: 0}, 20)
	if !hit.OK {
		t.Fatal("expected a hit")
	}
	if hit.X != 5 || hit.Y != 0 || hit.Z != 0 {
		t.Fatalf("hit cell = (%d,%d,%d), want (5,0,0)", hit.X, hit.Y, hit.Z)
	}
	if hit.Face != FaceNegX {
		t.Fatalf("entered face = %v, want FaceNegX", hit.Face)
	}
	if hit.Block != Stone {
		t.Fatalf("hit block = %v, want Stone", hit.Block)
	}
}

func TestCastMissReturnsNotOK(t *testing.T) {
	w := New(10, 10, 10)
	w.Set(9, 0, 0, Stone)
	origin := geom.Vec3{X: 0.5, Y: 0.5, Z: 0.5}
	// maxDist too short to reach the block at x=9.
	hit := w.Cast(origin, geom.Vec3{X: 1, Y: 0, Z: 0}, 3)
	if hit.OK {
		t.Fatalf("expected miss within maxDist, got hit at (%d,%d,%d)", hit.X, hit.Y, hit.Z)
	}
}

func TestCastEmptyWorldNoHit(t *testing.T) {
	w := New(8, 8, 8)
	hit := w.Cast(geom.Vec3{X: 4, Y: 4, Z: 4}, geom.Vec3{X: 0, Y: 1, Z: 0}, 100)
	if hit.OK {
		t.Fatal("empty world should produce no hit and must terminate")
	}
}

func TestCastFaceFromAbove(t *testing.T) {
	w := New(8, 8, 8)
	w.Set(4, 0, 4, Stone)
	// Drop straight down onto the top face.
	hit := w.Cast(geom.Vec3{X: 4.5, Y: 5, Z: 4.5}, geom.Vec3{X: 0, Y: -1, Z: 0}, 20)
	if !hit.OK || hit.Face != FacePosY {
		t.Fatalf("downward cast: ok=%v face=%v, want hit on FacePosY", hit.OK, hit.Face)
	}
}
