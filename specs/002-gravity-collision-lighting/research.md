# Phase 0 Research: Gravity, Collision & Ray-Traced Lighting

**Feature**: `002-gravity-collision-lighting`
**Date**: 2026-06-15
**User directive**: *"research ways to implement the specs, study from the best."*

This builds on feature 001 (the first-person voxel raycaster). Each section: **Decision → Rationale → Alternatives → Source(s) studied**.

---

## R1. Player physics: AABB body, fixed-timestep integration

**Decision**: Give the player an **axis-aligned bounding box** (AABB) — half-width ~0.3, height ~1.8 cells, eye at ~1.6 — tracked by a **feet-centered position + velocity + grounded flag + movement mode**. Integrate physics on the **existing fixed 30 FPS tick** (constant `dt`): every tick apply gravity to `vel.Y`, then resolve movement against the world. The camera eye is derived each tick as `feet + eyeHeight`.

**Rationale**:
- This is exactly how Minecraft-like games model the player: a moving AABB through a voxel field, integrated at a fixed step so behavior is deterministic and frame-rate independent.
- A **fixed timestep** is essential for stable physics — variable `dt` makes gravity/jumps inconsistent and can cause tunneling. We already run a fixed 30 FPS ticker (feature 001), so physics slots straight in.
- Keeping the body separate from the camera (camera stays pure geometry) preserves the clean 001 layering; `game` syncs `cam.Pos` from the body.

**Alternatives considered**:
- **Point camera + "is the cell solid" check** (current 001 behavior): no body, so you clip and can stand in walls. Rejected — that's the bug we're fixing.
- **Full rigid-body / physics engine**: massive overkill for an arcade voxel game; more allocation and tuning. Rejected (YAGNI).
- **Per-render-frame variable-dt integration**: non-deterministic, tunneling risk. Rejected in favor of fixed step.

**Sources**: [Gaffer "Fix Your Timestep"](https://gafferongames.com/post/fix_your_timestep/) (fixed-step rationale); [GameDev.net: collision on Minecraft voxel landscapes](https://www.gamedev.net/forums/topic/641699-collision-detection-on-minecraft-type-landscape-ie-voxels/5052610/).

---

## R2. Collision resolution: per-axis swept move against solid voxels

**Decision**: Resolve movement **one axis at a time** (Y, then X, then Z). For each axis: tentatively advance the AABB along that axis only, gather the integer voxel range the box now overlaps (broadphase), and if any of those cells is solid, **clamp the box flush against the blocking cell and zero the velocity on that axis**. Resolving axes independently is what lets the player **slide along walls** and prevents corner-snag ambiguity (Notch's "never let the boxes intersect in the first place"). After the Y step, if downward motion was stopped, set **grounded = true**; if upward motion was stopped, just zero `vel.Y` (head bump).

**Rationale**:
- Per-axis resolution is the standard, robust voxel-collision technique: simple, allocation-free (we iterate an integer range and read `world.Solid`), and it naturally yields wall-sliding (FR-004), ground rest (FR-003), and ceiling stop (FR-006).
- At our speeds (≤ ~0.5 cell/tick) and a fixed 30 FPS step, a single clamped step per axis cannot tunnel through 1-cell-thick terrain, so we don't need a full swept-ray broadphase. (A swept-AABB-vs-voxels raycast is the more accurate upgrade if speeds ever rise — noted below.)
- **Overlap recovery** (FR-008): if the body starts a tick already intersecting solids (spawn/edit), push it upward to the nearest free space before integrating, so the player is never trapped/suffocated.
- **Bounds** (FR-007): clamp the feet position to the world's horizontal extent and rest on the solid world floor, so the player can neither walk off the edge of the world nor fall forever.

**Alternatives considered**:
- **Swept AABB raycast along the leading corner** (fenomas `voxel-aabb-sweep`): more accurate for large per-tick moves, but more code; unnecessary at our step size. Kept as a documented future upgrade.
- **Resolve all axes simultaneously from a single penetration vector (MTV)**: order-dependent corner cases; the per-axis method sidesteps them. Rejected.
- **Continuous collision via Minkowski difference**: elegant but heavier; not needed here.

**Sources**: [fenomas/voxel-aabb-sweep](https://github.com/fenomas/voxel-aabb-sweep); [Andre Blunt — 3D AABB collision & resolution for voxel games](https://medium.com/@andrebluntindie/3d-aabb-collision-detection-and-resolution-for-voxel-games-5fcbfdb8cdb4); [Luis Reis — AABB collision handling](https://luisreis.net/blog/aabb_collision_handling/).

---

## R3. Lighting model: directional sun + hard shadow ray + voxel AO

**Decision**: At each **primary-ray hit** in the world pass, compute a light factor and multiply the block's base color by it:

```
light = ambient + (inShadow ? 0 : sunIntensity * max(0, faceNormal · sunDir))   // directional + hard shadow
final = shade(baseColor, light) * aoFactor                                       // × ambient occlusion
```

Three parts, all reusing machinery we already have:

1. **Directional ("sun") shading** — `max(0, n·L)` where `n` is the hit face normal (already known from `Face`) and `L` is a fixed sun direction. Replaces 001's flat per-face `faceFactor` with physically-meaningful directional light. Free (a dot product).
2. **Hard shadows** — from the hit point, offset slightly along `n`, fire **one DDA ray toward the sun** (reusing `world.Cast`); if it hits any solid before `shadowDist`, the point is shadowed. One extra ray per hit pixel.
3. **Ambient occlusion** — adapt the canonical 0fps voxel-AO rule. For the hit face, compute AO at its **4 corners** with `vertexAO(side1, side2, corner) = (side1 && side2) ? 0 : 3 - (side1+side2+corner)` (each term = is-that-neighbor-solid, sampled from the 3 voxels around the corner in the plane just outside the face), map 0..3 → a darkening factor, and **bilinearly interpolate** across the face using the hit point's in-face UV. Pure array reads, no rays.

**Rationale**:
- This is the **direct-lighting** subset the spec scoped to (sun + hard shadows + AO) — the highest visual impact achievable inside a terminal frame budget.
- Everything reuses 001: the normal comes from `Face`, the shadow ray is the same `Cast`, and AO is cheap neighbor lookups into the flat world. So it's **allocation-free** and integrates as a small addition to `RenderWorld`.
- It's **deterministic** — no stochastic sampling — so a static scene produces identical frames: no shimmer/flicker (FR-018, SC-010), and the dirty-cell `Present` optimization keeps working.
- Per-pixel AO (bilerp of the 4 corner values by the exact hit UV) gives **smooth** gradients rather than the per-face blockiness, because the raycaster gives us the exact hit point for free (the mesh-based 0fps version only gets per-vertex AO).

**Alternatives considered**:
- **Baked/flood-fill light propagation** (0fps "Voxel lighting"): great for large static worlds and indirect light, but needs a light volume that must be recomputed on edits and stored — more memory and complexity, and overkill for direct sun. Rejected for this feature (revisit if we want torches/indirect light).
- **Soft shadows / many shadow rays / path tracing**: too expensive for real-time terminal rendering and would add noise (shimmer). Explicitly out of scope.
- **Screen-space AO**: needs a depth prepass and convolution; we already have true geometry, so object-space voxel AO is simpler and exact.

**Sources**: [0fps — Ambient occlusion for Minecraft-like worlds](https://0fps.net/2013/07/03/ambient-occlusion-for-minecraft-like-worlds/) (the `vertexAO` rule + quad-flip anisotropy note); [0fps — Voxel lighting](https://0fps.net/2018/02/21/voxel-lighting/) (flood-fill alternative); [Deferred voxel shading — shadow rays toward the light](https://jose-villegas.github.io/post/deferred_voxel_shading/).

---

## R4. Performance & the frame budget

**Decision**: Keep the whole lit frame inside the 33 ms (30 FPS) budget and **zero per-frame allocation**. Budget plan:
- Primary rays (001): ~2.6 ms at 240×80.
- + One shadow ray per hit (capped at `shadowDist`, ~24 cells): roughly doubles ray work → ~5–6 ms.
- + AO: ~8 neighbor block reads per hit, cheap → sub-ms.
- Provide a **lighting on/off toggle** (key) as a guaranteed fallback and an A/B for the player; with lighting off the renderer behaves like 001.
- Keep everything value-typed (the shadow `Cast` returns a `Hit` by value; AO uses ints) so the render path stays **0 allocs/op** — the existing `BenchmarkFrame` guard is extended to the lit path.

**Rationale**: kitty is GPU-accelerated and fast; the measured 001 headroom (~620 fps-equivalent for the world pass) comfortably absorbs a second ray per pixel. Determinism keeps the dirty-cell present effective, so a still scene still costs ~nothing to display.

**Alternatives considered**: render lighting at half resolution / every other frame — adds shimmer and complexity; not needed given the headroom. Rejected.

---

## R5. Input → physics bridge, and QoL

**Decision**:
- **Physics ticks every frame** (gravity always integrates), independent of input. Because terminals report key *repeats* but no key-up, horizontal movement uses a **target-velocity-with-decay** model: a movement keypress sets the horizontal velocity toward walk speed in that direction; each tick it decays toward zero. Held keys (auto-repeat) sustain motion; it stops shortly after release. Gravity/jump live in `vel.Y`.
- **Walk vs fly toggle** (FR-010): a key flips `mode`. In **walk** mode gravity + collision apply and `Space` = **jump** (only when grounded). In **fly** mode gravity/collision are off and `Space`/`C` = up/down (the 001 behavior). Walking is the launch default.
- **Aiming cursor** (FR-011/012): the HUD crosshair gains two states — neutral when nothing is targeted, and a distinct "can interact" glyph/color when `target.OK` within reach. Trivial, driven by the per-tick center-ray cast we already do.
- **Status readout** (FR-013): mode + integer position drawn in the HUD.

**Rationale**: This reconciles continuous physics with discrete terminal input in the simplest way that still feels good, and folds the QoL items onto existing per-tick state (`target`, `mode`, position) so they're nearly free.

**Alternatives considered**: raw keyboard-protocol key-up events (kitty supports a keyboard protocol with release events) — would enable true hold-to-move, but it's terminal-specific and beyond scope; the decay model is portable. Noted as a future enhancement.

**Sources**: [Gaffer "Fix Your Timestep"](https://gafferongames.com/post/fix_your_timestep/) (decouple sim from frame/input).

---

## Summary of decisions

| # | Area | Decision |
|---|------|----------|
| R1 | Player physics | AABB body (feet pos + velocity + grounded + mode), fixed 30 FPS integration |
| R2 | Collision | Per-axis swept move vs `world.Solid`; clamp + zero-velocity; slide; overlap recovery; world-bounds clamp |
| R3 | Lighting | Directional sun `n·L` + one hard shadow ray (reuse `Cast`) + bilerp voxel AO (0fps rule); deterministic |
| R4 | Performance | ~5–6 ms lit frame, 0 allocs/op, lighting on/off toggle fallback |
| R5 | Input/QoL | Per-tick physics + target-velocity decay; walk/fly toggle (Space=jump/up); 2-state cursor; mode+pos HUD |

**All open questions resolved.** Ray-tracing scope is the spec's documented subset (direct light + hard shadows + AO; no reflections/GI/soft shadows/day-night).
