package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// IshtarGate is ancient Babylon's Ishtar Gate and Processional Way, reconstructed.
// The double gate faces +Z (north): a front gate with two crenellated towers flanking
// an arched passage, then the larger inner gate behind it, linked by a passage block.
// Their deep-blue glazed brick carries rows of golden aurochs and white sirrush
// dragons in relief, rosette friezes, white cornices and stepped Babylonian merlons.
// The Processional Way runs north between high blue walls with striding lions; the
// mud-brick double city wall extends east and west with towers to the Euphrates,
// which flows along the west (-X) edge. Inside the city stand flat-roofed houses,
// date palms and the stepped ziggurat Etemenanki with its blue top shrine. Flat desert
// sand slopes to a low rim.
type IshtarGate struct{}

func init() { Register(IshtarGate{}) }

func (IshtarGate) Name() string  { return "ishtar_gate" }
func (IshtarGate) Title() string { return "Ishtar Gate" }

const (
	ishtarGateG                      = 1.2 // Ground level.
	ishtarGateX0, ishtarGateX1       = -42.0, 40.0
	ishtarGateZ0, ishtarGateZ1       = -44.0, 42.0
	ishtarGateRim                    = 4.5   // Width of the sloping rim.
	ishtarGateRiver                  = -34.5 // Water lies west of this x.
	ishtarGateBank                   = -31.0 // The bank rises from the water to here.
	ishtarGateBull, ishtarGateDr     = 0, 1  // Animal kinds.
	ishtarGateLionG, ishtarGateLionW = 2, 3
)

// each visits every lattice cell whose world center lies in the box.
func (IshtarGate) each(g *Grid, b Box, f func(c Cell, p Vec3)) {
	s := g.Span()
	lo := func(v float64) int32 { return int32(math.Ceil(v * 2 / s)) }
	hi := func(v float64) int32 { return int32(math.Floor(v * 2 / s)) }
	y0, y1 := lo(b.Min.Y), hi(b.Max.Y)
	z0, z1 := lo(b.Min.Z), hi(b.Max.Z)
	x0, x1 := lo(b.Min.X), hi(b.Max.X)
	for y := y0; y <= y1; y++ {
		par := y & 1
		for z := z0 + ((z0&1)^par)&1; z <= z1; z += 2 {
			for x := x0 + ((x0&1)^par)&1; x <= x1; x += 2 {
				c := Cell{X: x, Y: y, Z: z}
				f(c, V(float64(x), float64(y), float64(z)).Scale(s/2))
			}
		}
	}
}

func (IshtarGate) set(g *Grid, c Cell, col Color) { g.Remove(c); g.Use(col).Add(c) }

// fill occupies cells whose centers lie in any solid (no half-unit snapping).
func (k IshtarGate) fill(g *Grid, col Color, solids ...Solid) {
	for _, s := range solids {
		k.each(g, s.Bounds(), func(c Cell, p Vec3) {
			if s.Contains(p) {
				k.set(g, c, col)
			}
		})
	}
}

func (k IshtarGate) carve(g *Grid, solids ...Solid) {
	for _, s := range solids {
		k.each(g, s.Bounds(), func(c Cell, p Vec3) {
			if s.Contains(p) {
				g.Remove(c)
			}
		})
	}
}

func (k IshtarGate) paint(g *Grid, col Color, solids ...Solid) {
	for _, s := range solids {
		k.each(g, s.Bounds(), func(c Cell, p Vec3) {
			if g.Has(c) && s.Contains(p) {
				k.set(g, c, col)
			}
		})
	}
}

// shade sets cells in a box from a per-cell function; existing=true only recolors.
func (k IshtarGate) shade(g *Grid, b Box, existing bool, f func(c Cell, p Vec3) (Color, bool)) {
	k.each(g, b, func(c Cell, p Vec3) {
		if existing && !g.Has(c) {
			return
		}
		if col, ok := f(c, p); ok {
			k.set(g, c, col)
		}
	})
}

// box is a hollow box: walls of thickness t, closed roof of thickness 1.
func (k IshtarGate) box(g *Grid, col Color, x0, x1, z0, z1, y0, y1, t float64) {
	k.fill(g, col, Block(V(x0, y0, z0), V(x1, y1, z1)))
	if x1-x0 > 2*t+0.8 && z1-z0 > 2*t+0.8 {
		k.carve(g, Block(V(x0+t, ishtarGateG, z0+t), V(x1-t, y1-1, z1-t)))
	}
}

func ishtarGateHash(c Cell, seed int) float64 {
	return Hash01(int(c.X), int(c.Y), int(c.Z), seed)
}

func ishtarGateSmooth(t float64) float64 {
	t = math.Max(0, math.Min(1, t))
	return t * t * (3 - 2*t)
}

func ishtarGateNoise(x, z float64, seed int) float64 {
	ix, iz := math.Floor(x), math.Floor(z)
	fx, fz := x-ix, z-iz
	sx, sz := fx*fx*(3-2*fx), fz*fz*(3-2*fz)
	i, j := int(ix), int(iz)
	a := Hash01(i, j, seed) + (Hash01(i+1, j, seed)-Hash01(i, j, seed))*sx
	b := Hash01(i, j+1, seed) + (Hash01(i+1, j+1, seed)-Hash01(i, j+1, seed))*sx
	return a + (b-a)*sz
}

// height is the ground surface; negative means none (river or outside the site).
func (IshtarGate) height(x, z float64) float64 {
	const g, run = ishtarGateG, ishtarGateRim
	dx := math.Max(ishtarGateX0+run-x, x-(ishtarGateX1-run))
	dz := math.Max(ishtarGateZ0+run-z, z-(ishtarGateZ1-run))
	d := math.Hypot(math.Max(dx, 0), math.Max(dz, 0))
	if dx > 0 || dz > 0 {
		d += 1.4 * (ishtarGateNoise(x/6, z/6, 3) - 0.5)
	}
	h := g
	if d > 0 {
		if d >= run {
			return -1
		}
		h = g - (g-0.35)*ishtarGateSmooth(d/run)
	}
	// The river bank.
	if x < ishtarGateBank {
		if x < ishtarGateRiver {
			return -1
		}
		t := (x - ishtarGateRiver) / (ishtarGateBank - ishtarGateRiver)
		h = math.Min(h, 0.45+(g-0.45)*ishtarGateSmooth(t))
	}
	return h
}

func (k IshtarGate) ground(g *Grid) {
	b := Box{V(ishtarGateX0, 0, ishtarGateZ0), V(ishtarGateX1, ishtarGateG+0.05, ishtarGateZ1)}
	k.shade(g, b, false, func(c Cell, p Vec3) (Color, bool) {
		h := k.height(p.X, p.Z)
		if h < 0 || p.Y > h || p.Y < math.Max(0, h-0.85) {
			return 0, false
		}
		r := ishtarGateHash(c, 11)
		switch {
		case r < 0.05:
			return Stone, true
		case r < 0.07:
			return Terracotta, true
		case r < 0.085:
			return DarkStone, true
		}
		return Sand, true
	})
	// The Euphrates.
	k.shade(g, Box{V(ishtarGateX0, 0, ishtarGateZ0+1.5), V(ishtarGateRiver+0.6, 0.55, ishtarGateZ1-1.5)}, false,
		func(c Cell, p Vec3) (Color, bool) {
			if g.Has(c) {
				return 0, false
			}
			if ishtarGateHash(c, 13) < 0.06 {
				return Glass, true
			}
			return Water, true
		})
	// Processional Way paving: limestone slabs in the middle, red breccia edges.
	pave := func(c Cell, p Vec3) (Color, bool) {
		ax := math.Abs(p.X)
		if ax > 3.6 {
			return Brick, true
		}
		fu, fv := math.Mod(p.X+40, 2.4), math.Mod(p.Z+80, 2.4)
		if fu < 0.4 || fv < 0.4 {
			return Stone, true
		}
		if ishtarGateHash(c, 17) < 0.12 {
			return Sand, true
		}
		return White, true
	}
	k.shade(g, Box{V(-6, ishtarGateG-1, -19), V(6, ishtarGateG+0.05, ishtarGateZ1-2)}, true, pave)
	k.shade(g, Box{V(-4.8, ishtarGateG-1, ishtarGateZ0+2), V(4.8, ishtarGateG+0.05, -19)}, true, pave)
}

// ---------------------------------------------------------------------------
// Relief animals, in a unit-height frame facing +x, from x = 0 (tail) to about 1.4.

func ishtarGateEll(x, y, cx, cy, rx, ry float64) bool {
	dx, dy := (x-cx)/rx, (y-cy)/ry
	return dx*dx+dy*dy <= 1
}

func ishtarGateCap(x, y, ax, ay, bx, by, r float64) bool {
	ux, uy := bx-ax, by-ay
	t := ((x-ax)*ux + (y-ay)*uy) / math.Max(ux*ux+uy*uy, 1e-9)
	t = math.Max(0, math.Min(1, t))
	dx, dy := x-(ax+ux*t), y-(ay+uy*t)
	return dx*dx+dy*dy <= r*r
}

func (IshtarGate) animal(kind int, x, y float64) (Color, bool) {
	e, c := ishtarGateEll, ishtarGateCap
	switch kind {
	case ishtarGateBull: // Aurochs: gold body, white horns, sand hooves.
		if c(x, y, 1.22, 0.7, 1.15, 0.86, 0.075) || c(x, y, 1.15, 0.86, 1.27, 1.0, 0.07) {
			return White, true
		}
		legs := c(x, y, 1.0, 0.45, 1.05, 0.03, 0.09) || c(x, y, 0.86, 0.45, 0.8, 0.03, 0.08) ||
			c(x, y, 0.36, 0.45, 0.28, 0.03, 0.1) || c(x, y, 0.5, 0.45, 0.53, 0.03, 0.08)
		if legs && y < 0.11 {
			return Sand, true
		}
		if legs || e(x, y, 0.68, 0.53, 0.47, 0.2) || e(x, y, 0.98, 0.63, 0.21, 0.16) ||
			c(x, y, 1.0, 0.62, 1.24, 0.62, 0.12) || e(x, y, 1.27, 0.56, 0.11, 0.13) ||
			c(x, y, 1.27, 0.55, 1.34, 0.42, 0.08) || c(x, y, 0.24, 0.6, 0.12, 0.3, 0.06) {
			return Gold, true
		}
		if e(x, y, 0.12, 0.26, 0.07, 0.08) {
			return Sand, true
		}
	case ishtarGateDr: // Sirrush: white serpent-dragon, gold horn and bird talons.
		if c(x, y, 1.1, 0.92, 1.0, 1.02, 0.07) || c(x, y, 0.36, 0.38, 0.3, 0.03, 0.09) {
			return Gold, true
		}
		if e(x, y, 0.62, 0.41, 0.37, 0.15) || c(x, y, 0.9, 0.46, 1.1, 0.83, 0.1) ||
			c(x, y, 1.08, 0.86, 1.36, 0.84, 0.085) || c(x, y, 0.9, 0.36, 0.97, 0.03, 0.085) ||
			c(x, y, 0.28, 0.44, 0.1, 0.56, 0.075) || c(x, y, 0.1, 0.56, 0.13, 0.8, 0.065) {
			return White, true
		}
	case ishtarGateLionG, ishtarGateLionW: // Striding lion.
		body, mane := Gold, White
		if kind == ishtarGateLionW {
			body, mane = White, Gold
		}
		if e(x, y, 1.1, 0.63, 0.18, 0.25) {
			return mane, true
		}
		if e(x, y, 0.72, 0.5, 0.47, 0.17) || e(x, y, 1.29, 0.63, 0.13, 0.13) ||
			c(x, y, 1.28, 0.55, 1.42, 0.52, 0.065) ||
			c(x, y, 1.06, 0.45, 1.17, 0.03, 0.085) || c(x, y, 0.96, 0.45, 0.87, 0.03, 0.075) ||
			c(x, y, 0.34, 0.45, 0.23, 0.03, 0.095) || c(x, y, 0.46, 0.45, 0.53, 0.03, 0.075) ||
			c(x, y, 0.27, 0.56, 0.1, 0.74, 0.055) {
			return body, true
		}
		if e(x, y, 0.07, 0.8, 0.07, 0.09) {
			return mane, true
		}
	}
	return 0, false
}

// ---------------------------------------------------------------------------
// Glazed faces.

// ishtarGateFace is a vertical rectangle on a wall plane with an outward normal along
// X or Z. u runs left to right as seen from outside, v up from Y0.
type ishtarGateFace struct {
	AlongX bool    // The face runs along X (its normal is ±Z).
	Sign   float64 // Outward normal sign.
	At     float64 // Plane coordinate.
	A0, A1 float64 // Extent along the face.
	Y0, Y1 float64
}

func (f ishtarGateFace) local(p Vec3) (u, v, d float64) {
	v = p.Y - f.Y0
	if f.AlongX {
		d = (p.Z - f.At) * f.Sign
		if f.Sign > 0 {
			u = p.X - f.A0
		} else {
			u = f.A1 - p.X
		}
		return
	}
	d = (p.X - f.At) * f.Sign
	if f.Sign > 0 {
		u = f.A1 - p.Z
	} else {
		u = p.Z - f.A0
	}
	return
}

func (f ishtarGateFace) bounds() Box {
	lo, hi := f.At-0.9, f.At+0.6
	if f.Sign < 0 {
		lo, hi = f.At-0.6, f.At+0.9
	}
	if f.AlongX {
		return Box{V(f.A0, f.Y0, lo), V(f.A1, f.Y1+0.6, hi)}
	}
	return Box{V(lo, f.Y0, f.A0), V(hi, f.Y1+0.6, f.A1)}
}

// ishtarGatePattern colors a face point; relief points also protrude from the wall.
type ishtarGatePattern func(u, v, h float64) (col Color, relief, ok bool)

func (k IshtarGate) decorate(g *Grid, f ishtarGateFace, pat ishtarGatePattern) {
	k.each(g, f.bounds(), func(c Cell, p Vec3) {
		u, v, d := f.local(p)
		col, relief, ok := pat(u, v, ishtarGateHash(c, 29))
		if !ok {
			return
		}
		if d > 0 {
			if relief && d <= 0.55 {
				k.set(g, c, col)
			}
		} else if d >= -0.85 && g.Has(c) {
			k.set(g, c, col)
		}
	})
}

// ishtarGateGlaze describes a blue-glazed face: rows of relief animals, an optional
// arch with a gold-and-white border, a rosette frieze and a white cornice at the top.
type ishtarGateGlaze struct {
	W, H         float64
	VStart, Ha   float64 // Lowest row and animal height; Ha = 0 means no animals.
	Kinds        []int   // Animal kind per row, cycled.
	Dir          float64 // +1: animals face +u.
	Ratio        float64 // Animal width / height (default 1.4).
	Pitch        float64 // Animal spacing along u; 0 fits them to the width.
	ArchU        float64 // Arch center along u.
	ArchHW       float64 // Arch half width; 0 = none.
	ArchSpring   float64
	Plinth       float64 // Height of a white limestone plinth.
	FriezeHeight float64 // Distance of the rosette frieze top below H (default 2.4).
	Cornice      Color   // Cornice color (Stone selects White).
}

func (s ishtarGateGlaze) pattern(k IshtarGate) ishtarGatePattern {
	const gap = 1.6
	ratio := s.Ratio
	if ratio == 0 {
		ratio = 1.4
	}
	fr := s.FriezeHeight
	if fr == 0 {
		fr = 2.4
	}
	top := s.H - fr - 2.4
	n := 0
	if s.Ha > 0 {
		n = int((top - s.VStart + gap) / (s.Ha + gap))
	}
	base := s.VStart
	if n > 0 {
		base += (top - s.VStart - (float64(n)*(s.Ha+gap) - gap)) / 2
	}
	wa := s.Ha * ratio
	pitch := s.Pitch
	m := 0
	off := 0.0
	if n > 0 {
		if pitch == 0 {
			m = int((s.W - 0.8) / (wa + 1.6))
			if m > 0 {
				pitch = s.W / float64(m)
			}
		} else {
			m = int(s.W / pitch)
			off = (s.W - float64(m)*pitch) / 2
		}
	}
	return func(u, v, h float64) (Color, bool, bool) {
		if u < -0.05 || u > s.W+0.05 || v < -0.05 || v > s.H+0.3 {
			return 0, false, false
		}
		switch {
		case v >= s.H-1.1:
			return s.Cornice.Or(White), true, true
		case v >= s.H-1.7:
			return Gold, false, true
		case v >= s.H-fr-1.9 && v < s.H-fr:
			cv := v - (s.H - fr - 0.95)
			pu := math.Mod(u+50, 2.2) - 1.1
			if pu*pu+cv*cv <= 0.66*0.66 {
				return White, true, true
			}
			return Navy, false, true
		case v >= s.H-fr-2.4 && v < s.H-fr-1.9:
			return Gold, false, true
		case v < s.Plinth:
			if h < 0.15 {
				return Stone, false, true
			}
			return White, false, true
		}
		if s.ArchHW > 0 {
			x := u - s.ArchU
			var d float64
			if v <= s.ArchSpring {
				d = math.Abs(x) - s.ArchHW
			} else {
				d = math.Hypot(x, v-s.ArchSpring) - s.ArchHW
			}
			switch {
			case d < -0.05:
				return 0, false, false
			case d < 0.75:
				return Gold, false, true
			case d < 1.75:
				a := math.Atan2(v-s.ArchSpring, x)
				if v > s.ArchSpring && math.Mod(a*(s.ArchHW+1.25)+50, 2.0) < 1.0 {
					return Navy, false, true
				}
				return White, true, true
			}
		}
		if n > 0 && m > 0 && v >= base {
			r := int((v - base) / (s.Ha + gap))
			lv := v - base - float64(r)*(s.Ha+gap)
			if r < n && lv <= s.Ha {
				i := int((u - off) / pitch)
				if i >= 0 && i < m {
					center := off + (float64(i)+0.5)*pitch
					x := (u-center)/s.Ha*s.Dir + ratio/2
					kind := s.Kinds[r%len(s.Kinds)]
					if col, ok := k.animal(kind, x, lv/s.Ha); ok {
						return col, true, true
					}
				}
			}
		}
		switch {
		case h < 0.08:
			return Slate, false, true
		case h < 0.09:
			return Glass, false, true
		}
		return Navy, false, true
	}
}

// mudBrick colors city walls: sun-dried brick with bitumen courses.
func (IshtarGate) mudBrick(u, v, h, H float64) (Color, bool, bool) {
	switch {
	case v > H-1.0:
		return Terracotta, false, true
	case math.Mod(v+0.2, 3.2) < 0.45:
		return Terracotta, false, true
	case h < 0.06:
		return Stone, false, true
	case h < 0.1:
		return Terracotta, false, true
	}
	return Sand, false, true
}

// merlons places stepped Babylonian merlons along one edge strip at height top.
func (k IshtarGate) merlons(g *Grid, col, tip Color, alongX bool, a0, a1, t0, t1, top, w, hm float64) {
	L := a1 - a0
	gap := w * 0.4
	n := int((L + gap) / (w + gap))
	if n < 1 {
		return
	}
	step := 0.0
	if n > 1 {
		step = (L - w) / float64(n-1)
	}
	for i := 0; i < n; i++ {
		c := a0 + w/2 + float64(i)*step
		if n == 1 {
			c = (a0 + a1) / 2
		}
		for t := 0; t < 3; t++ {
			hw := w * float64(3-t) / 6
			y0, y1 := top+float64(t)*hm, top+float64(t+1)*hm
			cc := col
			if t == 2 {
				cc = tip
			}
			if alongX {
				k.fill(g, cc, Block(V(c-hw, y0, t0), V(c+hw, y1, t1)))
			} else {
				k.fill(g, cc, Block(V(t0, y0, c-hw), V(t1, y1, c+hw)))
			}
		}
	}
}

// rectMerlons runs merlons around a rectangle's roof edge.
func (k IshtarGate) rectMerlons(g *Grid, col, tip Color, x0, x1, z0, z1, top, w, hm float64) {
	const t = 1.0
	k.merlons(g, col, tip, true, x0, x1, z1-t, z1, top, w, hm)
	k.merlons(g, col, tip, true, x0, x1, z0, z0+t, top, w, hm)
	k.merlons(g, col, tip, false, z0, z1, x0, x0+t, top, w, hm)
	k.merlons(g, col, tip, false, z0, z1, x1-t, x1, top, w, hm)
}

// ---------------------------------------------------------------------------
// The double gate.

// glazedBlock builds a hollow glazed block and decorates its four faces. faceDir maps
// a face index (0 = +Z, 1 = -Z, 2 = +X, 3 = -X) to its animal direction.
func (k IshtarGate) glazedBlock(g *Grid, x0, x1, z0, z1, h, ha float64, kinds []int, dirs [4]float64, merl bool) {
	const G = ishtarGateG
	k.box(g, Navy, x0, x1, z0, z1, 0.4, G+h, 1.0)
	faces := []ishtarGateFace{
		{AlongX: true, Sign: 1, At: z1, A0: x0, A1: x1, Y0: G, Y1: G + h},
		{AlongX: true, Sign: -1, At: z0, A0: x0, A1: x1, Y0: G, Y1: G + h},
		{AlongX: false, Sign: 1, At: x1, A0: z0, A1: z1, Y0: G, Y1: G + h},
		{AlongX: false, Sign: -1, At: x0, A0: z0, A1: z1, Y0: G, Y1: G + h},
	}
	for i, f := range faces {
		w := f.A1 - f.A0
		k.decorate(g, f, ishtarGateGlaze{W: w, H: h, VStart: 1.6, Ha: ha, Kinds: kinds, Dir: dirs[i]}.pattern(k))
	}
	if merl {
		k.rectMerlons(g, White, Gold, x0, x1, z0, z1, G+h, 2.7, 0.95)
	}
}

func (k IshtarGate) gate(g *Grid) {
	const G = ishtarGateG
	bd := []int{ishtarGateBull, ishtarGateDr}
	db := []int{ishtarGateDr, ishtarGateBull}
	// Inner (larger) gate: towers and the central block.
	for _, s := range []float64{-1, 1} {
		x0, x1 := 10.5, 23.0
		if s < 0 {
			x0, x1 = -23, -10.5
		}
		k.glazedBlock(g, x0, x1, -17, -3, 34, 5.2, db, [4]float64{-s, s, -1, 1}, true)
	}
	k.box(g, Navy, -10.5, 10.5, -16, -4, 0.4, G+28, 1.0)
	// Front gate towers and central block.
	for _, s := range []float64{-1, 1} {
		x0, x1 := 8.0, 18.0
		if s < 0 {
			x0, x1 = -18, -8
		}
		k.glazedBlock(g, x0, x1, 1.5, 11, 26, 4.6, bd, [4]float64{-s, s, -1, 1}, true)
	}
	k.box(g, Navy, -8, 8, 2.5, 10, 0.4, G+21, 1.0)
	// Connecting passage block between the two gates.
	k.box(g, Navy, -8.5, 8.5, -4.5, 2.5, 0.4, G+15, 1.0)

	// The arched passage: lower through the front gate, taller through the inner gate.
	front := Arch{Sill: V(0, G, 3), HalfWidth: 3, Spring: 7, Depth: 17, Through: AxisZ}
	inner := Arch{Sill: V(0, G, -10.25), HalfWidth: 3.5, Spring: 10, Depth: 13.5, Through: AxisZ}
	// Solid masonry around the passage, then the opening itself.
	k.fill(g, Navy,
		Intersect(Arch{Sill: V(0, G, 3), HalfWidth: 4.2, Spring: 7, Depth: 17, Through: AxisZ}, Block(V(-8, 0, -4.5), V(8, 30, 10))),
		Intersect(Arch{Sill: V(0, G, -10.25), HalfWidth: 4.7, Spring: 10, Depth: 13.5, Through: AxisZ}, Block(V(-10.5, 0, -16), V(10.5, 30, -4))))
	k.carve(g, front, inner)
	k.paint(g, Charcoal, Intersect(Union(
		Arch{Sill: V(0, G, 3), HalfWidth: 4.0, Spring: 7, Depth: 17, Through: AxisZ},
		Arch{Sill: V(0, G, -10.25), HalfWidth: 4.5, Spring: 10, Depth: 13.5, Through: AxisZ},
	), Block(V(-6, G+0.25, -17), V(6, 30, 12))))

	// Central block façades: arch, border and a row of animals facing the passage.
	for _, s := range []float64{-1, 1} {
		k.decorate(g, ishtarGateFace{AlongX: true, Sign: 1, At: 10, A0: -8, A1: 8, Y0: G, Y1: G + 21},
			ishtarGateGlaze{W: 16, H: 21, VStart: 12.2, Ha: 3.6, Kinds: []int{ishtarGateDr}, Dir: 1,
				ArchU: 8, ArchHW: 3, ArchSpring: 7}.pattern(k).half(s, 8))
		k.decorate(g, ishtarGateFace{AlongX: true, Sign: -1, At: -16, A0: -10.5, A1: 10.5, Y0: G, Y1: G + 28},
			ishtarGateGlaze{W: 21, H: 28, VStart: 15.6, Ha: 5.0, Kinds: []int{ishtarGateBull}, Dir: 1,
				ArchU: 10.5, ArchHW: 3.5, ArchSpring: 10}.pattern(k).half(s, 10.5))
	}
	k.decorate(g, ishtarGateFace{AlongX: true, Sign: 1, At: -4, A0: -10.5, A1: 10.5, Y0: G, Y1: G + 28},
		ishtarGateGlaze{W: 21, H: 28, VStart: 16, Ha: 5.0, Kinds: bd, Dir: 1}.pattern(k))
	for _, s := range []float64{-1, 1} {
		k.decorate(g, ishtarGateFace{AlongX: false, Sign: s, At: 8.5 * s, A0: -4.5, A1: 2.5, Y0: G, Y1: G + 15},
			ishtarGateGlaze{W: 7, H: 15, VStart: 1.6, Ha: 4.0, Pitch: 7, Kinds: []int{ishtarGateBull}, Dir: -s}.pattern(k))
	}
	k.rectMerlons(g, White, Gold, -8, 8, 2.5, 10, G+21, 2.7, 0.95)
	k.rectMerlons(g, White, Gold, -10.5, 10.5, -16, -4, G+28, 2.7, 0.95)
	k.rectMerlons(g, White, Gold, -8.5, 8.5, -4.5, 2.5, G+15, 2.7, 0.95)
}

// half mirrors a pattern so animals on each side of a central arch face the arch:
// s = -1 keeps the left half, +1 the right half, center is the face's middle u.
func (p ishtarGatePattern) half(s, center float64) ishtarGatePattern {
	return func(u, v, h float64) (Color, bool, bool) {
		if (u-center)*s < 0 {
			return 0, false, false
		}
		// Mirror the right half so its animals face left, toward the arch.
		if s > 0 {
			return p(2*center-u, v, h)
		}
		return p(u, v, h)
	}
}

// ---------------------------------------------------------------------------
// Walls.

// processional builds the high blue walls lining the Processional Way.
func (k IshtarGate) processional(g *Grid) {
	const G = ishtarGateG
	const z0, z1, h = 11.0, 37.0, 10.0
	lions := []int{ishtarGateLionG}
	for _, s := range []float64{-1, 1} {
		xi, xo := 6.5*s, 8.5*s
		lo, hi := math.Min(xi, xo), math.Max(xi, xo)
		k.fill(g, Navy, Block(V(lo, 0.4, z0), V(hi, G+h, z1)))
		// Buttress towers on the outer side.
		var towers []float64
		for z := z0 + 4.5; z < z1-1; z += 10 {
			towers = append(towers, z)
			tl, th := math.Min(xo, 10.5*s), math.Max(xo, 10.5*s)
			k.fill(g, Navy, Block(V(tl, 0.4, z-1.5), V(th, G+h+2, z+1.5)))
		}
		// Lions stride toward the gate on both faces.
		inner := ishtarGateFace{AlongX: false, Sign: -s, At: xi, A0: z0, A1: z1, Y0: G, Y1: G + h}
		k.decorate(g, inner, ishtarGateGlaze{W: z1 - z0, H: h, VStart: 1.0, Ha: 3.6, Ratio: 1.5,
			Kinds: lions, Dir: -s, Pitch: 6.5, FriezeHeight: 1.7, Cornice: Gold}.pattern(k))
		// Outer face segments between the towers.
		edges := append([]float64{z0 - 1.5}, towers...)
		edges = append(edges, z1+1.5)
		for i := 0; i+1 < len(edges); i++ {
			a0, a1 := edges[i]+1.5, edges[i+1]-1.5
			if a1-a0 < 2 {
				continue
			}
			f := ishtarGateFace{AlongX: false, Sign: s, At: xo, A0: a0, A1: a1, Y0: G, Y1: G + h}
			k.decorate(g, f, ishtarGateGlaze{W: a1 - a0, H: h, VStart: 1.0, Ha: 3.6, Ratio: 1.5,
				Kinds: lions, Dir: s, FriezeHeight: 1.7, Cornice: Gold}.pattern(k))
		}
		for _, z := range towers {
			for _, f := range []ishtarGateFace{
				{AlongX: false, Sign: s, At: 10.5 * s, A0: z - 1.5, A1: z + 1.5, Y0: G, Y1: G + h + 2},
				{AlongX: true, Sign: 1, At: z + 1.5, A0: math.Min(xo, 10.5*s), A1: math.Max(xo, 10.5*s), Y0: G, Y1: G + h + 2},
				{AlongX: true, Sign: -1, At: z - 1.5, A0: math.Min(xo, 10.5*s), A1: math.Max(xo, 10.5*s), Y0: G, Y1: G + h + 2},
			} {
				k.decorate(g, f, ishtarGateGlaze{W: f.A1 - f.A0, H: h + 2, Cornice: Gold}.pattern(k))
			}
			tl, th := math.Min(xo, 10.5*s), math.Max(xo, 10.5*s)
			k.rectMerlons(g, Navy, White, tl, th, z-1.5, z+1.5, G+h+2, 2.2, 0.8)
		}
		// End face toward the north.
		k.decorate(g, ishtarGateFace{AlongX: true, Sign: 1, At: z1, A0: lo, A1: hi, Y0: G, Y1: G + h},
			ishtarGateGlaze{W: 2, H: h, FriezeHeight: 1.7, Cornice: Gold}.pattern(k))
		// Parapet: merlons along the outer edge.
		ml, mh := math.Min(xo-s, xo), math.Max(xo-s, xo)
		k.merlons(g, Navy, White, false, z0, z1, ml, mh, G+h, 2.2, 0.8)
	}
}

// cityWall builds a mud-brick wall running along X with towers projecting toward +Z.
func (k IshtarGate) cityWall(g *Grid, xa, xb, z0, z1, h, th float64, towers []float64, tw, tp float64) {
	const G = ishtarGateG
	x0, x1 := math.Min(xa, xb), math.Max(xa, xb)
	k.fill(g, Sand, Block(V(x0, 0.4, z0), V(x1, G+h, z1)))
	k.carve(g, Block(V(x0+1.2, G, z0+1.0), V(x1-1.2, G+h-1, z1-1.0)))
	brick := func(H float64) ishtarGatePattern {
		return func(u, v, hh float64) (Color, bool, bool) { return k.mudBrick(u, v, hh, H) }
	}
	for _, f := range []ishtarGateFace{
		{AlongX: true, Sign: 1, At: z1, A0: x0, A1: x1, Y0: G, Y1: G + h},
		{AlongX: true, Sign: -1, At: z0, A0: x0, A1: x1, Y0: G, Y1: G + h},
		{AlongX: false, Sign: 1, At: x1, A0: z0, A1: z1, Y0: G, Y1: G + h},
		{AlongX: false, Sign: -1, At: x0, A0: z0, A1: z1, Y0: G, Y1: G + h},
	} {
		k.decorate(g, f, brick(h))
	}
	k.merlons(g, Sand, Sand, true, x0, x1, z1-1, z1, G+h, 2.4, 0.8)
	k.merlons(g, Sand, Sand, true, x0, x1, z0, z0+1, G+h, 2.4, 0.8)
	for _, x := range towers {
		a, b := x-tw/2, x+tw/2
		k.box(g, Sand, a, b, z0-0.5, z1+tp, 0.4, G+h+th, 1.0)
		for _, f := range []ishtarGateFace{
			{AlongX: true, Sign: 1, At: z1 + tp, A0: a, A1: b, Y0: G, Y1: G + h + th},
			{AlongX: true, Sign: -1, At: z0 - 0.5, A0: a, A1: b, Y0: G, Y1: G + h + th},
			{AlongX: false, Sign: 1, At: b, A0: z0 - 0.5, A1: z1 + tp, Y0: G, Y1: G + h + th},
			{AlongX: false, Sign: -1, At: a, A0: z0 - 0.5, A1: z1 + tp, Y0: G, Y1: G + h + th},
		} {
			k.decorate(g, f, brick(h+th))
		}
		k.rectMerlons(g, Sand, Sand, a, b, z0-0.5, z1+tp, G+h+th, 2.4, 0.8)
		// Dark arrow slits.
		for _, y := range []float64{G + h - 2.5, G + h + th - 3} {
			k.fill(g, Charcoal, Block(V(x-0.5, y-1.2, z1+tp-0.6), V(x+0.5, y, z1+tp+0.05)))
		}
	}
}

// ---------------------------------------------------------------------------
// The city behind the gate.

func (k IshtarGate) ziggurat(g *Grid) {
	const G = ishtarGateG
	cx, cz := 25.0, -32.0
	tiers := []struct {
		hw, y0, y1 float64
		body, rib  Color
	}{
		{9, 0.4, 7, Sand, Terracotta},
		{7, 7, 11.5, Terracotta, Brick},
		{5.2, 11.5, 15, Sand, Terracotta},
		{3.7, 15, 18, Terracotta, Brick},
	}
	for _, t := range tiers {
		y1 := G + t.y1
		k.box(g, t.body, cx-t.hw, cx+t.hw, cz-t.hw, cz+t.hw, t.y0, y1, 1.0)
		body, rib := t.body, t.rib
		top := y1
		k.shade(g, Box(Block(V(cx-t.hw-0.1, t.y0, cz-t.hw-0.1), V(cx+t.hw+0.1, y1, cz+t.hw+0.1))), true,
			func(c Cell, p Vec3) (Color, bool) {
				if p.Y > top-0.7 {
					return rib, true
				}
				// Vertical buttress ribs on each face.
				dx, dz := p.X-cx, p.Z-cz
				a := dx
				if math.Abs(dx) > math.Abs(dz) {
					a = dz
				}
				if math.Mod(a+60, 2.6) < 0.9 {
					return rib, true
				}
				if ishtarGateHash(c, 31) < 0.06 {
					return Stone, true
				}
				return body, true
			})
	}
	// Blue-glazed top shrine with gold trim and horned corners.
	sy := G + 18
	k.box(g, Navy, cx-2.8, cx+2.8, cz-2.4, cz+2.4, sy, sy+4, 0.9)
	k.fill(g, Gold, Block(V(cx-3.0, sy+3.2, cz-2.6), V(cx+3.0, sy+4.0, cz+2.6)))
	for _, sx := range []float64{-1, 1} {
		for _, sz := range []float64{-1, 1} {
			k.fill(g, Gold, Block(V(cx+sx*2.8-0.5, sy+4, cz+sz*2.4-0.5), V(cx+sx*2.8+0.5, sy+5.3, cz+sz*2.4+0.5)))
		}
	}
	k.fill(g, Charcoal, Block(V(cx-0.8, sy, cz+1.6), V(cx+0.8, sy+2.4, cz+2.5)))
	// The great perpendicular stair climbs from the south (+Z) to the first terrace.
	k.shade(g, Box{V(cx-1.8, 0.4, cz+7), V(cx+1.8, G+7, cz+15.2)}, false, func(c Cell, p Vec3) (Color, bool) {
		hgt := G + (cz+15-p.Z)*7/6
		hgt = math.Floor(hgt/0.8) * 0.8
		if p.Y > hgt {
			return 0, false
		}
		if math.Abs(p.X-cx) > 1.0 {
			return Terracotta, true
		}
		if hgt-p.Y < 0.45 {
			return White, true
		}
		return Sand, true
	})
}

func (k IshtarGate) house(g *Grid, cx, cz, wx, wz, h float64, seed int) {
	const G = ishtarGateG
	cols := []Color{White, Stone, Sand, White, Stone, Terracotta}
	body := cols[int(Hash01(seed, 1)*float64(len(cols)))]
	x0, x1, z0, z1 := cx-wx/2, cx+wx/2, cz-wz/2, cz+wz/2
	k.box(g, body, x0, x1, z0, z1, 0.4, G+h, 0.9)
	// Low parapet.
	k.fill(g, body, Subtract(Block(V(x0, G+h, z0), V(x1, G+h+0.7, z1)), Block(V(x0+0.8, G+h, z0+0.8), V(x1-0.8, G+h+1, z1-0.8))))
	// Interior shadow, door and small windows.
	k.paint(g, Charcoal, Block(V(x0+0.5, G-0.5, z0+0.5), V(x1-0.5, G+h-0.5, z1-0.5)))
	k.carve(g, Block(V(cx-0.7, G, z1-1), V(cx+0.7, G+2.2, z1+0.1)))
	if h > 4 {
		for _, wx2 := range []float64{x0 + 1.4, x1 - 1.4} {
			k.carve(g, Block(V(wx2-0.5, G+h-2.2, z1-1), V(wx2+0.5, G+h-1.2, z1+0.1)))
		}
	}
	if Hash01(seed, 7) < 0.35 {
		k.box(g, Terracotta, x0+0.8, x0+0.8+wx*0.45, z0+0.8, z0+0.8+wz*0.45, G+h, G+h+2.2, 0.8)
	}
}

func (k IshtarGate) city(g *Grid) {
	seed := 0
	for z := -38.5; z <= -25; z += 6.5 {
		for x := -26.0; x <= 16; x += 6.5 {
			seed++
			jx := (Hash01(seed, 2) - 0.5) * 1.5
			jz := (Hash01(seed, 3) - 0.5) * 1.5
			cx, cz := x+jx, z+jz
			wx := 3.8 + 1.8*Hash01(seed, 5)
			wz := 3.8 + 1.8*Hash01(seed, 6)
			h := 3.2 + 2.6*Hash01(seed, 8)
			switch {
			case math.Abs(cx) < 8.5:
				continue // The Processional Way continues south.
			case cx+wx/2 > 15.5:
				continue // Etemenanki.
			case Hash01(seed, 4) < 0.12:
				continue
			}
			k.house(g, cx, cz, wx, wz, h, seed)
		}
	}
}

func (k IshtarGate) palm(g *Grid, at Vec2, h float64, seed int) {
	const G = ishtarGateG
	lean := Hash01(seed, 1) * 2 * math.Pi
	amt := 0.4 + 0.9*Hash01(seed, 2)
	top := V(at.X+math.Cos(lean)*amt, G+h, at.Y+math.Sin(lean)*amt)
	k.fill(g, Wood, Capsule{A: at.At(0.6), B: top, Radius: 0.5})
	k.fill(g, Terracotta, Sphere(top.Add(V(0, -0.6, 0)), 0.75))
	n := 7
	for i := 0; i < n; i++ {
		a := 2*math.Pi*float64(i)/float64(n) + Hash01(seed, 3)
		d := V(math.Cos(a), 0, math.Sin(a))
		mid := top.Add(d.Scale(1.7)).Add(V(0, 0.55, 0))
		tip := top.Add(d.Scale(3.3)).Add(V(0, -0.9, 0))
		col := Foliage
		if i%3 == 0 {
			col = Grass
		}
		k.fill(g, col, Capsule{A: top, B: mid, Radius: 0.5}, Capsule{A: mid, B: tip, Radius: 0.45})
	}
}

func (k IshtarGate) palms(g *Grid) {
	seed := 100
	// Along the Euphrates bank.
	for z := -38.0; z <= 37; z += 5 {
		seed++
		if (z > -18 && z < -8) || (z > 1 && z < 10) || Hash01(seed, 9) < 0.15 {
			continue
		}
		k.palm(g, P(-28.4+2*Hash01(seed, 4), z), 7+2.5*Hash01(seed, 5), seed)
	}
	// Groves either side of the Processional Way.
	for _, s := range []float64{-1, 1} {
		for z := 15.0; z <= 36; z += 6.2 {
			for x := 15.0; x <= 34; x += 5.6 {
				seed++
				px := s * (x + 1.6*(Hash01(seed, 6)-0.5))
				pz := z + 1.6*(Hash01(seed, 7)-0.5)
				if px < -27 || Hash01(seed, 8) < 0.15 {
					continue
				}
				h := 6.5 + 2.5*Hash01(seed, 5)
				if s > 0 {
					// Keep the east grove low and clear of the gate in the hero view.
					if pz < 20 || (pz < 26 && px < 21) {
						continue
					}
					h -= 1.2
				}
				k.palm(g, P(px, pz), h, seed)
			}
		}
	}
	// A few in the city.
	for _, p := range []Vec2{P(-6, -23), P(6, -29), P(-6, -36), P(12, -21), P(35, -20), P(-26, -19.5), P(-12, -20)} {
		seed++
		k.palm(g, p, 7+2*Hash01(seed, 5), seed)
	}
}

func (k IshtarGate) Build(g *Grid) Order {
	k.ground(g)
	k.city(g)
	k.ziggurat(g)
	k.palms(g)
	k.gate(g)
	k.processional(g)
	// Outer (front) city wall and the larger inner wall, east and west.
	k.cityWall(g, 18.3, 35, 3, 7, 10, 3.5, []float64{26.5, 33.5}, 4.5, 2.0)
	k.cityWall(g, -18.3, ishtarGateBank, 3, 7, 10, 3.5, []float64{-25, -30}, 4.5, 2.0)
	k.cityWall(g, 23.3, 35, -16, -10, 14, 4.5, []float64{32}, 5.5, 2.5)
	k.cityWall(g, -23.3, ishtarGateBank, -16, -10, 14, 4.5, []float64{-29}, 5.5, 2.5)
	return Along(P(0, -1))
}
