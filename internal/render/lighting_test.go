package render

import (
	"testing"

	"github.com/tyler-void/TermiGoCraft/internal/geom"
	"github.com/tyler-void/TermiGoCraft/internal/world"
)

// topHit builds a Hit on the +Y (top) face of cell (x,y,z) at in-face UV center.
func topHit(x, y, z int) (world.Hit, geom.Vec3) {
	h := world.Hit{X: x, Y: y, Z: z, Face: world.FacePosY, Block: world.Grass, OK: true}
	hp := geom.Vec3{X: float64(x) + 0.5, Y: float64(y) + 1, Z: float64(z) + 0.5}
	return h, hp
}

func TestDirectionalTopVsBottom(t *testing.T) {
	w := world.New(8, 8, 8)
	w.Set(4, 2, 4, world.Grass)
	sun := DefaultSun() // shines from above
	top, tp := topHit(4, 2, 4)
	bottom := world.Hit{X: 4, Y: 2, Z: 4, Face: world.FaceNegY, Block: world.Grass, OK: true}
	bp := geom.Vec3{X: 4.5, Y: 2, Z: 4.5}
	lt := lightAt(w, top, tp, sun)
	lb := lightAt(w, bottom, bp, sun)
	if lt <= lb {
		t.Fatalf("top face (%.3f) should be brighter than bottom face (%.3f)", lt, lb)
	}
}

func TestShadowDarkensGround(t *testing.T) {
	w := world.New(20, 20, 20)
	// flat ground
	for z := 0; z < 20; z++ {
		for x := 0; x < 20; x++ {
			w.Set(x, 0, z, world.Grass)
		}
	}
	// a tall pillar that will cast a shadow toward -Dir (down-sun side)
	for y := 1; y < 8; y++ {
		w.Set(10, y, 10, world.Stone)
	}
	sun := DefaultSun()
	// open ground cell far from the pillar
	openH, openP := topHit(2, 0, 2)
	open := lightAt(w, openH, openP, sun)
	// ground cell on the shadow side of the pillar (sun.Dir has +X,+Z, so the
	// shadow falls toward -X,-Z of the pillar)
	shadowH, shadowP := topHit(9, 0, 9)
	shadow := lightAt(w, shadowH, shadowP, sun)
	if shadow >= open {
		t.Fatalf("shadowed ground (%.3f) should be darker than open ground (%.3f)", shadow, open)
	}
}

func TestAOInnerCornerDarker(t *testing.T) {
	w := world.New(8, 8, 8)
	// floor block whose top face we sample
	w.Set(4, 0, 4, world.Grass)
	// open-face AO
	h, _ := topHit(4, 0, 4)
	openAO := aoFactor(w, h, geom.Vec3{X: 4.5, Y: 1, Z: 4.5})
	// add neighbors that occlude one corner of the top face
	w.Set(5, 1, 4, world.Stone)
	w.Set(4, 1, 5, world.Stone)
	w.Set(5, 1, 5, world.Stone)
	cornerAO := aoFactor(w, h, geom.Vec3{X: 4.99, Y: 1, Z: 4.99}) // toward the occluded corner
	if cornerAO >= openAO {
		t.Fatalf("occluded corner AO (%.3f) should be darker than open (%.3f)", cornerAO, openAO)
	}
}

func TestLightAtDeterministic(t *testing.T) {
	w := world.New(8, 8, 8)
	w.Set(4, 2, 4, world.Grass)
	sun := DefaultSun()
	h, hp := topHit(4, 2, 4)
	first := lightAt(w, h, hp, sun)
	for i := 0; i < 100; i++ {
		if lightAt(w, h, hp, sun) != first {
			t.Fatal("lightAt is not deterministic across identical calls")
		}
	}
}
