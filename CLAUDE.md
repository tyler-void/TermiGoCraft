<!-- SPECKIT START -->
## Active feature: 001-first-person-voxel-world

Turning the single-file terminal 3D wireframe demo into a first-person, Minecraft-style
voxel world (explore + break/place blocks + entities/HUD), scoped to one sprint.

**Plan & design docs** (read these for technologies, structure, and conventions):
- `specs/001-first-person-voxel-world/plan.md` — implementation plan, tech context, package layout
- `specs/001-first-person-voxel-world/research.md` — key decisions (tcell/v2, DDA raycast, hybrid render, zero-alloc memory)
- `specs/001-first-person-voxel-world/data-model.md` — World/Camera/Framebuffer/Entity types
- `specs/001-first-person-voxel-world/contracts/` — controls + rendering-pipeline contracts
- `specs/001-first-person-voxel-world/quickstart.md` — build/run/verify

**Key technical guardrails**:
- Terminal I/O: `github.com/gdamore/tcell/v2` (truecolor). Drop termbox-go and unused go-gl/glfw.
- Render: per-pixel DDA voxel raycaster for the world + filled-triangle rasterizer for entities,
  both writing one shared color + **perpendicular camera-Z depth** framebuffer; half-block (`▀`) = 2 px/cell.
- Performance: **zero heap allocation in the steady-state frame**; flat preallocated buffers; flat `[]uint8` world;
  fixed ~30 FPS loop; input on its own goroutine. Test target terminal: **kitty**.
<!-- SPECKIT END -->
