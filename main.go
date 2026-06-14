// Command termigocraft is a first-person, Minecraft-style voxel world rendered
// in the terminal. The world is drawn by a per-pixel voxel raycaster and
// entities/overlays by a triangle rasterizer, both sharing one depth buffer.
//
// Controls: WASD move, SPACE/C up/down, arrows look, F/Enter break, R place,
// Q/E or 1-9 select block, H toggle help, ESC quit.
package main

import (
	"log"

	"github.com/gdamore/tcell/v2"

	"github.com/tyler-void/TermiGoCraft/internal/game"
)

func main() {
	scr, err := tcell.NewScreen()
	if err != nil {
		log.Fatalf("create screen: %v", err)
	}
	if err := scr.Init(); err != nil {
		log.Fatalf("init screen: %v", err)
	}
	// Always restore the terminal, even on panic.
	defer scr.Fini()

	scr.HideCursor()
	scr.Clear()

	cols, rows := scr.Size()
	g := game.New(scr, cols, rows)
	g.Run()
}
