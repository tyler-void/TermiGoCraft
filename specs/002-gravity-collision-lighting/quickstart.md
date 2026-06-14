# Quickstart: Gravity, Collision & Ray-Traced Lighting

**Feature**: `002-gravity-collision-lighting` | **Date**: 2026-06-15
Builds on feature 001. No new dependencies.

## Build & run

```bash
go build -o termigocraft .   # or: make build
./termigocraft               # walking mode, lighting on, fixed sun
```

## Controls (delta from 001)

```
Space   jump (walking)  /  up (flying)
C       down (flying)
G       toggle walk / fly
L       toggle lighting (sun + shadows)
(WASD move, arrows look, F/Enter break, R place, Q/E or 1-9 select, H help, Esc quit)
```

## Verifying against the spec

**P1 — Gravity / collision / jump (SC-001..004):**
1. Launch → you settle onto the ground instead of hovering.
2. Walk into a wall → stop; press a sideways key → slide along it (never enter the block).
3. Walk off a ledge → fall and land.
4. `Space` → jump; jump onto a 1-block step → land on top. Jump again mid-air → nothing.
5. Break the block under your feet → you fall.
6. Aim at your own feet/body and place → rejected.

**P2 — QoL (SC-005, SC-006):**
7. A cursor is always dead-center; aim at a block in reach → it switches to the "targeting" state; aim at sky → neutral.
8. `G` → toggle to fly (free movement, no gravity/collision); `G` again → walking. HUD shows mode + position.

**P3 — Ray-traced lighting (SC-007..010):**
9. Find a tree/raised terrain with the sun to one side → lit faces brighter than shaded faces; a shadow falls on the ground behind it.
10. Break/place a block → its shadow appears/disappears next frame.
11. Look into an inner corner → it's subtly darker (AO).
12. Stand still → lighting is steady (no shimmer). Press `L` → lighting off looks like 001; on again restores shadows.

## Performance / GC checks (project norm)

```bash
# Lit frame must stay zero-allocation
go test ./internal/render/ -bench=BenchmarkFrame -benchmem -run='^$'
#   expect: 0 B/op   0 allocs/op   (with lighting enabled in the benchmark)

# Confirm ≥30 FPS budget: ns/op for the lit 240x80 frame should be well under 33ms
# Long-session no-bleed:
GODEBUG=gctrace=1 ./termigocraft 2> gctrace.log   # flat heap over a 10-min session
```

## Headless tests (no terminal)

```bash
go test ./...        # adds: physics (rest/slide/jump/headbump/unstick/bounds),
                     #        lighting (shadow occlusion, directional, AO levels, determinism)
go vet ./...
```

If `BenchmarkFrame` shows allocations once lighting is on, the culprit is almost certainly a per-pixel `make`/closure in `lightAt` — hoist scratch into reused state or keep it value-typed (research R3/R4).
