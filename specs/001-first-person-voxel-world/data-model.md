# Phase 1 Data Model: First-Person Voxel World

**Feature**: `001-first-person-voxel-world` | **Date**: 2026-06-15
Derives the concrete data structures from the spec's Key Entities and the research decisions (flat, pointer-free, preallocated — R6). Types are described language-neutrally with the intended Go shape noted; field-level detail is for implementation.

---

## Entity: World

The finite 3D voxel grid (spec "World").

| Field | Type | Notes |
|-------|------|-------|
| `SX, SY, SZ` | int | Dimensions in cells (X=width, Y=height/up, Z=depth). Default 64×32×64. |
| `Blocks` | `[]uint8` | **Flat** array, length `SX*SY*SZ`, indexed `x + SX*(y + SY*z)`. Each byte is a `BlockType`. Pointer-free ⇒ GC scans nothing (R6). |

**Operations**:
- `InBounds(x,y,z) bool` — guards all access.
- `At(x,y,z) BlockType` — returns `Air` when out of bounds (so rays/edits past the edge are safe, not errors — Edge Cases).
- `Set(x,y,z, t)` — mutates in place; no allocation, never grows (no memory bleed).

**Invariants**:
- `len(Blocks) == SX*SY*SZ` for the whole program lifetime (allocated once at gen).
- Indices always validated by `InBounds` before raw slice access.

**Validation rules (from requirements)**:
- FR-001: every cell is exactly one `BlockType` (Air or a solid type).
- FR-002: generation produces a walkable ground layer + ≥ 2 solid types.

---

## Entity: BlockType + Palette

A category of material (spec "Block Type"). Implemented as a small integer enum into a static palette.

| Constant | Value | Solid? | Role |
|----------|-------|--------|------|
| `Air` | 0 | no | empty cell; rays pass through |
| `Grass` | 1 | yes | ground top |
| `Dirt` | 2 | yes | sub-surface / placeable |
| `Stone` | 3 | yes | placeable |
| `Wood` | 4 | yes | placeable feature |
| … | … | … | extensible; ≥ 2 required (FR-002), ≥ 4 recommended for a usable hotbar (FR-011) |

**Palette entry** (static, indexed by `BlockType`):

| Field | Type | Notes |
|-------|------|-------|
| `Base` | `RGB` | base color of the block |
| `Solid` | bool | `Air` ⇒ false; everything else true (`IsSolid(t)`) |

Per-face appearance is **derived**, not stored: shading = `Base × faceFactor(face) × distanceFactor(dist)` (see `render/color.go`). This keeps the palette tiny and avoids 6 stored colors per type.

---

## Entity: Camera / Player Viewpoint

First-person camera + player attributes (spec "Player / Viewpoint"). One struct; no physics this sprint (R: free-look).

| Field | Type | Notes |
|-------|------|-------|
| `Pos` | `geom.Vector3` | eye position in world space |
| `Yaw` | float64 | horizontal angle (radians) |
| `Pitch` | float64 | vertical angle (radians), clamped to ±~1.5 to avoid flips (FR-006) |
| `FOV` | float64 | field of view (radians) for ray spread |
| `Reach` | float64 | max edit/target distance in cells (e.g. 5.0) (FR-007) |
| `Selected` | `BlockType` | block type to place (FR-011) |

**Derived (computed, not stored)**:
- `Direction()`, `Right()`, `Up()` — orthonormal basis from Yaw/Pitch (reuse current math).
- `ScreenRay(px, py, W, H) geom.Vector3` — ray direction through a pixel, from FOV + basis (drives the raycaster).

**State transitions**:
- Movement adds `Right/Direction × (speed·dt)` to `Pos` (time-based, R5); forward flattened to the ground plane for walking feel.
- Look adds `turnSpeed·dt` to Yaw/Pitch (Pitch clamped).
- `Selected` cycles through the solid palette on select keys.

---

## Entity: Target (transient, recomputed each tick)

The block currently aimed at (spec implies via FR-007). Not stored long-term; produced by a raycast from the camera each update.

| Field | Type | Notes |
|-------|------|-------|
| `Cell` | `(x,y,z int)` | the solid cell hit |
| `Face` | `Face` | which face was entered (`+X,-X,+Y,-Y,+Z,-Z`); identifies the empty neighbor for placement |
| `Dist` | float64 | perpendicular distance to hit |
| `Ok` | bool | false when nothing solid within `Reach` (sky / out of range) |

Produced by `world.Cast(origin, dir, maxDist)`. Drives selection highlight (Pass C), break (the cell), and place (the `Face`-adjacent cell).

---

## Entity: Block edits (operations, not state)

Realize FR-008/009/010.

- `Break(w, target)`: if `target.Ok`, `w.Set(target.Cell, Air)`. No-op if `!Ok`.
- `Place(w, target, t)`: if `target.Ok`, compute neighbor `n = target.Cell + faceNormal(target.Face)`; reject if `!InBounds(n)`, if `At(n)` is solid, or if `n` equals the player's occupied cell (self-placement guard, FR-010); else `w.Set(n, t)`.

**Validation rules**: all edits are no-ops (never panics) when there is no valid target or the placement is blocked (Edge Cases, FR-010).

---

## Entity: Entity (non-block scene object)

Things drawn by the rasterizer, not the raycaster (spec "Entity").

| Field | Type | Notes |
|-------|------|-------|
| `Pos` | `geom.Vector3` | world position |
| `Mesh` | `[]Triangle` | triangles (reuse existing cube mesh builder) |
| `Color` | `RGB` | fill color |

The **held item** is a special entity rendered in view space (fixed relative to the camera) rather than world space. Both go through the depth-tested triangle pass (Pass B) so terrain occludes them correctly (FR-012, SC-005).

---

## Entity: Framebuffer (render target — the "single depth representation")

The shared buffer both renderers write to (research R4/R6). Allocated once; resized only on terminal resize.

| Field | Type | Notes |
|-------|------|-------|
| `W, H` | int | pixel dimensions = `cols × (2·rows)` (half-block, R2) |
| `Color` | `[]RGB` | flat, length `W*H`, indexed `y*W+x` |
| `Depth` | `[]float32` | flat, length `W*H`; **perpendicular** (camera-Z) distance; `+Inf` = empty (R3) |

**Operations**:
- `Clear(sky RGB)` — set Color to sky, Depth to `+Inf` via range-loop (compiler memset; no alloc).
- `TestSet(x,y, z float32, c RGB)` — write iff `z < Depth[i]`; the depth test shared by all 3D passes.
- `Resize(cols, rows)` — the **only** place buffers are reallocated.
- `Present(screen)` — pack pixel pairs → `SetContent('▀', fg=Color[top], bg=Color[bottom])`; then `screen.Show()`.

**Invariants**:
- `len(Color) == len(Depth) == W*H` at all times.
- No allocation outside `Resize`.

---

## Entity: HUD / Overlay

Flat layer drawn last (spec "HUD / Overlay", FR-013).

| Element | Source | Notes |
|---------|--------|-------|
| Crosshair | center cell | always drawn on top |
| Selected-block indicator | `Camera.Selected` | name/color swatch of current placeable |
| Help line | static | controls list (FR-016); toggle or always-on footer |

Drawn directly to tcell cells after `Present` (no depth), so it always sits on top of the scene.

---

## Relationships

```text
Game
 ├── World          (Blocks []uint8, dims)        ── Cast()/Break()/Place() ──┐
 ├── Camera         (Pos,Yaw,Pitch,Reach,Selected) ── ScreenRay() ─────────────┤ produces Target each tick
 ├── []Entity       (+ held item)                                              │
 └── Renderer
      └── Framebuffer (Color[], Depth[])  ◄── Pass A raycast(World,Camera)
                                          ◄── Pass B raster(Entities,Camera)   (depth-tested)
                                          ◄── Pass C raster(Target highlight)
                                          ── Present ──► tcell.Screen ──► Pass D HUD on top
```

**Dependency direction** (matches package layout): `geom` ← `world`/`camera`/`entity`/`render` ← `game`. No cycles. Pure-logic packages (`geom`, `world`, `camera`) have **no** tcell dependency and are unit-testable headlessly.

---

## Memory budget (constant over a session — R6 / SC-006)

| Buffer | Size (default world, 120×40 terminal) |
|--------|----------------------------------------|
| `World.Blocks` (64×32×64 `uint8`) | ~128 KB |
| `Framebuffer.Color` (240×80 RGB) | ~58 KB |
| `Framebuffer.Depth` (240×80 float32) | ~77 KB |
| Entity meshes (handful of cubes) | < 10 KB |
| **Total** | **well under 1 MB, fixed** |

All allocated at startup / resize; the per-frame path allocates nothing. No structure grows during play ⇒ no leak/bleed.
