# Feature Specification: First-Person Voxel World

**Feature Branch**: `001-first-person-voxel-world`  
**Created**: 2026-06-15  
**Status**: Draft  
**Input**: User description: "build the spec based on this, all of these should be available in 1 sprint of this sdd, specify - plan - task - implement" — turn the existing terminal 3D wireframe demo into a first-person, Minecraft-style voxel world the player can explore and edit, where the solid blocky terrain and the non-block elements (entities, held item, HUD) are drawn into the same scene with correct depth between them.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Explore a solid voxel world in first person (Priority: P1)

A player launches the program and is placed *inside* a small blocky world, seeing it through their own eyes (first-person view). The terrain is made of solid, opaque blocks — a walkable ground surface with a few visually distinct block types. The player can move forward/back/strafe and turn and look up/down; the view updates smoothly and immediately as they do.

**Why this priority**: This is the foundation and the single biggest visible change from the current wireframe demo. On its own it already delivers a recognizable "I'm standing in a Minecraft-like world" experience and is a complete, demonstrable MVP. Every other story builds on top of it.

**Independent Test**: Launch the program; confirm the view is first-person, the world appears as solid blocks (not wireframe), and that movement and looking keys change the view believably and responsively. No editing or entities required.

**Acceptance Scenarios**:

1. **Given** the program has started, **When** the world finishes loading, **Then** the player sees a first-person view of a solid blocky landscape with at least two distinguishable block types and a clear ground surface.
2. **Given** the player is standing still, **When** they press a movement key, **Then** the viewpoint moves in the corresponding direction relative to where they are looking, and the view redraws within a fraction of a second.
3. **Given** the player is looking forward, **When** they press the look/turn keys, **Then** the camera rotates horizontally and vertically (with a sensible up/down limit) and the world appears to pan accordingly.
4. **Given** a block is closer to the player than another, **When** the scene is drawn, **Then** the nearer block correctly hides the part of the farther block behind it (no see-through or incorrect overlap).

---

### User Story 2 - Break and place blocks (Priority: P2)

The player looks at a block within arm's reach, sees it highlighted as the current target, and can break it (the block disappears and the space becomes open) or place a new block against it (a new block appears in the adjacent open space). The world visibly updates the instant the action happens.

**Why this priority**: Editing the world is the core Minecraft gameplay loop. It turns a viewer into a game. It depends on P1 (you must be able to see and aim first) but delivers the defining interaction.

**Independent Test**: Aim at a block, confirm it is highlighted, break it and confirm the gap appears, then place a block and confirm it appears in the expected adjacent cell — all reflected in the view immediately.

**Acceptance Scenarios**:

1. **Given** the player is looking at a block within reach, **When** the scene is drawn, **Then** that specific block is clearly indicated as the current target (highlight/outline).
2. **Given** a block is targeted, **When** the player triggers "break", **Then** that block is removed, the space becomes empty, and the view updates within one frame.
3. **Given** a block is targeted, **When** the player triggers "place", **Then** a new block of the currently selected type appears in the open cell adjacent to the targeted face, and the view updates within one frame.
4. **Given** the player is looking at open sky or beyond reach (no target), **When** they trigger break or place, **Then** nothing changes and no error occurs.
5. **Given** placing a block would occupy the player's own position, **When** the player triggers "place", **Then** the placement is rejected and nothing changes.
6. **Given** more than one block type is available, **When** the player switches the selected block type, **Then** subsequent placements use the newly selected type.

---

### User Story 3 - Entities, held item, and HUD over the world with correct depth (Priority: P3)

Beyond the blocky terrain, the player sees non-block visual elements drawn into the same scene: a simple in-world object (e.g., a marker/entity) and the item currently "held". These are correctly occluded by terrain — an object behind a wall is hidden, an object in front is drawn over the wall. On top of everything sits a flat heads-up display (a center crosshair and an indicator of the currently selected block type) plus access to the controls/help.

**Why this priority**: This is what makes the renderer a true hybrid and a real game UI rather than just a terrain viewer. It is valuable but the world (P1) and editing (P2) are usable without it, so it comes last in the sprint.

**Independent Test**: Place a non-block object in the world and walk so terrain passes between it and the player; confirm it is hidden when behind a block and visible when in front. Confirm the crosshair and selected-block indicator are always visible on top of the scene.

**Acceptance Scenarios**:

1. **Given** a non-block object sits behind a solid block relative to the player, **When** the scene is drawn, **Then** the object is hidden by the block.
2. **Given** the same object is in front of all nearby terrain, **When** the scene is drawn, **Then** the object is drawn over the terrain.
3. **Given** the game is running, **When** any frame is drawn, **Then** a center crosshair and the currently selected block type are shown on top of the 3D scene and remain legible.
4. **Given** the player wants to learn the controls, **When** they request help or read the on-screen hints, **Then** the movement, look, break/place, block-select, and quit controls are listed.

---

### Edge Cases

- **No target in view**: Looking at sky or past the reach limit shows no highlight; break/place are no-ops.
- **Target at maximum reach**: A block exactly at the edge of reach is still targetable; one cell beyond is not.
- **Self-placement**: A placement that would land on the player's own cell is rejected.
- **World boundary**: Looking toward or moving to the edge of the finite world renders empty space gracefully without out-of-bounds errors.
- **Very small terminal**: A terminal below a usable size renders whatever fits without crashing.
- **Terminal resize mid-session**: The view re-adapts to the new size without crashing or leaving visual corruption.
- **Rapid/held input**: Spamming or holding movement/edit keys keeps the game responsive without freezing or unbounded lag.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST represent the game world as a finite, bounded three-dimensional grid of cells, where each cell is either empty (air) or holds a solid block of a specific type.
- **FR-002**: System MUST generate a starting world at launch that includes a walkable ground surface and at least two visually distinguishable solid block types.
- **FR-003**: System MUST render the world from a first-person point of view, with the viewer positioned inside the world looking out.
- **FR-004**: System MUST draw solid blocks as filled, opaque surfaces (not wireframe outlines) so the world reads as solid terrain.
- **FR-005**: System MUST convey depth and surface orientation visually (e.g., shading/contrast by distance and by face) so the 3D structure and individual blocks are legible at a glance.
- **FR-006**: Players MUST be able to move the viewpoint through the world (forward, backward, and strafe left/right relative to facing) and turn/look horizontally and vertically, with the vertical look constrained to avoid disorienting flips.
- **FR-007**: System MUST determine the specific block the player is aiming at within a limited reach distance (the "targeted" block) and visibly indicate it.
- **FR-008**: Players MUST be able to break the targeted block, after which that cell becomes empty and the view reflects the change within one update cycle.
- **FR-009**: Players MUST be able to place a block of the currently selected type into the empty cell adjacent to the targeted face, with the view reflecting the change within one update cycle.
- **FR-010**: System MUST reject placements that would occupy the player's own position, and MUST treat break/place with no valid target as a no-op without error.
- **FR-011**: Players MUST be able to select which block type to place from the available set.
- **FR-012**: System MUST draw non-block visual elements (at least one in-world object/entity and a representation of the held item) within the same 3D scene, correctly occluded by nearer terrain and drawn over farther terrain.
- **FR-013**: System MUST display a heads-up overlay on top of the 3D scene that includes at minimum a center aiming crosshair and an indicator of the currently selected block type.
- **FR-014**: System MUST continuously update and redraw the scene over time (not solely in response to input) to support smooth motion and dynamic elements.
- **FR-015**: System MUST adapt the rendered view to the current terminal dimensions and continue functioning correctly when the terminal is resized.
- **FR-016**: System MUST provide a discoverable list of controls and a way to quit the program cleanly, restoring the terminal to a normal state on exit.

### Key Entities *(include if feature involves data)*

- **World**: The finite 3D grid of cells defining the playable space; knows its bounds and what block (if any) occupies each cell.
- **Block Type**: A category of solid material (including the special "empty/air" case) that determines a cell's appearance and whether it is solid.
- **Player / Viewpoint**: The first-person camera — a position and look direction within the world — together with the player's reach distance and currently selected block type.
- **Entity**: A non-block object positioned in the world that is drawn into the scene (e.g., a marker/creature and the held item).
- **HUD / Overlay**: The flat informational layer drawn on top of the 3D scene (crosshair, selected block indicator, controls/help).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A new viewer, on launch, can identify the ground, individual blocks, and relative distance within the first-person view within 5 seconds — the world reads as solid 3D terrain, not a wireframe or flat noise.
- **SC-002**: Movement and look input is reflected in the view within 50 ms of the keypress, so control feels immediate.
- **SC-003**: The scene redraws continuously at a smooth rate (target at least 20 updates per second) at the default world and a typical terminal size (around 120×40 cells).
- **SC-004**: Breaking or placing a block is reflected in the view within a single update cycle with no perceptible delay, and the targeted block is clearly indicated before every edit.
- **SC-005**: In 100% of observed cases, in-world objects and the held item are occluded correctly relative to terrain (hidden when behind a block, visible when in front).
- **SC-006**: The program runs for a continuous 10-minute session — including movement, looking, at least one terminal resize, and at least 50 break/place actions — without crashing or corrupting the display.
- **SC-007**: A first-time player, using only the on-screen controls/help, can move, look, break a block, and place a block within 2 minutes.
- **SC-008**: The starting world spans at least 32×32 cells horizontally so that exploration and editing feel non-trivial.
- **SC-009**: On quit, the terminal is restored to its normal usable state (no leftover raw-mode artifacts or hidden cursor).

## Assumptions

- This feature builds directly on the existing terminal 3D renderer already in the repository; the existing triangle/mesh drawing and camera math are reused for entities and overlays, while a new world-rendering path handles the voxel terrain. The two share a single depth representation so they compose with correct occlusion.
- **Single-player, local, single terminal**; no networking or multiplayer.
- **No persistence this sprint**: the world is generated fresh on each launch; saving/loading is out of scope.
- **Movement is free-look / fly-style**: there is no gravity, falling, or physics simulation, and walking-collision response is out of scope this sprint (the player may pass freely through space; the focus is rendering and editing).
- **No textures or external art assets**: blocks and faces are distinguished by color and shading characters only.
- **No advanced lighting**: distance- and face-based shading only; global illumination, ambient occlusion, and cast shadows are out of scope.
- **Finite, bounded world** generated at startup; infinite or streaming chunk generation is out of scope.
- **Keyboard-only input**: mouse-look is out of scope given terminal input limitations.
- **Target environment**: a color-capable terminal at a reasonable size (roughly 80×24 or larger).
- The entire feature (specify → plan → tasks → implement) is scoped to fit within a single sprint; priorities P1→P3 are ordered so the sprint can stop after any completed story and still ship something demonstrable.

## Dependencies

- Reuses the project's existing rendering foundation (3D vector math, camera orientation, screen drawing, and depth comparison) rather than introducing a new graphics stack.
- Requires a terminal that supports color output and key input.
