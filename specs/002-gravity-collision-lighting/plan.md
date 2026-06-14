# Implementation Plan: Gravity, Collision & Ray-Traced Lighting

**Branch**: `002-gravity-collision-lighting` | **Date**: 2026-06-15 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `specs/002-gravity-collision-lighting/spec.md`

## Summary

Add three things to the feature-001 voxel game: (1) a **physical player** — an AABB body with gravity, per-axis voxel collision, jumping, and a walk/fly toggle — so the player stands on, bumps into, and jumps around the world instead of clipping through it; (2) **QoL** — a 2-state center aiming cursor and a mode/position readout; (3) **ray-traced lighting** — a directional sun with hard cast shadows (one shadow ray per hit, reusing the existing DDA) and bilinearly-interpolated voxel ambient occlusion (the 0fps `vertexAO` rule). Physics runs on the existing fixed 30 FPS tick; lighting plugs into the existing world raycast hit; both preserve the project's zero-per-frame-allocation discipline.

## Technical Context

**Language/Version**: Go 1.24 (unchanged).
**Primary Dependencies**: `github.com/gdamore/tcell/v2` (unchanged). No new dependencies — physics and lighting are hand-rolled, reusing feature 001's `geom`, `world` (incl. DDA `Cast`), `camera`, and `render` framebuffer.
**Storage**: N/A (no persistence).
**Testing**: Go `testing` (table tests beside code) + extended `0 allocs/op` benchmark over the **lit** frame + headless physics tests (gravity rest, wall slide, jump, overlap recovery, bounds) and lighting tests (shadow occlusion, AO levels, determinism).
**Target Platform**: kitty (truecolor) reference; portable.
**Project Type**: single-binary terminal game (unchanged).
**Performance Goals**: maintain ≥ 30 FPS at 120×40 with lighting on (budget ~5–6 ms/frame: primary + one shadow ray + AO lookups); input/physics response within 50 ms; deterministic, flicker-free lighting.
**Constraints**: **zero heap allocation in the steady-state lit frame**; fixed-timestep physics (deterministic, no tunneling at our speeds); player can never clip, fall out of the world, or get stuck; lighting on/off toggle as a perf fallback.
**Scale/Scope**: one player AABB; one directional sun; shadow ray cap ~24 cells; AO from the 3 neighbors per face corner. Ray-tracing scope = direct light + hard shadows + AO only (no reflections/GI/soft shadows/day-night).

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

The constitution (`.specify/memory/constitution.md`) remains the unfilled template — no project-specific gates. Held to the same general good-practice gates as feature 001:

| Gate | Status | Notes |
|------|--------|-------|
| **Simplicity / single project** | ✅ PASS | One module, one binary; adds one small `physics` package + one `render/lighting.go`. |
| **No speculative abstraction (YAGNI)** | ✅ PASS | Arcade physics, not an engine; direct lighting only; documented future upgrades (swept-ray, baked light) explicitly deferred. |
| **Dependency minimalism** | ✅ PASS | **Zero new dependencies.** |
| **Reuse over rewrite** | ✅ PASS | Reuses DDA `Cast` for shadow rays, `Face` normals for lighting, the framebuffer, the fixed-timestep loop, the HUD. |
| **Testability** | ✅ PASS | Physics and lighting are pure logic, headlessly testable; allocation benchmark extended to the lit path. |
| **Zero-allocation frame (project norm)** | ✅ PASS | Shadow `Cast` returns a value; AO uses ints; benchmark guards it. |

**Result**: PASS (Complexity Tracking empty).

## Project Structure

### Documentation (this feature)

```text
specs/002-gravity-collision-lighting/
├── plan.md              # This file
├── research.md          # Phase 0 (complete)
├── data-model.md        # Phase 1 (complete)
├── quickstart.md        # Phase 1 (complete)
├── contracts/           # Phase 1 (complete)
│   ├── physics.md           # player body, Step, collision/jump guarantees
│   ├── lighting.md          # sun + shadow ray + AO model
│   └── controls.md          # updated key map (jump, fly toggle, lighting toggle)
└── tasks.md             # Phase 2 (/speckit-tasks — NOT created here)
```

### Source Code (repository root) — changes from feature 001

```text
internal/
├── physics/                  # NEW — player body + gravity + voxel collision
│   ├── body.go               #   Body{Feet, Vel, Half, Height, EyeHeight, Grounded, Mode}; Eye()
│   ├── step.go               #   Step(w, b, intent, dt): gravity, per-axis resolve, jump, bounds, unstick
│   └── *_test.go             #   rest-on-ground, wall-slide, jump, head-bump, overlap recovery, bounds
├── render/
│   ├── lighting.go           # NEW — Sun, shadow ray (reuse world.Cast), voxel AO (vertexAO + bilerp), lightAt()
│   ├── raycaster.go          # MODIFIED — call lightAt() at each hit; lighting on/off flag
│   ├── hud.go                # MODIFIED — 2-state cursor; mode + position readout
│   └── raycaster_test.go     # MODIFIED — shadow/AO/determinism tests; lit-frame 0-alloc benchmark
├── camera/camera.go          # unchanged (eye stays pure geometry; game syncs from body)
├── input/input.go            # MODIFIED — add Jump, ToggleFly, ToggleLighting actions
└── game/
    ├── game.go               # MODIFIED — hold *physics.Body, Sun, lighting flag; eye derived from body
    ├── update.go             # MODIFIED — set horizontal intent (decay), jump, toggles; sync cam from body
    └── loop.go               # MODIFIED — call physics.Step every tick before render
```

**Structure Decision**: One new package (`physics`) and one new render file (`lighting.go`); everything else is a modification to existing 001 files. Dependency flow stays acyclic: `physics` → `geom`,`world`; `render/lighting` → `geom`,`world`,`camera`; `game` → all. The player body lives in `physics` (not in `camera`) to keep the camera pure geometry, matching the 001 layering. Pure-logic packages (`physics`, and the lighting helpers) remain headlessly testable with no tcell dependency.

## Complexity Tracking

> No constitution violations — section intentionally empty.

## Phase Outputs

- **Phase 0 — Research**: [research.md](./research.md) ✅ — AABB+fixed-step physics (R1), per-axis swept collision (R2), directional sun + hard shadow ray + voxel AO (R3), frame budget/0-alloc (R4), input→physics bridge + QoL (R5). Studied: 0fps, Gaffer, fenomas voxel-aabb-sweep, deferred voxel shading.
- **Phase 1 — Design & Contracts**: [data-model.md](./data-model.md), [contracts/physics.md](./contracts/physics.md), [contracts/lighting.md](./contracts/lighting.md), [contracts/controls.md](./contracts/controls.md), [quickstart.md](./quickstart.md) ✅. `CLAUDE.md` agent context updated.
- **Phase 2 — Tasks**: `/speckit-tasks` (not part of this command).

## Post-Design Constitution Re-Check

After Phase 1: no new dependencies, one small new package, no new abstractions beyond the body/lighting split justified above. Physics and lighting are pure and tested; the lit frame stays zero-allocation. **Still PASS.**
