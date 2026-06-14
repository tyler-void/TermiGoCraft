# Contract: Rendering Pipeline (Internal Interfaces)

**Feature**: `001-first-person-voxel-world` | **Date**: 2026-06-15

The internal "API surface" between the game and the renderer. These are the boundaries each implementation task must honor so the hybrid raycaster + rasterizer compose correctly and stay allocation-free. Signatures are illustrative Go.

## Core types

```go
type RGB struct{ R, G, B uint8 }

type Face uint8 // PosX, NegX, PosY, NegY, PosZ, NegZ

type Buffer struct {
    W, H  int       // pixel dims = cols × (2·rows)
    Color []RGB     // len W*H, index y*W+x
    Depth []float32 // len W*H, perpendicular (camera-Z) distance; +Inf = empty
}
```

## Framebuffer contract

```go
func NewBuffer(cols, rows int) *Buffer
func (b *Buffer) Resize(cols, rows int)            // ONLY place buffers (re)allocate
func (b *Buffer) Clear(sky RGB)                    // Color=sky, Depth=+Inf; no allocation
func (b *Buffer) TestSet(x, y int, z float32, c RGB) // write iff z < Depth[idx]; the shared depth test
func (b *Buffer) Present(scr tcell.Screen)         // pack '▀' pairs (fg=top,bg=bottom) + scr.Show()
```

**Guarantees**:
- **G-1 (single depth space)**: every 3D pass writes `Depth` in the **same** unit — perpendicular camera-Z distance. This is what makes cross-pass occlusion correct (FR-012, SC-005).
- **G-2 (zero per-frame alloc)**: `Clear`, `TestSet`, `Present`, and all passes perform **no** heap allocation. Only `Resize` allocates. Enforced by `0 allocs/op` benchmark (R6).
- **G-3 (bounds-safe)**: `TestSet` ignores out-of-range `(x,y)` (clipping) rather than panicking.
- **G-4 (present mapping)**: cell `(cx, cy)` shows `▀` with `fg = Color[(2cy)·W+cx]` (top) and `bg = Color[(2cy+1)·W+cx]` (bottom).

## Pass A — World raycaster

```go
func RenderWorld(b *Buffer, w *world.World, cam *camera.Camera)
```
- For each pixel `(x,y)`: `dir = cam.ScreenRay(x,y,b.W,b.H)`; `hit = w.Cast(cam.Pos, dir, viewDist)`.
- On hit: `z = perpendicular(hit.Dist, dir, cam.Forward)`; `c = shade(palette[hit.Type].Base, faceFactor(hit.Face), distFactor(z))`; `b.TestSet(x,y,z,c)`.
- On miss: leave sky (already cleared).
- **Contract**: must use perpendicular distance (G-1); must not allocate (G-2); reads world via `At`, never raw indexing past bounds.

## Pass B/C — Triangle rasterizer (entities, held item, selection)

```go
func RenderMesh(b *Buffer, mesh []geom.Triangle, modelPos geom.Vector3, color RGB, cam *camera.Camera)
func RenderHeld(b *Buffer, mesh []geom.Triangle, color RGB, cam *camera.Camera) // view-space
func RenderSelection(b *Buffer, target world.Target, cam *camera.Camera)        // highlight cube edges/face
```
- Project vertices → screen pixels + per-vertex camera-Z; backface-cull; **filled** rasterization (barycentric or scanline) with per-pixel interpolated depth; write via `TestSet` (depth-tested against Pass A — G-1).
- **Contract**: filled, not wireframe (FR-004 spirit for entities); depth-tested so terrain occludes entities and vice versa (FR-012); no allocation (preallocate any scratch in the renderer struct, reuse per frame).

## Pass D — HUD overlay

```go
func RenderHUD(scr tcell.Screen, st HUDState)
```
- Drawn to tcell cells **after** `Present`, no depth ⇒ always on top (FR-013).
- Includes crosshair + selected-block indicator (+ help when toggled).

## Frame contract (order is normative)

```go
buf.Clear(sky)
RenderWorld(buf, world, cam)                 // A
for _, e := range entities { RenderMesh(...) } // B
RenderHeld(buf, heldMesh, heldColor, cam)    // B (view space)
RenderSelection(buf, target, cam)            // C
buf.Present(screen)                          // pack + Show
RenderHUD(screen, hud)                       // D (on top)
screen.Show()                                // flush HUD (or single Show after D)
```

**Verification hooks**:
- A headless `Buffer` (no tcell) lets passes be unit-tested: assert a known block fills expected pixels at expected depth; assert an entity behind a block is **not** written (occlusion); assert an entity in front **is** (SC-005).
- Benchmark `BenchmarkFrame` over A–C asserts `0 allocs/op` (G-2, R6).
