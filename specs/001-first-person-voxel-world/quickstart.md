# Quickstart: First-Person Voxel World

**Feature**: `001-first-person-voxel-world` | **Date**: 2026-06-15
For a developer picking up implementation, and for verifying the result.

## Prerequisites

- **Go 1.24+** (module targets `go 1.24.2`).
- A **truecolor terminal** — **kitty** is the reference (test target). Any terminal exporting `COLORTERM=truecolor` works; others degrade to nearest palette.
- Terminal at least ~80×24; default scene tuned for ~120×40.

## One-time dependency switch

```bash
# Add the new terminal library
go get github.com/gdamore/tcell/v2@v2.13.10

# Remove the old/unused deps after code no longer imports them
go mod tidy   # drops nsf/termbox-go and the unused go-gl/gl + glfw once imports are gone
```

## Build & run

```bash
go build -o termigocraft .
./termigocraft
# or simply:
go run .
```

You should drop straight into a first-person view of a solid blocky world (User Story 1).

## Controls (see contracts/controls.md)

```
Move:   W/A/S/D        Up/Down: Space / Shift(or C)
Look:   Arrow keys     Break:   F/Enter     Place: R
Block:  Q/E or 1–9     Help:    H/?         Quit:  Esc / Ctrl+C
```

## Verifying against the spec (acceptance walkthrough)

**P1 — Explore (SC-001, SC-002, SC-003):**
1. Launch → confirm solid (not wireframe) blocks, visible ground, ≥ 2 block colors.
2. Press W/A/S/D and arrows → view moves/turns smoothly, responds within ~50 ms.
3. Walk so one block passes in front of another → nearer block correctly hides the farther one.

**P2 — Edit (SC-004):**
4. Aim at a block → it is highlighted.
5. Break (F) → block vanishes immediately; Place (R) → new block appears on the aimed face.
6. Aim at sky and Break/Place → nothing happens, no crash.
7. Try to place into your own position → rejected.
8. Cycle block type (E / number keys) → next placement uses the new type.

**P3 — Entities & HUD (SC-005):**
9. Find the in-world entity; walk so a wall comes between you and it → it disappears; step so it's in front → it draws over terrain.
10. Confirm crosshair + selected-block indicator always visible on top.

**Robustness (SC-006, SC-009):**
11. Resize the kitty window mid-session → view re-adapts, no corruption/crash.
12. Play ~10 min with movement + 50+ edits → no crash, no slowdown.
13. Quit (Esc) → terminal returns to normal (cursor back, no raw-mode artifacts).

## Performance & memory checks (user directives: smooth, good GC, no bleed)

```bash
# Allocation discipline: steady-state frame must be zero-alloc
go test ./internal/render/ -bench=BenchmarkFrame -benchmem
#   expect:  0 B/op   0 allocs/op

# Watch GC behavior during a live session (should be quiet / flat heap)
GODEBUG=gctrace=1 go run . 2> gctrace.log
#   expect: very infrequent GC, heap not growing over time

# Optional: relax GC since we barely allocate
GOGC=300 ./termigocraft
```

If `BenchmarkFrame` shows any allocs/op, find the per-frame `make`/`append`/boxing/closure and hoist it into a preallocated, reused field (research R6).

## Headless tests (no terminal needed)

```bash
go test ./...        # geom, world (DDA + edits + bounds), camera, render passes via headless Buffer
go vet ./...
```

Key headless assertions: DDA hits the expected cell/face; `Break`/`Place` validity rules; occlusion — an entity behind a block is not written to the framebuffer, in front it is.
