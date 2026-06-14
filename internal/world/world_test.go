package world

import "testing"

func TestBoundsAndSetGet(t *testing.T) {
	w := New(4, 4, 4)
	if w.At(0, 0, 0) != Air {
		t.Fatal("new world should be all Air")
	}
	w.Set(1, 2, 3, Stone)
	if w.At(1, 2, 3) != Stone {
		t.Fatal("Set/At round-trip failed")
	}
	// out of bounds reads are Air, writes are ignored (no panic)
	if w.At(-1, 0, 0) != Air || w.At(4, 0, 0) != Air {
		t.Fatal("OOB read should be Air")
	}
	w.Set(99, 99, 99, Stone) // must not panic
}

func TestPlaceableNonEmpty(t *testing.T) {
	if len(Placeable) < 2 {
		t.Fatalf("need >=2 placeable types, got %d", len(Placeable))
	}
	for _, b := range Placeable {
		if !b.IsSolid() {
			t.Fatalf("placeable %v is not solid", b)
		}
	}
}
