package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// ArcDeTriomphe is the triumphal arch on the Place Charles de Gaulle (the Étoile) in
// Paris. The arch is a hollow limestone block with the great barrel vault through its
// long faces (±Z) and the smaller transverse vaults through its short faces (±X), dark
// coffered intrados, four sculpture groups on the piers, relief panels, Fames in the
// spandrels, the frieze, cornice and attic with its 30 shields, and a parapeted roof. A
// huge tricolour hangs in the main arch above the Unknown Soldier's tomb and eternal
// flame. Around it: the cobbled roundabout with traffic and red buses, twelve
// tree-lined avenues (the Champs-Élysées runs to +Z), and Haussmann blocks with
// mansard roofs filling the wedges between them.
type ArcDeTriomphe struct{}

func init() { Register(ArcDeTriomphe{}) }

func (ArcDeTriomphe) Name() string  { return "arc_de_triomphe" }
func (ArcDeTriomphe) Title() string { return "Arc de Triomphe" }

const (
	arcDeTriompheM     = 0.85                                      // World units per meter.
	arcDeTriompheBase  = 0.8                                       // Arch floor: plinth top.
	arcDeTriompheHX    = 45 * arcDeTriompheM / 2                   // Half width along X.
	arcDeTriompheHZ    = 22 * arcDeTriompheM / 2                   // Half depth along Z.
	arcDeTriompheMainR = 14.62 * arcDeTriompheM / 2                // Main arch half span.
	arcDeTriompheMainS = 29.19*arcDeTriompheM - arcDeTriompheMainR // Main arch spring above base.
	arcDeTriompheSideR = 8.44 * arcDeTriompheM / 2
	arcDeTriompheSideS = 18.68*arcDeTriompheM - arcDeTriompheSideR
	arcDeTriompheRoof  = 41.3 // Roof level above base.
	arcDeTriompheWall  = 1.0

	arcDeTriompheIsland  = 25.0 // Paved central island.
	arcDeTriompheRoadOut = 42.5 // Outer edge of the roundabout.
	arcDeTriompheRin     = 44.0 // Building frontage on the place.
	arcDeTriompheRout    = 52.0 // Back of the blocks.
	arcDeTriompheEdge    = 54.0 // City plate.
	arcDeTriompheRim     = 56.0 // Lower stepped rim.

	arcDeTriompheGround = 2.8 // Ground floor height of the blocks.
	arcDeTriompheStorey = 2.0
	arcDeTriompheMans   = 2.8 // Mansard height.
)

func (a ArcDeTriomphe) Build(g *Grid) Order {
	a.ground(g)
	a.arch(g)
	a.trees(g)
	a.buildings(g)
	a.traffic(g)
	return Outward(P(0, 0))
}

// raster sets every cell whose center lies in the box and for which f reports a color,
// without half-unit snapping.
func (ArcDeTriomphe) raster(g *Grid, b Box, f func(Vec3) (Color, bool)) {
	s := g.Span()
	lo := func(v float64) int32 { return int32(math.Ceil(v * 2 / s)) }
	hi := func(v float64) int32 { return int32(math.Floor(v * 2 / s)) }
	previous := g.Brush()
	for y := lo(b.Min.Y); y <= hi(b.Max.Y); y++ {
		z0, x0 := lo(b.Min.Z), lo(b.Min.X)
		if (z0^y)&1 != 0 {
			z0++
		}
		if (x0^y)&1 != 0 {
			x0++
		}
		for z := z0; z <= hi(b.Max.Z); z += 2 {
			for x := x0; x <= hi(b.Max.X); x += 2 {
				c := Cell{X: x, Y: y, Z: z}
				if color, ok := f(g.World(c)); ok {
					g.Remove(c)
					g.Use(color).Add(c)
				}
			}
		}
	}
	g.Use(previous)
}

// recolor changes the color of occupied cells inside the box.
func (ArcDeTriomphe) recolor(g *Grid, b Box, f func(p Vec3, c Color) Color) {
	previous := g.Brush()
	for _, c := range g.Cells() {
		p := g.World(c)
		if !b.Contains(p) {
			continue
		}
		old, _ := g.ColorOf(c)
		if n := f(p, old); n != old {
			g.Remove(c)
			g.Use(n).Add(c)
		}
	}
	g.Use(previous)
}

// frac is the positive fractional part of v / period.
func (ArcDeTriomphe) frac(v, period float64) float64 {
	f := math.Mod(v/period, 1)
	if f < 0 {
		f++
	}
	return f
}

// onFace maps a solid authored in face-local coordinates (X along the face, Y up, Z
// outward from the face plane) onto one of the arch's faces: 0 +Z, 1 -Z, 2 +X, 3 -X.
func (ArcDeTriomphe) onFace(face int, s Solid) Solid {
	hx, hz := arcDeTriompheHX, arcDeTriompheHZ
	local := func(p Vec3) Vec3 {
		switch face {
		case 0:
			return V(p.X, p.Y, p.Z-hz)
		case 1:
			return V(-p.X, p.Y, -hz-p.Z)
		case 2:
			return V(-p.Z, p.Y, p.X-hx)
		}
		return V(p.Z, p.Y, -hx-p.X)
	}
	world := func(l Vec3) Vec3 {
		switch face {
		case 0:
			return V(l.X, l.Y, hz+l.Z)
		case 1:
			return V(-l.X, l.Y, -hz-l.Z)
		case 2:
			return V(hx+l.Z, l.Y, -l.X)
		}
		return V(-hx-l.Z, l.Y, l.X)
	}
	lb := s.Bounds()
	b := EmptyBox()
	for _, x := range []float64{lb.Min.X, lb.Max.X} {
		for _, z := range []float64{lb.Min.Z, lb.Max.Z} {
			w := world(V(x, lb.Min.Y, z))
			b = b.Union(Box{w, V(w.X, lb.Max.Y, w.Z)})
		}
	}
	return Func{b, func(p Vec3) bool { return s.Contains(local(p)) }}
}

// avenueHalf is the half width between facades of avenue k (k·30° in plan).
func (ArcDeTriomphe) avenueHalf(k int) float64 {
	k = ((k % 12) + 12) % 12
	if k == 3 || k == 9 { // Champs-Élysées (+Z) and Grande Armée (-Z).
		return 7.0
	}
	return 4.6
}

// avenue returns the nearest avenue, the distance along it and across it.
func (ArcDeTriomphe) avenue(q Vec2) (int, float64, float64) {
	k := int(math.Round(q.Angle() / (math.Pi / 6)))
	k = ((k % 12) + 12) % 12
	t := float64(k) * math.Pi / 6
	return k, q.Dot(P(math.Cos(t), math.Sin(t))), q.Dot(P(-math.Sin(t), math.Cos(t)))
}

// arcDeTriompheLot describes the Haussmann block around a plan point.
type arcDeTriompheLot struct {
	wedge, half int
	d, d2       float64 // Distance inside the street faces; second-nearest face.
	face        int     // 0 place, 1 back, 2/3 avenue sides.
	u           float64 // Coordinate along the nearest street face.
	dS          float64 // Distance from the party wall between the two halves.
	r           float64
	sc          float64 // Signed distance across the wedge bisector.
	top, other  float64 // Cornice heights of this half and the neighbouring half.
	storeys     int
	facade      Color
	cafe        bool
}

func (a ArcDeTriomphe) storeys(wedge, half int) int {
	return 5 + int(Hash01(wedge, half, 7)*3)
}

func (a ArcDeTriomphe) cornice(n int) float64 {
	return arcDeTriompheGround + float64(n-1)*arcDeTriompheStorey
}

func (a ArcDeTriomphe) lot(q Vec2) arcDeTriompheLot {
	ang := math.Mod(q.Angle()+2*math.Pi, 2*math.Pi)
	w := int(ang / (math.Pi / 6))
	if w > 11 {
		w = 11
	}
	a0 := float64(w) * math.Pi / 6
	a1 := a0 + math.Pi/6
	uA, nA := P(math.Cos(a0), math.Sin(a0)), P(-math.Sin(a0), math.Cos(a0))
	uB, nB := P(math.Cos(a1), math.Sin(a1)), P(-math.Sin(a1), math.Cos(a1))
	r := q.Length()
	rel := ang - a0
	ds := [4]float64{
		r - arcDeTriompheRin,
		arcDeTriompheRout - r,
		q.Dot(nA) - a.avenueHalf(w),
		-q.Dot(nB) - a.avenueHalf(w+1),
	}
	us := [4]float64{arcDeTriompheRin * rel, arcDeTriompheRout * rel, q.Dot(uA), q.Dot(uB)}
	l := arcDeTriompheLot{wedge: w, r: r}
	l.face = 0
	for i := 1; i < 4; i++ {
		if ds[i] < ds[l.face] {
			l.face = i
		}
	}
	l.d, l.u = ds[l.face], us[l.face]
	l.d2 = math.Inf(1)
	for i := 0; i < 4; i++ {
		if i != l.face && ds[i] < l.d2 {
			l.d2 = ds[i]
		}
	}
	am := a0 + math.Pi/12
	l.sc = q.Dot(P(-math.Sin(am), math.Cos(am)))
	if l.sc >= 0 {
		l.half = 1
	}
	l.dS = math.Abs(l.sc)
	l.storeys = a.storeys(w, l.half)
	l.top = a.cornice(l.storeys)
	l.other = a.cornice(a.storeys(w, 1-l.half))
	switch h := Hash01(w, l.half, 11); {
	case h < 0.5:
		l.facade = Sand
	case h < 0.85:
		l.facade = White
	default:
		l.facade = Stone
	}
	l.cafe = Hash01(w, l.half, 13) < 0.55
	return l
}

// ground lays the city plate: the paved island, the cobbled roundabout with lane
// markings, avenues with crossings and sidewalks, and a lower stepped rim.
func (a ArcDeTriomphe) ground(g *Grid) {
	R := arcDeTriompheRim
	a.raster(g, Box{V(-R, -0.45, -R), V(R, 0.9, R)}, func(p Vec3) (Color, bool) {
		q := p.Plan()
		r := q.Length()
		if r > arcDeTriompheRim {
			return 0, false
		}
		if r > arcDeTriompheEdge {
			if p.Y > 0.05 {
				return 0, false
			}
			return DarkStone, true
		}
		if p.Y < -0.05 {
			return 0, false
		}
		curb := r >= arcDeTriompheIsland && r < arcDeTriompheIsland+0.9
		if p.Y > 0.45 && !curb {
			return 0, false
		}
		if l := a.lot(q); l.d >= 0.8 && l.dS >= 0 {
			return 0, false // Under a building.
		}
		cx, cz := int(math.Floor(p.X/0.8)), int(math.Floor(p.Z/0.8))
		h := Hash01(cx, cz, 3)
		k, along, perp := a.avenue(q)
		H := a.avenueHalf(k)
		switch {
		case curb:
			return Stone, true
		case r < arcDeTriompheIsland:
			if r > 22.2 && r < 23.0 {
				return DarkStone, true
			}
			if h < 0.1 {
				return DarkStone, true
			}
			return Stone, true
		case r < arcDeTriompheRoadOut:
			for _, lane := range []float64{29.8, 33.8, 37.8} {
				if math.Abs(r-lane) < 0.42 && a.frac(q.Angle()*lane, 4.4) < 0.5 {
					return White, true
				}
			}
			if h < 0.09 {
				return Slate, true
			}
			return Charcoal, true
		case math.Abs(perp) < H:
			if math.Abs(perp) > H-1.6 {
				return Stone, true
			}
			if along > arcDeTriompheRoadOut+0.6 && along < arcDeTriompheRoadOut+2.8 {
				if a.frac(perp, 1.4) < 0.5 {
					return White, true
				}
				return Charcoal, true
			}
			if math.Abs(perp) < 0.3 && a.frac(along, 3.0) < 0.5 {
				return White, true
			}
			if h < 0.15 {
				return Slate, true
			}
			return Charcoal, true
		}
		if h < 0.15 {
			return DarkStone, true
		}
		return Stone, true
	})
}

// mainArch and sideArch are the arch openings; grow widens them by a lining thickness.
func (ArcDeTriomphe) mainArch(grow float64) Arch {
	return Arch{V(0, arcDeTriompheBase-1, 0), arcDeTriompheMainR + grow, arcDeTriompheMainS + 1, 80, AxisZ}
}

func (ArcDeTriomphe) sideArch(grow float64) Arch {
	return Arch{V(0, arcDeTriompheBase-1, 0), arcDeTriompheSideR + grow, arcDeTriompheSideS + 1, 80, AxisX}
}

// ring returns an outward ring around the arch block from y0 to y1 (above base),
// protruding out beyond the faces and reaching in by 0.6.
func (a ArcDeTriomphe) ring(y0, y1, out float64) Solid {
	hx, hz, b := arcDeTriompheHX, arcDeTriompheHZ, arcDeTriompheBase
	return Subtract(
		Block(V(-hx-out, b+y0, -hz-out), V(hx+out, b+y1, hz+out)),
		Block(V(-hx+0.6, b+y0-1, -hz+0.6), V(hx-0.6, b+y1+1, hz-0.6)),
	)
}

func (a ArcDeTriomphe) arch(g *Grid) {
	hx, hz, b := arcDeTriompheHX, arcDeTriompheHZ, arcDeTriompheBase
	top := b + arcDeTriompheRoof
	wall := arcDeTriompheWall
	openings := []Solid{a.mainArch(0), a.sideArch(0)}

	// Plinth step under the whole arch.
	g.Use(Stone).Fill(Block(V(-hx-1.6, 0.4, -hz-1.6), V(hx+1.6, b, hz+1.6)))

	// Hollow masonry block pierced by the vaults, with lined passages.
	outer := Block(V(-hx, b, -hz), V(hx, top, hz))
	inner := Subtract(Block(V(-hx+wall, b, -hz+wall), V(hx-wall, top-1.4, hz-wall)), a.mainArch(1.3), a.sideArch(1.3))
	g.Use(Sand).Fill(Subtract(outer, append(openings, inner)...))

	// Horizontal courses: socle, imposts, architrave, frieze, cornice, attic cornice.
	g.Fill(Subtract(a.ring(0, 2.4, 0.8), openings...))
	g.Fill(Subtract(a.ring(arcDeTriompheMainS, arcDeTriompheMainS+0.8, 0.8), openings...))
	g.Fill(Subtract(Union(
		Block(V(hx-0.6, b+arcDeTriompheSideS, -hz), V(hx+0.8, b+arcDeTriompheSideS+0.7, hz)),
		Block(V(-hx-0.8, b+arcDeTriompheSideS, -hz), V(-hx+0.6, b+arcDeTriompheSideS+0.7, hz)),
	), openings...))
	g.Fill(a.ring(27.0, 27.6, 0.8))
	g.Fill(a.ring(27.6, 30.0, 0.8))
	g.Fill(a.ring(30.0, 30.7, 1.3))
	g.Fill(a.ring(30.7, 31.5, 1.8))
	g.Fill(a.ring(35.0, 35.5, 0.8))
	g.Fill(a.ring(39.9, 40.6, 0.9))
	g.Fill(a.ring(40.6, 41.3, 1.4))
	// Roof parapet and the small rooftop pavilion.
	g.Fill(Subtract(Block(V(-hx-0.6, top, -hz-0.6), V(hx+0.6, top+1.0, hz+0.6)),
		Block(V(-hx+0.6, top-1, -hz+0.6), V(hx-0.6, top+2, hz-0.6))))
	g.Fill(Block(V(-3, top, -2.2), V(3, top+1.8, 2.2)))

	// Archivolts around both arches.
	for f := 0; f < 4; f++ {
		r, s := arcDeTriompheMainR, arcDeTriompheMainS
		if f >= 2 {
			r, s = arcDeTriompheSideR, arcDeTriompheSideS
		}
		g.Fill(a.onFace(f, Func{Box{V(-r-1.4, b+s, -0.6), V(r+1.4, b+s+r+1.4, 0.8)}, func(p Vec3) bool {
			d := math.Hypot(p.X, p.Y-b-s)
			return d >= r && d <= r+1.3
		}}))
	}
	// Socle pedestals of the four groups.
	pier := (arcDeTriompheMainR + hx) / 2
	for _, f := range []int{0, 1} {
		for _, side := range []float64{-1, 1} {
			u := side * pier
			g.Fill(a.onFace(f, Block(V(u-4.8, b, -0.6), V(u+4.8, b+4.2, 1.6))))
			g.Fill(a.onFace(f, Block(V(u-5.1, b+3.7, -0.6), V(u+5.1, b+4.3, 2.0))))
		}
	}

	// Coursed limestone mottling.
	a.recolor(g, Box{V(-hx-3, b, -hz-3), V(hx+3, top+3, hz+3)}, func(p Vec3, c Color) Color {
		if c != Sand {
			return c
		}
		h := Hash01(int(math.Floor(p.X/1.7)), int(math.Floor((p.Y-b)/0.8)), int(math.Floor(p.Z/1.7)), 5)
		switch {
		case h < 0.2:
			return Stone
		case h < 0.23:
			return White
		}
		return Sand
	})

	// Coffered vaults and inscribed jambs.
	coffer := func(arc, along float64) Color {
		if a.frac(arc, 2.2) < 0.36 || a.frac(along, 2.2) < 0.36 {
			return DarkStone
		}
		return Charcoal
	}
	a.recolor(g, Box{V(-hx, b, -hz), V(hx, top, hz)}, func(p Vec3, c Color) Color {
		// Main vault along Z.
		if math.Abs(p.Z) < hz-0.5 && math.Abs(p.X) < arcDeTriompheMainR+1.3 {
			rise := p.Y - b - arcDeTriompheMainS
			if rise > 0 {
				d := math.Hypot(p.X, rise)
				if d >= arcDeTriompheMainR-0.2 && d < arcDeTriompheMainR+1.3 {
					return coffer(math.Atan2(rise, p.X)*arcDeTriompheMainR, p.Z+hz)
				}
			} else if math.Abs(p.X) >= arcDeTriompheMainR-0.2 {
				if p.Y > b+6 && p.Y < b+arcDeTriompheMainS-0.8 && a.frac(p.Y-b, 0.8) < 0.5 {
					return DarkStone
				}
				return Stone
			}
		}
		// Transverse vaults along X.
		if math.Abs(p.X) < hx-0.5 && math.Abs(p.X) > arcDeTriompheMainR && math.Abs(p.Z) < arcDeTriompheSideR+1.3 {
			rise := p.Y - b - arcDeTriompheSideS
			if rise > 0 {
				d := math.Hypot(p.Z, rise)
				if d >= arcDeTriompheSideR-0.2 && d < arcDeTriompheSideR+1.3 {
					return coffer(math.Atan2(rise, p.Z)*arcDeTriompheSideR, p.X+hx)
				}
			} else if math.Abs(p.Z) >= arcDeTriompheSideR-0.2 {
				return DarkStone
			}
		}
		return c
	})

	a.sculpture(g)
	a.memorial(g)
}

// sculpture adds the relief groups, panels, Fames, frieze figures and attic shields.
func (a ArcDeTriomphe) sculpture(g *Grid) {
	hx, hz, b := arcDeTriompheHX, arcDeTriompheHZ, arcDeTriompheBase
	pier := (arcDeTriompheMainR + hx) / 2
	// The four great groups on the pier fronts of the long faces.
	seed := 0
	for _, f := range []int{0, 1} {
		for _, side := range []float64{-1, 1} {
			seed++
			uc, vb := side*pier, b+4.3
			g.Use(DarkStone).Fill(a.onFace(f, Block(V(uc-4.4, vb, -0.4), V(uc+4.4, vb+10.6, 0.5))))
			g.Use(White)
			for i := 0; i < 3; i++ {
				u := uc + float64(i-1)*2.9 + (Hash01(seed, i, 1)-0.5)*0.8
				lean := (Hash01(seed, i, 2) - 0.5) * 1.6
				hgt := 5.6 + Hash01(seed, i, 3)*1.2
				g.Use(Stone).Fill(a.onFace(f, Capsule{V(u-0.5, vb+0.5, 1.1), V(u+lean, vb+hgt, 1.2), 0.8}))
				g.Fill(a.onFace(f, Capsule{V(u+0.5, vb+0.5, 1.1), V(u+lean, vb+hgt*0.6, 1.2), 0.7}))
				g.Use(White).Fill(a.onFace(f, Sphere(V(u+lean*1.15, vb+hgt+1.0, 1.2), 0.75)))
				arm := 1.0
				if Hash01(seed, i, 4) < 0.5 {
					arm = -1
				}
				g.Fill(a.onFace(f, Capsule{V(u+lean, vb+hgt-0.6, 1.3), V(u+lean+arm*1.7, vb+hgt+1.6, 1.4), 0.5}))
			}
			// Winged genius above the figures.
			dir := side
			g.Fill(a.onFace(f, Capsule{V(uc-2.8*dir, vb+9.4, 1.1), V(uc+2.2*dir, vb+8.0, 1.2), 0.75}))
			g.Fill(a.onFace(f, Sphere(V(uc+2.8*dir, vb+8.0, 1.2), 0.7)))
			g.Fill(a.onFace(f, Capsule{V(uc-0.4*dir, vb+8.8, 1.0), V(uc-2.6*dir, vb+10.6, 1.0), 0.6}))
			g.Fill(a.onFace(f, Capsule{V(uc+0.6*dir, vb+8.6, 1.0), V(uc+1.6*dir, vb+10.6, 1.0), 0.6}))
			g.Fill(a.onFace(f, Capsule{V(uc+2.4*dir, vb+8.2, 1.3), V(uc+3.6*dir, vb+10.2, 1.3), 0.45}))

			// Upper battle relief panel above the impost.
			v0 := b + arcDeTriompheMainS + 1.2
			g.Use(Stone).Fill(a.onFace(f, Subtract(Block(V(uc-4.7, v0, -0.4), V(uc+4.7, v0+5.2, 0.9)),
				Block(V(uc-3.9, v0+0.8, 0.3), V(uc+3.9, v0+4.4, 1.0)))))
			g.Use(White)
			for i := 0; i < 5; i++ {
				u := uc - 3.1 + float64(i)*1.55
				h := 2.2 + Hash01(seed, i, 9)*1.0
				g.Fill(a.onFace(f, Capsule{V(u, v0+1.0, 0.6), V(u+(Hash01(seed, i, 8)-0.5)*1.2, v0+1.0+h, 0.7), 0.55}))
			}
		}
		// Fames in the spandrels of the main arch.
		for _, side := range []float64{-1, 1} {
			v := b + arcDeTriompheMainS
			g.Use(White)
			g.Fill(a.onFace(f, Capsule{V(side*2.6, v+8.0, 0.7), V(side*7.4, v+3.6, 0.7), 0.7}))
			g.Fill(a.onFace(f, Sphere(V(side*2.3, v+8.2, 0.8), 0.65)))
			g.Fill(a.onFace(f, Capsule{V(side*4.6, v+6.4, 0.6), V(side*6.6, v+8.2, 0.6), 0.55}))
			g.Fill(a.onFace(f, Capsule{V(side*2.6, v+7.6, 0.9), V(side*0.9, v+7.0, 0.9), 0.45}))
		}
	}
	// Relief panels above the transverse arches.
	for _, f := range []int{2, 3} {
		v0 := b + arcDeTriompheMainS + 1.2
		g.Use(Stone).Fill(a.onFace(f, Subtract(Block(V(-6.0, v0, -0.4), V(6.0, v0+5.2, 0.9)),
			Block(V(-5.2, v0+0.8, 0.3), V(5.2, v0+4.4, 1.0)))))
		g.Use(White)
		for i := 0; i < 7; i++ {
			u := -4.5 + float64(i)*1.5
			h := 2.0 + Hash01(f, i, 9)*1.2
			g.Fill(a.onFace(f, Capsule{V(u, v0+1.0, 0.6), V(u+(Hash01(f, i, 8)-0.5)*1.2, v0+1.0+h, 0.7), 0.55}))
		}
	}
	// Frieze procession and attic shields on all four faces.
	for f := 0; f < 4; f++ {
		half := hx
		n := 10
		if f >= 2 {
			half, n = hz, 5
		}
		for i, u := 0, -half+1.0; u <= half-1.0; i, u = i+1, u+1.4 {
			g.Use([]Color{White, Stone}[i%2])
			g.Fill(a.onFace(f, Capsule{V(u, b+28.1, 1.0), V(u, b+29.5, 1.0), 0.45}))
		}
		spacing := (2*half - 2) / float64(n)
		for i := 0; i < n; i++ {
			u := -half + 1 + spacing*(float64(i)+0.5)
			g.Use(White).Fill(a.onFace(f, Func{Box{V(u-1.2, b+36.0, -0.4), V(u+1.2, b+38.4, 0.8)}, func(p Vec3) bool {
				return math.Hypot(p.X-u, p.Y-b-37.2) <= 1.15
			}}))
			g.Use(Stone).Fill(a.onFace(f, Sphere(V(u, b+37.2, 0.8), 0.5)))
		}
	}
}

// memorial adds the tricolour hanging in the main arch, the tomb and eternal flame.
func (a ArcDeTriomphe) memorial(g *Grid) {
	b := arcDeTriompheBase
	y0, y1, w := b+7.6, b+arcDeTriompheMainS+3.4, 4.8
	a.raster(g, Box{V(-w, y0, -1.2), V(w, y1, 1.2)}, func(p Vec3) (Color, bool) {
		wave := 0.4 * math.Sin(p.X*0.9+p.Y*0.15)
		if math.Abs(p.Z-wave) > 0.45 {
			return 0, false
		}
		switch {
		case p.X < -w/3:
			return Navy, true
		case p.X > w/3:
			return Red, true
		}
		return White, true
	})
	g.Use(Charcoal).Fill(Capsule{V(-w-0.6, y1+0.3, 0), V(w+0.6, y1+0.3, 0), 0.45})
	// Tomb of the Unknown Soldier and the flame at its head.
	g.Use(DarkStone).Fill(Block(V(-1.6, 0.4, -1.4), V(1.6, b+0.4, 2.0)))
	g.Use(Charcoal).Fill(Block(V(-1.1, b, -1.0), V(1.1, b+0.4, 1.6)))
	g.Use(DarkStone).Fill(Cylinder{Center: P(0, 3.4), Radius: 1.1, Y0: b, Y1: b + 0.5})
	g.Use(Gold).Fill(Cylinder{Center: P(0, 3.4), Radius: 1.1, Inner: 0.5, Y0: b + 0.5, Y1: b + 0.9})
	g.Use(Red).Fill(Cone(P(0, 3.4), 0.75, b+0.5, 2.4))
	g.Use(Gold).Fill(Intersect(Cone(P(0, 3.4), 0.75, b+0.5, 2.4), Block(V(-1, b+1.5, 2), V(1, b+4, 5))))
}

// trees plants plane trees along both sidewalks of every avenue.
func (a ArcDeTriomphe) trees(g *Grid) {
	for k := 0; k < 12; k++ {
		t := float64(k) * math.Pi / 6
		u, n := P(math.Cos(t), math.Sin(t)), P(-math.Sin(t), math.Cos(t))
		H := a.avenueHalf(k)
		for _, side := range []float64{-1, 1} {
			for i, along := range []float64{45.6, 48.8, 52.0} {
				at := u.Scale(along).Add(n.Scale(side * (H - 0.9)))
				h := 5.2 + Hash01(k, i, int(side))*1.0
				g.Use(Wood).Fill(Cylinder{Center: at, Radius: 0.45, Y0: 0.4, Y1: 0.4 + h*0.6})
				g.Use(Foliage).FillVolume(Ellipsoid{at.At(0.4 + h*0.75), V(1.55, 1.65, 1.55)})
				g.Paint(Grass, Ellipsoid{at.At(0.4 + h*0.75 + 0.6), V(1.0, 1.0, 1.0)})
			}
		}
	}
}

// buildings rasterizes the Haussmann blocks between the avenues.
func (a ArcDeTriomphe) buildings(g *Grid) {
	R := arcDeTriompheRout + 1.2
	a.raster(g, Box{V(-R, 0, -R), V(R, a.cornice(7)+arcDeTriompheMans+1.6, R)}, func(p Vec3) (Color, bool) {
		l := a.lot(p.Plan())
		if l.d < -1.0 {
			return 0, false
		}
		t := p.Y - l.top
		if l.d < 0 {
			return a.ornament(l, p.Y)
		}
		owner := l.top > l.other
		if t < 0 {
			if l.d < 1.0 {
				return a.facade(l, p, l.d), true
			}
			if owner && l.dS < 1.0 && p.Y > l.other+arcDeTriompheMans-1.0 {
				return l.facade, true
			}
			return 0, false
		}
		// Dormers project from the lower mansard slope.
		if t < 1.9 && l.d >= 0.2 && l.d < 2.0 && l.d2 >= 1.5 {
			m := a.frac(l.u, 4.4) * 4.4
			if m >= 1.3 && m <= 3.1 {
				switch {
				case t >= 1.5:
					return Slate, true
				case l.d < 0.9 && m >= 1.75 && m <= 2.65 && t >= 0.3 && t <= 1.4:
					return Charcoal, true
				}
				return White, true
			}
		}
		inset := t * 0.38
		if t > 2.0 {
			inset = 0.76 + (t-2.0)*1.6
		}
		if t <= arcDeTriompheMans {
			if l.d < inset {
				if owner && l.dS < 1.0 && l.d >= 0 {
					return l.facade, true // Gable party wall above the lower neighbour.
				}
				return 0, false
			}
			if l.d < inset+1.0 || t >= arcDeTriompheMans-0.8 {
				return Slate, true
			}
			if owner && l.dS < 1.0 {
				return l.facade, true
			}
			return 0, false
		}
		// Chimney stacks on the flat top.
		if t <= arcDeTriompheMans+1.4 && l.d >= 2.4 {
			pot := t > arcDeTriompheMans+1.0
			if owner && l.dS < 0.9 && a.frac(l.r, 2.4) < 0.7 {
				if pot {
					return Terracotta, true
				}
				return l.facade, true
			}
			for j := 0; j < 2; j++ {
				rc := arcDeTriompheRin + 3.0 + float64(j)*3.6
				sc := (1.4 + Hash01(l.wedge, l.half, j)*1.6) * float64(2*l.half-1)
				if math.Abs(l.r-rc) < 0.7 && math.Abs(l.sc-sc) < 0.9 {
					if pot {
						return Terracotta, true
					}
					return Sand, true
				}
			}
		}
		return 0, false
	})
}

// facade colors a wall cell: shopfronts, window grid and corner piers.
func (a ArcDeTriomphe) facade(l arcDeTriompheLot, p Vec3, d float64) Color {
	if l.d2 < 1.0 {
		return l.facade
	}
	if p.Y < arcDeTriompheGround {
		m := a.frac(l.u, 3.6) * 3.6
		if p.Y < 2.4 && m >= 0.6 && m <= 3.0 {
			return Charcoal
		}
		return Stone
	}
	yy := math.Mod(p.Y-arcDeTriompheGround, arcDeTriompheStorey)
	m := a.frac(l.u, 2.2) * 2.2
	if m >= 0.55 && m <= 1.65 && yy >= 0.45 && yy <= 1.85 {
		return Charcoal
	}
	return l.facade
}

// ornament returns projecting cells just outside a facade: balconies, cornices,
// string course and café awnings.
func (a ArcDeTriomphe) ornament(l arcDeTriompheLot, y float64) (Color, bool) {
	if l.d2 < -0.2 {
		return 0, false
	}
	switch {
	case y >= l.top-0.6 && y <= l.top && l.d >= -0.8:
		return Stone, true
	case y >= arcDeTriompheGround-0.3 && y <= arcDeTriompheGround+0.2 && l.d >= -0.5:
		return Stone, true
	case l.cafe && l.face == 0 && y >= 2.0 && y <= 2.6 && l.d >= -1.0:
		m := a.frac(l.u, 3.6) * 3.6
		if m >= 0.4 && m <= 3.2 {
			return Red, true
		}
	case y >= arcDeTriompheGround && y < l.top-0.6 && l.d >= -0.8:
		i := int((y - arcDeTriompheGround) / arcDeTriompheStorey)
		yy := math.Mod(y-arcDeTriompheGround, arcDeTriompheStorey)
		if (i == 1 || i == l.storeys-2) && yy < 0.75 {
			return Charcoal, true
		}
	}
	return 0, false
}

// traffic places cars and red sightseeing buses around the roundabout and avenues.
func (a ArcDeTriomphe) traffic(g *Grid) {
	colors := []Color{White, Charcoal, Slate, Navy, Red, Stone, Glass, White, DarkStone, Gold}
	car := func(at Vec2, heading float64, c Color) {
		place := func(s Solid) Solid { return Translate(RotateY(s, P(0, 0), heading), at.At(0)) }
		g.Use(c).Fill(place(Block(V(-1.5, 0.4, -0.75), V(1.5, 1.2, 0.75))))
		g.Use(Charcoal).Fill(place(Block(V(-0.8, 1.2, -0.65), V(0.7, 1.9, 0.65))))
	}
	bus := func(at Vec2, heading float64) {
		place := func(s Solid) Solid { return Translate(RotateY(s, P(0, 0), heading), at.At(0)) }
		g.Use(Red).Fill(place(Block(V(-4.4, 0.4, -1.05), V(4.4, 3.8, 1.05))))
		for _, band := range [][2]float64{{1.5, 2.3}, {2.8, 3.4}} {
			g.Paint(Charcoal, place(Subtract(Block(V(-4.6, band[0], -1.3), V(4.6, band[1], 1.3)),
				Block(V(-3.9, 0, -0.5), V(3.9, 5, 0.5)))))
		}
		g.Paint(White, place(Block(V(-4.6, 2.4, -1.3), V(4.6, 2.7, 1.3))))
	}
	type slot struct{ lane, angle float64 }
	buses := []slot{{31.8, 0.55}, {35.8, 2.75}, {39.8, 4.4}}
	for _, s := range buses {
		bus(Polar(P(0, 0), s.lane, s.angle), s.angle+math.Pi/2)
	}
	for li, lane := range []float64{27.8, 31.8, 35.8, 39.8} {
		n := int(2 * math.Pi * lane / 6.5)
		for i := 0; i < n; i++ {
			if Hash01(li, i, 21) > 0.5 {
				continue
			}
			ang := (float64(i) + Hash01(li, i, 22)*0.3) * 2 * math.Pi / float64(n)
			clear := true
			for _, s := range buses {
				d := math.Abs(math.Remainder(ang-s.angle, 2*math.Pi)) * lane
				if math.Abs(s.lane-lane) < 1 && d < 8 {
					clear = false
				}
			}
			if clear {
				car(Polar(P(0, 0), lane, ang), ang+math.Pi/2, colors[int(Hash01(li, i, 23)*float64(len(colors)))])
			}
		}
	}
	for k := 0; k < 12; k++ {
		t := float64(k) * math.Pi / 6
		u, n := P(math.Cos(t), math.Sin(t)), P(-math.Sin(t), math.Cos(t))
		H := a.avenueHalf(k)
		for i, along := range []float64{47.5, 51.5} {
			if Hash01(k, i, 31) < 0.4 {
				continue
			}
			side := 1.0
			if i == 1 {
				side = -1
			}
			at := u.Scale(along).Add(n.Scale(side * (H - 3.0)))
			car(at, t, colors[int(Hash01(k, i, 32)*float64(len(colors)))])
		}
	}
}
