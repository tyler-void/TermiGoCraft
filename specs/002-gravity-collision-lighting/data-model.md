# Phase 1 Data Model: Gravity, Collision & Ray-Traced Lighting

**Feature**: `002-gravity-collision-lighting` | **Date**: 2026-06-15
New/changed data structures only; everything from feature 001 (World, Block, Camera, Hit, Buffer) is reused as-is unless noted. Go shapes are indicative.

---

## Entity: Body (player physics) — NEW (`internal/physics`)

The player's physical presence: an AABB integrated under gravity.

| Field | Type | Notes |
|-------|------|-------|
| `Feet` | `geom.Vec3` | bottom-center of the box (X,Z centered; Y = box bottom) |
| `Vel` | `geom.Vec3` | velocity in cells/second |
| `Half` | float64 | half-width/depth of the box (≈0.3) |
| `Height` | float64 | total body height (≈1.8) |
| `EyeHeight` | float64 | eye offset above feet (≈1.6) |
| `Grounded` | bool | true when resting on a solid surface this tick |
| `Mode` | `Mode` | `Walk` or `Fly` |

**Derived**: `Eye() geom.Vec3 = Feet + {0, EyeHeight, 0}` — fed into `camera.Pos` each tick.

**Constants** (tunable): `Gravity` (cells/s², e.g. −28), `JumpSpeed` (e.g. +8.5 ⇒ ~1.25-block jump), `WalkSpeed` (e.g. 4.5 cells/s), `MaxFall` (terminal velocity clamp), `MoveDecay` (horizontal velocity decay per tick).

**State transitions** (per `Step`, fixed `dt`):
- `Walk`: `Vel.Y += Gravity*dt` (clamp to `MaxFall`); apply horizontal intent; resolve collision per axis (R2); set `Grounded`.
- `Fly`: no gravity/collision; `Vel` follows intent (incl. vertical) and decays; `Grounded=false`.
- Jump: if `Walk && Grounded && jumpRequested` → `Vel.Y = JumpSpeed`, `Grounded=false`.

**Invariants**:
- After a `Step`, the body **never overlaps** a solid cell in `Walk` mode (collision guarantees it).
- `Feet` stays within world horizontal bounds and at/above the world floor (FR-007).
- A body that *starts* a tick overlapping solids is pushed to free space before integration (FR-008).

---

## Entity: Mode — NEW

```
type Mode uint8   // Walk (default), Fly
```
Toggled by the player (FR-010). Walking is the launch default.

---

## Entity: Intent — NEW (transient, per tick)

What the input layer asks physics to do this tick; consumed and not stored.

| Field | Type | Notes |
|-------|------|-------|
| `Forward`,`Strafe` | float64 | −1..1 desired horizontal move (camera-relative) |
| `Up`,`Down` | bool | fly-mode vertical |
| `Jump` | bool | requested this tick (walk mode) |

`game` builds the Intent from drained key events (with horizontal decay so held/auto-repeat keys sustain motion — R5), then calls `physics.Step(world, body, intent, camYaw, dt)`.

---

## Entity: Sun / Light — NEW (`internal/render`)

The single directional light.

| Field | Type | Notes |
|-------|------|-------|
| `Dir` | `geom.Vec3` | unit vector pointing **toward** the sun (fixed, e.g. normalize{0.4, 0.85, 0.3}) |
| `Intensity` | float64 | diffuse strength (≈0.7) |
| `Ambient` | float64 | base light for unlit/back faces (≈0.3) so nothing is pure black |

**Used by** `lightAt(world, hitPoint, face) float64` → a brightness multiplier combining:
`ambient + (shadowed ? 0 : intensity*max(0, faceNormal·Dir))`, then `× aoFactor`.

**Tunables**: `ShadowDist` (max shadow-ray length, ≈24 cells), `ShadowBias` (offset along normal to avoid self-shadow acne), `AOmin` (darkest AO multiplier, ≈0.55).

---

## Entity: Ambient-occlusion sample — derived (no stored state)

Per hit face, AO is computed on the fly (R3), not stored:
- For each of the face's **4 corners**, read the **3 neighbor cells** in the plane just outside the face → `vertexAO(side1,side2,corner) = (side1&&side2)?0 : 3-(side1+side2+corner)` ∈ {0,1,2,3}.
- Map level→multiplier (e.g. `AOmin + (1-AOmin)*level/3`).
- **Bilinearly interpolate** the 4 corner multipliers using the hit point's in-face UV → smooth per-pixel AO.

---

## Entity: Aiming cursor state — extends HUD (feature 001)

| State | When | Shown |
|-------|------|-------|
| Neutral | `target.OK == false` | dim/thin reticle |
| Targeting | `target.OK == true` (block in reach) | bright/distinct reticle (e.g. brackets or bold color) |

Plus a status line: current `Mode` (Walk/Fly), lighting on/off, and integer player position.

---

## Changed: Camera (feature 001) — usage only

`camera.Camera` is **unchanged structurally**. Each tick `game` sets `cam.Pos = body.Eye()`. Look (yaw/pitch) still comes from input directly. The camera no longer "owns" movement — physics moves the body, camera follows.

---

## Relationships

```text
Game
 ├── Body (physics)   ──Step(world,intent,dt)──► gravity + per-axis collision ──► Feet/Vel/Grounded
 │        └── Eye() ──► Camera.Pos (each tick)
 ├── Camera           (yaw/pitch from input; Pos from Body)
 ├── Sun (render)
 └── render.RenderWorld(buf, world, cam, sun, lightingOn)
            └── per hit: base = palette[block]
                         light = lightAt(world, hitPoint, face, sun)   // n·L + shadow ray + AO
                         TestSet(x,y, depth, shade(base, light))
     render.RenderHUD(scr, target.OK, mode, pos, lightingOn)           // 2-state cursor + status
```

Dependency direction (acyclic): `geom` ← `world`,`camera`,`physics`,`render` ← `game`. `physics` depends on `geom`+`world` only (no render/tcell) → headlessly testable. `render/lighting` depends on `geom`+`world`+`camera`.

---

## Memory budget (still constant — project norm)

No new per-frame allocations: `Body`, `Intent`, `Sun` are tiny fixed structs; the shadow ray reuses `world.Cast` (returns a value `Hit`); AO uses ints and a handful of `world.At` reads. Footprint over a session stays flat (no growth) — preserving feature 001's no-bleed guarantee.
