package game

import (
	"testing"

	"github.com/gdamore/tcell/v2"

	"github.com/tyler-void/TermiGoCraft/internal/input"
	"github.com/tyler-void/TermiGoCraft/internal/physics"
	"github.com/tyler-void/TermiGoCraft/internal/world"
)

func newSimGame(t *testing.T) (*Game, tcell.SimulationScreen) {
	t.Helper()
	scr := tcell.NewSimulationScreen("")
	if err := scr.Init(); err != nil {
		t.Fatalf("sim screen init: %v", err)
	}
	scr.SetSize(120, 40)
	return New(scr, 120, 40), scr
}

func step(g *Game, n int) {
	for i := 0; i < n; i++ {
		physics.Step(g.world, g.body, physics.Intent{}, g.cam.Yaw, dt)
		g.cam.Pos = g.body.Eye()
	}
}

func TestNewGeneratesWorld(t *testing.T) {
	g, scr := newSimGame(t)
	defer scr.Fini()
	solid := 0
	for _, b := range g.world.Blocks {
		if b.IsSolid() {
			solid++
		}
	}
	if solid == 0 {
		t.Fatal("generated world has no solid blocks")
	}
}

func TestPlayerFallsAndLands(t *testing.T) {
	g, scr := newSimGame(t)
	defer scr.Fini()
	step(g, 180) // a few seconds of physics
	if !g.body.Grounded {
		t.Fatal("player should land (be grounded) after falling onto the generated ground")
	}
	if physics.Overlaps(g.world, g.body.Feet, g.body.Half, g.body.Height) {
		t.Fatal("landed player overlaps terrain")
	}
}

func TestFramePipeline(t *testing.T) {
	g, scr := newSimGame(t)
	defer scr.Fini()
	step(g, 180)
	for i := 0; i < 3; i++ {
		g.render() // must not panic with lighting on
	}
	g.cam.Pitch = -1.2
	g.recomputeTarget()
	if !g.target.OK {
		t.Fatal("looking down at the ground should yield a target block")
	}
}

func TestEditFlow(t *testing.T) {
	g, scr := newSimGame(t)
	defer scr.Fini()
	step(g, 180)
	g.cam.Pitch = -1.4
	g.recomputeTarget()
	if !g.target.OK {
		t.Skip("no ground target at this spawn (generation-dependent)")
	}
	bx, by, bz := g.target.X, g.target.Y, g.target.Z
	var in physics.Intent
	g.apply(input.Break, 0, &in)
	if g.world.Solid(bx, by, bz) {
		t.Fatal("Break did not clear the targeted block")
	}
	if g.apply(input.Quit, 0, &in) != true {
		t.Fatal("Quit should return true")
	}
}

func TestFlyToggle(t *testing.T) {
	g, scr := newSimGame(t)
	defer scr.Fini()
	var in physics.Intent
	g.apply(input.ToggleFly, 0, &in)
	if g.body.Mode != physics.Fly {
		t.Fatal("ToggleFly should switch to Fly mode")
	}
	g.apply(input.ToggleFly, 0, &in)
	if g.body.Mode != physics.Walk {
		t.Fatal("ToggleFly should switch back to Walk mode")
	}
}

func TestSelectCycling(t *testing.T) {
	g, scr := newSimGame(t)
	defer scr.Fini()
	var in physics.Intent
	start := g.selected
	g.apply(input.SelectNext, 0, &in)
	if g.selected == start && len(world.Placeable) > 1 {
		t.Fatal("SelectNext did not change the selected block")
	}
	g.apply(input.SelectSlot, 2, &in)
	if g.selected != world.Placeable[2] {
		t.Fatalf("SelectSlot(2) -> %v, want %v", g.selected, world.Placeable[2])
	}
}

func TestPlaceRejectedInsidePlayer(t *testing.T) {
	g, scr := newSimGame(t)
	defer scr.Fini()
	step(g, 180)
	// Aim straight down at the block under the feet; placing on its top face
	// would land in the player's own body and must be rejected.
	g.cam.Pitch = -1.5
	g.recomputeTarget()
	if !g.target.OK {
		t.Skip("no target under feet")
	}
	dx, dy, dz := g.target.Face.Normal()
	nx, ny, nz := g.target.X+dx, g.target.Y+dy, g.target.Z+dz
	before := g.world.At(nx, ny, nz)
	if g.body.OccupiesCell(nx, ny, nz) {
		var in physics.Intent
		g.apply(input.Place, 0, &in)
		if g.world.At(nx, ny, nz) != before {
			t.Fatal("placed a block inside the player's body")
		}
	}
}
