# Contract: Lighting (directional sun, hard shadows, voxel AO)

**Feature**: `002-gravity-collision-lighting` | **Date**: 2026-06-15
Internal API in `internal/render` (new `lighting.go`), called from the world raycast pass.

## Types

```go
type Sun struct {
    Dir       geom.Vec3 // unit, points TOWARD the sun (fixed)
    Intensity float64   // diffuse weight (~0.7)
    Ambient   float64   // base light (~0.3), never fully black
}
```

Tunables: `ShadowDist` (~24), `ShadowBias` (~1e-3..1e-2 along normal), `AOmin` (~0.55).

## Core function

```go
func lightAt(w *world.World, hit world.Hit, hitPoint geom.Vec3, sun Sun) float64
```

Returns a brightness multiplier in roughly `[AOmin*Ambient, 1]`, computed as:

```
n          = hit.Face normal
diffuse    = max(0, n · sun.Dir)
shadowed   = world.Cast(hitPoint + n*ShadowBias, sun.Dir, ShadowDist).OK
direct     = shadowed ? 0 : sun.Intensity * diffuse
ao         = aoFactor(w, hit, hitPoint)        // bilinear over 4 corner AO values
return (sun.Ambient + direct) * ao
```

**Guarantees**:
- **L-1 (directional)**: faces with `n·Dir > 0` are brighter than faces with `n·Dir ≤ 0`; back faces fall back to `Ambient*ao` (FR-014).
- **L-2 (hard shadows)**: exactly **one** shadow ray per lit pixel, reusing `world.Cast`; if a solid lies between the point and the sun within `ShadowDist`, `direct=0` (FR-015). `ShadowBias` prevents self-shadow acne.
- **L-3 (ambient occlusion)**: `aoFactor` uses the 0fps rule `vertexAO(side1,side2,corner)=(side1&&side2)?0:3-(side1+side2+corner)` at the face's 4 corners (3 neighbor reads each), mapped `0..3 → AOmin..1`, then **bilinearly interpolated** by the hit point's in-face UV → smooth per-pixel darkening in crevices (FR-016).
- **L-4 (live updates)**: lighting reads the live world each frame (no bake), so break/place changes shadows/AO next frame (FR-017).
- **L-5 (deterministic / stable)**: no random sampling; identical inputs → identical output, so a static scene does not shimmer (FR-018, SC-010) and the dirty-cell `Present` stays effective.
- **L-6 (zero alloc)**: value math + `world.At` int reads + a value-returning `Cast`; no heap allocation (project norm).

## Integration into the world pass

```go
func RenderWorld(b *Buffer, w *world.World, cam *camera.Camera, sun Sun, lit bool)
```
At each primary hit: `hitPoint = cam.Pos + dir*hit.Dist`; if `lit`, `f = lightAt(...)` else `f = faceFactor(hit.Face)*distFactor(...)` (the 001 fallback). Then `TestSet(x,y, perpDepth, shade(BlockColor(hit.Block), f))`. The `lit` flag is the lighting on/off toggle (perf fallback, R4).

## HUD contract delta (cursor + status)

```go
func RenderHUD(scr tcell.Screen, targeted bool, mode physics.Mode, pos geom.Vec3, lit bool, showHelp bool)
```
- Center cursor has two states: neutral (dim) vs targeting (bright/distinct) keyed by `targeted` (FR-011/012).
- Status line shows mode (Walk/Fly), lighting on/off, and integer position (FR-013).

## Verification hooks (headless)

- Shadow: a tall pillar beside flat ground → ground cells on the anti-sun side return a shadowed (lower) light than open cells.
- Directional: a top face (`n=+Y`) under a sun above is brighter than a `-Y` face.
- AO: an inner corner (two adjacent neighbors solid) returns a lower multiplier than an open face.
- Determinism: `lightAt` over identical inputs is bit-identical across calls.
- Allocation: `BenchmarkFrame` over `Clear`+`RenderWorld(lit=true)` stays `0 allocs/op`.
