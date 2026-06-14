// Package world holds the voxel grid, its generation, the DDA ray traversal
// used for both rendering and aiming, and the break/place edit operations.
// It depends only on geom and stores blocks as a flat, pointer-free []Block
// so the garbage collector never has to scan the world.
package world

// Block identifies the material in a cell. Air is the empty/non-solid case.
type Block uint8

const (
	Air Block = iota
	Grass
	Dirt
	Stone
	Wood
	Leaves
	Sand
	// Count is the number of distinct block types; keep it last.
	Count
)

// IsSolid reports whether the block stops a ray and collides.
func (b Block) IsSolid() bool { return b != Air }

// Placeable lists the solid block types in hotbar order (FR-011).
var Placeable = []Block{Grass, Dirt, Stone, Wood, Leaves, Sand}

// Name returns a short human label for the HUD indicator.
func (b Block) Name() string {
	switch b {
	case Air:
		return "Air"
	case Grass:
		return "Grass"
	case Dirt:
		return "Dirt"
	case Stone:
		return "Stone"
	case Wood:
		return "Wood"
	case Leaves:
		return "Leaves"
	case Sand:
		return "Sand"
	}
	return "?"
}
