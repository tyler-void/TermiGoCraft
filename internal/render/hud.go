package render

import (
	"strconv"

	"github.com/gdamore/tcell/v2"

	"github.com/tyler-void/TermiGoCraft/internal/geom"
	"github.com/tyler-void/TermiGoCraft/internal/world"
)

var helpLines = []string{
	"WASD move  SPACE jump/up  C down  Arrows look",
	"F/Enter break  R place  Q/E or 1-9 select",
	"G walk/fly  L lighting  H help  ESC quit",
}

// HUDState is the per-frame overlay state.
type HUDState struct {
	Selected world.Block
	Targeted bool      // a block is in reach (cursor "can interact" state)
	Flying   bool      // movement mode
	Lit      bool      // lighting on/off
	Pos      geom.Vec3 // player position
	ShowHelp bool
}

// RenderHUD draws the flat overlay directly to terminal cells after the 3D scene
// has been presented, so it always sits on top (Pass D).
func RenderHUD(scr tcell.Screen, st HUDState) {
	w, h := scr.Size()
	if w == 0 || h == 0 {
		return
	}
	cx, cy := w/2, h/2

	// Two-state aiming cursor: bright brackets when a block is targeted,
	// a dim dot when nothing is in reach.
	if st.Targeted {
		bright := tcell.StyleDefault.Foreground(tcell.ColorYellow).Bold(true)
		scr.SetContent(cx-1, cy, '[', nil, bright)
		scr.SetContent(cx, cy, '+', nil, bright)
		scr.SetContent(cx+1, cy, ']', nil, bright)
	} else {
		dim := tcell.StyleDefault.Foreground(tcell.ColorGray)
		scr.SetContent(cx, cy, '·', nil, dim)
	}

	// Selected-block swatch + status line (mode, lighting, position).
	col := BlockColor(st.Selected)
	swatch := tcell.StyleDefault.Background(tcell.NewRGBColor(int32(col.R), int32(col.G), int32(col.B)))
	label := tcell.StyleDefault.Foreground(tcell.ColorWhite)
	scr.SetContent(1, h-1, ' ', nil, swatch)
	scr.SetContent(2, h-1, ' ', nil, swatch)

	mode := "WALK"
	if st.Flying {
		mode = "FLY"
	}
	lit := "off"
	if st.Lit {
		lit = "on"
	}
	status := st.Selected.Name() + "  [" + mode + "]  light:" + lit + "  " +
		itoa(st.Pos.X) + "," + itoa(st.Pos.Y) + "," + itoa(st.Pos.Z)
	drawString(scr, 4, h-1, status, label)

	if st.ShowHelp {
		hint := tcell.StyleDefault.Foreground(tcell.ColorYellow)
		for i, line := range helpLines {
			drawString(scr, 1, i, line, hint)
		}
	} else {
		drawString(scr, 1, 0, "H: help", tcell.StyleDefault.Foreground(tcell.ColorGray))
	}
}

func itoa(f float64) string { return strconv.Itoa(int(f)) }

func drawString(scr tcell.Screen, x, y int, s string, st tcell.Style) {
	for i, r := range s {
		scr.SetContent(x+i, y, r, nil, st)
	}
}
