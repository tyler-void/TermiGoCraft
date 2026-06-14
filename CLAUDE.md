<!-- SPECKIT START -->
## Active feature: 002-gravity-collision-lighting

Add gravity + AABB voxel collision + jump (walk/fly toggle), a 2-state aiming cursor &
mode/position HUD, and ray-traced lighting (directional sun + hard cast shadows + voxel
ambient occlusion). Builds on feature 001.

**Plan & design docs** (002):
- `specs/002-gravity-collision-lighting/plan.md` — plan, tech context, package changes
- `specs/002-gravity-collision-lighting/research.md` — decisions (AABB+fixed-step physics, per-axis
  swept collision, sun + shadow ray + 0fps voxel AO); sources studied
- `specs/002-gravity-collision-lighting/data-model.md` — Body/Mode/Intent/Sun types
- `specs/002-gravity-collision-lighting/contracts/` — physics, lighting, controls contracts
- `specs/002-gravity-collision-lighting/quickstart.md` — build/run/verify

**Key technical guardrails (002)**:
- New `internal/physics`: AABB `Body` (feet pos + vel + grounded + Walk/Fly mode); `Step` integrates
  gravity on the fixed 30 FPS tick + per-axis collision vs `world.Solid` (clamp + zero-vel → wall slide);
  jump only when grounded; unstick + world-bounds clamp. Camera eye = `Body.Eye()` each tick.
- New `render/lighting.go`: directional sun `n·L` + **one** hard shadow ray (reuse `world.Cast`) +
  voxel AO (`vertexAO(s1,s2,c)=(s1&&s2)?0:3-(s1+s2+c)`, bilerp by hit UV). Deterministic (no shimmer);
  lighting on/off toggle (key L) as perf fallback.
- Preserve the **zero-allocation lit frame** (shadow Cast returns a value; AO uses ints) — extend
  `BenchmarkFrame` to lit path. ≥30 FPS budget (~5-6ms). Test terminal: **kitty**.

**Feature 001 (shipped, foundation)**: tcell/v2 truecolor half-block (`▀`) renderer; per-pixel DDA
voxel raycaster + filled-triangle rasterizer sharing one color+perpendicular-depth framebuffer; flat
`[]uint8` world; input goroutine. See `specs/001-first-person-voxel-world/`.
<!-- SPECKIT END -->
