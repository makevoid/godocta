package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// SaturnVLaunch is Apollo-era Launch Complex 39A: the Saturn V (S-IC with its black and
// white roll pattern, five F-1 bells, engine fairings and fins; S-II, S-IVB, instrument
// unit, adapter, Apollo CSM under its cover and the launch escape tower) standing on the
// Mobile Launcher over the flame trench, beside the red Launch Umbilical Tower with its
// open truss, nine swing arms and hammerhead crane. The raised concrete hardstand is cut
// by the flame trench and its deflector; the crawlerway climbs the ramp from +Z, where the
// crawler-transporter is parked. Around it: LOX and LH2 spheres with their pipelines,
// water tower, tanks, service buildings, Florida scrub with cabbage palms and a lagoon.
type SaturnVLaunch struct{}

func init() { Register(SaturnVLaunch{}) }

func (SaturnVLaunch) Name() string  { return "saturn_v_launch" }
func (SaturnVLaunch) Title() string { return "Saturn V Launch" }

const (
	saturnVPad    = 7.0   // Hardstand top above the surrounding ground.
	saturnVSlope  = 15.0  // Horizontal run of the pad embankment.
	saturnVCrawl  = -8.0  // Crawlerway center line (X).
	saturnVDeck0  = 12.0  // Mobile Launcher underside.
	saturnVDeck1  = 17.7  // Mobile Launcher deck top.
	saturnVBase   = 20.2  // Bottom of the S-IC aft skirt.
	saturnVR      = 3.8   // S-IC and S-II radius.
	saturnVR3     = 2.5   // S-IVB radius.
	saturnVRcsm   = 1.46  // Service module radius.
	saturnVLUTx   = -19.0 // Launch Umbilical Tower center (X); it is centered on Z = 0.
	saturnVLUTh   = 4.55  // LUT half width.
	saturnVLUTtop = 108.7 // LUT roof.
	saturnVTrench = 6.5   // Flame trench half width.
	saturnVX0     = -76.0 // Site extent.
	saturnVX1     = 66.0
	saturnVZ0     = -64.0
	saturnVZ1     = 86.0
)

// Stage stations (world Y).
const (
	saturnVICTop  = saturnVBase + 31.5
	saturnVIITop  = saturnVICTop + 18.6
	saturnVAdTop  = saturnVIITop + 3.5
	saturnVIVBTop = saturnVAdTop + 9.9
	saturnVIUTop  = saturnVIVBTop + 0.7
	saturnVSLATop = saturnVIUTop + 6.4
	saturnVSMTop  = saturnVSLATop + 5.5
	saturnVCMTop  = saturnVSMTop + 2.7
)

func (SaturnVLaunch) clamp01(t float64) float64 { return math.Max(0, math.Min(1, t)) }

// noise is smooth value noise in [-1, 1].
func (SaturnVLaunch) noise(x, z float64, seed int) float64 {
	x0, z0 := math.Floor(x), math.Floor(z)
	fx, fz := x-x0, z-z0
	ix, iz := int(x0), int(z0)
	sx, sz := fx*fx*(3-2*fx), fz*fz*(3-2*fz)
	a, b := Hash01(ix, iz, seed), Hash01(ix+1, iz, seed)
	c, d := Hash01(ix, iz+1, seed), Hash01(ix+1, iz+1, seed)
	ab, cd := a+(b-a)*sx, c+(d-c)*sx
	return (ab+(cd-ab)*sz)*2 - 1
}

// raster sets every cell whose center lies in the box and for which f reports a color,
// without half-unit snapping, replacing existing cells.
func (SaturnVLaunch) raster(g *Grid, b Box, f func(Vec3) (Color, bool)) {
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

// plateau is the distance outside the hardstand's rounded rectangle (≤ 0 inside).
func (SaturnVLaunch) plateau(x, z float64) float64 {
	const cx, cz, hx, hz, r = -4.0, -1.0, 36.0, 30.0, 12.0
	dx := math.Max(math.Abs(x-cx)-(hx-r), 0)
	dz := math.Max(math.Abs(z-cz)-(hz-r), 0)
	return math.Hypot(dx, dz) - r
}

// mound is the pad embankment height with the crawlerway ramp, ignoring the trench.
func (s SaturnVLaunch) mound(x, z float64) float64 {
	h := saturnVPad * s.clamp01(1-s.plateau(x, z)/saturnVSlope)
	if z > 17 {
		r := saturnVPad*s.clamp01(1-(z-29)/22) - math.Max(math.Abs(x-saturnVCrawl)-17, 0)*saturnVPad/saturnVSlope
		h = math.Max(h, r)
	}
	return h
}

func (SaturnVLaunch) inTrench(x, z float64) bool {
	return math.Abs(z) < saturnVTrench && x > -62 && x < 54
}

// lagoon reports 2 for open water, 1 for its sandy shore, 0 elsewhere.
func (s SaturnVLaunch) lagoon(x, z float64) int {
	d := P(x-60, z-80)
	a := d.Angle()
	r := 25 + 3*math.Sin(3*a+1) + 2*math.Cos(5*a) + 2*s.noise(x/7, z/7, 41)
	switch l := d.Length(); {
	case l < r-3:
		return 2
	case l < r:
		return 1
	}
	return 0
}

func (s SaturnVLaunch) Build(g *Grid) Order {
	s.terrain(g)
	s.trench(g)
	s.launcher(g)
	s.rocket(g)
	s.tower(g)
	s.crawler(g)
	s.facilities(g)
	s.vegetation(g)
	return Outward(P(0, 0))
}

// terrain rasterizes the ground layer, the hollow pad embankment and ramp, the
// crawlerway, the lagoon and the trench floor.
func (s SaturnVLaunch) terrain(g *Grid) {
	xc := saturnVCrawl
	b := Box{V(saturnVX0, -0.6, saturnVZ0), V(saturnVX1, saturnVPad+0.1, saturnVZ1)}
	s.raster(g, b, func(p Vec3) (Color, bool) {
		x, z, y := p.X, p.Z, p.Y
		ix, iz := int(math.Floor(x)), int(math.Floor(z))
		switch s.lagoon(x, z) {
		case 2:
			return Water, y < -0.1 && y > -0.6
		case 1:
			return Sand, math.Abs(y) < 0.05
		}
		raw := s.mound(x, z)
		h := raw
		trench := s.inTrench(x, z)
		if trench {
			h = 0
		}
		t := 0.25
		if h > 0.3 {
			t = 0.65
		}
		if y > h+0.02 || y < math.Max(h-t, 0)-0.02 {
			return 0, false
		}
		lane := math.Abs(x-xc) > 5.5 && math.Abs(x-xc) < 14.5
		top := y > h-0.5
		switch {
		case trench && raw > 0.5:
			if Hash01(ix, iz, 3) < 0.3 {
				return DarkStone, true
			}
			return Charcoal, true
		case trench || (math.Abs(z) < 12 && (x < -54 || x > 46) && s.noise(x/4, z/4, 5) > -0.2):
			// Blast apron beyond the trench mouths.
			if Hash01(ix, iz, 4) < 0.25 {
				return DarkStone, true
			}
			return Sand, true
		case raw >= saturnVPad-0.01 && s.plateau(x, z) <= 0 && !(z > 15 && lane):
			// Concrete hardstand, scorched around the launcher.
			r := math.Hypot(x, z)
			if top && Hash01(ix, iz, 7) < 0.85-r/22 {
				return DarkStone, true
			}
			if math.Mod(x+400, 10) < 0.5 || math.Mod(z+400, 10) < 0.5 {
				return DarkStone, true
			}
			return Stone, true
		case z > 15 && lane:
			return Sand, true
		case raw > 0.3:
			if Hash01(ix, iz, 9) < 0.12 {
				return Foliage, true
			}
			return Grass, true
		}
		// Flat ground: roads, ditches beside the crawlerway, scrub and sand.
		if z > 51 && math.Abs(math.Abs(x-xc)-18) < 0.9 {
			return Water, true
		}
		if z > 52 && z < 56 && x < xc-14 {
			return DarkStone, true
		}
		// Perimeter road around the foot of the pad.
		if d := s.plateau(x, z); d > saturnVSlope+1.5 && d < saturnVSlope+4.5 {
			return DarkStone, true
		}
		n := s.noise(x/9, z/9, 11) + 0.3*s.noise(x/3, z/3, 12)
		switch {
		case n > 0.38:
			return Foliage, true
		case n < -0.62:
			return Sand, true
		}
		return Grass, true
	})
}

// trench builds the flame trench walls through the embankment and the deflector.
func (s SaturnVLaunch) trench(g *Grid) {
	b := Box{V(-62, 0, -saturnVTrench-1.6), V(54, saturnVPad, saturnVTrench+1.6)}
	s.raster(g, b, func(p Vec3) (Color, bool) {
		a := math.Abs(p.Z)
		if a < saturnVTrench || a > saturnVTrench+1.0 {
			return 0, false
		}
		return DarkStone, p.Y <= s.mound(p.X, p.Z)+0.02
	})
	// The wedge deflector under the launcher's exhaust hole.
	outer := Gable{Min: P(-6, -saturnVTrench), Max: P(6, saturnVTrench), Base: 0.3, Ridge: AxisZ}
	inner := Gable{Min: P(-4.8, -saturnVTrench-1), Max: P(4.8, saturnVTrench+1), Base: 0.3, Ridge: AxisZ}
	g.Use(Charcoal).Fill(Subtract(outer, inner))
	g.Paint(DarkStone, Block(V(-6, 0, -saturnVTrench), V(6, 1.2, saturnVTrench)))
}

// launcher is the Mobile Launcher: a hollow two-storey deck on six pedestals with the
// square exhaust opening, hold-down arms and tail service masts.
func (s SaturnVLaunch) launcher(g *Grid) {
	d0, d1 := saturnVDeck0, saturnVDeck1
	g.Use(DarkStone)
	for _, x := range []float64{-22, -8, 6} {
		for _, z := range []float64{-12.5, 12.5} {
			g.Fill(BlockAt(P(x, z), 3, 3, saturnVPad-0.5, d0))
		}
	}
	ml := Block(V(-26.5, d0, -15.4), V(10.1, d1, 15.4))
	inner := Block(V(-25.7, d0+0.6, -14.6), V(9.3, d1-0.8, 14.6))
	g.Fill(Subtract(ml, inner))
	g.Fill(Block(V(-6.4, d0, -6.4), V(6.4, d1, 6.4)))
	g.Carve(Block(V(-5.15, d0-1, -5.15), V(5.15, d1+1, 5.15)))
	g.Paint(Stone, Block(V(-26.5, d1-0.45, -15.4), V(10.1, d1, 15.4)))
	// Doors and windows of the two internal levels.
	g.Paint(Charcoal, Func{Box: ml.Bounds(), F: func(p Vec3) bool {
		band := math.Abs(p.Y-(d0+3.2)) < 0.7
		edgeZ := math.Abs(p.Z) > 14.9
		edgeX := p.X < -26 || p.X > 9.6
		along := p.X
		if edgeX {
			along = p.Z
		}
		return band && (edgeZ || edgeX) && math.Mod(along+100, 4) < 1.8
	}})
	// Hold-down arms between the fins.
	g.Use(Charcoal)
	for k := 0; k < 4; k++ {
		arm := BlockAt(P(5.2, 0), 3.2, 1.6, d1, saturnVBase+0.6)
		g.Fill(RotateY(arm, P(0, 0), float64(k)*math.Pi/2))
	}
	// Tail service masts.
	g.Use(DarkStone)
	for _, c := range []Vec2{P(7.5, -7.5), P(7.5, 7.5), P(-7.5, 8)} {
		g.Fill(BlockAt(c, 2, 2, d1, d1+4))
		g.Fill(BlockAt(c.Scale(0.85), 1.2, 1.2, d1+3, d1+4))
	}
}

// radius is the vehicle's outer radius at height y (0 outside the stack).
func (s SaturnVLaunch) radius(y float64) float64 {
	lerp := func(a, b, y0, y1 float64) float64 { return a + (b-a)*(y-y0)/(y1-y0) }
	switch {
	case y < saturnVBase:
		return 0
	case y <= saturnVIITop:
		return saturnVR
	case y <= saturnVAdTop:
		return lerp(saturnVR, saturnVR3, saturnVIITop, saturnVAdTop)
	case y <= saturnVIUTop:
		return saturnVR3
	case y <= saturnVSLATop:
		return lerp(saturnVR3, saturnVRcsm, saturnVIUTop, saturnVSLATop)
	case y <= saturnVSMTop:
		return saturnVRcsm
	case y <= saturnVCMTop:
		return lerp(saturnVRcsm, 0.35, saturnVSMTop, saturnVCMTop)
	}
	return 0
}

// rocket builds the Saturn V stack.
func (s SaturnVLaunch) rocket(g *Grid) {
	c := P(0, 0)
	rb := saturnVBase
	stack := Func{Box: Box{V(-saturnVR, rb, -saturnVR), V(saturnVR, saturnVCMTop, saturnVR)}, F: func(p Vec3) bool {
		r := s.radius(p.Y)
		d := p.Plan().Length()
		return d <= r && (r < 2.2 || d >= r-1.3 || p.Y < rb+1)
	}}
	g.Use(White).FillVolume(stack)

	// F-1 engines and the fairings and fins of the outer four.
	g.Use(Charcoal).FillVolume(Frustum{c, 1.39, 0.75, rb - 4.3, rb + 0.2})
	for k := 0; k < 4; k++ {
		a := math.Pi/4 + float64(k)*math.Pi/2
		e := Polar(c, 2.6, a)
		g.Use(Charcoal).FillVolume(Frustum{e, 1.39, 0.75, rb - 4.3, rb + 0.2})
		g.Use(White).FillVolume(Frustum{Polar(c, 3.05, a), 1.75, 0.2, rb - 0.8, rb + 6.5})
		dir := Polar(P(0, 0), 1, a)
		perp := P(-dir.Y, dir.X)
		fin := Func{Box: Box{V(-7.4, rb-1.4, -7.4), V(7.4, rb+6, 7.4)}, F: func(p Vec3) bool {
			q := p.Plan()
			r, t := q.Dot(dir), q.Dot(perp)
			if math.Abs(t) > 0.55 || r < 4.0 || r > 7.2 {
				return false
			}
			f := (r - 4.0) / 3.2
			return p.Y >= rb-0.8-0.5*f && p.Y <= rb+5.5-3.6*f
		}}
		g.Use(White).Fill(fin)
		g.Paint(Charcoal, Intersect(fin, Func{Box: fin.Box, F: func(p Vec3) bool { return p.Plan().Dot(dir) > 6.2 }}))
	}

	// Roll pattern: alternating black and white panels.
	checker := func(y0, y1, r float64, n int, odd bool) Solid {
		return Func{Box: Box{V(-r-0.5, y0, -r-0.5), V(r+0.5, y1, r+0.5)}, F: func(p Vec3) bool {
			d := p.Plan().Length()
			if d < r-1.6 || d > r+0.4 {
				return false
			}
			a := math.Atan2(p.Z, p.X) + 2*math.Pi
			k := int(a/(2*math.Pi/float64(n))) % n
			return (k%2 == 1) == odd
		}}
	}
	g.Paint(Charcoal,
		checker(rb, rb+4.5, saturnVR, 8, false),
		checker(rb+14.3, rb+19.3, saturnVR, 8, true),
		checker(rb+29, saturnVICTop, saturnVR, 8, false),
		checker(saturnVAdTop, saturnVAdTop+2.4, saturnVR3, 8, true),
	)
	// Separation plane between S-IC and S-II.
	g.Paint(Charcoal, Func{Box: Box{V(-5, saturnVICTop+0.2, -5), V(5, saturnVICTop+0.7, 5)}, F: func(p Vec3) bool {
		return p.Plan().Length() > saturnVR-0.6
	}})
	// Instrument unit and service module.
	g.Paint(Stone, Block(V(-3, saturnVIVBTop, -3), V(3, saturnVIUTop, 3)))
	g.Paint(Stone, Block(V(-2, saturnVSLATop+0.4, -2), V(2, saturnVSMTop, 2)))
	// Flag on the S-IC, facing the hero view.
	fa := 0.9
	flag := func(y0, y1, a0, a1 float64) Solid {
		return Func{Box: Box{V(-5, y0, -5), V(5, y1, 5)}, F: func(p Vec3) bool {
			d := p.Plan().Length()
			a := math.Atan2(p.Z, p.X)
			return d > saturnVR-0.7 && a >= a0 && a <= a1
		}}
	}
	fy := rb + 21.5
	g.Paint(Red, flag(fy, fy+3.6, fa-0.42, fa+0.42))
	g.Paint(White, flag(fy+0.8, fy+1.4, fa-0.42, fa+0.42), flag(fy+2.2, fy+2.8, fa-0.42, fa+0.42))
	g.Paint(Navy, flag(fy+1.8, fy+3.6, fa-0.42, fa-0.02))

	// Launch escape system: truss legs, motor, nose with pitch motor.
	ct := saturnVCMTop
	g.Use(Red)
	for k := 0; k < 4; k++ {
		a := math.Pi/4 + float64(k)*math.Pi/2
		g.Fill(Capsule{Polar(c, 0.95, a).At(ct - 1.4), Polar(c, 0.4, a).At(ct + 2.9), 0.42})
	}
	g.Use(White).Fill(Cylinder{Center: c, Radius: 0.62, Y0: ct + 2.6, Y1: ct + 7.3})
	g.Use(Charcoal).Fill(Frustum{c, 0.62, 0.3, ct + 7.3, ct + 8.6})
}

// tower is the Launch Umbilical Tower: an open red truss with floors, swing arms and
// the hammerhead crane.
func (s SaturnVLaunch) tower(g *Grid) {
	x0, x1 := saturnVLUTx-saturnVLUTh, saturnVLUTx+saturnVLUTh
	z0, z1 := -saturnVLUTh, saturnVLUTh
	w := 2 * saturnVLUTh
	const panel = 4.55
	member := func(u, y float64) bool {
		v := math.Mod(y-saturnVDeck1, panel)
		if u < 1.1 || u > w-1.1 || v < 0.45 || v > panel-0.45 {
			return true
		}
		// One X brace over each pair of levels: a square panel.
		v = math.Mod(y-saturnVDeck1, 2*panel) * w / (2 * panel)
		return math.Abs(u-v)/math.Sqrt2 < 0.5 || math.Abs(u+v-w)/math.Sqrt2 < 0.5
	}
	b := Box{V(x0, saturnVDeck1, z0), V(x1, saturnVLUTtop, z1)}
	s.raster(g, b, func(p Vec3) (Color, bool) {
		face := false
		if p.Z-z0 < 1.0 || z1-p.Z < 1.0 {
			face = face || member(p.X-x0, p.Y)
		}
		if p.X-x0 < 1.0 || x1-p.X < 1.0 {
			face = face || member(p.Z-z0, p.Y)
		}
		if face {
			return Red, true
		}
		// Floors every fourth level.
		lv := math.Mod(p.Y-saturnVDeck1, 4*panel)
		if lv < 0.2 || lv > 4*panel-0.2 {
			return DarkStone, true
		}
		return 0, false
	})
	g.Use(DarkStone).Fill(Block(V(x0, saturnVLUTtop-0.5, z0), V(x1, saturnVLUTtop, z1)))

	// Swing arms from the LUT face to the vehicle.
	type arm struct{ y, r float64 }
	arms := []arm{
		{saturnVBase + 16.5, saturnVR}, {saturnVBase + 29.8, saturnVR},
		{saturnVICTop + 3.2, saturnVR}, {saturnVICTop + 10, saturnVR}, {saturnVIITop - 1.8, saturnVR},
		{saturnVAdTop + 1.8, saturnVR3}, {saturnVIVBTop - 0.6, saturnVR3},
		{saturnVSLATop + 2.6, saturnVRcsm}, {saturnVSMTop + 1.6, 1.0},
	}
	for i, a := range arms {
		xa, xb := -a.r+0.3, x1-0.6
		g.Use(Red).Fill(Block(V(xa, a.y-1.2, -1.3), V(xb, a.y+1.2, 1.3)))
		for x := xa + 2.2; x < xb-1.2; x += 2.4 {
			g.Carve(Block(V(x-0.6, a.y-0.45, -2), V(x+0.6, a.y+0.45, 2)))
		}
		if i == len(arms)-1 {
			// White room around the command module hatch.
			g.Use(White).Fill(Block(V(-5.2, a.y-1.6, -1.9), V(-0.7, a.y+1.9, 1.9)))
		} else {
			g.Use(Charcoal).Fill(Block(V(xa-0.3, a.y-1.4, -1.4), V(xa+0.7, a.y+1.4, 1.4)))
		}
	}

	// Hammerhead crane.
	top := saturnVLUTtop
	g.Use(Red).Fill(BlockAt(P(saturnVLUTx, 0), 3, 3, top, top+3.2))
	g.Fill(Block(V(-31, top+3.2, -1.0), V(-6, top+4.8, 1.0)))
	for x := -29.0; x < -7; x += 2.4 {
		g.Carve(Block(V(x-0.6, top+3.7, -2), V(x+0.6, top+4.3, 2)))
	}
	g.Use(DarkStone).Fill(Block(V(-32, top+1.8, -1.6), V(-28.5, top+4.8, 1.6)))
	g.Use(Charcoal).Fill(Block(V(-16.5, top+1.2, -2.2), V(-14, top+3.2, 0)))
	g.Fill(Capsule{V(-7, top+3.2, 0), V(-7, top+0.6, 0), 0.45})
}

// crawler is the crawler-transporter parked on the crawlerway.
func (s SaturnVLaunch) crawler(g *Grid) {
	xc := saturnVCrawl
	zA, zB := 59.0, 84.0
	for _, lx := range []float64{xc - 10, xc + 10} {
		for _, tz := range []float64{zA + 4.2, zB - 4.2} {
			for _, side := range []float64{-2.1, 2.1} {
				track := BlockAt(P(lx+side, tz), 3.6, 8.4, 0.2, 2.2)
				g.Use(Charcoal).Fill(track)
				g.Paint(DarkStone, Func{Box: track.Bounds(), F: func(p Vec3) bool {
					return math.Mod(p.Z+100, 1.6) < 0.6
				}})
			}
			g.Use(DarkStone).Fill(BlockAt(P(lx, tz), 6.4, 6, 2.2, 3.2))
		}
	}
	body := Block(V(xc-11.5, 3.2, zA+3), V(xc+11.5, 5.8, zB-3))
	g.Use(Stone).Fill(Subtract(body, Block(V(xc-10.7, 3.8, zA+3.8), V(xc+10.7, 5.2, zB-3.8))))
	g.Paint(DarkStone, Func{Box: body.Bounds(), F: func(p Vec3) bool { return math.Abs(p.Y-4.4) < 0.45 }})
	// Control cabs at diagonal corners, engine houses and stacks on the deck.
	for _, c := range [][2]float64{{xc - 10, zA - 0.4}, {xc + 10, zB + 0.4}} {
		g.Use(White).Fill(BlockAt(P(c[0], c[1]), 4, 2.6, 2.6, 5.4))
		g.Paint(Glass, Block(V(c[0]-2, 4.0, c[1]-1.3), V(c[0]+2, 4.8, c[1]+1.3)))
	}
	g.Use(DarkStone).Fill(Block(V(xc-8.5, 5.8, zA+7), V(xc-2.5, 7.8, zB-7)))
	g.Fill(Block(V(xc+2.5, 5.8, zA+7), V(xc+8.5, 7.8, zB-7)))
	g.Paint(Charcoal, Func{Box: Box{V(xc-9, 6.6, zA), V(xc+9, 7.2, zB)}, F: func(p Vec3) bool {
		return math.Mod(p.Z+100, 2.4) < 1.2
	}})
	g.Use(Charcoal)
	for _, c := range []Vec2{P(xc-5.5, zA+9), P(xc+5.5, zB-9)} {
		g.Fill(Cylinder{Center: c, Radius: 0.7, Y0: 7.8, Y1: 9.8})
	}
	g.Use(Red)
	for _, c := range []Vec2{P(xc-10, zA+5), P(xc+10, zA+5), P(xc-10, zB-5), P(xc+10, zB-5)} {
		g.Fill(BlockAt(c, 2.4, 2.4, 5.8, 7.0))
	}
}

// pipe follows the terrain between two plan points, 0.9 above the surface.
func (s SaturnVLaunch) pipe(g *Grid, a, b Vec2, z float64) {
	n := int(b.Sub(a).Length()/3) + 1
	prev := a.At(s.mound(a.X, a.Y) + z)
	for i := 1; i <= n; i++ {
		q := a.Add(b.Sub(a).Scale(float64(i) / float64(n)))
		cur := q.At(s.mound(q.X, q.Y) + z)
		g.Fill(Capsule{prev, cur, 0.55})
		prev = cur
	}
}

// facilities places the propellant spheres, pipelines, water tower, tanks and buildings.
func (s SaturnVLaunch) facilities(g *Grid) {
	// LOX (west) and LH2 (east) spheres in their spill basins.
	for _, c := range []Vec2{P(-60, -46), P(50, -48)} {
		g.Use(Stone).Fill(Cylinder{Center: c, Radius: 13, Inner: 12, Y0: 0.2, Y1: 1.4})
		g.Use(DarkStone)
		for k := 0; k < 8; k++ {
			a := float64(k) * math.Pi / 4
			g.Fill(Capsule{Polar(c, 6.2, a).At(0.2), Polar(c, 5.6, a).At(6.5), 0.6})
		}
		g.Fill(Cylinder{Center: c, Radius: 6.3, Inner: 5.3, Y0: 5.2, Y1: 6.0})
		g.Use(White).FillVolume(Subtract(Sphere(c.At(11.2), 8), Sphere(c.At(11.2), 7)))
		g.Paint(Stone, Func{Box: Box{c.At(10.6).Sub(V(9, 0, 9)), c.At(11.8).Add(V(9, 0, 9))}, F: func(Vec3) bool { return true }})
		g.Use(Stone).Fill(BlockAt(c.Add(P(0, 10)), 4, 4, 0.2, 3))
	}
	g.Use(White)
	s.pipe(g, P(-56, -36), P(-24, -15.4), 0.9)
	s.pipe(g, P(-57, -35), P(-25.5, -14.6), 2.0)
	s.pipe(g, P(47, -38), P(6, -15.4), 0.9)

	// Water tower.
	wt := P(56, 42)
	g.Use(DarkStone)
	for k := 0; k < 4; k++ {
		a := math.Pi/4 + float64(k)*math.Pi/2
		g.Fill(Capsule{Polar(wt, 6.5, a).At(0), Polar(wt, 3.2, a).At(25), 0.55})
		for _, y := range []float64{8, 17} {
			r := 6.5 - 3.3*y/25
			g.Fill(Capsule{Polar(wt, r, a).At(y), Polar(wt, r, a+math.Pi/2).At(y), 0.45})
		}
	}
	g.Use(White).FillVolume(Frustum{wt, 2.0, 5.2, 22.5, 25.5})
	g.FillVolume(Subtract(Cylinder{Center: wt, Radius: 5.2, Y0: 25.5, Y1: 31.5}, Cylinder{Center: wt, Radius: 4.0, Y0: 25.5, Y1: 31}))
	g.Use(Charcoal).FillVolume(Frustum{wt, 5.4, 0.6, 31.5, 34})
	g.Paint(Navy, Block(wt.At(27.8).Sub(V(6, 0, 6)), wt.At(29.2).Add(V(6, 0, 6))))

	// Water tanks for the deluge system on the hardstand.
	for _, c := range []Vec2{P(24, -22), P(24, -14.5), P(17, -22)} {
		g.Use(Slate).Fill(Cylinder{Center: c, Radius: 2.8, Y0: saturnVPad - 0.4, Y1: saturnVPad + 6})
		g.Use(White).FillVolume(Ellipsoid{c.At(saturnVPad + 6), V(2.8, 1.2, 2.8)})
	}

	// Buildings: pad terminal on the hardstand, service buildings beyond.
	type shed struct {
		min, max Vec2
		base, h  float64
		wall     Color
	}
	for _, b := range []shed{
		{P(13, 10), P(27, 20), saturnVPad - 0.4, 5, Sand},
		{P(-72, 62), P(-52, 76), 0, 6.5, Sand},
		{P(-68, 36), P(-56, 46), 0, 5, Stone},
		{P(-46, 64), P(-38, 74), 0, 4, Stone},
		{P(52, -30), P(63, -18), 0, 5, Sand},
	} {
		outer := Block(b.min.At(b.base), b.max.At(b.base+b.h))
		inner := Block(b.min.Add(P(1, 1)).At(b.base+0.4), b.max.Sub(P(1, 1)).At(b.base+b.h-0.8))
		g.Use(b.wall).Fill(Subtract(outer, inner))
		g.Paint(Glass, Func{Box: outer.Bounds(), F: func(p Vec3) bool {
			nearX := p.X < b.min.X+1 || p.X > b.max.X-1
			nearZ := p.Z < b.min.Y+1 || p.Z > b.max.Y-1
			along := p.X
			if nearX {
				along = p.Z
			}
			return nearX != nearZ && math.Abs(p.Y-b.base-b.h*0.5) < 0.7 && math.Mod(along+100, 3) < 2
		}})
		g.Use(DarkStone).Fill(Block(b.min.At(b.base+b.h-0.4), b.max.At(b.base+b.h)))
		g.Use(Stone).Fill(BlockAt(b.min.Add(b.max).Scale(0.5), 3, 3, b.base+b.h, b.base+b.h+1.2))
	}
}

// vegetation scatters palmetto scrub and cabbage palms over the open ground.
func (s SaturnVLaunch) vegetation(g *Grid) {
	blocked := func(x, z float64) bool {
		if s.plateau(x, z) < saturnVSlope+6 || s.mound(x, z) > 0 || s.lagoon(x, z) != 0 {
			return true
		}
		if math.Abs(x-saturnVCrawl) < 21 && z > 25 {
			return true
		}
		if (z > 50 && z < 58 && x < -20) || (math.Abs(z) < 16 && (x < -52 || x > 44)) {
			return true
		}
		for _, c := range []Vec2{P(-60, -46), P(50, -48), P(56, 42)} {
			if P(x, z).Sub(c).Length() < 15 {
				return true
			}
		}
		for _, r := range [][4]float64{{-74, 60, -50, 78}, {-70, 34, -54, 48}, {-48, 62, -36, 76}, {50, -32, 65, -16}} {
			if x > r[0] && x < r[2] && z > r[1] && z < r[3] {
				return true
			}
		}
		return x < saturnVX0+2 || x > saturnVX1-2 || z < saturnVZ0+2 || z > saturnVZ1-2
	}
	const step = 6.0
	for i := -13; i <= 11; i++ {
		for j := -11; j <= 14; j++ {
			x := float64(i)*step + Hash01(i, j, 21)*step
			z := float64(j)*step + Hash01(i, j, 22)*step
			if blocked(x, z) || Hash01(i, j, 23) < 0.25 {
				continue
			}
			h := Hash01(i, j, 24)
			if Hash01(i, j, 25) < 0.22 {
				// Cabbage palm.
				top := V(x+(h-0.5)*2, 5.5+3*h, z+(Hash01(i, j, 26)-0.5)*2)
				g.Use(Wood).Fill(Capsule{V(x, 0, z), top, 0.5})
				g.Use(Foliage).Fill(Sphere(top, 1.0))
				for k := 0; k < 6; k++ {
					a := float64(k)*math.Pi/3 + h
					tip := Polar(top.Plan(), 2.6, a).At(top.Y - 1.2)
					g.Fill(Capsule{top, tip, 0.5})
				}
				continue
			}
			g.Use(Foliage).Fill(Ellipsoid{V(x, 0.3, z), V(1.4+h*1.6, 0.9+h*0.6, 1.2+h*1.4)})
		}
	}
}
