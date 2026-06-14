package render

import (
	"github.com/gdamore/tcell/v2"

	"github.com/tyler-void/TermiGoCraft/internal/world"
)

var helpLines = []string{
	"WASD move  SPACE/C up/down  Arrows look",
	"F/Enter break  R place  Q/E or 1-9 select block",
	"H toggle help  ESC quit",
}

// RenderHUD draws the flat overlay directly to terminal cells after the 3D scene
// has been presented, so it always sits on top (Pass D). It uses one color per
// cell (not half-blocks), which is fine for text and the crosshair.
func RenderHUD(scr tcell.Screen, selected world.Block, showHelp bool) {
	w, h := scr.Size()
	if w == 0 || h == 0 {
		return
	}

	// Center crosshair.
	cross := tcell.StyleDefault.Foreground(tcell.ColorWhite)
	scr.SetContent(w/2, h/2, '+', nil, cross)

	// Selected-block indicator (bottom-left), colored with the block's tint.
	col := BlockColor(selected)
	swatch := tcell.StyleDefault.Background(tcell.NewRGBColor(int32(col.R), int32(col.G), int32(col.B)))
	label := tcell.StyleDefault.Foreground(tcell.ColorWhite)
	scr.SetContent(1, h-1, ' ', nil, swatch)
	scr.SetContent(2, h-1, ' ', nil, swatch)
	drawString(scr, 4, h-1, "Block: "+selected.Name(), label)

	if showHelp {
		hint := tcell.StyleDefault.Foreground(tcell.ColorYellow)
		for i, line := range helpLines {
			drawString(scr, 1, i, line, hint)
		}
	} else {
		hint := tcell.StyleDefault.Foreground(tcell.ColorGray)
		drawString(scr, 1, 0, "H: help", hint)
	}
}

func drawString(scr tcell.Screen, x, y int, s string, st tcell.Style) {
	for i, r := range s {
		scr.SetContent(x+i, y, r, nil, st)
	}
}
