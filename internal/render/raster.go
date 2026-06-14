package render

import (
	"math"

	"github.com/tyler-void/TermiGoCraft/internal/camera"
	"github.com/tyler-void/TermiGoCraft/internal/geom"
	"github.com/tyler-void/TermiGoCraft/internal/world"
)

// projV is a vertex projected to pixel space with its camera-Z depth.
type projV struct {
	x, y, z float64
	ok      bool
}

// basis bundles the per-frame projection constants so triangle projection does
// not recompute trig per vertex.
type basis struct {
	pos            geom.Vec3
	fwd, right, up geom.Vec3
	tanF, aspect   float64
	w, h           int
}

func newBasis(cam *camera.Camera, b *Buffer) basis {
	return basis{
		pos:    cam.Pos,
		fwd:    cam.Direction(),
		right:  cam.Right(),
		up:     cam.Up(),
		tanF:   math.Tan(cam.FOV / 2),
		aspect: float64(b.W) / float64(b.H),
		w:      b.W,
		h:      b.H,
	}
}

func (bs basis) project(p geom.Vec3) projV {
	rel := p.Sub(bs.pos)
	cz := rel.Dot(bs.fwd)
	if cz <= 0.01 {
		return projV{ok: false}
	}
	cx := rel.Dot(bs.right)
	cy := rel.Dot(bs.up)
	ndcX := cx / (cz * bs.tanF * bs.aspect)
	ndcY := cy / (cz * bs.tanF)
	sx := (ndcX+1)*0.5*float64(bs.w) - 0.5
	sy := (1-ndcY)*0.5*float64(bs.h) - 0.5
	return projV{x: sx, y: sy, z: cz, ok: true}
}

// RenderMesh draws a filled, depth-tested triangle mesh at modelPos (Pass B).
// It writes through TestSet, so terrain already in the buffer occludes the mesh
// and vice versa.
func RenderMesh(b *Buffer, mesh []geom.Triangle, modelPos geom.Vec3, color geom.RGB, cam *camera.Camera) {
	bs := newBasis(cam, b)
	for _, t := range mesh {
		a := t.A.Add(modelPos)
		bb := t.B.Add(modelPos)
		cc := t.C.Add(modelPos)

		// Backface cull: skip triangles facing away from the camera.
		normal := bb.Sub(a).Cross(cc.Sub(a))
		if normal.Dot(bs.pos.Sub(a)) < 0 {
			continue
		}

		pa := bs.project(a)
		pb := bs.project(bb)
		pc := bs.project(cc)
		if !pa.ok || !pb.ok || !pc.ok {
			continue
		}
		fillTriangle(b, pa, pb, pc, color)
	}
}

// RenderHeld draws the held-item mesh fixed relative to the camera so it sits in
// the lower-right of the view (Pass B, view space).
func RenderHeld(b *Buffer, mesh []geom.Triangle, color geom.RGB, cam *camera.Camera) {
	d := cam.Direction()
	r := cam.Right()
	u := cam.Up()
	pos := cam.Pos.Add(d.Scale(1.4)).Add(r.Scale(0.7)).Add(u.Scale(-0.55))
	RenderMesh(b, mesh, pos, color, cam)
}

// RenderSelection outlines the targeted block as a white wireframe cube biased
// slightly toward the camera so it reads on top of its own surface (Pass C).
func RenderSelection(b *Buffer, h world.Hit, cam *camera.Camera) {
	if !h.OK {
		return
	}
	bs := newBasis(cam, b)
	ox, oy, oz := float64(h.X), float64(h.Y), float64(h.Z)
	corners := [8]geom.Vec3{
		{X: ox, Y: oy, Z: oz}, {X: ox + 1, Y: oy, Z: oz},
		{X: ox + 1, Y: oy, Z: oz + 1}, {X: ox, Y: oy, Z: oz + 1},
		{X: ox, Y: oy + 1, Z: oz}, {X: ox + 1, Y: oy + 1, Z: oz},
		{X: ox + 1, Y: oy + 1, Z: oz + 1}, {X: ox, Y: oy + 1, Z: oz + 1},
	}
	var pc [8]projV
	for i, c := range corners {
		pc[i] = bs.project(c)
	}
	edges := [12][2]int{
		{0, 1}, {1, 2}, {2, 3}, {3, 0},
		{4, 5}, {5, 6}, {6, 7}, {7, 4},
		{0, 4}, {1, 5}, {2, 6}, {3, 7},
	}
	white := geom.RGB{R: 255, G: 255, B: 255}
	for _, e := range edges {
		p0, p1 := pc[e[0]], pc[e[1]]
		if !p0.ok || !p1.ok {
			continue
		}
		z := float32(math.Min(p0.z, p1.z) * 0.97)
		drawLine(b, int(p0.x), int(p0.y), int(p1.x), int(p1.y), z, white)
	}
}

// fillTriangle rasterizes a triangle with barycentric coverage and per-pixel
// interpolated depth.
func fillTriangle(b *Buffer, p0, p1, p2 projV, color geom.RGB) {
	minX := int(math.Floor(min3(p0.x, p1.x, p2.x)))
	maxX := int(math.Ceil(max3(p0.x, p1.x, p2.x)))
	minY := int(math.Floor(min3(p0.y, p1.y, p2.y)))
	maxY := int(math.Ceil(max3(p0.y, p1.y, p2.y)))
	if minX < 0 {
		minX = 0
	}
	if minY < 0 {
		minY = 0
	}
	if maxX >= b.W {
		maxX = b.W - 1
	}
	if maxY >= b.H {
		maxY = b.H - 1
	}
	area := edge(p0.x, p0.y, p1.x, p1.y, p2.x, p2.y)
	if area == 0 {
		return
	}
	inv := 1 / area
	for y := minY; y <= maxY; y++ {
		py := float64(y) + 0.5
		for x := minX; x <= maxX; x++ {
			px := float64(x) + 0.5
			w0 := edge(p1.x, p1.y, p2.x, p2.y, px, py)
			w1 := edge(p2.x, p2.y, p0.x, p0.y, px, py)
			w2 := edge(p0.x, p0.y, p1.x, p1.y, px, py)
			// inside if all weights share the triangle's winding sign
			if (w0 >= 0 && w1 >= 0 && w2 >= 0) || (w0 <= 0 && w1 <= 0 && w2 <= 0) {
				l0 := w0 * inv
				l1 := w1 * inv
				l2 := w2 * inv
				z := l0*p0.z + l1*p1.z + l2*p2.z
				b.TestSet(x, y, float32(z), color)
			}
		}
	}
}

// drawLine is a depth-aware Bresenham line in pixel space.
func drawLine(b *Buffer, x0, y0, x1, y1 int, z float32, c geom.RGB) {
	dx := abs(x1 - x0)
	dy := abs(y1 - y0)
	sx := 1
	if x0 >= x1 {
		sx = -1
	}
	sy := 1
	if y0 >= y1 {
		sy = -1
	}
	err := dx - dy
	for {
		b.TestSet(x0, y0, z, c)
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x0 += sx
		}
		if e2 < dx {
			err += dx
			y0 += sy
		}
	}
}

// edge is twice the signed area of the triangle (ax,ay)(bx,by)(cx,cy); used both
// for the full-triangle area and per-pixel barycentric weights.
func edge(ax, ay, bx, by, cx, cy float64) float64 {
	return (bx-ax)*(cy-ay) - (by-ay)*(cx-ax)
}

func min3(a, b, c float64) float64 { return math.Min(a, math.Min(b, c)) }
func max3(a, b, c float64) float64 { return math.Max(a, math.Max(b, c)) }

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
