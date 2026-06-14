package render

import (
	"github.com/tyler-void/TermiGoCraft/internal/geom"
	"github.com/tyler-void/TermiGoCraft/internal/world"
)

// Sun is the single directional light.
type Sun struct {
	Dir       geom.Vec3 // unit, points toward the sun
	Intensity float64   // diffuse weight
	Ambient   float64   // base light so nothing is pure black
}

// DefaultSun is a fixed sun from the upper-front-right.
func DefaultSun() Sun {
	return Sun{
		Dir:       geom.Vec3{X: 0.4, Y: 0.85, Z: 0.3}.Normalize(),
		Intensity: 0.75,
		Ambient:   0.35,
	}
}

const (
	shadowDist = 24.0
	shadowBias = 0.02
	aoMin      = 0.55
)

func faceNormal(f world.Face) geom.Vec3 {
	x, y, z := f.Normal()
	return geom.Vec3{X: float64(x), Y: float64(y), Z: float64(z)}
}

// lightAt returns a brightness multiplier for a surface hit: directional sun
// (Lambert) gated by a hard shadow ray, scaled by ambient occlusion. Pure,
// deterministic, allocation-free.
func lightAt(w *world.World, hit world.Hit, hitPoint geom.Vec3, sun Sun) float64 {
	n := faceNormal(hit.Face)
	diffuse := n.Dot(sun.Dir)
	if diffuse < 0 {
		diffuse = 0
	}
	direct := 0.0
	if diffuse > 0 {
		// Shadow ray toward the sun; offset off the surface to avoid acne.
		origin := hitPoint.Add(n.Scale(shadowBias))
		if !w.Cast(origin, sun.Dir, shadowDist).OK {
			direct = sun.Intensity * diffuse
		}
	}
	light := (sun.Ambient + direct) * aoFactor(w, hit, hitPoint)
	if light > 1 {
		light = 1
	}
	return light
}

// aoFactor computes per-pixel ambient occlusion by evaluating the 0fps vertex-AO
// rule at the hit face's four corners and bilinearly interpolating by the hit
// point's in-face UV.
func aoFactor(w *world.World, hit world.Hit, hp geom.Vec3) float64 {
	nx, ny, nz := hit.Face.Normal()
	cx, cy, cz := hit.X, hit.Y, hit.Z

	// Tangent integer axes of the face, and the in-face UV of the hit point.
	var t1, t2 [3]int
	var u, v float64
	switch {
	case nx != 0:
		t1, t2 = [3]int{0, 1, 0}, [3]int{0, 0, 1}
		u, v = hp.Y-float64(cy), hp.Z-float64(cz)
	case ny != 0:
		t1, t2 = [3]int{1, 0, 0}, [3]int{0, 0, 1}
		u, v = hp.X-float64(cx), hp.Z-float64(cz)
	default:
		t1, t2 = [3]int{1, 0, 0}, [3]int{0, 1, 0}
		u, v = hp.X-float64(cx), hp.Y-float64(cy)
	}
	u, v = clamp01(u), clamp01(v)

	// Occluders live in the cell layer just outside the face.
	ox, oy, oz := cx+nx, cy+ny, cz+nz
	corner := func(du, dv int) float64 {
		s1 := solidInt(w, ox+t1[0]*du, oy+t1[1]*du, oz+t1[2]*du)
		s2 := solidInt(w, ox+t2[0]*dv, oy+t2[1]*dv, oz+t2[2]*dv)
		cc := solidInt(w, ox+t1[0]*du+t2[0]*dv, oy+t1[1]*du+t2[1]*dv, oz+t1[2]*du+t2[2]*dv)
		return aoMult(vertexAO(s1, s2, cc))
	}
	a := corner(-1, -1)
	b := corner(1, -1)
	c := corner(-1, 1)
	d := corner(1, 1)
	return a*(1-u)*(1-v) + b*u*(1-v) + c*(1-u)*v + d*u*v
}

// vertexAO is the canonical 0fps rule: 3 = fully open, 0 = most occluded.
func vertexAO(s1, s2, corner int) int {
	if s1 == 1 && s2 == 1 {
		return 0
	}
	return 3 - (s1 + s2 + corner)
}

func aoMult(level int) float64 { return aoMin + (1-aoMin)*float64(level)/3.0 }

func solidInt(w *world.World, x, y, z int) int {
	if w.Solid(x, y, z) {
		return 1
	}
	return 0
}

func clamp01(x float64) float64 {
	if x < 0 {
		return 0
	}
	if x > 1 {
		return 1
	}
	return x
}
