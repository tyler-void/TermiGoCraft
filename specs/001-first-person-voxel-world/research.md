# Phase 0 Research: First-Person Voxel World

**Feature**: `001-first-person-voxel-world`
**Date**: 2026-06-15
**Inputs that shaped this research**: user is free to add an external library if it doesn't clutter the project (or reinvent freely); rendering must run **smoothly**, be **GC-friendly**, and **not bleed memory**; testing terminal is **kitty**.

This document resolves every open technical question before design. Each section: **Decision → Rationale → Alternatives considered**.

---

## R1. Terminal output library: replace termbox-go with tcell/v2

**Decision**: Swap the current `github.com/nsf/termbox-go` for **`github.com/gdamore/tcell/v2`** (v2.13.10). Remove the unused `go-gl/gl` and `go-gl/glfw` dependencies left over from an abandoned OpenGL approach.

**Rationale**:
- **24-bit truecolor** — tcell renders full RGB via `tcell.NewRGBColor(r,g,b)`; kitty advertises truecolor, so we get the whole color space for block tints and depth shading. termbox-go is limited to 256 colors at best and is awkward about it.
- **Minimal redraw** — tcell diffs its internal cell buffer and only emits escape sequences for cells that actually changed between `Show()` calls, "avoiding repeated sequences or drawing the same cell on refresh updates." That is exactly the property a 30 FPS terminal renderer needs.
- **Low/zero-allocation draw path** — `SetContent` writes into a preallocated cell grid; no per-call heap allocation. We keep our own pixel/depth buffers and only touch tcell at present time.
- **Maintenance & ecosystem** — tcell is the actively maintained successor to termbox (which is effectively frozen). It even ships a termbox compatibility shim, but we will use the native API for RGB.
- **One dependency, not a cluster** — this is a *net reduction*: termbox + 2 unused gl/glfw deps → a single tcell dependency.

**Alternatives considered**:
- **Keep termbox-go**: no truecolor, stagnant. Rejected — color is central to making solid blocks legible.
- **tcell/v3 (v3.4.0)**: current major as of May 2026, but newer with far less reference material and some API churn vs v2. Pinning v2 maximizes implementation confidence this sprint. *Revisit v3 later; the migration is small.*
- **Hand-rolled ANSI writer** (reinvent the wheel): we'd reimplement truecolor sequences, cell diffing, resize/SIGWINCH handling, raw-mode setup, and key decoding. High effort, high bug surface, no payoff over tcell. Rejected — this is the "clutter by reinvention" the user warned against.
- **bubbletea/lipgloss**: Elm-architecture TUI framework — great for forms/dashboards, wrong model for a per-frame 3D renderer. Rejected.

---

## R2. Pixel model: Unicode half-block (▀) for 2× vertical resolution

**Decision**: Treat each terminal **cell as two stacked pixels**. Render the upper-half-block glyph **`▀` (U+2580)** where **foreground = top pixel color** and **background = bottom pixel color**. The internal framebuffer is a pixel grid of size `W × (2·H)` (W = terminal columns, H = terminal rows).

**Rationale**:
- Terminal cells are ~2:1 tall; one color per cell yields stretched, chunky pixels. The half-block trick gives **two independently-colored pixels per cell**, producing near-square pixels and dramatically smoother 3D — the same technique used by the `minecraftty` reference renderer.
- Requires only truecolor, which kitty has. Pure glyph rendering — no terminal graphics protocol needed (keeps it portable).
- Present step is a tight, allocation-free loop pairing `pixel[y]` (top→fg) with `pixel[y+1]` (bottom→bg).

**Alternatives considered**:
- **One pixel per cell** (current approach): simplest but blocky and vertically stretched. Acceptable fallback if a terminal lacks truecolor; we keep it as a degradation path but default to half-block.
- **Quadrant/sextant block glyphs** (▖▘ / U+1FB00 block sextants): up to 2×3 sub-cells, but each cell can carry only one fg + one bg color, so sub-cells inside a cell can't be independently colored — it only helps monochrome. Worse for a colored voxel scene. Rejected.
- **Kitty graphics protocol (real images)**: would look best but is kitty-specific and pulls us toward "GPU pixels piped to terminal" (the minecraftty model). Out of scope and non-portable. Rejected for this sprint.

---

## R3. World rendering: DDA voxel raycasting (Amanatides & Woo)

**Decision**: Render the voxel world with a **per-pixel ray cast** using the **Amanatides & Woo "Fast Voxel Traversal"** (3D DDA) algorithm. For each screen pixel, build a ray from the camera through that pixel, step cell-by-cell through the grid until the first solid block is hit (or reach max view distance / world bound), then shade that pixel from the block type, the **face** that was hit, and the distance.

**Rationale**:
- **Inherently first-person** — the camera is the ray origin; the required POV falls out for free (satisfies FR-003).
- **No meshing** — the world stays a plain 3D array of block IDs; we never build/sort triangles for terrain. This is *less* code than fixing the existing triangle rasterizer to fill + cull + sort, and it is the approach the working terminal clones use.
- **Cost scales with screen pixels, not block count** — a bigger world is free; cost is bounded by `W·2H·maxSteps`. Predictable frame time.
- **Face detection is free** — the last axis stepped (X/Y/Z) tells us which face was hit, giving per-face shading (FR-005) at no extra cost.
- **Editing is the same ray** — break = the hit cell; place = the empty cell on the hit-face side (FR-007/8/9). One algorithm powers rendering *and* interaction.

**Key correctness detail — perpendicular distance**: depth/shading must use the **camera-space Z (perpendicular) distance**, i.e. the ray parameter projected onto the camera forward axis — **not** raw Euclidean ray length. Using Euclidean length produces fisheye distortion and wrong occlusion at screen edges (the classic Wolfenstein fix). This same camera-Z metric is what the rasterizer pass writes, so the two passes share one depth scale (see R4).

**Alternatives considered**:
- **Triangle rasterization of meshed chunks**: general (supports arbitrary meshes) but needs greedy meshing, face culling, triangle fill, and per-triangle depth — more code, more allocation, more bugs for a cube world. Kept only for entities/overlay (R4). Rejected as the *terrain* renderer.
- **Sparse voxel octree / sphere tracing**: overkill for a small bounded world; premature optimization. Rejected.

---

## R4. Hybrid architecture: one framebuffer + one depth buffer, drawn in passes

**Decision**: Both renderers write into **one shared pixel framebuffer** (`color []RGB`) and **one shared depth buffer** (`depth []float32`), then a single present step packs pixels into cells. Render order per frame:

1. **Clear** color (to sky) and depth (to +∞).
2. **Pass A — World (raycaster)**: fill terrain pixels, write camera-Z depth.
3. **Pass B — Entities & held item (rasterizer)**: fill triangles, **depth-tested** against Pass A (`if z < depth[i] { write }`). Reuses the existing triangle/projection code, upgraded from wireframe to filled + RGB.
4. **Pass C — Selection highlight**: outline the targeted block (depth-tested or drawn slightly biased so it reads on top of its own face).
5. **Pass D — HUD/overlay**: crosshair + selected-block indicator + help, written directly to cells (no depth) **after** present, or to pixels with depth = −∞.
6. **Present**: pack vertical pixel pairs → `tcell.SetContent('▀', fg=top, bg=bottom)`, then `screen.Show()` (diffed flush).

**Rationale**: The shared depth buffer is the entire trick that makes occlusion between terrain and entities correct (FR-012, SC-005). Because both passes agree on the camera-Z metric (R3), a mob behind a wall loses the depth test and a mob in front wins it — automatically. This realizes the "single depth representation" the spec calls for, keeps the existing rasterizer useful, and cleanly separates "world" from "stuff in the world".

**Alternatives considered**:
- **Two separate buffers composited afterward**: needs a manual depth compare at composite time anyway — same logic, extra memory and a copy. Rejected.
- **Raycast everything including entities**: entities aren't axis-aligned voxels; raycasting arbitrary meshes is hard. Rasterizing a handful of small meshes is trivial and cheap. Rejected.

---

## R5. Smoothness: fixed-timestep loop, time-based movement, synchronized output

**Decision**:
- Run a **fixed-timestep render loop** targeting **~30 FPS** (≈33 ms/frame) driven by a `time.Ticker`.
- **Decouple input from rendering**: a goroutine runs `screen.PollEvent()` and pushes onto a buffered channel; the main loop drains pending input each tick. (Replaces the current "only redraw on keypress" model — satisfies FR-014.)
- **Time-based movement**: scale movement/turn by real elapsed `dt` so speed is frame-rate independent and feels smooth even if a frame runs long.
- **Synchronized output / no tearing**: wrap each frame's terminal writes in the DEC **synchronized update** sequence (mode `2026`: `ESC[?2026h` … `ESC[?2026l`), which kitty supports, so the terminal swaps the whole frame atomically instead of showing a half-drawn frame. If the tcell version in use doesn't expose this, emit the two sequences around `Show()` directly.

**Rationale**: Fixed timestep + diffed redraw + atomic frame swap = smooth, flicker-free motion (SC-002, SC-003). Input on its own goroutine keeps controls responsive regardless of render cost. kitty offloads rendering to the GPU and is very fast, so 30 FPS at our resolution is comfortably achievable.

**Alternatives considered**:
- **Render only on input** (current): no continuous motion/animation, can't meet FR-014. Rejected.
- **Unbounded busy loop**: pegs a CPU core, wastes battery, no benefit past the terminal's refresh. Rejected — cap with the ticker.
- **Variable timestep with interpolation**: nicer for physics, unnecessary without gravity/physics this sprint. Deferred.

---

## R6. Memory discipline: preallocate, flat arrays, zero per-frame allocation (GC concern)

**Decision** — directly addressing "good GC, don't bleed memory":
- **Flat 1D buffers, indexed `y*W + x`** for color and depth — *not* `[][]float64`. The current code's slice-of-slices allocates one slice per row (many small heap objects, pointer-chasing, poor cache locality). One contiguous backing array is faster *and* gives the GC almost nothing to scan.
- **Allocate once, reuse every frame.** Buffers are created at startup and on resize **only**. The steady-state render path performs **zero heap allocations** — no `make`, no `append`-growth, no closures-in-loops, no interface boxing, no per-frame `fmt.Sprintf` in the hot path.
- **Clear by overwrite, not realloc**: reset depth/color with a `for i := range buf` loop (the Go compiler lowers slice-clear to a fast memset).
- **World = flat `[]uint8` of block IDs** for the bounded grid. Contiguous, **pointer-free** → the GC scans none of it; memory is fixed-size and cannot "bleed" as you play (editing mutates bytes in place, never grows the world).
- **Bounded everything**: fixed world dims and buffers sized to the terminal mean total footprint is a few hundred KB and constant over a session (supports SC-006: 10-minute session, no growth).
- **GC tuning is secondary**: with near-zero steady-state allocation the GC essentially never fires. We may set `debug.SetGCPercent` higher (e.g. 200–400) or rely on `GOGC` as a tunable, but correctness comes from *not allocating*, not from tuning the collector.

**Rationale**: In a Go render loop, frame-time spikes come overwhelmingly from GC pauses triggered by per-frame allocations. Eliminating allocations in the hot path removes the cause rather than treating the symptom. Flat, pointer-free buffers minimize both allocation count and GC scan work — the two things the user explicitly asked us to protect.

**Validation approach**: a benchmark / `go test -bench` over the render path asserting **0 allocs/op** for a steady-state frame (no resize), plus `pprof`/`GODEBUG=gctrace=1` spot-checks during a long session to confirm a flat heap.

**Alternatives considered**:
- **`sync.Pool` for frame buffers**: pooling suits churn of short-lived objects; our buffers live for the whole program, so a plain preallocated field is simpler and strictly better. Rejected (would be cargo-culting).
- **Keep `[][]float64`**: simplest to leave alone but it's the allocation/locality anti-pattern we're explicitly avoiding. Rejected.

---

## R7. kitty-specific setup & truecolor enablement

**Decision**: Target kitty's defaults; make truecolor robust:
- tcell auto-detects truecolor from terminfo (`RGB`/`Tc`) or from `COLORTERM`. kitty sets `TERM=xterm-kitty`; ensure RGB is honored by relying on `COLORTERM=truecolor` (kitty sets it) and **not** forcing `TCELL_TRUECOLOR=disable`.
- Use kitty's strengths: it's GPU-accelerated and fast, so the half-block + 30 FPS plan has ample headroom.
- Keep a **graceful degradation path**: if truecolor is unavailable, fall back to tcell's nearest-palette mapping automatically (tcell handles this), and optionally to one-pixel-per-cell.
- Quit must **restore the terminal** (`screen.Fini()`), clearing raw mode / alt-screen / cursor (SC-009, FR-016).

**Rationale**: Matches the actual test environment, uses standard env-var detection (no kitty-only code paths required), and stays portable to other truecolor terminals.

**Alternatives considered**:
- **Hard-code kitty escape sequences**: ties us to one terminal and duplicates tcell. Rejected.
- **Require a config flag for truecolor**: unnecessary; detection works. Rejected.

---

## Summary of decisions

| # | Area | Decision |
|---|------|----------|
| R1 | Terminal lib | tcell/v2 (drop termbox + unused gl/glfw) |
| R2 | Pixel model | Half-block `▀`, 2 px/cell, truecolor |
| R3 | World render | Per-pixel DDA voxel raycast; perpendicular (camera-Z) depth |
| R4 | Architecture | Hybrid: shared color+depth framebuffer, passes World→Entities→Selection→HUD→Present |
| R5 | Loop | Fixed 30 FPS ticker, input goroutine, time-based movement, synchronized output (2026) |
| R6 | Memory/GC | Flat preallocated buffers, flat `[]uint8` world, zero per-frame allocation, bounded footprint |
| R7 | kitty | Truecolor via COLORTERM auto-detect; graceful degradation; clean terminal restore |

**All NEEDS CLARIFICATION resolved.** No open questions remain for design.
