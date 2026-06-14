package input

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestRuneBindings(t *testing.T) {
	cases := map[rune]Action{
		'w': MoveForward,
		's': MoveBack,
		'a': StrafeLeft,
		'd': StrafeRight,
		' ': Jump,
		'c': MoveDown,
		'f': Break,
		'r': Place,
		'e': SelectNext,
		'q': SelectPrev,
		'g': ToggleFly,
		'l': ToggleLighting,
		'h': ToggleHelp,
	}
	for r, want := range cases {
		ev := tcell.NewEventKey(tcell.KeyRune, r, tcell.ModNone)
		if got, _ := Map(ev); got != want {
			t.Errorf("rune %q -> %v, want %v", r, got, want)
		}
	}
}

func TestSpecialKeys(t *testing.T) {
	cases := map[tcell.Key]Action{
		tcell.KeyEscape: Quit,
		tcell.KeyLeft:   LookLeft,
		tcell.KeyRight:  LookRight,
		tcell.KeyUp:     LookUp,
		tcell.KeyDown:   LookDown,
		tcell.KeyEnter:  Break,
	}
	for k, want := range cases {
		ev := tcell.NewEventKey(k, 0, tcell.ModNone)
		if got, _ := Map(ev); got != want {
			t.Errorf("key %v -> %v, want %v", k, got, want)
		}
	}
}

func TestSelectSlotPayload(t *testing.T) {
	ev := tcell.NewEventKey(tcell.KeyRune, '3', tcell.ModNone)
	got, slot := Map(ev)
	if got != SelectSlot || slot != 2 {
		t.Fatalf("'3' -> (%v,%d), want (SelectSlot,2)", got, slot)
	}
}

func TestUnknownKeyIsNone(t *testing.T) {
	ev := tcell.NewEventKey(tcell.KeyRune, 'z', tcell.ModNone)
	if got, _ := Map(ev); got != None {
		t.Fatalf("'z' -> %v, want None", got)
	}
}
