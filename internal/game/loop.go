package game

import (
	"time"

	"github.com/gdamore/tcell/v2"

	"github.com/tyler-void/TermiGoCraft/internal/input"
	"github.com/tyler-void/TermiGoCraft/internal/render"
)

// frameRate is the fixed render cadence (~30 FPS).
const frameRate = 30

// Run drives the game until the player quits. Input is polled on a separate
// goroutine and drained each tick, so rendering never blocks on input and a
// burst of keypresses cannot stall a frame.
func (g *Game) Run() {
	events := make(chan tcell.Event, 128)
	quit := make(chan struct{})

	go func() {
		for {
			ev := g.scr.PollEvent()
			if ev == nil { // screen finalized
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
		// Drain all input that has arrived since the last tick.
		draining := true
		for draining {
			select {
			case ev := <-events:
				if g.handle(ev) {
					close(quit)
					return
				}
			default:
				draining = false
			}
		}

		g.recomputeTarget()
		g.render()
		<-ticker.C
	}
}

// handle processes a single event; returns true on quit.
func (g *Game) handle(ev tcell.Event) bool {
	switch e := ev.(type) {
	case *tcell.EventResize:
		g.scr.Sync()
		cols, rows := g.scr.Size()
		g.buf.Resize(cols, rows)
	case *tcell.EventKey:
		act, payload := input.Map(e)
		return g.apply(act, payload)
	}
	return false
}

// render runs the full pass order into the shared buffer, presents it, then
// draws the HUD on top.
func (g *Game) render() {
	g.buf.Clear(render.Sky)
	render.RenderWorld(g.buf, g.world, g.cam)
	for i := range g.entities {
		render.RenderMesh(g.buf, g.entities[i].Mesh, g.entities[i].Pos, g.entities[i].Color, g.cam)
	}
	render.RenderHeld(g.buf, g.heldMesh, render.BlockColor(g.selected), g.cam)
	render.RenderSelection(g.buf, g.target, g.cam)
	g.buf.Present(g.scr)
	render.RenderHUD(g.scr, g.selected, g.showHelp)
	g.scr.Show()
}
