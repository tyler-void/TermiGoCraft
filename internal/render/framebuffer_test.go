package render

import (
	"math"
	"testing"

	"github.com/gdamore/tcell/v2"

	"github.com/tyler-void/TermiGoCraft/internal/geom"
)

func TestClearResetsColorAndDepth(t *testing.T) {
	b := NewBuffer(4, 4)
	b.TestSet(1, 1, 1.0, geom.RGB{R: 9})
	b.Clear(Sky)
	for i := range b.Color {
		if b.Color[i] != Sky {
			t.Fatalf("pixel %d not cleared to sky", i)
		}
		if !math.IsInf(float64(b.Depth[i]), 1) {
			t.Fatalf("depth %d not reset to +Inf", i)
		}
	}
}

func TestTestSetDepthOrdering(t *testing.T) {
	b := NewBuffer(4, 4)
	b.Clear(Sky)
	near := geom.RGB{R: 1}
	far := geom.RGB{R: 2}
	b.TestSet(0, 0, 5, far)
	b.TestSet(0, 0, 2, near) // nearer wins
	if b.Color[0] != near {
		t.Fatal("nearer depth should overwrite")
	}
	b.TestSet(0, 0, 9, far) // farther loses
	if b.Color[0] != near {
		t.Fatal("farther depth should not overwrite")
	}
}

func TestResizeReusesCapacity(t *testing.T) {
	b := NewBuffer(120, 40)
	c0 := cap(b.Color)
	b.Resize(80, 20) // smaller: must not reallocate
	if cap(b.Color) != c0 {
		t.Fatal("resize down should reuse the existing backing array")
	}
}

// BenchmarkClearPresent is the zero-allocation guard for the present path
// (research R6 / contract G-2). Expect 0 allocs/op.
func BenchmarkClearPresent(b *testing.B) {
	scr := tcell.NewSimulationScreen("")
	if err := scr.Init(); err != nil {
		b.Fatal(err)
	}
	defer scr.Fini()
	scr.SetSize(120, 40)
	buf := NewBuffer(120, 40)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Clear(Sky)
		buf.Present(scr)
	}
}
