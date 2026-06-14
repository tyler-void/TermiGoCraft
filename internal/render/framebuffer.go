package render

import (
	"math"

	"github.com/gdamore/tcell/v2"

	"github.com/tyler-void/TermiGoCraft/internal/geom"
)

// upperHalfBlock renders the top half of a cell in the foreground color and the
// bottom half in the background color, giving two independently-colored pixels
// per terminal cell.
const upperHalfBlock = '▀'

var infDepth = float32(math.Inf(1))

// Buffer is the shared render target. Color and Depth are flat slices indexed
// y*W+x, allocated once and reused every frame (only Resize reallocates). Depth
// holds perpendicular (camera-Z) distance; smaller is nearer.
type Buffer struct {
	W, H  int // pixel dimensions; H == rows*2
	Color []geom.RGB
	Depth []float32

	// prev holds the colors last pushed to the screen, so Present can skip
	// cells that did not change. tcell's SetContent allocates internally on
	// every call, so skipping unchanged cells keeps an idle view at zero
	// allocations and makes motion pay only for cells that actually changed.
	prev []geom.RGB
	full bool // force a full redraw on the next Present (after resize)
}

// NewBuffer creates a buffer sized for a cols×rows terminal.
func NewBuffer(cols, rows int) *Buffer {
	b := &Buffer{}
	b.Resize(cols, rows)
	return b
}

// Resize (re)sizes the buffer to cols×rows terminal cells. This is the only
// method that allocates; it reuses the existing backing array when it is large
// enough so resizing down never re-allocates.
func (b *Buffer) Resize(cols, rows int) {
	if cols < 1 {
		cols = 1
	}
	if rows < 1 {
		rows = 1
	}
	b.W = cols
	b.H = rows * 2
	n := b.W * b.H
	if cap(b.Color) < n {
		b.Color = make([]geom.RGB, n)
		b.Depth = make([]float32, n)
		b.prev = make([]geom.RGB, n)
	} else {
		b.Color = b.Color[:n]
		b.Depth = b.Depth[:n]
		b.prev = b.prev[:n]
	}
	b.full = true // dimensions changed: redraw everything next Present
}

// Clear resets every pixel to sky and every depth to +Inf. The range loops
// lower to a fast memset and allocate nothing.
func (b *Buffer) Clear(sky geom.RGB) {
	for i := range b.Color {
		b.Color[i] = sky
	}
	for i := range b.Depth {
		b.Depth[i] = infDepth
	}
}

// TestSet writes color c at pixel (x,y) iff depth z is nearer than what is
// already there. Out-of-range pixels are ignored (clipping).
func (b *Buffer) TestSet(x, y int, z float32, c geom.RGB) {
	if x < 0 || x >= b.W || y < 0 || y >= b.H {
		return
	}
	i := y*b.W + x
	if z < b.Depth[i] {
		b.Depth[i] = z
		b.Color[i] = c
	}
}

// Present packs vertical pixel pairs into terminal cells and pushes them to the
// screen. tcell diffs against its own previous frame, so only changed cells are
// actually written to the terminal.
func (b *Buffer) Present(scr tcell.Screen) {
	rows := b.H / 2
	for cy := 0; cy < rows; cy++ {
		topRow := (2 * cy) * b.W
		botRow := (2*cy + 1) * b.W
		for cx := 0; cx < b.W; cx++ {
			ti := topRow + cx
			bi := botRow + cx
			top := b.Color[ti]
			bot := b.Color[bi]
			if !b.full && b.prev[ti] == top && b.prev[bi] == bot {
				continue // unchanged cell: skip the allocating SetContent
			}
			b.prev[ti] = top
			b.prev[bi] = bot
			st := tcell.StyleDefault.
				Foreground(tcell.NewRGBColor(int32(top.R), int32(top.G), int32(top.B))).
				Background(tcell.NewRGBColor(int32(bot.R), int32(bot.G), int32(bot.B)))
			scr.SetContent(cx, cy, upperHalfBlock, nil, st)
		}
	}
	b.full = false
}
