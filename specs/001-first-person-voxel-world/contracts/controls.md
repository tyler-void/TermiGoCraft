# Contract: Controls (Input → Action)

**Feature**: `001-first-person-voxel-world` | **Date**: 2026-06-15

The user-facing contract for an interactive terminal app is its **control scheme**. Input events from tcell are mapped to a closed set of `Action`s; the game loop applies them each tick. Keyboard-only this sprint (no mouse-look — R7 / Assumptions).

## Action set

| Action | Meaning |
|--------|---------|
| `MoveForward` / `MoveBack` | move eye along (flattened) facing direction |
| `StrafeLeft` / `StrafeRight` | move eye along the right vector |
| `MoveUp` / `MoveDown` | move eye along world up (free-fly) |
| `LookLeft` / `LookRight` | yaw the camera |
| `LookUp` / `LookDown` | pitch the camera (clamped) |
| `Break` | remove the targeted block |
| `Place` | place `Selected` block against the targeted face |
| `SelectNext` / `SelectPrev` | cycle the selected block type |
| `SelectSlot(n)` | pick block type by hotbar index |
| `ToggleHelp` | show/hide controls overlay |
| `Quit` | exit cleanly, restoring the terminal |

## Default key bindings

| Key(s) | Action | Requirement |
|--------|--------|-------------|
| `W` / `S` | MoveForward / MoveBack | FR-006 |
| `A` / `D` | StrafeLeft / StrafeRight | FR-006 |
| `Space` / `Shift` (or `C`) | MoveUp / MoveDown | FR-006 |
| `←` `→` | LookLeft / LookRight | FR-006 |
| `↑` `↓` | LookUp / LookDown | FR-006 |
| `F` or `Enter` | Break | FR-008 |
| `R` or `Space`* | Place | FR-009 |
| `Q` / `E` | SelectPrev / SelectNext | FR-011 |
| `1`–`9` | SelectSlot(n) | FR-011 |
| `H` or `?` | ToggleHelp | FR-016 |
| `Esc` or `Ctrl+C` | Quit | FR-016 |

\* If `Space` is used for jump/up, Place is bound to `R`/`F`. Final binding chosen at implementation to avoid conflicts; the *contract* is that every Action above is reachable and listed in the on-screen help.

## Behavioral contract (testable)

- **C-1**: Every `Action` is bound to at least one key and appears in the help overlay (FR-016).
- **C-2**: `Break`/`Place` with no valid target (`Target.Ok == false`) are **no-ops**, never errors (FR-010, Edge Cases).
- **C-3**: `Place` that would occupy the player's own cell or an out-of-bounds/occupied cell is rejected (FR-010).
- **C-4**: Movement/look magnitude scales with frame `dt`, so behavior is frame-rate independent (R5, SC-002).
- **C-5**: `Quit` always restores the terminal (cooked mode, cursor shown, alt-screen exited) — SC-009.
- **C-6**: Input is read on a separate goroutine and never blocks rendering; a burst of keypresses cannot stall a frame (Edge Cases: rapid input).
- **C-7**: Pitch is clamped so the view cannot flip over (FR-006).

## Invocation contract

- Runs as a single binary with no required arguments: launching it drops the player straight into a generated world (User Story 1).
- Honors `COLORTERM`/terminfo for truecolor detection; no flags required for kitty (R7).
