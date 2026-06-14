package game

import (
	"github.com/tyler-void/TermiGoCraft/internal/input"
	"github.com/tyler-void/TermiGoCraft/internal/physics"
	"github.com/tyler-void/TermiGoCraft/internal/world"
)

// apply handles one action. Movement actions accumulate into the per-tick
// intent; discrete actions take effect immediately. Returns true on quit.
func (g *Game) apply(a input.Action, payload int, in *physics.Intent) bool {
	switch a {
	case input.Quit:
		return true

	// Movement intent (consumed by physics this tick).
	case input.MoveForward:
		in.Forward += 1
	case input.MoveBack:
		in.Forward -= 1
	case input.StrafeRight:
		in.Strafe += 1
	case input.StrafeLeft:
		in.Strafe -= 1
	case input.Jump:
		if g.body.Mode == physics.Fly {
			in.Up = true
		} else {
			in.Jump = true
		}
	case input.MoveDown:
		if g.body.Mode == physics.Fly {
			in.Down = true
		}

	// Look (applied directly).
	case input.LookLeft:
		g.cam.Yaw -= g.turnSpeed
	case input.LookRight:
		g.cam.Yaw += g.turnSpeed
	case input.LookUp:
		g.cam.Pitch += g.turnSpeed
		g.cam.ClampPitch()
	case input.LookDown:
		g.cam.Pitch -= g.turnSpeed
		g.cam.ClampPitch()

	// World edits.
	case input.Break:
		g.world.Break(g.target)
	case input.Place:
		g.placeBlock()

	// Selection.
	case input.SelectNext:
		g.cycleSelection(1)
	case input.SelectPrev:
		g.cycleSelection(-1)
	case input.SelectSlot:
		if payload >= 0 && payload < len(world.Placeable) {
			g.selSlot = payload
			g.selected = world.Placeable[g.selSlot]
		}

	// Toggles.
	case input.ToggleFly:
		if g.body.Mode == physics.Walk {
			g.body.Mode = physics.Fly
		} else {
			g.body.Mode = physics.Walk
		}
		g.body.Vel = g.body.Vel.Scale(0) // stop cleanly on mode switch
	case input.ToggleLighting:
		g.lit = !g.lit
	case input.ToggleHelp:
		g.showHelp = !g.showHelp
	}
	return false
}

// placeBlock puts the selected block against the targeted face, unless that cell
// is out of bounds, occupied, or would intersect the player's body (FR-009).
func (g *Game) placeBlock() {
	if !g.target.OK {
		return
	}
	dx, dy, dz := g.target.Face.Normal()
	nx, ny, nz := g.target.X+dx, g.target.Y+dy, g.target.Z+dz
	if !g.world.InBounds(nx, ny, nz) || g.world.Solid(nx, ny, nz) {
		return
	}
	if g.body.OccupiesCell(nx, ny, nz) {
		return
	}
	g.world.Set(nx, ny, nz, g.selected)
}

func (g *Game) cycleSelection(dir int) {
	n := len(world.Placeable)
	g.selSlot = (g.selSlot + dir + n) % n
	g.selected = world.Placeable[g.selSlot]
}

// recomputeTarget casts the center ray to find the block the player is aiming at.
func (g *Game) recomputeTarget() {
	g.target = g.world.Cast(g.cam.Pos, g.cam.Direction(), g.reach)
}
