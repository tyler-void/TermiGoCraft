---
description: "Task list for First-Person Voxel World"
---

# Tasks: First-Person Voxel World

**Input**: Design documents from `specs/001-first-person-voxel-world/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Targeted tests ARE included — the plan/quickstart explicitly require DDA correctness, cross-pass occlusion, and a `0 allocs/op` frame benchmark (the GC/no-bleed guarantee from research R6). These are not blanket TDD; only the design-mandated tests are tasked.

**Organization**: Tasks are grouped by user story (P1→P3) so each is an independently shippable increment.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependency on an incomplete task)
- **[Story]**: US1 / US2 / US3 (Setup, Foundational, Polish carry no story label)

## Path Conventions

Single Go module, single binary. Source at repo root: `main.go` + `internal/<pkg>/`. Tests live beside code (`*_test.go`). Per plan.md "Source Code" tree.

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization, dependency switch, package skeleton.

- [X] T001 Add `github.com/gdamore/tcell/v2@v2.13.10` via `go get`, and create the empty package skeleton directories `internal/{geom,world,render,camera,entity,input,game}` (each with a placeholder `doc.go` declaring the package)
- [X] T002 [P] Add a `Makefile` at repo root with `build`, `run`, `test`, `bench` (→ `go test ./internal/render/ -bench=. -benchmem`), and `vet` targets; confirm `gofmt`/`go vet ./...` baseline is clean

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: The shared engine substrate every user story needs — math, world data + DDA traversal, camera, the shared color+depth framebuffer, tcell screen lifecycle, the fixed-timestep loop, and input mapping.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete.

- [X] T003 [P] Extract `Vector3` math from current `main.go` into `internal/geom/vec.go` (Add, Sub, Scale, Normalize, Cross, Dot, Length)
- [X] T004 [P] Unit tests for vector math in `internal/geom/vec_test.go` (normalize zero-vector guard, cross/dot identities)
- [X] T005 [P] Define `BlockType` enum + static palette in `internal/world/block.go` (Air=0 + ≥4 solid types with RGB `Base`; `IsSolid(t)`)
- [X] T006 World grid in `internal/world/world.go` — flat `[]uint8` (index `x+SX*(y+SY*z)`), `NewWorld(sx,sy,sz)`, `InBounds`, `At` (Air when OOB), `Set` (in-place, no growth) — depends on T005
- [X] T007 DDA voxel traversal in `internal/world/raycast.go` — `Face` type (±X/±Y/±Z) and `Cast(origin, dir geom.Vector3, maxDist float64) (cell [3]int, face Face, dist float64, ok bool)` using Amanatides & Woo (research R3) — depends on T003, T006
- [X] T008 [P] World tests in `internal/world/world_test.go` + `internal/world/raycast_test.go` (bounds/At/Set; DDA hits the expected cell and reports the correct entry face; miss past maxDist returns ok=false) — depends on T006, T007
- [X] T009 [P] First-person camera in `internal/camera/camera.go` (`Pos, Yaw, Pitch, FOV, Reach, Selected`; `Direction/Right/Up`; `ScreenRay(px,py,W,H) geom.Vector3`) — depends on T003
- [X] T010 [P] Camera tests in `internal/camera/camera_test.go` (orthonormal basis; pitch clamp; center ray ≈ forward) — depends on T009
- [X] T011 [P] `RGB` type + shading helpers in `internal/render/color.go` (`shade(rgb, factor)`, `faceFactor(Face)`, `distFactor(dist)`)
- [X] T012 Shared framebuffer in `internal/render/framebuffer.go` — `Buffer{W,H,Color []RGB,Depth []float32}`, `NewBuffer`, `Resize` (only alloc site), `Clear(sky)`, `TestSet(x,y,z,c)` (write iff `z<Depth`), `Present(scr)` packing `▀` pairs (fg=top, bg=bottom) + `scr.Show()` — perpendicular camera-Z depth space (contracts/rendering.md G-1) — depends on T011
- [X] T013 [P] Allocation benchmark in `internal/render/framebuffer_test.go` asserting `Clear`+`Present` are `0 allocs/op` (research R6, contract G-2) — depends on T012
- [X] T014 [P] Input mapping in `internal/input/input.go` — map `tcell.Event` → `Action` set per contracts/controls.md (move/strafe/up-down/look/break/place/select/help/quit)
- [X] T015 [P] Input mapping tests in `internal/input/input_test.go` (every Action reachable from a default key; unknown keys ignored) — depends on T014
- [X] T016 Terminal lifecycle + loop skeleton — rewrite `main.go` to init tcell (truecolor via terminfo/COLORTERM, alt-screen), spawn an input-polling goroutine onto a buffered channel, run a fixed `time.Ticker` ~30 FPS loop that drains input, `Clear`s the framebuffer to sky and `Present`s it, and calls `screen.Fini()` to restore the terminal on quit; loop body in `internal/game/loop.go`. **Remove all termbox-based code.** — depends on T012, T014
- [X] T017 `go mod tidy` to drop `nsf/termbox-go` and the unused `go-gl/gl` + `go-gl/glfw`; verify `go build ./...` is clean — depends on T016

**Checkpoint**: Program launches into kitty showing a flat sky, holds ~30 FPS, redraws continuously, survives a resize, and quits cleanly restoring the terminal. No termbox/gl deps remain.

---

## Phase 3: User Story 1 - Explore a solid voxel world in first person (Priority: P1) 🎯 MVP

**Goal**: Player is dropped inside a generated blocky world, sees solid opaque blocks in first person, and can move + look smoothly with correct depth between blocks.

**Independent Test**: Launch → solid (not wireframe) blocks with ≥2 colors and a clear ground; W/A/S/D + arrows move/turn responsively; a nearer block correctly hides a farther one.

- [X] T018 [P] [US1] Initial terrain generation in `internal/world/gen.go` (`Generate(w)`: flat/rolling ground layer using ≥2 solid block types + a few raised features for parallax; spawn the camera above ground) — depends on T005, T006
- [X] T019 [P] [US1] World render pass in `internal/render/raycaster.go` — `RenderWorld(buf, w, cam)`: for each pixel cast `cam.ScreenRay` via `w.Cast`, convert hit distance to **perpendicular** camera-Z, shade by block base × `faceFactor` × `distFactor`, write with `TestSet`; misses leave sky (contracts/rendering.md Pass A) — depends on T007, T009, T011, T012
- [X] T020 [P] [US1] Headless test in `internal/render/raycaster_test.go` (camera facing a known block: expected pixels filled with expected color at expected perpendicular depth; facing empty: sky retained; no fisheye — center vs edge depth of a flat wall) — depends on T019
- [X] T021 [P] [US1] Movement + look in `internal/game/update.go` (dt-scaled forward/back/strafe on the ground plane, fly up/down, arrow-key yaw/pitch with pitch clamp) — depends on T009, T014
- [X] T022 [US1] Assemble game state + frame in `internal/game/game.go` and `internal/game/loop.go` — hold `World`, `Camera`; per frame `Clear`→`RenderWorld`→`Present`; apply drained input via update; wire up in `main.go` — depends on T016, T018, T019, T021
- [X] T023 [P] [US1] `BenchmarkFrame` in `internal/render/raycaster_test.go` over `Clear`+`RenderWorld`+`Present` asserting `0 allocs/op` and recording ns/op for the 240×80 default (research R6, SC-003) — depends on T019

**Checkpoint**: 🎯 **MVP** — walk and look around a solid first-person voxel world; meets SC-001/002/003 and FR-001..006. Shippable.

---

## Phase 4: User Story 2 - Break and place blocks (Priority: P2)

**Goal**: Player aims at a block (highlighted), breaks it or places the selected block type against the aimed face, with the world updating immediately.

**Independent Test**: Aim → highlight appears; Break removes the block; Place adds one on the aimed face; aiming at sky and acting is a no-op; placing into own cell is rejected; switching block type changes what gets placed.

- [X] T024 [P] [US2] Block edits in `internal/world/edit.go` — `Break(w, cell)` → set Air; `Place(w, cell, face, t, playerCell)` → compute face-adjacent neighbor, reject if OOB / occupied / equals `playerCell`, else set `t` (data-model "Block edits", FR-008/009/010) — depends on T006, T007
- [X] T025 [P] [US2] Edit tests in `internal/world/edit_test.go` (break clears the cell; place fills the correct neighbor; rejects occupied, OOB, and self-placement; no-op semantics never panic) — depends on T024
- [X] T026 [P] [US2] Per-tick targeting in `internal/game/update.go` — cast a single center-screen ray (`w.Cast` to `cam.Reach`) → store `Target{cell, face, ok}` on the game state each tick — depends on T007, T021
- [X] T027 [US2] Selection highlight in `internal/render/raster.go` — `RenderSelection(buf, target, cam)` outlining the targeted block face/edges, depth-aware so it reads on top of its own surface — depends on T012, T026
- [X] T028 [US2] Wire actions in `internal/game/update.go` + `internal/input` — Break/Place call `world.edit` against the current `Target`; `SelectNext/Prev` (Q/E) and `SelectSlot` (1–9) change `cam.Selected` (FR-011) — depends on T024, T026, T014

**Checkpoint**: Aim-highlight + break/place + block-type selection work; meets SC-004 and FR-007..011. US1 still works.

---

## Phase 5: User Story 3 - Entities, held item, and HUD with correct depth (Priority: P3)

**Goal**: Non-block objects (an entity + the held item) are rasterized into the same scene with correct occlusion vs terrain, under a HUD (crosshair + selected-block indicator + help).

**Independent Test**: Walk so a wall passes between you and the entity → it hides; in front → it draws over terrain. Crosshair + selected-block indicator always visible on top; help toggles.

- [X] T029 [P] [US3] Entity package in `internal/entity/entity.go` (`Entity{Pos, Mesh []geom.Triangle, Color RGB}`, a unit-cube mesh builder, and a held-item mesh) — depends on T003
- [X] T030 [P] [US3] Filled triangle rasterizer in `internal/render/raster.go` — `RenderMesh(buf, mesh, modelPos, color, cam)`: project vertices→pixels + per-vertex camera-Z, backface-cull, **filled** (barycentric) fill with interpolated depth via `TestSet` (upgrades the old wireframe path; contracts Pass B) — depends on T009, T011, T012
- [X] T031 [P] [US3] Occlusion test in `internal/render/raster_test.go` — entity behind a solid block (rendered after `RenderWorld`) is NOT written; same entity in front IS written (SC-005, the hybrid depth proof) — depends on T019, T030
- [X] T032 [US3] View-space held item `RenderHeld(buf, mesh, color, cam)` in `internal/render/raster.go` (fixed relative to camera, depth-tested) — depends on T030
- [X] T033 [P] [US3] HUD overlay in `internal/render/hud.go` — `RenderHUD(scr, state)` drawing a center crosshair, the selected-block name/swatch, and a toggleable controls/help line, written to tcell cells after `Present` (no depth, always on top; FR-013/016) — depends on T012, T016
- [X] T034 [US3] Add entities to game state and run passes in order in `internal/game/loop.go` + `game.go` — Clear→RenderWorld→RenderMesh(entities)→RenderHeld→RenderSelection→Present→RenderHUD (contracts "Frame contract") — depends on T027, T029, T030, T032, T033

**Checkpoint**: All three stories work; meets SC-005 and FR-012..016. Full feature complete.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Tuning, verification of the smoothness/GC/no-bleed directives, and docs.

- [X] T035 [P] Update `README.md` (and confirm `CLAUDE.md`) with controls, build/run, and kitty truecolor note
- [X] T036 [P] Tune defaults for kitty at 120×40 (FOV, world size, view distance, FPS cap) and confirm ≥30 FPS via `BenchmarkFrame` ns/op (SC-003)
- [ ] T037 Memory/GC verification per quickstart.md — `GODEBUG=gctrace=1 go run .` for a ~10-minute session with movement + 50+ edits + a resize; confirm flat heap / infrequent GC / no growth (SC-006, research R6)
- [ ] T038 Run the quickstart.md acceptance walkthrough end-to-end in kitty (P1+P2+P3 + robustness + clean-quit SC-009)
- [X] T039 [P] `gofmt`, `go vet ./...`, and remove any dead code left from the original wireframe demo

---

## Dependencies & Execution Order

### Phase dependencies
- **Setup (P1)** → no deps.
- **Foundational (P2)** → after Setup. **Blocks all stories.** Internal order: T003/T005/T009/T011/T014 start in parallel → T006→T007(→T008); T012(→T013); T010, T015; then T016→T017.
- **US1 (P3)** → after Foundational. The MVP.
- **US2 (P4)** → after Foundational; independently testable. (Reuses `world.Cast` from T007.)
- **US3 (P5)** → after Foundational; independently testable. (Adds the rasterizer; relies on `RenderWorld` from US1 only for the occlusion *test* T031, not for its own functionality.)
- **Polish (P6)** → after the stories you intend to ship.

### Story independence
- US1, US2, US3 each sit on the Foundational substrate and can be built/tested on their own. Sequential priority order P1→P2→P3 is recommended for solo work; with multiple devs they can proceed in parallel after Phase 2.

### Within a story
- Tests for pure logic (T020, T025, T031) can be written alongside their implementation; the alloc benchmarks (T013, T023) follow their target code.

### Parallel opportunities
- Setup: T002 ∥ T001's tail.
- Foundational kickoff (all different packages): **T003, T005, T009, T011, T014 in parallel**, plus their tests (T004, T008, T010, T013, T015) as each unblocks.
- US1: T018 ∥ T019 ∥ T021 (different files) → T022 joins them; T020, T023 [P] after T019.
- US2: T024 ∥ T025 ∥ T026 → T027 → T028.
- US3: T029 ∥ T030 ∥ T033 → T032 → T034; T031 [P] after T030.
- Polish: T035 ∥ T036 ∥ T039.

---

## Parallel Example: Foundational kickoff

```bash
# Five independent packages can be started simultaneously:
Task: "T003 Extract Vector3 into internal/geom/vec.go"
Task: "T005 BlockType + palette in internal/world/block.go"
Task: "T009 Camera in internal/camera/camera.go"
Task: "T011 RGB + shading in internal/render/color.go"
Task: "T014 Input mapping in internal/input/input.go"
```

## Parallel Example: User Story 1

```bash
Task: "T018 Terrain generation in internal/world/gen.go"
Task: "T019 World raycast pass in internal/render/raycaster.go"
Task: "T021 Movement + look in internal/game/update.go"
# then T022 wires them together; T020 + T023 validate.
```

---

## Implementation Strategy

### MVP first (User Story 1)
1. Phase 1 Setup → 2. Phase 2 Foundational (the bulk of the engine) → 3. Phase 3 US1 → **STOP & VALIDATE**: walk a solid voxel world in kitty. This alone is a demoable product and the biggest visible leap from the wireframe demo.

### Incremental delivery
- Foundational + US1 → MVP (explore). 
- + US2 → editable world (the Minecraft loop). 
- + US3 → entities/held item + HUD (the full hybrid renderer). 
- Each increment is independently testable and doesn't break the prior one.

### Notes
- `[P]` = different files, no incomplete dependency. `[Story]` = traceability to spec.md user stories.
- Keep the steady-state frame allocation-free (T013/T023 guard this); commit after each task or logical group.
- Stop at any checkpoint to validate a story independently.
