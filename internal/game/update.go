package game

import (
	"math"

	"github.com/tyler-void/TermiGoCraft/internal/input"
	"github.com/tyler-void/TermiGoCraft/internal/world"
)

// apply mutates state for a single action. It returns true when the player asked
// to quit. SelectSlot carries its slot index in the payload.
func (g *Game) apply(a input.Action, payload int) bool {
	switch a {
	case input.Quit:
		return true
	case input.MoveForward:
		g.moveHoriz(g.moveSpeed)
	case input.MoveBack:
		g.moveHoriz(-g.moveSpeed)
	case input.StrafeLeft:
		g.strafe(-g.moveSpeed)
	case input.StrafeRight:
		g.strafe(g.moveSpeed)
	case input.MoveUp:
		g.cam.Pos.Y += g.moveSpeed
	case input.MoveDown:
		g.cam.Pos.Y -= g.moveSpeed
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
	case input.Break:
		g.world.Break(g.target)
	case input.Place:
		px, py, pz := g.playerCell()
		if g.world.Place(g.target, g.selected, px, py, pz) {
			// keep aiming feedback fresh after a change
			g.recomputeTarget()
		}
	case input.SelectNext:
		g.cycleSelection(1)
	case input.SelectPrev:
		g.cycleSelection(-1)
	case input.SelectSlot:
		if payload >= 0 && payload < len(world.Placeable) {
			g.selSlot = payload
			g.setSelected()
		}
	case input.ToggleHelp:
		g.showHelp = !g.showHelp
	}
	return false
}

func (g *Game) moveHoriz(amt float64) {
	d := g.cam.Direction()
	d.Y = 0
	d = d.Normalize()
	g.cam.Pos = g.cam.Pos.Add(d.Scale(amt))
}

func (g *Game) strafe(amt float64) {
	g.cam.Pos = g.cam.Pos.Add(g.cam.Right().Scale(amt))
}

func (g *Game) cycleSelection(dir int) {
	n := len(world.Placeable)
	g.selSlot = (g.selSlot + dir + n) % n
	g.setSelected()
}

func (g *Game) setSelected() {
	g.selected = world.Placeable[g.selSlot]
}

func (g *Game) playerCell() (int, int, int) {
	return int(math.Floor(g.cam.Pos.X)), int(math.Floor(g.cam.Pos.Y)), int(math.Floor(g.cam.Pos.Z))
}

// recomputeTarget casts the center ray to find the block the player is aiming at.
func (g *Game) recomputeTarget() {
	g.target = g.world.Cast(g.cam.Pos, g.cam.Direction(), g.reach)
}
