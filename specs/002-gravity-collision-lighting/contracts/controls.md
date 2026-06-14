# Contract: Controls (delta from feature 001)

**Feature**: `002-gravity-collision-lighting` | **Date**: 2026-06-15
Only the changes/additions to the feature-001 control scheme. Unchanged keys (movement WASD, arrow look, F/Enter break, R place, Q/E or 1–9 select, H help, Esc quit) carry over.

## New / changed actions

| Action | Key | Behavior | Requirement |
|--------|-----|----------|-------------|
| `Jump` | `Space` (in **Walk** mode) | jump when grounded; ignored mid-air | FR-005 |
| `MoveUp` / `MoveDown` | `Space` / `C` (in **Fly** mode) | free vertical (001 behavior) | FR-010 |
| `ToggleFly` | `G` | switch Walk ⇄ Fly (gravity on/off) | FR-010 |
| `ToggleLighting` | `L` | sun lighting + shadows on/off (perf/preference) | R4 |

**Note**: `Space` is **context-dependent** — jump while walking, ascend while flying. `C` descends only while flying. This is the one behavioral overload; everything else is a fixed binding.

## Behavioral contract (testable)

- **C-1**: In Walk mode `Space` triggers a jump only when grounded; in Fly mode `Space`/`C` move vertically.
- **C-2**: `G` flips the movement mode; the active mode is shown in the HUD; walking is the default at launch.
- **C-3**: `L` toggles lighting; with lighting off the world renders as in feature 001 (flat per-face shading), with no shadows/AO.
- **C-4**: Horizontal movement keys set a decaying velocity (held/auto-repeat sustains motion; it stops shortly after release) — physics still ticks (gravity) every frame regardless of input.
- **C-5**: Every new action appears in the on-screen help (FR-013/016 spirit).
- **C-6**: All bindings remain reachable; unknown keys are ignored (carried from 001).

## Invocation

Unchanged: single binary, no args, drops into a generated world in **Walk** mode under a fixed sun with lighting **on**.
