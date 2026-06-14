package game

import (
	"time"

	"github.com/gdamore/tcell/v2"

	"github.com/tyler-void/TermiGoCraft/internal/input"
	"github.com/tyler-void/TermiGoCraft/internal/physics"
	"github.com/tyler-void/TermiGoCraft/internal/render"
)

// frameRate is the fixed render + physics cadence.
const frameRate = 30
const dt = 1.0 / float64(frameRate)

// Run drives the game until the player quits. Input is polled on a separate
// goroutine and drained each tick; physics integrates every tick (so gravity
// applies regardless of input).
func (g *Game) Run() {
	events := make(chan tcell.Event, 128)
	quit := make(chan struct{})

	go func() {
		for {
			ev := g.scr.PollEvent()
			if ev == nil {
				return
			}
			select {
			case events <- ev:
			case <-quit:
				return
			}
		}
	}()

	ticker := time.NewTicker(time.Second / frameRate)
	defer ticker.Stop()

	g.recomputeTarget()
	g.render()

	for {
		var intent physics.Intent
		draining := true
		for draining {
			select {
			case ev := <-events:
				if g.handle(ev, &intent) {
					close(quit)
					return
				}
			default:
				draining = false
			}
		}

		physics.Step(g.world, g.body, intent, g.cam.Yaw, dt)
		g.cam.Pos = g.body.Eye()
		g.recomputeTarget()
		g.render()
		<-ticker.C
	}
}

// handle processes a single event; returns true on quit.
func (g *Game) handle(ev tcell.Event, intent *physics.Intent) bool {
	switch e := ev.(type) {
	case *tcell.EventResize:
		g.scr.Sync()
		cols, rows := g.scr.Size()
		g.buf.Resize(cols, rows)
	case *tcell.EventKey:
		act, payload := input.Map(e)
		return g.apply(act, payload, intent)
	}
	return false
}

// render runs the full pass order into the shared buffer, presents it, then
// draws the HUD on top.
func (g *Game) render() {
	g.buf.Clear(render.Sky)
	render.RenderWorld(g.buf, g.world, g.cam, g.sun, g.lit)
	for i := range g.entities {
		render.RenderMesh(g.buf, g.entities[i].Mesh, g.entities[i].Pos, g.entities[i].Color, g.cam)
	}
	render.RenderHeld(g.buf, g.heldMesh, render.BlockColor(g.selected), g.cam)
	render.RenderSelection(g.buf, g.target, g.cam)
	g.buf.Present(g.scr)
	render.RenderHUD(g.scr, render.HUDState{
		Selected: g.selected,
		Targeted: g.target.OK,
		Flying:   g.body.Mode == physics.Fly,
		Lit:      g.lit,
		Pos:      g.cam.Pos,
		ShowHelp: g.showHelp,
	})
	g.scr.Show()
}
