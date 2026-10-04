package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// ForbiddenCity is the central axis of Beijing's imperial palace: the U-shaped Meridian
// Gate with its five pavilions on a massive red base, the bow-shaped Golden Water River
// crossed by five white marble bridges, the Gate of Supreme Harmony, the great paved
// court and the three-tier white marble Dragon Terrace carrying the Halls of Supreme,
// Central and Preserving Harmony under gold glazed roofs. Side galleries and halls,
// the purple-red outer wall with its complex corner towers and gates, and the moat
// surround it. North is -Z; the Meridian Gate faces south (+Z).
type ForbiddenCity struct{}

func init() { Register(ForbiddenCity{}) }

func (ForbiddenCity) Name() string  { return "forbidden_city" }
func (ForbiddenCity) Title() string { return "Forbidden City" }

const (
	forbiddenCityG       = 0.5   // Ground top.
	forbiddenCityWallX   = 32.0  // Outer face of the east/west walls.
	forbiddenCityWallS   = 22.0  // Outer face of the south wall.
	forbiddenCityWallN   = -64.0 // Outer face of the north wall.
	forbiddenCityWallT   = 2.4   // Wall thickness.
	forbiddenCityWallH   = 5.6   // Wall height above the ground.
	forbiddenCityStrip   = 2.0   // Road between the wall and the moat.
	forbiddenCityMoatS   = 39.0  // South moat inner edge (beyond the Meridian wings).
	forbiddenCityMoatW   = 5.0   // Moat width.
	forbiddenCityRim     = 2.5   // Land beyond the moat.
	forbiddenCityCourtX  = 21.0  // Inner face of the side galleries.
	forbiddenCityGallery = 2.6   // Side gallery depth.
	forbiddenCityMeridN  = 15.5  // Meridian Gate base, inner (north) face.
	forbiddenCityMeridS  = 23.5  // Meridian Gate base, outer face.
	forbiddenCityWingS   = 37.0  // Meridian Gate wing tips.
	forbiddenCityRiverC  = -38.0 // Golden Water River arc center (z).
	forbiddenCityRiverR  = 49.5  // Golden Water River arc radius.
	forbiddenCityTierH   = 2.0   // Dragon Terrace tier height.
	forbiddenCityTierIn  = 2.0   // Dragon Terrace tier setback.
)

// forbiddenCityRect is a plan rectangle by center and half sizes.
type forbiddenCityRect struct{ X, Z, Hx, Hz float64 }

// forbiddenCityRoof is a glazed roof described as a height field over a rectangle:
// a concave hip (or, with Gable, a hip-and-gable xieshan) rising from the eave, with
// upturned corners. Hole > 0 makes it the lower skirt of a double-eave roof.
type forbiddenCityRoof struct {
	C            Vec2
	Hx, Hz       float64 // Eave half sizes.
	Ye, Rise     float64 // Eave height; rise to the ridge (or to the hole edge).
	Reach        float64 // Horizontal run of the rise; 0 = min(Hx, Hz).
	Gable        float64 // Xieshan: gable inset from the short ends; 0 = hip.
	Lift         float64 // Corner upturn.
	HoleX, HoleZ float64 // Skirt: no roof inside this rectangle.
	Ridges       bool    // Raise hip and main ridges.
}

func (ForbiddenCity) clamp01(v float64) float64 { return math.Max(0, math.Min(1, v)) }

func (ForbiddenCity) in(x, z float64, r forbiddenCityRect) bool {
	return math.Abs(x-r.X) <= r.Hx && math.Abs(z-r.Z) <= r.Hz
}

// at is the roof's top surface height over a plan point.
func (r forbiddenCityRoof) at(x, z float64) (float64, bool) {
	ax, az := math.Abs(x-r.C.X), math.Abs(z-r.C.Y)
	dX, dZ := r.Hx-ax, r.Hz-az
	if dX < 0 || dZ < 0 {
		return 0, false
	}
	if r.HoleX > 0 && ax < r.HoleX && az < r.HoleZ {
		return 0, false
	}
	reach := r.Reach
	if reach <= 0 {
		reach = math.Min(r.Hx, r.Hz)
	}
	d := math.Min(dX, dZ)
	hip := true
	if r.Gable > 0 {
		if r.Hx >= r.Hz && dX >= r.Gable {
			d, hip = dZ, false
		} else if r.Hz > r.Hx && dZ >= r.Gable {
			d, hip = dX, false
		}
	}
	u := math.Max(0, math.Min(1, d/reach))
	y := r.Ye + r.Rise*math.Pow(u, 1.7)
	if u < 0.12 {
		y += 0.3 * (1 - u/0.12) * (1 - u/0.12)
	}
	cr := math.Min(3.5, reach)
	c := math.Max(0, 1-(dX+dZ)/cr)
	y += r.Lift * c * c
	if r.Ridges {
		if hip && math.Abs(dX-dZ) < 0.55 && d < reach {
			y += 0.4
		}
		if d >= reach-0.5 {
			y += 0.5
		}
	}
	return y, true
}

// box bounds the roof shell.
func (r forbiddenCityRoof) box(th float64) Box {
	return Box{V(r.C.X-r.Hx-0.5, r.Ye-3*th-0.6, r.C.Y-r.Hz-0.5), V(r.C.X+r.Hx+0.5, r.Ye+r.Rise+r.Lift+1.2, r.C.Y+r.Hz+0.5)}
}

// surface combines roofs by taking the highest one at each plan point.
func (ForbiddenCity) surface(roofs ...forbiddenCityRoof) func(x, z float64) (float64, bool) {
	return func(x, z float64) (float64, bool) {
		best, any := 0.0, false
		for _, r := range roofs {
			if y, ok := r.at(x, z); ok && (!any || y > best) {
				best, any = y, true
			}
		}
		return best, any
	}
}

// shell fills a roof shell of vertical-normal thickness th under a surface, paints
// tile rows and a blue-green painted soffit near the eave.
func (f ForbiddenCity) shell(g *Grid, color Color, surf func(x, z float64) (float64, bool), box Box, th float64) {
	thick := func(x, z, s float64) float64 {
		const e = 0.35
		slope := func(a, b float64, okA, okB bool) float64 {
			switch {
			case okA && okB:
				return (b - a) / (2 * e)
			case okB:
				return (b - s) / e
			case okA:
				return (s - a) / e
			}
			return 0
		}
		x0, ok0 := surf(x-e, z)
		x1, ok1 := surf(x+e, z)
		z0, ok2 := surf(x, z-e)
		z1, ok3 := surf(x, z+e)
		gx, gz := slope(x0, x1, ok0, ok1), slope(z0, z1, ok2, ok3)
		return math.Min(th*math.Sqrt(1+gx*gx+gz*gz), 3*th)
	}
	in := func(p Vec3) bool {
		s, ok := surf(p.X, p.Z)
		return ok && p.Y <= s && p.Y >= s-thick(p.X, p.Z, s)
	}
	g.Use(color).FillVolume(Func{box, in})
	// Anything poking through the tiles (gable fills, wall tops) takes the tile color.
	g.Paint(color, Func{box, func(p Vec3) bool {
		s, ok := surf(p.X, p.Z)
		return ok && p.Y <= s+0.6 && p.Y >= s-0.7
	}})
	// Soffit: the underside layer near the eaves is painted.
	g.Paint(Grass, Func{box, func(p Vec3) bool {
		s, ok := surf(p.X, p.Z)
		if !ok {
			return false
		}
		v := thick(p.X, p.Z, s)
		return p.Y <= s && p.Y >= s-v && p.Y < s-v+0.4
	}})
}

// roof fills a single roof, its gable triangles and ridge ornaments.
func (f ForbiddenCity) roof(g *Grid, r forbiddenCityRoof, th float64) {
	f.gables(g, r)
	f.shell(g, Gold, f.surface(r), r.box(th), th)
	f.ornaments(g, r)
}

// gables fills the vertical gable triangles of a xieshan roof in red with a gold rim.
func (f ForbiddenCity) gables(g *Grid, r forbiddenCityRoof) {
	if r.Gable <= 0 {
		return
	}
	long := r.Hx >= r.Hz
	box := r.box(1)
	g.Use(Red).Fill(Func{box, func(p Vec3) bool {
		ax, az := math.Abs(p.X-r.C.X), math.Abs(p.Z-r.C.Y)
		dX, dZ := r.Hx-ax, r.Hz-az
		if !long {
			dX, dZ = dZ, dX
		}
		if dX < r.Gable || dX > r.Gable+0.9 || dZ < 0.6 {
			return false
		}
		s, ok := r.at(p.X, p.Z)
		low := r.Ye + r.Rise*math.Pow(f.clamp01(math.Min(r.Gable, dZ)/math.Max(r.Reach, math.Min(r.Hx, r.Hz))), 1.7)
		return ok && p.Y <= s-0.3 && p.Y >= low-0.6
	}})
}

// ornaments places the gold ridge, chiwen ridge-end dragons or a finial ball.
func (f ForbiddenCity) ornaments(g *Grid, r forbiddenCityRoof) {
	top := r.Ye + r.Rise + 0.5
	long, short := r.Hx, r.Hz
	if r.Hz > r.Hx {
		long, short = r.Hz, r.Hx
	}
	run := long - short
	if r.Gable > 0 {
		run = long - r.Gable
	}
	g.Use(Gold)
	if run < 0.6 {
		// Pyramid: a gilded finial.
		g.Fill(Cylinder{Center: r.C, Radius: 0.5, Y0: top - 1, Y1: top + 0.8})
		g.Fill(Sphere(r.C.At(top+1.3), 0.85))
		return
	}
	dir := V(1, 0, 0)
	if r.Hz > r.Hx {
		dir = V(0, 0, 1)
	}
	a := r.C.At(top).Add(dir.Scale(-run))
	b := r.C.At(top).Add(dir.Scale(run))
	g.Fill(Capsule{a, b, 0.55})
	for _, e := range []Vec3{a, b} {
		g.Fill(Capsule{e.Add(V(0, -0.2, 0)), e.Add(V(0, 1.3, 0)), 0.55})
		g.Fill(Capsule{e.Add(V(0, 1.3, 0)), e.Add(V(0, 1.3, 0)).Sub(dir.Scale(0.7 * math.Copysign(1, e.Sub(r.C.At(top)).Dot(dir)))), 0.5})
	}
}

// forbiddenCityHall is a red timber hall with lattice doors, painted brackets and a
// gold glazed roof, single or double eaved.
type forbiddenCityHall struct {
	C        Vec2
	Hx, Hz   float64 // Body half sizes.
	Y0, Wall float64 // Floor height and wall height.
	Ov       float64 // Eave overhang.
	Double   bool    // Double eaves.
	Upper    float64 // Upper storey height above the skirt for double eaves.
	Gable    float64 // Xieshan gable inset; 0 = hip (or pyramid when square).
	Rise     float64 // Upper roof rise as a fraction of its short half size.
	Doors    bool    // Lattice doors on the long faces (else windows).
}

// body builds a red rectangular wall shell with columns, lattice openings and the
// painted bracket frieze under the eaves.
func (f ForbiddenCity) body(g *Grid, c Vec2, hx, hz, y0, top float64, doors bool) {
	g.Use(Red).Fill(Subtract(BlockAt(c, 2*hx, 2*hz, y0, top), BlockAt(c, 2*hx-2, 2*hz-2, y0+0.3, top+10)))
	long, alongZ := hx, false
	if hz > hx {
		long, alongZ = hz, true
	}
	n := int(math.Max(1, math.Round(2*long/2.7)))
	bay := 2 * long / float64(n)
	lo := y0 + 0.4
	if !doors {
		lo = y0 + 0.4 + (top-y0)*0.35
	}
	hi := top - 1.3
	for k := 0; k < n; k++ {
		pos := -long + bay*(float64(k)+0.5)
		for _, sd := range []float64{-1, 1} {
			var at Vec2
			w, d := bay-0.9, 1.2
			if alongZ {
				at = P(c.X+sd*hx, c.Y+pos)
				w, d = d, w
			} else {
				at = P(c.X+pos, c.Y+sd*hz)
			}
			if hi-lo < 0.8 {
				continue
			}
			g.Paint(Gold, BlockAt(at, w+0.1, d+0.1, lo-0.4, hi+0.4))
			inner := BlockAt(at, w-0.7, d+0.2, lo, hi)
			if alongZ {
				inner = BlockAt(at, d+0.2, w-0.7, lo, hi)
			}
			g.Carve(Func{inner.Bounds(), func(p Vec3) bool {
				a := p.X
				if alongZ {
					a = p.Z
				}
				return int(math.Floor((a+200)/0.8))%2 == 0 && int(math.Floor((p.Y+200)/0.8))%2 == 1
			}})
			g.Paint(Red, inner)
		}
		// Columns at the bay boundaries, proud of the wall.
		if k > 0 {
			for _, sd := range []float64{-1, 1} {
				var at Vec2
				if alongZ {
					at = P(c.X+sd*(hx+0.15), c.Y-long+bay*float64(k))
				} else {
					at = P(c.X-long+bay*float64(k), c.Y+sd*(hz+0.15))
				}
				if doors {
					g.Use(Red).Fill(Cylinder{Center: at, Radius: 0.5, Y0: y0, Y1: top})
				}
			}
		}
	}
	// Bracket frieze: blue and green panels over a gold beam.
	frieze := Subtract(BlockAt(c, 2*hx+1.4, 2*hz+1.4, top-1.1, top+0.5), BlockAt(c, 2*hx-2, 2*hz-2, top-3, top+3))
	g.Use(Navy).Fill(frieze)
	g.Paint(Grass, Func{frieze.Bounds(), func(p Vec3) bool {
		return frieze.Contains(p) && p.Y > top-0.6 && int(math.Floor((p.X+p.Z+200)/1.6))%2 == 0
	}})
	g.Paint(Gold, Func{frieze.Bounds(), func(p Vec3) bool { return frieze.Contains(p) && p.Y <= top-0.7 }})
}

// hall builds a complete hall and returns its roof's ridge height.
func (f ForbiddenCity) hall(g *Grid, h forbiddenCityHall) float64 {
	top := h.Y0 + h.Wall
	f.body(g, h.C, h.Hx, h.Hz, h.Y0, top, h.Doors)
	rise := h.Rise
	if rise <= 0 {
		rise = 1.0
	}
	hx, hz, ye := h.Hx, h.Hz, top+0.4
	if h.Double {
		const ins = 1.6
		skirt := forbiddenCityRoof{C: h.C, Hx: hx + h.Ov, Hz: hz + h.Ov, Ye: ye, Rise: 0.55 * (h.Ov + ins), Reach: h.Ov + ins,
			Lift: 1.0, HoleX: hx - ins, HoleZ: hz - ins}
		f.shell(g, Gold, f.surface(skirt), skirt.box(1.1), 1.1)
		hx, hz = hx-ins, hz-ins
		top2 := ye + skirt.Rise + h.Upper
		f.body(g, h.C, hx, hz, ye, top2, false)
		ye = top2 + 0.4
	}
	r := forbiddenCityRoof{C: h.C, Hx: hx + h.Ov, Hz: hz + h.Ov, Ye: ye, Gable: h.Gable, Lift: 1.3, Ridges: true}
	r.Rise = rise * math.Min(r.Hx, r.Hz)
	if r.Gable > 0 {
		r.Gable += h.Ov
	}
	f.roof(g, r, 1.1)
	return r.Ye + r.Rise
}

// rectSDF is the signed distance to a union of rectangles grown by -inset, with the
// coordinate along the nearest edge.
func (ForbiddenCity) rectSDF(x, z float64, rects []forbiddenCityRect, inset float64) (float64, float64) {
	best, along := math.Inf(1), 0.0
	for _, r := range rects {
		qx := math.Abs(x-r.X) - (r.Hx - inset)
		qz := math.Abs(z-r.Z) - (r.Hz - inset)
		out := math.Hypot(math.Max(qx, 0), math.Max(qz, 0))
		d := out + math.Min(math.Max(qx, qz), 0)
		if d < best {
			best = d
			if qx > qz {
				along = z
			} else {
				along = x
			}
		}
	}
	return best, along
}

// balustrade fills white marble posts and panels in a band of a signed-distance
// field, at deck height y, with dragon-head spouts below the posts.
func (f ForbiddenCity) balustrade(g *Grid, box Box, sdf func(x, z float64) (float64, float64, bool), y float64, spouts bool) {
	g.Use(White).Fill(Func{Box{V(box.Min.X-1.2, y-1, box.Min.Z-1.2), V(box.Max.X+1.2, y+1.6, box.Max.Z+1.2)}, func(p Vec3) bool {
		d, along, ok := sdf(p.X, p.Z)
		if !ok {
			return false
		}
		post := math.Mod(along+300, 2.0) < 0.8
		if d >= -0.85 && d <= 0 {
			if p.Y > y && p.Y <= y+0.45 {
				return true
			}
			return post && p.Y > y && p.Y <= y+1.25
		}
		return spouts && post && d > 0 && d <= 1.1 && p.Y >= y-0.8 && p.Y <= y-0.3
	}})
}

// stair fills a straight flight descending along +Z (dir 1) or -Z (dir -1) from a
// deck edge at zTop, height top, to the ground, with a carved central ramp and
// balustraded cheeks; it first clears whatever stands over the flight.
func (f ForbiddenCity) stair(g *Grid, x, halfW, zTop, top float64, dir float64, ramp bool) {
	const run, step = 0.7, 0.4
	n := math.Ceil((top - forbiddenCityG) / step)
	length := n * run
	z0, z1 := zTop, zTop+dir*length
	box := Box{V(x-halfW-1.2, 0, math.Min(z0, z1)-1.2), V(x+halfW+1.2, top+2, math.Max(z0, z1)+0.1)}
	surf := func(z float64) float64 {
		t := (z - zTop) * dir
		if t <= 0 {
			return top
		}
		return top - math.Ceil(t/run)*step
	}
	g.Carve(Func{box, func(p Vec3) bool {
		t := (p.Z - zTop) * dir
		return math.Abs(p.X-x) <= halfW+1 && t > -1.2 && t <= length && p.Y > surf(p.Z)+0.05 && p.Y > forbiddenCityG
	}})
	g.Use(White).Fill(Func{box, func(p Vec3) bool {
		t := (p.Z - zTop) * dir
		return math.Abs(p.X-x) <= halfW && t > -0.6 && t <= length && p.Y <= surf(p.Z)
	}})
	if ramp {
		g.Use(Stone).Fill(Func{box, func(p Vec3) bool {
			t := (p.Z - zTop) * dir
			s := top - t/run*step
			return math.Abs(p.X-x) <= 1.3 && t > 0 && t <= length && p.Y <= s+0.2 && p.Y >= s-1.2
		}})
	}
	g.Use(White).Fill(Func{box, func(p Vec3) bool {
		t := (p.Z - zTop) * dir
		ax := math.Abs(p.X - x)
		s := top - t/run*step
		return ax > halfW && ax <= halfW+0.9 && t > -0.6 && t <= length+0.3 && p.Y <= s+1.2 && p.Y >= forbiddenCityG-0.5
	}})
}

func (f ForbiddenCity) Build(g *Grid) Order {
	f.site(g)
	f.river(g)
	f.outerWall(g)
	f.meridianGate(g)
	f.gateOfSupremeHarmony(g)
	f.galleries(g)
	f.dragonTerrace(g)
	f.sideHalls(g)
	f.trees(g)
	f.hollow(g)
	return Along(P(0, -1))
}

// hollow removes buried masonry cells (every face neighbor occupied): solid walls,
// bases and terraces keep their look as closed shells at a fraction of the cells.
func (ForbiddenCity) hollow(g *Grid) {
	var buried []Cell
	for _, c := range g.Cells() {
		color, _ := g.ColorOf(c)
		switch color {
		case Brick, Red, White, Stone, DarkStone:
		default:
			continue
		}
		enclosed := true
		for _, n := range Neighbors {
			if !g.Has(c.Add(n)) {
				enclosed = false
				break
			}
		}
		if enclosed {
			buried = append(buried, c)
		}
	}
	for _, c := range buried {
		g.Remove(c)
	}
}

// moatIn reports whether a plan point lies in the moat channel.
func (ForbiddenCity) moatIn(x, z, grow float64) bool {
	ix := forbiddenCityWallX + forbiddenCityStrip
	in := math.Abs(x) < ix-grow && z > forbiddenCityWallN-forbiddenCityStrip+grow && z < forbiddenCityMoatS-grow
	out := math.Abs(x) < ix+forbiddenCityMoatW+grow && z > forbiddenCityWallN-forbiddenCityStrip-forbiddenCityMoatW-grow &&
		z < forbiddenCityMoatS+forbiddenCityMoatW+grow
	return out && !in
}

// siteIn reports whether a plan point lies on the site, inset by d.
func (ForbiddenCity) siteIn(x, z, d float64) bool {
	e := forbiddenCityMoatW + forbiddenCityRim - d
	ix := forbiddenCityWallX + forbiddenCityStrip + e
	return math.Abs(x) < ix && z > forbiddenCityWallN-forbiddenCityStrip-e && z < forbiddenCityMoatS+e
}

// site lays the paved ground, the moat with its stone embankments and the green rim.
func (f ForbiddenCity) site(g *Grid) {
	e := forbiddenCityMoatW + forbiddenCityRim
	box := Box{V(-forbiddenCityWallX-forbiddenCityStrip-e, -1.3, forbiddenCityWallN-forbiddenCityStrip-e),
		V(forbiddenCityWallX+forbiddenCityStrip+e, forbiddenCityG, forbiddenCityMoatS+e)}
	g.Use(Stone).Fill(Func{box, func(p Vec3) bool {
		return p.Y >= 0 && f.siteIn(p.X, p.Z, 0) && !f.moatIn(p.X, p.Z, 0)
	}})
	// Embankments and water.
	g.Use(DarkStone).Fill(Func{box, func(p Vec3) bool {
		return p.Y >= -1.25 && f.moatIn(p.X, p.Z, 0) && !f.moatIn(p.X, p.Z, -0.9)
	}})
	g.Use(Water).Fill(Func{box, func(p Vec3) bool {
		return p.Y >= -0.85 && p.Y <= -0.35 && f.moatIn(p.X, p.Z, -0.9)
	}})
	top := func(fn func(x, z float64) bool) Solid {
		return Func{Box{V(box.Min.X, forbiddenCityG-0.1, box.Min.Z), V(box.Max.X, forbiddenCityG, box.Max.Z)}, func(p Vec3) bool {
			return fn(p.X, p.Z)
		}}
	}
	inWalls := func(x, z float64) bool {
		return math.Abs(x) < forbiddenCityWallX && z > forbiddenCityWallN && z < forbiddenCityWallS
	}
	// Green rim with a dark curb beyond the moat; a road between moat and wall.
	g.Paint(Grass, top(func(x, z float64) bool {
		return !f.siteIn(x, z, e-0.1) && f.siteIn(x, z, 0.8)
	}))
	g.Paint(DarkStone, top(func(x, z float64) bool {
		return !inWalls(x, z) && f.siteIn(x, z, e-0.1) && !(math.Abs(x) < 15.5 && z > forbiddenCityWallS)
	}))
	// Groves flanking the Meridian Gate forecourt (Ancestral Temple and Altar parks).
	g.Paint(Grass, top(func(x, z float64) bool {
		ax := math.Abs(x)
		return ax > 16.6 && ax < forbiddenCityWallX+0.6 && z > forbiddenCityWallS+0.8 && z < forbiddenCityMoatS-0.8
	}))
	// Side strips inside the wall: lawns.
	g.Paint(Grass, top(func(x, z float64) bool {
		ax := math.Abs(x)
		return inWalls(x, z) && ax > forbiddenCityCourtX+forbiddenCityGallery+0.6 && ax < forbiddenCityWallX-forbiddenCityWallT-0.6 &&
			z < forbiddenCityMeridN-1 && z > forbiddenCityWallN+forbiddenCityWallT+0.6
	}))
	// Court paving joints and the white imperial way on the axis.
	g.Paint(DarkStone, top(func(x, z float64) bool {
		ax := math.Abs(x)
		joint := math.Mod(x+300, 3.2) < 0.5 || math.Mod(z+300, 3.2) < 0.5
		return joint && (ax < forbiddenCityCourtX && z < forbiddenCityWallS-1 && z > forbiddenCityWallN+forbiddenCityWallT ||
			ax < 10.5 && z > forbiddenCityWallS && z < forbiddenCityMoatS)
	}))
	g.Paint(White, top(func(x, z float64) bool {
		return math.Abs(x) < 1.6 && z > forbiddenCityWallN && z < forbiddenCityMoatS+e-0.8
	}))
	// Bridge over the south moat on the axis.
	g.Use(White).Fill(Func{Box{V(-4.5, -0.9, forbiddenCityMoatS-0.5), V(4.5, forbiddenCityG+1.3, forbiddenCityMoatS+forbiddenCityMoatW+0.5)}, func(p Vec3) bool {
		t := (p.Z - forbiddenCityMoatS - forbiddenCityMoatW/2) / (forbiddenCityMoatW/2 + 0.5)
		s := forbiddenCityG + 0.5*(1-t*t)
		ax := math.Abs(p.X)
		if ax <= 3.6 {
			return p.Y <= s && p.Y >= s-0.9
		}
		return p.Y <= s+0.9 && p.Y >= s-0.9
	}})
	g.Paint(Stone, Func{Box{V(-1.6, 0, forbiddenCityMoatS-0.5), V(1.6, forbiddenCityG+1.3, forbiddenCityMoatS+forbiddenCityMoatW+0.5)}, func(p Vec3) bool {
		return true
	}})
}

// riverDist is the signed offset of a plan point from the Golden Water River's arc.
func (ForbiddenCity) riverDist(x, z float64) (float64, float64) {
	dz := z - forbiddenCityRiverC
	return math.Hypot(x, dz) - forbiddenCityRiverR, math.Atan2(x, dz) * forbiddenCityRiverR
}

// river carves the bow-shaped Golden Water River across the outer court, lines it with
// marble balustrades and spans it with five white marble bridges.
func (f ForbiddenCity) river(g *Grid) {
	const water, bank = 1.0, 1.6
	xr := forbiddenCityCourtX
	box := Box{V(-xr, -1.3, 4), V(xr, forbiddenCityG+2, forbiddenCityRiverC+forbiddenCityRiverR+3)}
	channel := func(p Vec3, w float64) bool {
		d, _ := f.riverDist(p.X, p.Z)
		return math.Abs(p.X) < xr && math.Abs(d) <= w
	}
	g.Carve(Func{box, func(p Vec3) bool { return channel(p, bank) && p.Y > -2 }})
	g.Use(DarkStone).Fill(Func{box, func(p Vec3) bool { return channel(p, bank) && p.Y >= -1.25 && p.Y <= forbiddenCityG }})
	g.Carve(Func{box, func(p Vec3) bool { return channel(p, water) && p.Y > -0.9 }})
	g.Use(Water).Fill(Func{box, func(p Vec3) bool { return channel(p, water) && p.Y >= -0.85 && p.Y <= -0.35 }})
	g.Paint(White, Func{box, func(p Vec3) bool { return channel(p, bank) && p.Y > forbiddenCityG-0.45 }})
	bridges := []struct{ X, W float64 }{{0, 3.0}, {-4.6, 2.4}, {4.6, 2.4}, {-8.8, 2.2}, {8.8, 2.2}}
	sdf := func(x, z float64) (float64, float64, bool) {
		d, along := f.riverDist(x, z)
		if math.Abs(x) > xr-0.4 {
			return 0, 0, false
		}
		for _, b := range bridges {
			if math.Abs(x-b.X) < b.W/2+0.3 {
				return 0, 0, false
			}
		}
		return bank - math.Abs(d) - 0.85, along, true
	}
	// Balustrades just outside the channel on both banks.
	g.Use(White).Fill(Func{box, func(p Vec3) bool {
		s, along, ok := sdf(p.X, p.Z)
		if !ok || s < -0.85 || s > 0 {
			return false
		}
		post := math.Mod(along+300, 2.0) < 0.8
		return p.Y > forbiddenCityG && (p.Y <= forbiddenCityG+0.45 || post && p.Y <= forbiddenCityG+1.25)
	}})
	for _, b := range bridges {
		zc := forbiddenCityRiverC + math.Sqrt(forbiddenCityRiverR*forbiddenCityRiverR-b.X*b.X)
		const half = 3.0
		bb := Box{V(b.X-b.W/2-0.2, -1, zc-half-0.2), V(b.X+b.W/2+0.2, forbiddenCityG+2.5, zc+half+0.2)}
		deck := func(z float64) float64 {
			t := (z - zc) / half
			return forbiddenCityG + 0.9*(1-t*t)
		}
		g.Use(White).Fill(Func{bb, func(p Vec3) bool {
			if math.Abs(p.X-b.X) > b.W/2 || math.Abs(p.Z-zc) > half {
				return false
			}
			s := deck(p.Z)
			edge := math.Abs(p.X-b.X) > b.W/2-0.75
			post := math.Mod(p.Z+300, 1.4) < 0.6
			if edge {
				return p.Y >= s-0.8 && (p.Y <= s+0.6 || post && p.Y <= s+1.2)
			}
			return p.Y >= s-0.8 && p.Y <= s
		}})
		if b.X == 0 {
			g.Paint(Stone, Func{bb, func(p Vec3) bool { return math.Abs(p.X) < 0.8 && p.Y > deck(p.Z)-0.45 && p.Y <= deck(p.Z) }})
		}
	}
}

// outerWall builds the purple-red outer wall with its crenellated outer parapet, gold
// coped inner parapet, the four corner towers and the east, west and north gates.
func (f ForbiddenCity) outerWall(g *Grid) {
	x, n, s, t := forbiddenCityWallX, forbiddenCityWallN, forbiddenCityWallS, forbiddenCityWallT
	h := forbiddenCityG + forbiddenCityWallH
	box := Box{V(-x, 0, n), V(x, h+1.2, s)}
	ring := func(p Vec3, out, in float64) (bool, float64) {
		inOuter := math.Abs(p.X) <= x-out && p.Z >= n+out && p.Z <= s-out
		inInner := math.Abs(p.X) < x-in && p.Z > n+in && p.Z < s-in
		along := p.X
		if math.Abs(p.X) > x-t-0.01 {
			along = p.Z
		}
		return inOuter && !inInner, along
	}
	g.Use(Brick).Fill(Func{box, func(p Vec3) bool {
		ok, _ := ring(p, 0, t)
		return ok && p.Y <= h
	}})
	g.Paint(DarkStone, Func{box, func(p Vec3) bool {
		ok, _ := ring(p, 0, t)
		return ok && p.Y <= forbiddenCityG+0.5
	}})
	g.Paint(Stone, Func{box, func(p Vec3) bool {
		ok, _ := ring(p, 0, t)
		return ok && p.Y > h-0.4
	}})
	// Crenellated outer parapet and low inner parapet with gold coping.
	g.Use(Brick).Fill(Func{box, func(p Vec3) bool {
		ok, along := ring(p, 0, 0.75)
		return ok && p.Y > h && (p.Y <= h+0.5 || p.Y <= h+1.1 && math.Mod(along+300, 1.6) < 0.9)
	}})
	g.Use(Gold).Fill(Func{box, func(p Vec3) bool {
		ok, _ := ring(p, t-0.6, t)
		return ok && p.Y > h && p.Y <= h+0.5
	}})
	// Corner towers.
	for _, sx := range []float64{-1, 1} {
		for _, z := range []float64{n + t/2, s - t/2} {
			f.cornerTower(g, P(sx*(x-t/2), z), h)
		}
	}
	// Gate of Divine Prowess (north), East and West Flowery Gates.
	f.wallGate(g, P(0, n+t/2), 7.5, 2.6, h, false, 5.4)
	for _, sx := range []float64{-1, 1} {
		f.wallGate(g, P(sx*(x-t/2), 5), 2.6, 5.5, h, true, 3.4)
	}
}

// wallGate is a gate tower straddling the wall: a red base with arched passages and a
// double-eave hall on top.
func (f ForbiddenCity) wallGate(g *Grid, c Vec2, hx, hz, h float64, alongZ bool, body float64) {
	g.Use(Brick).Fill(BlockAt(c, 2*hx, 2*hz, 0, h))
	g.Paint(DarkStone, BlockAt(c, 2*hx, 2*hz, 0, forbiddenCityG+0.5))
	g.Paint(Stone, BlockAt(c, 2*hx, 2*hz, h-0.4, h))
	if alongZ {
		g.Carve(Arch{c.At(forbiddenCityG), 1.3, 1.6, 2*hx + 1, AxisX})
		f.hall(g, forbiddenCityHall{C: c, Hx: hx - 0.8, Hz: body, Y0: h, Wall: 3, Ov: 1.6, Double: true, Upper: 1.6, Rise: 0.85, Doors: true})
		return
	}
	for _, px := range []float64{-3.4, 0, 3.4} {
		g.Carve(Arch{V(c.X+px, forbiddenCityG, c.Y), 1.1, 1.6, 2*hz + 1, AxisZ})
	}
	f.hall(g, forbiddenCityHall{C: c, Hx: body, Hz: hz - 0.6, Y0: h, Wall: 3, Ov: 1.6, Double: true, Upper: 1.6, Rise: 0.85, Doors: true})
}

// cornerTower is a watchtower on a wall corner with the famous cross-ridged roof:
// a skirt eave around a cross plan, four crossing xieshan gables and a gold finial.
func (f ForbiddenCity) cornerTower(g *Grid, c Vec2, h float64) {
	const base, ax, az = 7.6, 3.2, 2.0 // Platform size; cross-arm half sizes.
	g.Use(Brick).Fill(BlockAt(c, base, base, 0, h))
	g.Paint(Stone, BlockAt(c, base, base, h-0.4, h))
	sdf := func(x, z float64) (float64, float64, bool) {
		d, along := f.rectSDF(x, z, []forbiddenCityRect{{c.X, c.Y, base / 2, base / 2}}, 0)
		return d, along, true
	}
	f.balustrade(g, BlockAt(c, base, base, h, h).Bounds(), sdf, h, false)
	inCross := func(x, z, a, b float64) bool {
		dx, dz := math.Abs(x-c.X), math.Abs(z-c.Y)
		return dx < a && dz < b || dx < b && dz < a
	}
	y0 := h
	top := y0 + 3.0
	g.Use(White).Fill(BlockAt(c, 2*ax+0.8, 2*az+0.8, y0-0.2, y0+0.3), BlockAt(c, 2*az+0.8, 2*ax+0.8, y0-0.2, y0+0.3))
	g.Use(Red).Fill(BlockAt(c, 2*ax, 2*az, y0, top), BlockAt(c, 2*az, 2*ax, y0, top))
	// Lattice windows on the four arm ends.
	for k := 0; k < 4; k++ {
		dir := Polar(P(0, 0), 1, float64(k)*math.Pi/2)
		at := c.Add(dir.Scale(ax))
		w, d := 0.8, 2*az-0.8
		if math.Abs(dir.X) < 0.5 {
			w, d = d, w
		}
		g.Carve(Func{BlockAt(at, w, d-0.6, y0+0.8, top-1.0).Bounds(), func(p Vec3) bool {
			return int(math.Floor((p.X+p.Z+200)/0.8))%2 == 0 && int(math.Floor((p.Y+200)/0.8))%2 == 0
		}})
	}
	g.Use(Navy).Fill(Func{BlockAt(c, 2*ax+1, 2*ax+1, top-0.9, top+0.4).Bounds(), func(p Vec3) bool {
		return inCross(p.X, p.Z, ax+0.5, az+0.5) && !inCross(p.X, p.Z, ax-1, az-1)
	}})
	g.Paint(Grass, Func{BlockAt(c, 2*ax+1, 2*ax+1, top-0.4, top+0.4).Bounds(), func(p Vec3) bool {
		return int(math.Floor((p.X+p.Z+200)/1.6))%2 == 0
	}})
	ye := top + 0.3
	sa := forbiddenCityRoof{C: c, Hx: ax + 1.6, Hz: az + 1.6, Ye: ye, Rise: 1.2, Reach: 1.6, Lift: 0.9}
	sb := forbiddenCityRoof{C: c, Hx: az + 1.6, Hz: ax + 1.6, Ye: ye, Rise: 1.2, Reach: 1.6, Lift: 0.9}
	both := f.surface(sa, sb)
	skirt := func(x, z float64) (float64, bool) {
		if inCross(x, z, ax-0.6, az-0.6) {
			return 0, false
		}
		y, ok := both(x, z)
		return math.Min(y, ye+1.2), ok
	}
	f.shell(g, Gold, skirt, sa.box(1).Union(sb.box(1)), 1.0)
	// Upper storey and the cross of xieshan roofs with four gables.
	const ux, uz = 2.6, 1.5
	top2 := ye + 1.2 + 1.5
	g.Use(Red).Fill(BlockAt(c, 2*ux, 2*uz, ye, top2), BlockAt(c, 2*uz, 2*ux, ye, top2))
	g.Use(Navy).Fill(Func{BlockAt(c, 2*ux+1, 2*ux+1, top2-0.7, top2+0.3).Bounds(), func(p Vec3) bool {
		return inCross(p.X, p.Z, ux+0.4, uz+0.4) && !inCross(p.X, p.Z, ux-1, uz-1)
	}})
	ra := forbiddenCityRoof{C: c, Hx: ux + 1.6, Hz: uz + 1.3, Ye: top2 + 0.3, Rise: 3.0, Gable: 1.5, Lift: 0.9, Ridges: true}
	rb := forbiddenCityRoof{C: c, Hx: uz + 1.3, Hz: ux + 1.6, Ye: top2 + 0.3, Rise: 3.0, Gable: 1.5, Lift: 0.9, Ridges: true}
	f.gables(g, ra)
	f.gables(g, rb)
	f.shell(g, Gold, f.surface(ra, rb), ra.box(1).Union(rb.box(1)), 1.0)
	apex := ra.Ye + ra.Rise + 0.5
	run := ra.Hx - ra.Gable
	g.Use(Gold).Fill(
		Capsule{V(c.X-run, apex, c.Y), V(c.X+run, apex, c.Y), 0.5},
		Capsule{V(c.X, apex, c.Y-run), V(c.X, apex, c.Y+run), 0.5},
		Cylinder{Center: c, Radius: 0.45, Y0: apex, Y1: apex + 1.2},
		Sphere(c.At(apex+1.7), 0.8),
	)
	for k := 0; k < 4; k++ {
		e := Polar(c, run, float64(k)*math.Pi/2)
		g.Fill(Capsule{e.At(apex - 0.2), e.At(apex + 1.0), 0.5})
	}
}

// meridianGate builds the U-shaped Meridian Gate: a massive red base with three arched
// passages, a nine-bay double-eave hall and four pavilions on the projecting wings
// joined by galleries.
func (f ForbiddenCity) meridianGate(g *Grid) {
	nz, sz, tip := forbiddenCityMeridN, forbiddenCityMeridS, forbiddenCityWingS
	h := forbiddenCityG + forbiddenCityWallH
	const ox, wx = 15.0, 10.5 // Outer and inner x of the wings.
	parts := []forbiddenCityRect{
		{0, (nz + sz) / 2, ox, (sz - nz) / 2},
		{-(ox + wx) / 2, (sz + tip) / 2, (ox - wx) / 2, (tip-sz)/2 + 0.6},
		{(ox + wx) / 2, (sz + tip) / 2, (ox - wx) / 2, (tip-sz)/2 + 0.6},
	}
	box := Box{V(-ox-1.5, 0, nz-1.5), V(ox+1.5, h+2, tip+1.5)}
	sd := func(x, z float64) (float64, float64) { return f.rectSDF(x, z, parts, 0) }
	g.Use(Red).Fill(Func{box, func(p Vec3) bool {
		d, _ := sd(p.X, p.Z)
		return d <= 0 && (d >= -1.3 || p.Y >= h-0.5) && p.Y <= h
	}})
	// White marble plinth course and cap; a dark foot.
	g.Use(White).Fill(Func{box, func(p Vec3) bool {
		d, _ := sd(p.X, p.Z)
		return d <= 0.3 && d >= -1.3 && p.Y >= forbiddenCityG-0.5 && p.Y <= forbiddenCityG+0.9
	}})
	g.Paint(White, Func{box, func(p Vec3) bool {
		d, _ := sd(p.X, p.Z)
		return d <= 0 && p.Y > h-0.45
	}})
	g.Paint(Stone, Func{box, func(p Vec3) bool {
		d, _ := sd(p.X, p.Z)
		return d <= -1.0 && p.Y > h-0.45
	}})
	// Low crenellated parapet around the top edge.
	g.Use(Red).Fill(Func{box, func(p Vec3) bool {
		d, along := sd(p.X, p.Z)
		return d <= 0 && d >= -0.7 && p.Y > h && (p.Y <= h+0.5 || p.Y <= h+1.0 && math.Mod(along+300, 1.6) < 0.9)
	}})
	// Three passages: the central one for the emperor.
	g.Carve(Arch{V(0, forbiddenCityG, (nz+sz)/2), 1.5, 2.0, sz - nz + 2, AxisZ})
	for _, sx := range []float64{-1, 1} {
		g.Carve(Arch{V(sx*5, forbiddenCityG, (nz+sz)/2), 1.1, 1.6, sz - nz + 2, AxisZ})
	}
	g.Paint(DarkStone, Func{Box{V(-7, 0, nz-1), V(7, 5, sz+1)}, func(p Vec3) bool {
		ax := math.Abs(p.X)
		return p.Y < forbiddenCityG+0.5 && (ax < 1.6 || math.Abs(ax-5) < 1.2)
	}})
	// Main hall.
	f.hall(g, forbiddenCityHall{C: P(0, (nz+sz)/2), Hx: 9.5, Hz: 2.4, Y0: h, Wall: 3.6, Ov: 1.8, Double: true, Upper: 2.0, Rise: 0.95, Doors: true})
	// Wing pavilions and connecting galleries.
	for _, sx := range []float64{-1, 1} {
		xc := sx * (ox + wx) / 2
		for _, z := range []float64{sz + 1.5, tip - 1.3} {
			f.hall(g, forbiddenCityHall{C: P(xc, z), Hx: 1.9, Hz: 1.9, Y0: h, Wall: 2.6, Ov: 1.4, Double: true, Upper: 1.2, Rise: 1.1})
		}
		f.hall(g, forbiddenCityHall{C: P(xc, (sz+tip)/2+0.1), Hx: 1.2, Hz: (tip - sz - 7.0) / 2, Y0: h, Wall: 2.4, Ov: 1.0, Gable: 0.1, Rise: 0.9})
		f.hall(g, forbiddenCityHall{C: P(sx*7.5, sz-1.6), Hx: 2.2, Hz: 1.0, Y0: h, Wall: 2.2, Ov: 0.9, Gable: 0.1, Rise: 0.9})
	}
}

// gateOfSupremeHarmony builds the gate hall on its marble base with stairs, the two
// flanking side gates and the walls that close the outer court.
func (f ForbiddenCity) gateOfSupremeHarmony(g *Grid) {
	c := P(0, 0)
	base := forbiddenCityG + 1.6
	rects := []forbiddenCityRect{{0, 0, 12, 4}}
	box := Box{V(-13.5, 0, -5.5), V(13.5, base+2, 5.5)}
	sdf := func(x, z float64) (float64, float64, bool) {
		d, along := f.rectSDF(x, z, rects, 0)
		return d, along, true
	}
	g.Use(White).Fill(Func{box, func(p Vec3) bool {
		d, _, _ := sdf(p.X, p.Z)
		return d <= 0 && p.Y >= 0 && p.Y <= base
	}})
	g.Paint(Stone, Func{box, func(p Vec3) bool {
		d, _, _ := sdf(p.X, p.Z)
		return d <= -0.9 && p.Y > base-0.45
	}})
	f.balustrade(g, box, sdf, base, true)
	for _, dir := range []float64{1, -1} {
		f.stair(g, 0, 3.2, dir*4, base, dir, true)
		for _, sx := range []float64{-1, 1} {
			f.stair(g, sx*7.5, 1.4, dir*4, base, dir, false)
		}
	}
	f.hall(g, forbiddenCityHall{C: c, Hx: 8.2, Hz: 2.6, Y0: base, Wall: 3.6, Ov: 2.0, Double: true, Upper: 1.8, Gable: 2.0, Rise: 1.0, Doors: true})
	// Bronze lions on plinths in front.
	for _, sx := range []float64{-1, 1} {
		at := P(sx*4.3, 6.4)
		g.Use(White).Fill(BlockAt(at, 1.6, 1.6, 0, forbiddenCityG+0.9))
		g.Use(DarkStone).Fill(Ellipsoid{at.At(forbiddenCityG + 1.6), V(0.7, 0.9, 0.9)})
	}
	// Side gates and connecting walls.
	for _, sx := range []float64{-1, 1} {
		at := P(sx*16.5, 0)
		g.Use(White).Fill(BlockAt(at, 6.4, 5, 0, forbiddenCityG+0.8))
		f.hall(g, forbiddenCityHall{C: at, Hx: 2.4, Hz: 1.6, Y0: forbiddenCityG + 0.8, Wall: 3.0, Ov: 1.3, Gable: 0.8, Rise: 0.95, Doors: true})
		g.Carve(Arch{V(at.X, forbiddenCityG+0.8, 0), 0.9, 1.4, 5, AxisZ})
		for _, seg := range [][2]float64{{12, 14.1}, {18.9, forbiddenCityCourtX}} {
			a, b := sx*seg[0], sx*seg[1]
			f.lowWall(g, P(math.Min(a, b), -0.6), P(math.Max(a, b), 0.6), 3.4, AxisX)
		}
	}
}

// lowWall is a red courtyard wall with a gold tiled coping.
func (f ForbiddenCity) lowWall(g *Grid, a, b Vec2, h float64, ridge Axis) {
	y0 := forbiddenCityG
	g.Use(Red).Fill(Block(a.At(0), b.At(y0+h)))
	g.Paint(DarkStone, Block(a.At(0), b.At(y0+0.5)))
	lo, hi := P(a.X-0.7, a.Y-0.7), P(b.X+0.7, b.Y+0.7)
	roof := Gable{lo, hi, y0 + h - 0.3, ridge}
	g.Use(Gold).Fill(Subtract(roof, Translate(roof, V(0, -0.9, 0))))
}

// galleries builds the long side galleries enclosing the courts on east and west,
// with gate halls in the outer court and the two-storey pavilions of the great court.
func (f ForbiddenCity) galleries(g *Grid) {
	hx := forbiddenCityGallery / 2
	xc := forbiddenCityCourtX + hx
	north := forbiddenCityWallN + forbiddenCityWallT
	segs := [][2]float64{{forbiddenCityMeridN - 0.6, 12.6}, {7.6, 0.8}, {-0.8, -6.2}, {-14.2, north + 0.5}}
	for _, sx := range []float64{-1, 1} {
		for _, s := range segs {
			z0, z1 := s[1], s[0]
			f.hall(g, forbiddenCityHall{C: P(sx*xc, (z0+z1)/2), Hx: hx, Hz: (z1 - z0) / 2, Y0: forbiddenCityG, Wall: 3.0, Ov: 0.9, Gable: 0.1, Rise: 0.85})
		}
		// Xiehe / Xihe gates in the outer court.
		at := P(sx*xc, 10.1)
		f.hall(g, forbiddenCityHall{C: at, Hx: 2.0, Hz: 2.4, Y0: forbiddenCityG, Wall: 3.6, Ov: 1.3, Gable: 0.8, Rise: 0.95, Doors: true})
		g.Carve(Arch{V(at.X, forbiddenCityG, at.Y), 1.0, 1.6, 6, AxisX})
		// Tiren / Hongyi pavilions on white bases.
		pc := P(sx*(xc-0.4), -10.2)
		g.Use(White).Fill(BlockAt(pc, 7.2, 9.6, 0, forbiddenCityG+1.0))
		g.Paint(Stone, BlockAt(pc, 7.2, 9.6, forbiddenCityG+0.6, forbiddenCityG+1.0))
		f.hall(g, forbiddenCityHall{C: pc, Hx: 2.8, Hz: 3.6, Y0: forbiddenCityG + 1.0, Wall: 3.4, Ov: 1.4, Double: true, Upper: 2.0, Gable: 1.0, Rise: 1.0, Doors: true})
	}
	// Gilded water vats in the great court.
	for _, sx := range []float64{-1, 1} {
		for _, z := range []float64{-3.2, -12} {
			at := P(sx*18.5, z)
			g.Use(Gold).Fill(Cylinder{Center: at, Radius: 0.9, Y0: forbiddenCityG, Y1: forbiddenCityG + 1.4})
		}
	}
}

// dragonTerrace builds the three-tier, I-shaped white marble terrace with sumeru
// mouldings, balustrades, spouts and stairs, and the three great halls on top.
func (f ForbiddenCity) dragonTerrace(g *Grid) {
	rects := []forbiddenCityRect{
		{0, -26, 17, 10},  // Front block: Hall of Supreme Harmony.
		{0, -35, 11, 17},  // Waist.
		{0, -48, 15, 7.5}, // Rear block: Hall of Preserving Harmony.
	}
	box := Box{V(-18, 0, -57), V(18, 12, -15)}
	var tops []float64
	for i := 0; i < 3; i++ {
		inset := float64(i) * forbiddenCityTierIn
		yb, yt := forbiddenCityG+float64(i)*forbiddenCityTierH, forbiddenCityG+float64(i+1)*forbiddenCityTierH
		tops = append(tops, yt)
		sd := func(x, z float64) (float64, float64) { return f.rectSDF(x, z, rects, inset) }
		g.Use(White).Fill(Func{box, func(p Vec3) bool {
			d, _ := sd(p.X, p.Z)
			return d <= 0 && d >= -1.3 && p.Y >= yb-0.4 && p.Y <= yt
		}})
		g.Use(Stone).Fill(Func{box, func(p Vec3) bool {
			d, _ := sd(p.X, p.Z)
			return d < -1.3 && p.Y > yt-0.45 && p.Y <= yt
		}})
		// Sumeru base: a projecting foot moulding and a recessed waist.
		g.Use(White).Fill(Func{box, func(p Vec3) bool {
			d, _ := sd(p.X, p.Z)
			return d > 0 && d <= 0.4 && p.Y >= yb-0.2 && p.Y <= yb+0.35
		}})
		g.Carve(Func{box, func(p Vec3) bool {
			d, _ := sd(p.X, p.Z)
			return d > -0.4 && d <= 0 && p.Y > yb+0.5 && p.Y < yt-0.75
		}})
		g.Paint(Stone, Func{box, func(p Vec3) bool {
			d, _ := sd(p.X, p.Z)
			return d > -1.3 && d <= -0.4 && p.Y > yb+0.5 && p.Y < yt-0.75
		}})
		f.balustrade(g, box, func(x, z float64) (float64, float64, bool) {
			d, along := sd(x, z)
			return d, along, true
		}, yt, true)
	}
	top := tops[2]
	// Stairs: three flights on the south, one on the north.
	f.stair(g, 0, 3.6, -16, top, 1, true)
	for _, sx := range []float64{-1, 1} {
		f.stair(g, sx*10.5, 1.5, -16, top, 1, false)
	}
	f.stair(g, 0, 3.0, -55.5, top, -1, true)
	// Hall of Supreme Harmony.
	f.hall(g, forbiddenCityHall{C: P(0, -27), Hx: 11, Hz: 4.6, Y0: top + 0.4, Wall: 5.6, Ov: 2.6, Double: true, Upper: 3.0, Rise: 1.2, Doors: true})
	g.Use(White).Fill(BlockAt(P(0, -27), 24.4, 11.6, top-0.2, top+0.4))
	// Bronze incense burners, crane and tortoise on the front deck.
	for _, sx := range []float64{-1, 1} {
		for k, x := range []float64{4, 8} {
			at := P(sx*x, -20.8+float64(k)*0.4)
			g.Use(DarkStone).Fill(Cylinder{Center: at, Radius: 0.6, Y0: top, Y1: top + 1.4})
			g.Use(Gold).Fill(Sphere(at.At(top+1.7), 0.5))
		}
	}
	// Hall of Central Harmony: square, single eave, pyramid roof with gilt finial.
	f.hall(g, forbiddenCityHall{C: P(0, -38.5), Hx: 3.0, Hz: 3.0, Y0: top, Wall: 4.2, Ov: 1.8, Rise: 1.15, Doors: true})
	// Hall of Preserving Harmony: double-eave xieshan.
	f.hall(g, forbiddenCityHall{C: P(0, -49), Hx: 8.4, Hz: 2.8, Y0: top + 0.2, Wall: 4.6, Ov: 2.2, Double: true, Upper: 2.0, Gable: 2.2, Rise: 1.1, Doors: true})
}

// sideHalls fills the lawns between the galleries and the outer wall with smaller
// halls: the Hall of Literary Glory (east), Martial Valor (west) and storehouses.
func (f ForbiddenCity) sideHalls(g *Grid) {
	xc := (forbiddenCityCourtX + forbiddenCityGallery + forbiddenCityWallX - forbiddenCityWallT) / 2
	for _, sx := range []float64{-1, 1} {
		for k, z := range []float64{-6, -27, -47} {
			at := P(sx*xc, z)
			hz := 4.2
			if k == 0 {
				hz = 5.4
			}
			g.Use(White).Fill(BlockAt(at, 4.4, 2*hz+1.4, 0, forbiddenCityG+0.6))
			f.hall(g, forbiddenCityHall{C: at, Hx: 1.5, Hz: hz, Y0: forbiddenCityG + 0.6, Wall: 3.0, Ov: 1.0, Gable: 1.0, Rise: 0.95, Doors: true})
		}
	}
}

// trees plants cypresses on the side lawns and willows along the moat's outer rim.
func (f ForbiddenCity) trees(g *Grid) {
	xc := (forbiddenCityCourtX + forbiddenCityGallery + forbiddenCityWallX - forbiddenCityWallT) / 2
	cypress := func(at Vec2, h float64) {
		base := forbiddenCityG
		g.Use(Wood).Fill(Cylinder{Center: at, Radius: 0.45, Y0: base, Y1: base + 1.4})
		g.Use(Foliage).FillVolume(Frustum{at, 1.4, 0.2, base + 1.0, base + h}, Ellipsoid{at.At(base + 1.7), V(1.4, 0.9, 1.4)})
	}
	for _, sx := range []float64{-1, 1} {
		for _, z := range []float64{9, 4, -14, -17, -35.5, -55, -59} {
			k := int(z*10) + int(sx)
			cypress(P(sx*xc+(Hash01(k, 1)-0.5)*1.4, z+(Hash01(k, 2)-0.5)), 4+2*Hash01(k, 3))
		}
	}
	// Groves beside the Meridian Gate forecourt.
	for x := 18.6; x < forbiddenCityWallX; x += 3.3 {
		for z := forbiddenCityWallS + 2.6; z < forbiddenCityMoatS-1.5; z += 3.3 {
			for _, sx := range []float64{-1, 1} {
				k := int(x*10)*7 + int(z*10) + int(sx)
				if Hash01(k, 5) < 0.3 {
					continue
				}
				cypress(P(sx*x+(Hash01(k, 1)-0.5)*1.2, z+(Hash01(k, 2)-0.5)*1.2), 3.6+2.4*Hash01(k, 3))
			}
		}
	}
	// Willows on the green rim beyond the moat.
	e := forbiddenCityMoatW + forbiddenCityRim
	ix := forbiddenCityWallX + forbiddenCityStrip + e - 1.3
	zn, zs := forbiddenCityWallN-forbiddenCityStrip-e+1.3, forbiddenCityMoatS+e-1.3
	var spots []Vec2
	for z := zn + 3; z < zs-2; z += 7 {
		spots = append(spots, P(-ix, z), P(ix, z))
	}
	for x := -ix + 4; x < ix-2; x += 7 {
		if math.Abs(x) > 8 {
			spots = append(spots, P(x, zs))
		}
		spots = append(spots, P(x, zn))
	}
	for i, at := range spots {
		if Hash01(i, 7) < 0.25 {
			continue
		}
		h := 2.6 + 1.2*Hash01(i, 8)
		g.Use(Wood).Fill(Cylinder{Center: at, Radius: 0.45, Y0: forbiddenCityG, Y1: forbiddenCityG + h})
		g.Use(Foliage).FillVolume(Ellipsoid{at.At(forbiddenCityG + h + 0.4), V(1.6, 1.5, 1.6)})
		g.Paint(Grass, Ellipsoid{at.At(forbiddenCityG + h + 0.9), V(1.7, 1.0, 1.7)})
	}
}
