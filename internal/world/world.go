package world

// World is a finite, bounded voxel grid. Blocks is a single flat slice indexed
// x + SX*(y + SY*z); it is allocated once and mutated in place (never grown),
// so the footprint is constant for the whole session — no memory bleed.
type World struct {
	SX, SY, SZ int
	Blocks     []Block
}

// New allocates an all-Air world of the given dimensions.
func New(sx, sy, sz int) *World {
	return &World{SX: sx, SY: sy, SZ: sz, Blocks: make([]Block, sx*sy*sz)}
}

func (w *World) idx(x, y, z int) int { return x + w.SX*(y+w.SY*z) }

// InBounds reports whether (x,y,z) is inside the grid.
func (w *World) InBounds(x, y, z int) bool {
	return x >= 0 && x < w.SX && y >= 0 && y < w.SY && z >= 0 && z < w.SZ
}

// At returns the block at (x,y,z), or Air when out of bounds so rays and edits
// past the edge are safe rather than errors.
func (w *World) At(x, y, z int) Block {
	if !w.InBounds(x, y, z) {
		return Air
	}
	return w.Blocks[w.idx(x, y, z)]
}

// Set writes a block at (x,y,z); out-of-bounds writes are ignored.
func (w *World) Set(x, y, z int, b Block) {
	if !w.InBounds(x, y, z) {
		return
	}
	w.Blocks[w.idx(x, y, z)] = b
}

// Solid is a convenience for At(x,y,z).IsSolid().
func (w *World) Solid(x, y, z int) bool { return w.At(x, y, z).IsSolid() }
