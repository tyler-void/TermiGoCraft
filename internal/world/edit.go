package world

// Break removes the block at the hit cell. It is a no-op (returns false) when
// the hit is invalid, so aiming at the sky and breaking does nothing.
func (w *World) Break(h Hit) bool {
	if !h.OK {
		return false
	}
	w.Set(h.X, h.Y, h.Z, Air)
	return true
}

// Place puts block b into the empty cell adjacent to the hit face. It is
// rejected (returns false) when there is no valid target, when the neighbor is
// out of bounds or already solid, or when it would occupy the player's own cell
// (px,py,pz) — satisfying the self-placement guard (FR-010).
func (w *World) Place(h Hit, b Block, px, py, pz int) bool {
	if !h.OK {
		return false
	}
	dx, dy, dz := h.Face.Normal()
	nx, ny, nz := h.X+dx, h.Y+dy, h.Z+dz
	if !w.InBounds(nx, ny, nz) {
		return false
	}
	if w.Solid(nx, ny, nz) {
		return false
	}
	if nx == px && ny == py && nz == pz {
		return false
	}
	w.Set(nx, ny, nz, b)
	return true
}
