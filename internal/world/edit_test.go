package world

import "testing"

func TestBreakClearsCell(t *testing.T) {
	w := New(5, 5, 5)
	w.Set(2, 2, 2, Stone)
	h := Hit{X: 2, Y: 2, Z: 2, Face: FacePosY, Block: Stone, OK: true}
	if !w.Break(h) {
		t.Fatal("Break should succeed")
	}
	if w.At(2, 2, 2) != Air {
		t.Fatal("cell should be Air after Break")
	}
}

func TestBreakNoTargetIsNoOp(t *testing.T) {
	w := New(5, 5, 5)
	if w.Break(Hit{OK: false}) {
		t.Fatal("Break with no target should return false")
	}
}

func TestPlaceFillsFaceNeighbor(t *testing.T) {
	w := New(5, 5, 5)
	w.Set(2, 2, 2, Stone)
	h := Hit{X: 2, Y: 2, Z: 2, Face: FacePosY, Block: Stone, OK: true}
	if !w.Place(h, Dirt, 0, 0, 0) {
		t.Fatal("Place on free top face should succeed")
	}
	if w.At(2, 3, 2) != Dirt {
		t.Fatal("expected Dirt placed above the targeted block")
	}
}

func TestPlaceRejectsOccupied(t *testing.T) {
	w := New(5, 5, 5)
	w.Set(2, 2, 2, Stone)
	w.Set(2, 3, 2, Stone) // neighbor already solid
	h := Hit{X: 2, Y: 2, Z: 2, Face: FacePosY, Block: Stone, OK: true}
	if w.Place(h, Dirt, 0, 0, 0) {
		t.Fatal("Place into occupied cell should be rejected")
	}
}

func TestPlaceRejectsSelf(t *testing.T) {
	w := New(5, 5, 5)
	w.Set(2, 2, 2, Stone)
	h := Hit{X: 2, Y: 2, Z: 2, Face: FacePosY, Block: Stone, OK: true}
	// player occupies the neighbor cell (2,3,2)
	if w.Place(h, Dirt, 2, 3, 2) {
		t.Fatal("Place into the player's own cell should be rejected")
	}
}

func TestPlaceRejectsOutOfBounds(t *testing.T) {
	w := New(5, 5, 5)
	w.Set(4, 0, 0, Stone)
	h := Hit{X: 4, Y: 0, Z: 0, Face: FacePosX, Block: Stone, OK: true}
	if w.Place(h, Dirt, -1, -1, -1) {
		t.Fatal("Place past the world edge should be rejected")
	}
}
