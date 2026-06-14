// Package game wires the world, camera, renderer, and input together and runs
// the fixed-timestep loop. It owns player state (selected block, reach) that is
// deliberately kept out of the pure camera.
package game

import (
	"github.com/gdamore/tcell/v2"

	"github.com/tyler-void/TermiGoCraft/internal/camera"
	"github.com/tyler-void/TermiGoCraft/internal/entity"
	"github.com/tyler-void/TermiGoCraft/internal/geom"
	"github.com/tyler-void/TermiGoCraft/internal/render"
	"github.com/tyler-void/TermiGoCraft/internal/world"
)

// Game holds all mutable run state.
type Game struct {
	scr    tcell.Screen
	world  *world.World
	cam    *camera.Camera
	buf    *render.Buffer
	target world.Hit

	entities []entity.Entity
	heldMesh []geom.Triangle

	selected  world.Block
	selSlot   int
	reach     float64
	moveSpeed float64
	turnSpeed float64
	showHelp  bool
}

// New builds a game with a freshly generated world sized to the terminal.
func New(scr tcell.Screen, cols, rows int) *Game {
	w := world.New(64, 32, 64)
	spawnY := world.Generate(w)

	cam := &camera.Camera{
		Pos:   geom.Vec3{X: float64(w.SX) / 2, Y: float64(spawnY), Z: float64(w.SZ) / 2},
		Yaw:   0,
		Pitch: -0.2,
		FOV:   1.3,
	}

	g := &Game{
		scr:       scr,
		world:     w,
		cam:       cam,
		buf:       render.NewBuffer(cols, rows),
		selected:  world.Placeable[0],
		selSlot:   0,
		reach:     5.0,
		moveSpeed: 0.45,
		turnSpeed: 0.08,
		showHelp:  true,
		heldMesh:  entity.Cube(0.35),
	}

	// A simple wandering marker entity sitting on the surface near spawn.
	ex, ez := w.SX/2+4, w.SZ/2+2
	ey := surfaceTop(w, ex, ez) + 1
	g.entities = []entity.Entity{
		{
			Pos:   geom.Vec3{X: float64(ex) + 0.5, Y: float64(ey) + 0.5, Z: float64(ez) + 0.5},
			Mesh:  entity.Cube(0.9),
			Color: geom.RGB{R: 220, G: 60, B: 200},
		},
	}
	return g
}

func surfaceTop(w *world.World, x, z int) int {
	for y := w.SY - 1; y >= 0; y-- {
		if w.Solid(x, y, z) {
			return y
		}
	}
	return 0
}
