---
description: "Task list for Gravity, Collision & Ray-Traced Lighting"
---

# Tasks: Gravity, Collision & Ray-Traced Lighting

**Input**: Design documents from `specs/002-gravity-collision-lighting/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/
**Builds on**: feature 001 (shipped) — reuses `geom`, `world` (incl. DDA `Cast`), `camera`, `render`, `game`.

**Tests**: Design-mandated tests are included — physics behaviors, lighting (shadow/directional/AO/determinism), and the extended `0 allocs/op` lit-frame benchmark (project norm). Not blanket TDD.

**Organization**: Grouped by user story (P1→P3), each independently shippable.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: parallelizable (different file, no incomplete dependency)
- **[Story]**: US1 / US2 / US3 (Setup, Foundational, Polish carry no story label)

## Path Conventions

Single Go module at repo root: `main.go` + `internal/<pkg>/`; tests beside code. Per plan.md.

---

## Phase 1: Setup

- [X] T001 [P] Create the `internal/physics` package skeleton (`internal/physics/doc.go` declaring the package); confirm `go build ./...` and the feature-001 test suite are green before changes.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Input actions shared by all three stories.

**⚠️ CRITICAL**: Complete before the user-story phases.

- [X] T002 Extend the input layer in `internal/input/input.go`: add actions `Jump`, `ToggleFly`, `ToggleLighting`; map `Space`→`Jump` (the game routes it to jump in Walk and ascend in Fly), `C`→`MoveDown` (fly descend), `G`→`ToggleFly`, `L`→`ToggleLighting` (keep all 001 bindings).
- [X] T003 [P] Update `internal/input/input_test.go` to cover the new bindings (Space→Jump, G→ToggleFly, L→ToggleLighting) and that unknown keys remain `None`.

**Checkpoint**: input compiles and maps the new actions; nothing else changed yet.

---

## Phase 3: User Story 1 - Gravity, collision, and jumping (Priority: P1) 🎯 MVP

**Goal**: The player has a body, falls under gravity, stands on/【cannot pass through】blocks, slides along walls, and can jump. Fixes the clip-through.

**Independent Test**: Stop moving → settle on the ground. Walk into a wall → stop but slide sideways. Walk off a ledge → fall/land. `Space` → jump; jump onto a 1-block step → land on top; mid-air jump → nothing. Break the block underfoot → fall.

- [X] T004 [P] [US1] Define `Body` (Feet, Vel, Half, Height, EyeHeight, Grounded, Mode), `Mode` (Walk/Fly), `Intent`, `Eye()`, and tunable consts (Gravity, JumpSpeed, WalkSpeed, MaxFall, MoveDecay) in `internal/physics/body.go`.
- [X] T005 [P] [US1] Implement `Overlaps(w, feet, half, height) bool` (AABB-vs-solid over the integer cell range, allocation-free) in `internal/physics/collide.go`.
- [X] T006 [US1] Implement `Step(w, b, in, yaw, dt)` in `internal/physics/step.go`: apply gravity (clamp MaxFall), build world-space horizontal velocity from camera-relative intent, resolve motion **per axis** (Y,X,Z) against `Overlaps` with clamp + zero-velocity (wall slide), set `Grounded`, head-bump stop, jump only when grounded, world-bounds clamp, start-of-tick unstick, and a Fly branch (no gravity/collision) — per contracts/physics.md — depends on T004, T005.
- [X] T007 [P] [US1] Physics tests in `internal/physics/step_test.go`: rest-on-ground (Grounded, no sink), wall slide keeps parallel axis, jump sets +Vel.Y once and mid-air jump is a no-op, head-bump zeroes upward Vel, unstick relocates a body spawned in stone, walking toward the edge clamps in-bounds — depends on T006.
- [X] T008 [US1] Wire physics into the loop: hold `*physics.Body` (+ `Mode` default Walk) in `internal/game/game.go`; build `Intent` from drained input with horizontal decay and handle `Jump` per mode in `internal/game/update.go`; call `physics.Step` every tick and set `cam.Pos = body.Eye()` in `internal/game/loop.go`; spawn the body above ground — depends on T002, T006.
- [X] T009 [US1] Extend the placement guard so `world.Place` rejects a block intersecting any part of the player body (use the body's occupied cell range / `Overlaps`) in `internal/world/edit.go` + the call site in `internal/game/update.go` — depends on T005, T008.

**Checkpoint**: 🎯 **MVP** — gravity, collision, wall-slide, and jump work; meets SC-001..004 and FR-001..009. Shippable; fixes the clipping.

---

## Phase 4: User Story 2 - QoL: aiming cursor & movement mode (Priority: P2)

**Goal**: A 2-state center cursor (neutral vs targeting), a walk/fly toggle, and a mode+position readout.

**Independent Test**: Cursor always dead-center; aim at a reachable block → "targeting" state; aim at sky → neutral. `G` toggles walk/fly (fly = free, no gravity/collision). HUD shows mode + position.

- [X] T010 [P] [US2] Give the center cursor two states in `internal/render/hud.go` — neutral (dim/thin) vs targeting (bright/distinct) — driven by a `targeted bool` param — depends on (foundational only).
- [X] T011 [US2] Handle `ToggleFly` (`G`) in `internal/game/update.go`: flip `body.Mode`; ensure Fly path in `Step` ignores gravity/collision and `Space`/`C` move vertically while flying — depends on T008.
- [X] T012 [US2] Add a status readout (current mode Walk/Fly + integer player position) to `internal/render/hud.go`, with `game` passing mode + position — depends on T010, T008.

**Checkpoint**: cursor reflects targeting, fly toggle works, status shown; meets SC-005, SC-006 and FR-010..013. US1 still works.

---

## Phase 5: User Story 3 - Ray-traced lighting and shadows (Priority: P3)

**Goal**: Directional sun + hard cast shadows + voxel ambient occlusion, updating live with edits/movement, deterministic (no shimmer). Independent of US1/US2 (touches only the world render pass).

**Independent Test**: Sun-facing faces brighter than away-facing; a tree/pillar casts a ground shadow; break/place updates shadows next frame; inner corners darker (AO); static scene doesn't shimmer; `L` toggles lighting (off == feature 001 look).

- [X] T013 [P] [US3] Add `Sun` type + `lightAt` (directional `max(0,n·Dir)` + ambient) + one hard shadow ray via `world.Cast` (with `ShadowBias`, `ShadowDist`) in `internal/render/lighting.go` — per contracts/lighting.md.
- [X] T014 [US3] Add voxel ambient occlusion to `internal/render/lighting.go`: `vertexAO(s1,s2,c)=(s1&&s2)?0:3-(s1+s2+c)` at the 4 face corners (3 neighbor reads each), map 0..3→AOmin..1, bilinearly interpolate by the hit point's in-face UV — depends on T013.
- [X] T015 [US3] Integrate lighting into `internal/render/raycaster.go`: compute `hitPoint = cam.Pos + dir*hit.Dist`, add a `sun Sun, lit bool` param to `RenderWorld`, call `lightAt` when `lit` else the 001 fallback shading; `game` holds a `Sun` and a lighting flag toggled by `L` — depends on T013, T014, T002.
- [X] T016 [P] [US3] Lighting tests in `internal/render/lighting_test.go`: pillar casts a shadow (shadowed cell darker than open), top face brighter than bottom under an overhead sun, inner corner darker than open face (AO), `lightAt` deterministic across identical calls — depends on T015.
- [X] T017 [P] [US3] Extend `BenchmarkFrame` in `internal/render/raycaster_test.go` to run with lighting enabled and assert `0 allocs/op` (project norm; record ns/op vs the 33ms budget) — depends on T015.
- [X] T018 [US3] Update `internal/render/hud.go`: show lighting on/off in the status line and refresh the help text with the new keys (Space jump, G fly, L lighting) — depends on T012, T015.

**Checkpoint**: lighting + shadows + AO render, toggleable, zero-alloc; meets SC-007..010 and FR-014..019. All stories complete.

---

## Phase 6: Polish & Cross-Cutting

- [X] T019 [P] Tune defaults for feel/look in kitty (gravity, jump height, walk speed, decay; sun direction/intensity/ambient, AOmin, shadow bias/dist) in `internal/physics/body.go` + `internal/render/lighting.go`.
- [ ] T020 Memory/GC verification: confirm lit `BenchmarkFrame` is `0 allocs/op` and run `GODEBUG=gctrace=1` for a ~10-minute walk/jump/edit session — flat heap, no bleed (SC + project norm).
- [ ] T021 Run the quickstart.md acceptance walkthrough end-to-end in kitty (P1 gravity/collision/jump, P2 cursor/fly, P3 lighting + clean quit).
- [X] T022 [P] `gofmt`, `go vet ./...`, and update `README.md` controls (jump, fly toggle, lighting toggle).

---

## Dependencies & Execution Order

### Phase dependencies
- **Setup (T001)** → none.
- **Foundational (T002–T003)** → after Setup; blocks stories that use the new actions.
- **US1 (T004–T009)** → after Foundational. The MVP; the base for US2.
- **US2 (T010–T012)** → T010 (cursor) needs only Foundational; T011/T012 need US1's `Body`.
- **US3 (T013–T018)** → **independent of US1/US2** (only the world render pass + input toggle). Could be built first or in parallel; T015 needs T002 for the `L` toggle.
- **Polish (T019–T022)** → after the stories you intend to ship.

### Within stories
- US1: T004 ∥ T005 → T006 → T007; T008 (wiring) after T006; T009 after T008.
- US2: T010 ∥ (US1) → T011 → T012.
- US3: T013 → T014 → T015 → (T016 ∥ T017) → T018.

### Parallel opportunities
- Foundation: T003 after T002.
- US1 kickoff: **T004 ∥ T005** (body.go ∥ collide.go); T007 after T006.
- Cross-story: **US3 (T013/T014) can proceed in parallel with US1** — different files, no shared state — since lighting doesn't depend on physics.
- US3: T016 ∥ T017 (different test files) after T015.
- Polish: T019 ∥ T022.

---

## Parallel Example: US1 kickoff

```bash
Task: "T004 Body/Mode/Intent in internal/physics/body.go"
Task: "T005 Overlaps(AABB vs solid) in internal/physics/collide.go"
# then T006 Step joins them; T007 tests.
```

## Parallel Example: lighting alongside physics

```bash
# US3 is independent of US1 — run a lighting track in parallel:
Task: "T013 Sun + lightAt + shadow ray in internal/render/lighting.go"
Task: "T004 physics Body in internal/physics/body.go"
```

---

## Implementation Strategy

### MVP first (User Story 1)
Setup → Foundational → US1 → **STOP & VALIDATE** in kitty: you now stand on, bump into, and jump around the world (the clip-through fix). Shippable on its own.

### Incremental delivery
- + US2 → cursor feedback + fly toggle (keep building easily).
- + US3 → the ray-traced lighting payoff (do last; can be tuned or `L`-toggled off without affecting gameplay).

### Notes
- Physics runs every tick (gravity) regardless of input; horizontal motion uses decay since terminals lack key-up (research R5).
- Keep the lit frame allocation-free (T017 guards it). Commit after each task or logical group; stop at any checkpoint to validate independently.
