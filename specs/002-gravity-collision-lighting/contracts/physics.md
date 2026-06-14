# Contract: Physics (player body, gravity, collision)

**Feature**: `002-gravity-collision-lighting` | **Date**: 2026-06-15
Internal API between `game` and the new `internal/physics` package. Signatures illustrative.

## Types

```go
type Mode uint8
const (Walk Mode = iota; Fly)

type Body struct {
    Feet      geom.Vec3 // bottom-center
    Vel       geom.Vec3
    Half      float64   // half width/depth (~0.3)
    Height    float64   // (~1.8)
    EyeHeight float64   // (~1.6)
    Grounded  bool
    Mode      Mode
}
func (b Body) Eye() geom.Vec3 // Feet + {0,EyeHeight,0}

type Intent struct {
    Forward, Strafe float64 // -1..1, camera-relative (yaw)
    Up, Down, Jump  bool
}
```

## Core entry point

```go
func Step(w *world.World, b *Body, in Intent, yaw, dt float64)
```

**Guarantees**:
- **P-1 (gravity)**: in `Walk`, `Vel.Y` accelerates downward by `Gravity*dt` each call (clamped to `MaxFall`); the body falls when unsupported (FR-002) and rests on the first solid surface, with `Grounded=true` and `Vel.Y=0` on landing (FR-003).
- **P-2 (no clipping)**: after `Step`, in `Walk` the body's AABB overlaps **no** solid cell. Motion into a solid is stopped on that axis only (velocity zeroed on it), so remaining axes still move → wall sliding (FR-004). Axes resolved independently in order Y, X, Z.
- **P-3 (jump)**: `in.Jump` causes a jump **only** when `Walk && Grounded`; sets `Vel.Y=JumpSpeed`, clears `Grounded`. Mid-air jump is ignored (FR-005).
- **P-4 (head bump)**: upward motion that meets a solid above zeroes `Vel.Y` and stops (FR-006).
- **P-5 (bounds)**: `Feet` is clamped to the world's horizontal extent and never descends below the world floor; the body cannot leave the world or fall forever (FR-007).
- **P-6 (unstick)**: if the body begins a `Step` overlapping solids, it is lifted to the nearest free space before integrating; the player is never trapped/suffocated (FR-008).
- **P-7 (fly)**: in `Fly`, gravity and collision are skipped; `Vel` follows `Intent` (including `Up`/`Down`) and decays; `Grounded=false` (FR-010).
- **P-8 (determinism)**: `Step` is pure given `(w, b, in, yaw, dt)` and allocation-free — no RNG, no `make`.

## Helper contract

```go
func Overlaps(w *world.World, feet geom.Vec3, half, height float64) bool
```
True iff any solid cell intersects the AABB. Iterates the integer cell range `[floor(min)..floor(max)]` on each axis and checks `w.Solid`. Allocation-free.

## Placement guard (consumed by game/world edit)

The body exposes its occupied cell range so `world.Place` can reject a block that would intersect **any** part of the body (FR-009) — extending 001's single-cell self-placement guard.

## Verification hooks (headless, no terminal)

- Drop test: body above ground falls and `Grounded` becomes true at the surface; never below it.
- Wall slide: moving diagonally into an axis-aligned wall keeps the parallel component.
- Jump: from grounded, `Vel.Y` becomes positive once; a second jump while airborne is a no-op.
- Head bump: jumping under a 1-gap ceiling zeroes upward velocity, no pass-through.
- Unstick: a body initialized inside stone is relocated to free space by the first `Step`.
- Bounds: walking toward the world edge clamps; the body never leaves the grid.
