package physics

import (
	"math"
	"testing"

	"github.com/tyler-void/TermiGoCraft/internal/geom"
	"github.com/tyler-void/TermiGoCraft/internal/world"
)

const dt = 1.0 / 30

// ground returns a world with a solid floor at cell y=0 (top surface y=1).
func ground() *world.World {
	w := world.New(20, 20, 20)
	for z := 0; z < 20; z++ {
		for x := 0; x < 20; x++ {
			w.Set(x, 0, z, world.Stone)
		}
	}
	return w
}

func TestRestOnGround(t *testing.T) {
	w := ground()
	b := NewBody(geom.Vec3{X: 10, Y: 6, Z: 10})
	for i := 0; i < 120; i++ {
		Step(w, b, Intent{}, 0, dt)
	}
	if !b.Grounded {
		t.Fatal("body should be grounded after falling")
	}
	if math.Abs(b.Feet.Y-1) > 0.1 {
		t.Fatalf("rest Y = %v, want ~1 (floor top)", b.Feet.Y)
	}
	if Overlaps(w, b.Feet, b.Half, b.Height) {
		t.Fatal("resting body overlaps the ground")
	}
}

func TestWallSlide(t *testing.T) {
	w := ground()
	// wall plane at x=12 above the floor
	for y := 1; y < 5; y++ {
		for z := 0; z < 20; z++ {
			w.Set(12, y, z, world.Stone)
		}
	}
	b := NewBody(geom.Vec3{X: 11, Y: 1, Z: 10})
	// yaw=PI/2 ⇒ forward=+X (into wall), right=-Z (slide).
	for i := 0; i < 60; i++ {
		Step(w, b, Intent{Forward: 1, Strafe: 1}, math.Pi/2, dt)
	}
	if b.Feet.X >= 12-b.Half+0.01 {
		t.Fatalf("body entered the wall: x=%v", b.Feet.X)
	}
	if b.Feet.X < 11.5 {
		t.Fatalf("body did not advance to the wall: x=%v", b.Feet.X)
	}
	if math.Abs(b.Feet.Z-10) < 0.5 {
		t.Fatalf("body did not slide along the wall: z=%v", b.Feet.Z)
	}
}

func TestJump(t *testing.T) {
	w := ground()
	b := NewBody(geom.Vec3{X: 10, Y: 6, Z: 10})
	for i := 0; i < 120; i++ {
		Step(w, b, Intent{}, 0, dt)
	}
	if !b.Grounded {
		t.Fatal("not grounded before jump")
	}
	restY := b.Feet.Y
	Step(w, b, Intent{Jump: true}, 0, dt)
	if b.Vel.Y <= 0 {
		t.Fatalf("jump should produce upward velocity, got %v", b.Vel.Y)
	}
	vy := b.Vel.Y
	Step(w, b, Intent{Jump: true}, 0, dt) // mid-air jump: must not re-boost
	if b.Vel.Y > vy {
		t.Fatal("mid-air jump re-boosted velocity (double jump)")
	}
	maxY := b.Feet.Y
	for i := 0; i < 30; i++ {
		Step(w, b, Intent{}, 0, dt)
		if b.Feet.Y > maxY {
			maxY = b.Feet.Y
		}
	}
	if maxY < restY+0.7 {
		t.Fatalf("jump gained too little height: max=%v rest=%v", maxY, restY)
	}
}

func TestHeadBump(t *testing.T) {
	w := ground()
	for z := 0; z < 20; z++ {
		for x := 0; x < 20; x++ {
			w.Set(x, 3, z, world.Stone) // ceiling, bottom at y=3
		}
	}
	b := NewBody(geom.Vec3{X: 10, Y: 1, Z: 10})
	for i := 0; i < 60; i++ {
		Step(w, b, Intent{}, 0, dt)
	}
	Step(w, b, Intent{Jump: true}, 0, dt)
	for i := 0; i < 30; i++ {
		Step(w, b, Intent{}, 0, dt)
	}
	if b.Feet.Y+b.Height > 3.05 {
		t.Fatalf("head passed through the ceiling: feet=%v head=%v", b.Feet.Y, b.Feet.Y+b.Height)
	}
}

func TestUnstick(t *testing.T) {
	w := ground()
	for y := 0; y < 5; y++ {
		w.Set(10, y, 10, world.Stone) // solid column where the body spawns
	}
	b := NewBody(geom.Vec3{X: 10.5, Y: 1, Z: 10.5})
	if !Overlaps(w, b.Feet, b.Half, b.Height) {
		t.Fatal("test precondition: body should start stuck")
	}
	Step(w, b, Intent{}, 0, dt)
	if Overlaps(w, b.Feet, b.Half, b.Height) {
		t.Fatalf("body still stuck after Step: feet=%v", b.Feet)
	}
}

func TestBoundsClamp(t *testing.T) {
	w := ground()
	b := NewBody(geom.Vec3{X: 1, Y: 1, Z: 1})
	// yaw=-PI/2 ⇒ forward=-X; push into the world edge for a long time.
	for i := 0; i < 200; i++ {
		Step(w, b, Intent{Forward: 1}, -math.Pi/2, dt)
	}
	if b.Feet.X < b.Half-1e-6 {
		t.Fatalf("body left the world on -X: x=%v", b.Feet.X)
	}
}

func TestFlyIgnoresGravity(t *testing.T) {
	w := ground()
	b := NewBody(geom.Vec3{X: 10, Y: 8, Z: 10})
	b.Mode = Fly
	y0 := b.Feet.Y
	for i := 0; i < 60; i++ {
		Step(w, b, Intent{}, 0, dt)
	}
	if math.Abs(b.Feet.Y-y0) > 1e-9 {
		t.Fatalf("fly mode should not fall: y0=%v now=%v", y0, b.Feet.Y)
	}
}
