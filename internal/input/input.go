// Package input maps raw tcell key events to the game's closed Action set.
// Keeping the mapping here means the game loop deals only in intent, and the
// control scheme is testable without a terminal.
package input

import "github.com/gdamore/tcell/v2"

// Action is a single intent produced from a key event.
type Action int

const (
	None Action = iota
	MoveForward
	MoveBack
	StrafeLeft
	StrafeRight
	MoveUp
	MoveDown
	LookLeft
	LookRight
	LookUp
	LookDown
	Break
	Place
	SelectNext
	SelectPrev
	SelectSlot // payload = slot index (0-based)
	ToggleHelp
	Quit
)

// Map converts a key event to an Action. For SelectSlot the second return value
// is the 0-based hotbar slot; otherwise it is 0.
func Map(ev *tcell.EventKey) (Action, int) {
	switch ev.Key() {
	case tcell.KeyEscape, tcell.KeyCtrlC:
		return Quit, 0
	case tcell.KeyLeft:
		return LookLeft, 0
	case tcell.KeyRight:
		return LookRight, 0
	case tcell.KeyUp:
		return LookUp, 0
	case tcell.KeyDown:
		return LookDown, 0
	case tcell.KeyEnter:
		return Break, 0
	case tcell.KeyRune:
		switch ev.Rune() {
		case 'w', 'W':
			return MoveForward, 0
		case 's', 'S':
			return MoveBack, 0
		case 'a', 'A':
			return StrafeLeft, 0
		case 'd', 'D':
			return StrafeRight, 0
		case ' ':
			return MoveUp, 0
		case 'c', 'C':
			return MoveDown, 0
		case 'f', 'F':
			return Break, 0
		case 'r', 'R':
			return Place, 0
		case 'e', 'E':
			return SelectNext, 0
		case 'q', 'Q':
			return SelectPrev, 0
		case 'h', 'H', '?':
			return ToggleHelp, 0
		case '1', '2', '3', '4', '5', '6', '7', '8', '9':
			return SelectSlot, int(ev.Rune() - '1')
		}
	}
	return None, 0
}
