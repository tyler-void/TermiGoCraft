# Feature Specification: Gravity, Collision & Ray-Traced Lighting

**Feature Branch**: `002-gravity-collision-lighting`  
**Created**: 2026-06-15  
**Status**: Draft  
**Input**: User description: "gravity and collision and qol update, like cursor indicator at the middle of the screen, and ray tracing :0"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Gravity, collision, and jumping (Priority: P1)

The player no longer floats and clips through everything. They have a real body that is pulled down by gravity, stands on top of the ground, cannot walk through blocks, and can jump. Walking off a ledge makes them fall; walking into a wall stops them (while letting them slide along it); jumping lets them get up onto raised terrain.

**Why this priority**: This is the headline fix — today the player flies and passes through the world, which breaks the illusion of a solid place. Making the world solid to stand on, bump into, and jump around is the single biggest improvement and the foundation the rest of the feature sits on.

**Independent Test**: Spawn, stop pressing movement → the player settles onto the ground instead of hovering. Walk toward a wall → forward motion stops but sliding sideways works. Walk off an edge → fall and land. Jump → rise and come back down; jump onto a one-block step → land on top.

**Acceptance Scenarios**:

1. **Given** the player is in the air, **When** no movement is pressed, **Then** the player falls and comes to rest standing on the first solid surface below, and does not sink into or pass through it.
2. **Given** the player is walking toward a solid wall, **When** they continue pressing forward, **Then** they stop at the wall and never enter it, and pressing a sideways direction lets them slide along the wall.
3. **Given** the player is standing on solid ground, **When** they jump, **Then** they rise to a limited height and fall back; pressing jump again while still in the air does nothing.
4. **Given** the player is jumping upward beneath an overhang, **When** their head reaches the block above, **Then** upward motion stops and they do not pass through it.
5. **Given** the player is standing on a block, **When** that block is broken out from under them, **Then** they begin to fall.
6. **Given** a block placement would overlap the player's body, **When** the player tries to place it, **Then** the placement is rejected.

---

### User Story 2 - Quality-of-life: aiming cursor and movement mode (Priority: P2)

The player gets clearer feedback and comfort controls: a distinct aiming cursor pinned to the center of the view that also tells them when a block is in reach to interact with, plus the ability to toggle between walking (with gravity) and a free-fly mode for easy building and exploration, and a small status readout (mode + position).

**Why this priority**: Small, low-risk comfort wins that make the game much more pleasant to play and build in. They depend on movement existing (US1) but deliver value on their own and are quick to verify.

**Independent Test**: A cursor is always visible dead-center; aim at a block within reach → the cursor changes to its "can interact" state; aim at sky → it returns to its neutral state. Toggle movement mode → switch between walking (falls/collides) and flying (free movement); the current mode and position are shown on screen.

**Acceptance Scenarios**:

1. **Given** the game is running, **When** any frame is shown, **Then** a clear aiming cursor is present at the exact center of the view.
2. **Given** the player is aiming at a block within reach, **When** the frame is shown, **Then** the cursor is in a distinct "targeted / can interact" state, visibly different from when nothing is targeted.
3. **Given** the player is in walking mode, **When** they toggle movement mode, **Then** they enter free-fly mode (no gravity, no collision) and can move freely in all directions; toggling again returns them to walking.
4. **Given** the game is running, **When** the player checks the screen, **Then** the current movement mode and the player's position are shown.

---

### User Story 3 - Ray-traced lighting and shadows (Priority: P3)

The world gains realistic directional lighting: a sun lights the scene so surfaces facing it are bright and surfaces facing away are dark, raised terrain and trees cast shadows onto the ground and onto each other, and crevices/corners are subtly darker. The world reads as a lit, three-dimensional place with real depth rather than flat-colored blocks.

**Why this priority**: The exciting visual leap the player is most enthusiastic about, but also the most ambitious and the riskiest for performance. It depends only on the existing world view, so it is built last and can be tuned (or trimmed) without blocking the gameplay improvements in US1/US2.

**Independent Test**: Look across raised terrain or a tree with the sun to one side → the lit faces are clearly brighter than the shaded faces, and a visible shadow falls on the ground behind the obstacle. Break/place a block → the shadow it casts appears/disappears immediately. Stand still → the lighting holds steady with no shimmering.

**Acceptance Scenarios**:

1. **Given** the world is shown with lighting, **When** the player views surfaces at different orientations, **Then** surfaces facing the sun are visibly brighter than surfaces facing away.
2. **Given** a raised block or tree stands between the sun and the ground, **When** the scene is shown, **Then** a shadow is cast on the ground in the direction away from the sun.
3. **Given** the player breaks or places a block, **When** the next frame is shown, **Then** the shadows and lighting in the affected area update to match.
4. **Given** inner corners where blocks meet, **When** the scene is shown, **Then** those recesses appear gradually darker than open, exposed surfaces.
5. **Given** the player and world are not moving, **When** consecutive frames are shown, **Then** the lighting is stable with no flicker or shimmer.

---

### Edge Cases

- **Spawn or generation inside terrain**: if the player's body overlaps solid blocks (e.g., a block appears at their feet), they are nudged to the nearest free space rather than trapped or suffocated.
- **World edges / bottom**: the player cannot fall out of the world or below it indefinitely; they stay within the playable space.
- **Block under the feet removed**: the player begins to fall that same moment.
- **Standing in a 1-block-tall gap / under a low ceiling**: the player fits and does not jitter or get pushed through.
- **Fast fall onto a thin floor**: the player lands on it rather than tunneling through.
- **Aiming with no target** (sky / beyond reach): the cursor shows its neutral, non-interactive state.
- **Lighting at the world boundary / when looking at the sky**: the sky and edges render cleanly without artifacts.
- **Toggling to fly while falling**: motion stops cleanly and free movement takes over.

## Requirements *(mandatory)*

### Functional Requirements

#### Movement, gravity & collision

- **FR-001**: The player MUST have a physical body of fixed height and width (an upright box around the viewpoint) instead of being a dimensionless point.
- **FR-002**: System MUST continuously pull the player downward under gravity so they fall whenever unsupported.
- **FR-003**: The player MUST come to rest on top of the first solid surface beneath them and stop falling while supported.
- **FR-004**: The player MUST NOT pass through solid blocks; motion into a block MUST be stopped on the blocked axis while remaining motion along other axes is preserved (sliding along walls).
- **FR-005**: The player MUST be able to jump while supported, rising to a limited height before falling back; a jump MUST NOT be possible while airborne.
- **FR-006**: Upward motion MUST stop when the player's head meets a solid block.
- **FR-007**: System MUST prevent the player from leaving the world bounds or falling indefinitely.
- **FR-008**: If the player's body ever overlaps solid blocks, the system MUST resolve the overlap by relocating the player to the nearest free space rather than leaving them stuck.
- **FR-009**: Block placement MUST be rejected when the new block would intersect any part of the player's body.

#### Quality-of-life

- **FR-010**: The player MUST be able to toggle between walking mode (gravity and collision apply) and free-fly mode (no gravity, no collision); walking is the default at launch.
- **FR-011**: System MUST display a clear aiming cursor at the exact center of the view at all times.
- **FR-012**: The aiming cursor MUST visually distinguish "a block is targeted within reach" from "nothing is targeted".
- **FR-013**: System MUST display the current movement mode and the player's position as on-screen status.

#### Lighting & shadows

- **FR-014**: Surfaces MUST be shaded by a directional light (a sun) such that faces oriented toward the light are brighter than faces oriented away from it.
- **FR-015**: Solid blocks MUST cast shadows: a surface that has another solid block between it and the light MUST appear darker than an equivalent unobstructed surface.
- **FR-016**: Recessed areas, inner corners, and the seams where blocks meet MUST appear gradually darker than open, exposed surfaces, conveying depth.
- **FR-017**: Lighting and shadows MUST update to match the world as it is edited (break/place) and as the player moves, with no manual refresh.
- **FR-018**: Lighting MUST be stable frame-to-frame: a static scene MUST NOT shimmer, flicker, or crawl.

#### Cross-cutting

- **FR-019**: All new behavior MUST preserve the existing smooth, responsive feel — continuous redraw and prompt response to input.

### Key Entities *(include if feature involves data)*

- **Player body**: the player's physical presence — viewpoint/eye position, body height and width, current velocity, whether currently supported (grounded), and current movement mode (walking / flying).
- **Light (Sun)**: the directional light source — its direction, brightness, and a base ambient level for surfaces not directly lit.
- **Aiming cursor**: the center-of-view reticle and its two states (neutral vs. targeting a reachable block).
- *(Reuses the existing World, Camera, Block, Target, and HUD entities from feature 001.)*

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A player released in mid-air falls and comes to rest on the ground within ~2 seconds and never passes through the floor.
- **SC-002**: In 100% of attempts, walking into a wall stops forward progress without ever placing the player inside a solid block, and sideways input still slides them along the wall.
- **SC-003**: From standing, the player can jump onto a block one level higher and land stably on top of it.
- **SC-004**: Across a continuous 10-minute session of walking, jumping, and editing, there are zero occurrences of clipping through terrain, infinite falling, or becoming permanently stuck.
- **SC-005**: The center aiming cursor is visible in every frame and reflects the correct targeting state within one frame of aiming onto or off of a reachable block.
- **SC-006**: The player can switch between walking and flying instantly, and each mode behaves correctly (flying ignores gravity and collision; walking obeys both).
- **SC-007**: With lighting on, a first-time viewer can identify the light direction and point out at least one cast shadow within 10 seconds of looking at raised terrain or a tree.
- **SC-008**: Breaking or placing a block updates the lighting and shadows in the affected area within one frame (no perceptible delay).
- **SC-009**: With lighting and shadows enabled, the game stays real-time and responsive at the default world and a typical terminal size — movement and looking show no perceptible stutter.
- **SC-010**: When the player and world are static, lighting is steady across frames with no visible shimmer or flicker.

## Assumptions

- This feature builds directly on feature 001 (the first-person voxel renderer) and reuses its world, camera, rendering, input, and HUD.
- **Arcade-simple physics**: constant gravity, a single fixed jump impulse, and straightforward axis-by-axis collision — not a full physics engine. No momentum/friction tuning, swimming, ladders, slabs, or partial-height blocks.
- **Free-fly mode is retained** as a toggle (creative-style building/exploration); **walking is the default** at launch. (Change if you'd prefer flying as the default.)
- **"Ray tracing" is scoped to direct lighting**: a single directional **sun** with **hard cast shadows** and **ambient-occlusion-style darkening** in crevices. **Out of scope** this feature: reflections, refraction/transparency, water, multi-bounce/indirect (global illumination) lighting, colored bounce light, and soft (penumbra) shadows. The sun direction is **fixed** (no day/night cycle this sprint).
- **Lighting is computed for what the player currently sees** (not precomputed/baked into the world), so edits and movement update it immediately.
- The aiming cursor builds on the existing center crosshair, adding a clear targeted/neutral distinction.
- Single-player, local, keyboard-only, no persistence — unchanged from feature 001.
- Performance target is unchanged: the game must remain real-time and smooth on the kitty reference terminal, with the lighting cost fitting inside the existing per-frame budget.

## Dependencies

- Depends on feature 001's code: the player viewpoint (camera position), the movement/update path, and the world raycaster + shared framebuffer that lighting and the cursor build upon.
