package geom

import (
	"math"
	"testing"
)

func TestNormalizeZero(t *testing.T) {
	if got := (Vec3{}).Normalize(); got != (Vec3{}) {
		t.Fatalf("Normalize of zero = %v, want zero", got)
	}
}

func TestNormalizeUnit(t *testing.T) {
	v := Vec3{3, 0, 4}.Normalize()
	if math.Abs(v.Length()-1) > 1e-9 {
		t.Fatalf("normalized length = %v, want 1", v.Length())
	}
}

func TestCrossDot(t *testing.T) {
	x := Vec3{1, 0, 0}
	y := Vec3{0, 1, 0}
	z := x.Cross(y)
	if z != (Vec3{0, 0, 1}) {
		t.Fatalf("x×y = %v, want {0,0,1}", z)
	}
	// cross product is perpendicular to both inputs
	if x.Dot(z) != 0 || y.Dot(z) != 0 {
		t.Fatalf("cross not perpendicular: x·z=%v y·z=%v", x.Dot(z), y.Dot(z))
	}
}

func TestAddSubScale(t *testing.T) {
	a := Vec3{1, 2, 3}
	b := Vec3{4, 5, 6}
	if a.Add(b) != (Vec3{5, 7, 9}) {
		t.Fatal("Add wrong")
	}
	if b.Sub(a) != (Vec3{3, 3, 3}) {
		t.Fatal("Sub wrong")
	}
	if a.Scale(2) != (Vec3{2, 4, 6}) {
		t.Fatal("Scale wrong")
	}
}
