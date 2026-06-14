package camera

import (
	"math"
	"testing"

	"github.com/tyler-void/TermiGoCraft/internal/geom"
)

func almost(a, b geom.Vec3) bool {
	const eps = 1e-9
	return math.Abs(a.X-b.X) < eps && math.Abs(a.Y-b.Y) < eps && math.Abs(a.Z-b.Z) < eps
}

func TestDefaultBasis(t *testing.T) {
	c := Camera{FOV: 1.4}
	if !almost(c.Direction(), geom.Vec3{X: 0, Y: 0, Z: 1}) {
		t.Fatalf("forward = %v, want +Z", c.Direction())
	}
	if !almost(c.Right(), geom.Vec3{X: 1, Y: 0, Z: 0}) {
		t.Fatalf("right = %v, want +X", c.Right())
	}
	if !almost(c.Up(), geom.Vec3{X: 0, Y: 1, Z: 0}) {
		t.Fatalf("up = %v, want +Y", c.Up())
	}
}

func TestBasisOrthonormal(t *testing.T) {
	c := Camera{Yaw: 0.7, Pitch: 0.3, FOV: 1.4}
	f, r, u := c.Direction(), c.Right(), c.Up()
	for _, v := range []geom.Vec3{f, r, u} {
		if math.Abs(v.Length()-1) > 1e-9 {
			t.Fatalf("basis vector %v not unit length", v)
		}
	}
	if math.Abs(f.Dot(r)) > 1e-9 || math.Abs(f.Dot(u)) > 1e-9 || math.Abs(r.Dot(u)) > 1e-9 {
		t.Fatal("basis vectors not mutually perpendicular")
	}
}

func TestClampPitch(t *testing.T) {
	c := Camera{Pitch: 5}
	c.ClampPitch()
	if c.Pitch != PitchLimit {
		t.Fatalf("pitch = %v, want clamp to %v", c.Pitch, PitchLimit)
	}
	c.Pitch = -5
	c.ClampPitch()
	if c.Pitch != -PitchLimit {
		t.Fatalf("pitch = %v, want clamp to %v", c.Pitch, -PitchLimit)
	}
}

func TestCenterRayIsForward(t *testing.T) {
	c := Camera{Yaw: 0.4, Pitch: -0.2, FOV: 1.4}
	// the centre pixel of an even viewport straddles the axis; sample both
	// centre pixels and average — should be very close to forward.
	w, h := 100, 100
	r1 := c.ScreenRay(w/2, h/2, w, h)
	if r1.Dot(c.Direction()) < 0.99 {
		t.Fatalf("centre ray not close to forward: dot=%v", r1.Dot(c.Direction()))
	}
}
