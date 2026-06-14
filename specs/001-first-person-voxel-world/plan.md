# Implementation Plan: First-Person Voxel World

**Branch**: `001-first-person-voxel-world` | **Date**: 2026-06-15 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `specs/001-first-person-voxel-world/spec.md`

## Summary

Transform the existing single-file terminal 3D **wireframe** demo into a first-person, Minecraft-style **voxel** world the player can explore and edit. The world becomes a bounded 3D grid of solid blocks rendered in first person by a **per-pixel DDA raycaster**; the existing triangle renderer is repurposed (wireframe → filled) to draw **entities, the held item, and a selection highlight** into the **same shared color + depth framebuffer**, giving correct occlusion between terrain and non-block objects. Output uses **tcell/v2** with **truecolor half-block (`▀`) pixels** (2 px per terminal cell), driven by a **fixed ~30 FPS loop** with input on a separate goroutine. The render path is engineered for **zero per-frame allocation** (flat, preallocated, pointer-free buffers) so it stays smooth and GC-quiet over long sessions. Scope is one sprint, delivered as P1 (explore) → P2 (edit) → P3 (entities/HUD).

## Technical Context

**Language/Version**: Go 1.24 (module already targets `go 1.24.2`; installed toolchain is newer and compatible)
**Primary Dependencies**: `github.com/gdamore/tcell/v2` (v2.13.10) for terminal I/O + truecolor. **Remove** `github.com/nsf/termbox-go` and the unused `github.com/go-gl/gl` + `github.com/go-gl/glfw`. Standard library only otherwise (`math`, `time`, `os`, `runtime/debug`).
**Storage**: N/A — world is generated in memory each launch; no persistence this sprint.
**Testing**: Go's built-in `testing` (table tests beside code) + `go test -bench` allocation assertions (`0 allocs/op` on the steady-state render path) + `go vet`.
**Target Platform**: Linux terminal; **kitty** is the reference/test terminal (truecolor, GPU-accelerated). Portable to any truecolor terminal with graceful palette degradation.
**Project Type**: Single-binary desktop CLI application (interactive TUI game).
**Performance Goals**: ≥ 30 FPS at a 120×40 terminal (≈ 240×80 px ⇒ ~19,200 rays/frame); input reflected within 50 ms (SC-002); continuous redraw (SC-003).
**Constraints**: **Zero heap allocations in the steady-state frame**; bounded, constant memory footprint (a few hundred KB) over a session — no leaks/growth (user directive + SC-006); clean terminal restore on quit (SC-009).
**Scale/Scope**: Bounded world ≥ 32×32 horizontally (default e.g. 64×64×32 = ~128 K cells, flat `[]uint8`); ≥ 2 block types; ≥ 1 entity + held item; single player, local.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

The project constitution (`.specify/memory/constitution.md`) is still the **unfilled template** — no ratified principles exist, so there are no project-specific gates to enforce. In their absence, this plan is held to general SDD good-practice gates:

| Gate | Status | Notes |
|------|--------|-------|
| **Simplicity / single project** | ✅ PASS | One Go module, one binary. Internal packages are a thin, idiomatic split, not microservices. |
| **No speculative abstraction (YAGNI)** | ✅ PASS | No plugin systems, ECS, or chunk-streaming. Concrete structs; interfaces only where a pass boundary genuinely needs one. |
| **Dependency minimalism** | ✅ PASS | Net dependency count **drops** (3 → 1): tcell replaces termbox and removes dead gl/glfw. |
| **Reuse over rewrite** | ✅ PASS | Existing vector math, camera, projection, and Bresenham/triangle code are extracted and reused, not discarded. |
| **Testability** | ✅ PASS | Pure logic (math, DDA, world edits, projection) is unit-testable headlessly; render path has allocation benchmarks. |

**Result**: PASS (no violations; Complexity Tracking left empty).

## Project Structure

### Documentation (this feature)

```text
specs/001-first-person-voxel-world/
├── plan.md              # This file
├── research.md          # Phase 0 output (complete)
├── data-model.md        # Phase 1 output (complete)
├── quickstart.md        # Phase 1 output (complete)
├── contracts/           # Phase 1 output (complete)
│   ├── controls.md          # Keyboard input → action contract
│   └── rendering.md         # Internal render-pass / framebuffer contract
└── tasks.md             # Phase 2 output (/speckit-tasks — NOT created here)
```

### Source Code (repository root)

The current monolithic `main.go` (643 lines) is decomposed into a small set of focused internal packages. Tests live beside the code they cover (Go convention), so there is no separate `tests/` tree.

```text
main.go                       # entry point: tcell setup, build game, run loop, restore terminal on exit
internal/
├── geom/                     # 3D math (extracted from current main.go)
│   ├── vec.go                #   Vector3 + Add/Sub/Scale/Normalize/Cross/Dot
│   └── vec_test.go
├── world/                    # the voxel world
│   ├── block.go              #   BlockType enum, palette (RGB + per-face shade), IsSolid
│   ├── world.go              #   World: flat []uint8 grid, At/Set, bounds, InBounds
│   ├── gen.go                #   initial terrain generation (ground + features)
│   ├── raycast.go            #   DDA voxel traversal: Cast(origin, dir) -> (hitCell, face, dist, ok)
│   ├── edit.go               #   Break(target) / Place(target, face, type) with validity rules
│   └── *_test.go             #   raycast, edit, bounds tests
├── render/                   # the hybrid renderer (no game logic)
│   ├── framebuffer.go        #   Buffer: flat color []RGB + depth []float32; Clear; Resize; Present (half-block → tcell)
│   ├── color.go              #   RGB type, shade(rgb, factor), face/distance shading helpers
│   ├── raycaster.go          #   Pass A: world → framebuffer (per-pixel DDA, perpendicular depth)
│   ├── raster.go             #   Pass B/C: filled triangles (entities/held/selection), depth-tested
│   ├── hud.go                #   Pass D: crosshair, selected-block indicator, help overlay
│   └── *_test.go
├── camera/                   # first-person camera
│   ├── camera.go             #   Position, Yaw, Pitch; Direction/Right/Up; ScreenRay(px,py,W,H)->dir
│   └── camera_test.go
├── entity/                   # non-block scene objects
│   └── entity.go             #   Entity (pos + mesh + color), held-item mesh, cube mesh builder
├── input/                    # input mapping
│   └── input.go              #   tcell event → Action (move/look/break/place/select/quit)
└── game/                     # state + loop glue
    ├── game.go               #   Game state: world, camera, player (selected block, reach), entities
    ├── update.go             #   per-tick update: apply actions with dt, recompute target via raycast
    └── loop.go               #   fixed-timestep ticker, input drain, calls render passes in order
```

**Structure Decision**: Single Go module, single binary, **6 internal packages** plus `main.go`. The split mirrors the render-pass architecture (R4) and the data model: `geom` (math) → `world` (data + DDA + edits) → `camera`/`entity` (scene) → `render` (passes + framebuffer) → `input`/`game` (control + loop). Dependencies flow one way (`game` → everything; `render` → `geom`/`world`/`camera`/`entity`; `world` → `geom`); no cycles. This keeps each piece independently testable (pure-logic packages have no tcell dependency) while staying small enough to finish in one sprint.

## Complexity Tracking

> No constitution violations — section intentionally empty.

## Phase Outputs

- **Phase 0 — Research**: [research.md](./research.md) ✅ — tcell/v2, half-block pixels, DDA raycast, hybrid shared-depth architecture, fixed-timestep loop, zero-allocation memory discipline, kitty truecolor setup. All NEEDS CLARIFICATION resolved.
- **Phase 1 — Design & Contracts**: [data-model.md](./data-model.md), [contracts/controls.md](./contracts/controls.md), [contracts/rendering.md](./contracts/rendering.md), [quickstart.md](./quickstart.md) ✅. Agent context (`CLAUDE.md`) updated to reference this plan.
- **Phase 2 — Tasks**: generated by `/speckit-tasks` (not part of this command).

## Post-Design Constitution Re-Check

Re-evaluated after Phase 1: the design introduces no new dependencies beyond tcell, no new projects, and no abstractions beyond the package split justified above. Pure-logic packages remain headlessly testable; the render path carries allocation benchmarks. **Still PASS.**
