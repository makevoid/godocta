package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// Borobudur is the 9th-century Buddhist temple-mountain in central Java: a stepped
// stone mandala on a low hill. A broad redented foot carries four square galleries
// whose main walls are lined with relief panels and whose balustrades are crowned by
// small stupas over Buddha niches. Four axial stairways climb through arched gateways
// on every level to three circular terraces ringed by 72 perforated bell stupas (32,
// 24 and 16) with seated Buddhas inside, around the great central stupa and its spire.
// Grey andesite with moss, on a lawn with palms and approach paths. Sides face ±X/±Z.
type Borobudur struct{}

func init() { Register(Borobudur{}) }

func (Borobudur) Name() string  { return "borobudur" }
func (Borobudur) Title() string { return "Borobudur" }

const (
	borobudurG     = 4.5  // Hill-top ground level.
	borobudurSite  = 50.0 // Site superellipse half-size.
	borobudurPlat  = 41.5 // Plateau half-size before the hill slopes away.
	borobudurNotch = 1.3  // Corner redent step.
	borobudurWall  = 1.2  // Terrace wall thickness.
	borobudurHW    = 1.4  // Stair half width.
	borobudurBal   = 1.5  // Balustrade height.
)

var (
	borobudurR   = [5]float64{38, 34, 31, 28, 25}            // Square level half-sizes.
	borobudurY   = [5]float64{7.1, 9.9, 12.7, 15.5, 18.3}    // Square floor tops.
	borobudurCR  = [3]float64{23.0, 18.2, 13.6}              // Circular terrace radii.
	borobudurCY  = [3]float64{19.5, 20.7, 21.9}              // Circular terrace tops.
	borobudurRng = [3]float64{21.0, 16.4, 11.8}              // Bell stupa ring radii.
	borobudurNum = [3]int{32, 24, 16}                        // Bell stupas per ring.
	borobudurDir = [4]Vec2{{0, 1}, {1, 0}, {0, -1}, {-1, 0}} // Outward side normals.
)

func (Borobudur) mod(v, m float64) float64 { return math.Mod(math.Mod(v, m)+m, m) }

func (Borobudur) smooth(a, b, x float64) float64 {
	t := math.Max(0, math.Min(1, (x-a)/(b-a)))
	return t * t * (3 - 2*t)
}

// redent tests a plan offset against a square of half-size r whose corners step in
// three times, the Borobudur terrace outline.
func (Borobudur) redent(dx, dz, r float64) bool {
	const d = borobudurNotch
	ax, az := math.Abs(dx), math.Abs(dz)
	if ax > r || az > r {
		return false
	}
	return az <= r-3*d || ax <= r-3*d || (ax <= r-d && az <= r-2*d) || (ax <= r-2*d && az <= r-d)
}

// worn picks pale replaced flagstones on terrace floors.
func (Borobudur) worn(p Vec3) bool {
	return Hash01(int(math.Floor(p.X/1.6)), int(math.Floor(p.Z/1.6)), 51) < 0.1
}

func (b Borobudur) shell(p Vec3, r, t float64) bool {
	return b.redent(p.X, p.Z, r) && !b.redent(p.X, p.Z, r-t)
}

// edge is the outward distance of a redented side at along-coordinate v.
func (Borobudur) edge(r, v float64) float64 {
	const d = borobudurNotch
	av := math.Abs(v)
	switch {
	case av <= r-3*d:
		return r
	case av <= r-2*d:
		return r - d
	case av <= r-d:
		return r - 2*d
	}
	return r - 3*d
}

// local converts a world point to side coordinates: u outward, v along the side.
func (Borobudur) local(side int, p Vec3) (float64, float64) {
	n := borobudurDir[side]
	return p.X*n.X + p.Z*n.Y, p.X*n.Y - p.Z*n.X
}

// at converts side coordinates back to a world point.
func (Borobudur) at(side int, u, v, y float64) Vec3 {
	n := borobudurDir[side]
	return V(u*n.X+v*n.Y, y, u*n.Y-v*n.X)
}

// box bounds a side-local rectangle in world space.
func (b Borobudur) box(side int, u0, u1, v0, v1, y0, y1 float64) Box {
	return BoxOf(b.at(side, u0, v0, y0), b.at(side, u1, v1, y1)).Union(BoxOf(b.at(side, u0, v1, y0), b.at(side, u1, v0, y1)))
}

// sideFunc is a solid in side-local coordinates over a local rectangle.
func (b Borobudur) sideFunc(side int, u0, u1, v0, v1, y0, y1 float64, f func(u, v, y float64) bool) Solid {
	return Func{b.box(side, u0, u1, v0, v1, y0, y1).Grow(0.01), func(p Vec3) bool {
		u, v := b.local(side, p)
		return u >= u0 && u <= u1 && v >= v0 && v <= v1 && p.Y >= y0 && p.Y <= y1 && f(u, v, p.Y)
	}}
}

// ground is the hill height at a plan point, or false outside the site.
func (b Borobudur) ground(x, z float64) (float64, bool) {
	S := borobudurSite
	rho := math.Pow(math.Pow(math.Abs(x)/S, 6)+math.Pow(math.Abs(z)/S, 6), 1.0/6)
	if rho > 1 {
		return 0, false
	}
	t := b.smooth(borobudurPlat, S, rho*S)
	h := borobudurG - (borobudurG-0.4)*t
	h += 0.35 * t * math.Sin(x*0.21+1.3) * math.Cos(z*0.17-0.4)
	return h, true
}

// terrain lays the grassy hill as a surface shell, skipping the temple's footprint.
func (b Borobudur) terrain(g *Grid) {
	s := g.Span()
	lo := func(v float64) int32 { return int32(math.Floor(v * 2 / s)) }
	S := borobudurSite
	for par := int32(0); par < 2; par++ {
		for X := lo(-S - 1); X <= lo(S)+1; X++ {
			if X&1 != par {
				continue
			}
			for Z := lo(-S - 1); Z <= lo(S)+1; Z++ {
				if Z&1 != par {
					continue
				}
				x, z := float64(X)*s/2, float64(Z)*s/2
				if b.redent(x, z, borobudurR[0]-1.6) {
					continue
				}
				h, ok := b.ground(x, z)
				if !ok {
					continue
				}
				low := h
				for _, o := range [4]Vec2{{0.8, 0}, {-0.8, 0}, {0, 0.8}, {0, -0.8}} {
					nh, nok := b.ground(x+o.X, z+o.Y)
					if !nok {
						nh = 0.0
					}
					low = math.Min(low, nh)
				}
				col := Grass
				ax, az := math.Abs(x), math.Abs(z)
				m := math.Max(ax, az)
				switch {
				case math.Min(ax, az) < 1.9 && m > borobudurR[0]-0.5:
					col = Sand // Axial approach paths.
				case m > borobudurR[0]+1.3 && m < borobudurR[0]+3.0 && h > borobudurG-0.05:
					col = Sand // Circumambulation path round the foot.
				case math.Sin(x*0.23+2*math.Sin(z*0.13))+math.Sin(z*0.27+1.5*math.Sin(x*0.11))+0.4*Hash01(int(X), int(Z), 3) > 1.05:
					col = Foliage
				}
				for Y := lo(low - 0.8); Y <= lo(h)+1; Y++ {
					y := float64(Y) * s / 2
					if Y&1 != par || y < low-0.8 || y > h {
						continue
					}
					c := Cell{X, Y, Z}
					g.Remove(c)
					g.Use(col)
					g.Add(c)
				}
			}
		}
	}
}

// levels raises the foot, the four square galleries and their balustrades.
func (b Borobudur) levels(g *Grid) {
	R, Y := borobudurR, borobudurY
	big := Box{V(-R[0]-1, borobudurG-2, -R[0]-1), V(R[0]+1, Y[4]+borobudurBal+1, R[0]+1)}
	for k := 0; k < 5; k++ {
		k := k
		y0 := borobudurG - 1.2
		if k > 0 {
			y0 = Y[k-1] - 1.0
		}
		// Main wall.
		g.Use(DarkStone).Fill(Func{big, func(p Vec3) bool {
			return p.Y >= y0 && p.Y <= Y[k] && b.shell(p, R[k], borobudurWall)
		}})
		// Floor annulus up to the next wall (or the first circular terrace).
		g.Use(DarkStone).Fill(Func{big, func(p Vec3) bool {
			if p.Y < Y[k]-1.0 || p.Y > Y[k] || !b.redent(p.X, p.Z, R[k]) {
				return false
			}
			if k < 4 {
				return !b.redent(p.X, p.Z, R[k+1]-borobudurWall)
			}
			return math.Hypot(p.X, p.Z) > borobudurCR[0]-borobudurWall
		}})
		g.Paint(Stone, Func{big, func(p Vec3) bool {
			return p.Y > Y[k]-1.0 && p.Y <= Y[k] && b.redent(p.X, p.Z, R[k]-borobudurWall) && b.worn(p)
		}})
		// Balustrade with a pale coping.
		g.Use(DarkStone).Fill(Func{big, func(p Vec3) bool {
			return p.Y > Y[k] && p.Y <= Y[k]+borobudurBal && b.shell(p, R[k], 1.0)
		}})
		g.Paint(Stone, Func{big, func(p Vec3) bool {
			return p.Y > Y[k]+borobudurBal-0.3 && p.Y <= Y[k]+borobudurBal && b.shell(p, R[k], 1.0)
		}})
	}
}

// reliefs paints the main walls as rows of carved panels between pilasters, with pale
// mouldings, then scatters moss over the stonework.
func (b Borobudur) reliefs(g *Grid) {
	R, Y := borobudurR, borobudurY
	big := Box{V(-R[0]-1, borobudurG-2, -R[0]-1), V(R[0]+1, Y[4]+borobudurBal+1, R[0]+1)}
	for k := 0; k < 5; k++ {
		k := k
		y0 := borobudurG
		if k > 0 {
			y0 = Y[k-1]
		}
		g.Paint(Stone, Func{big, func(p Vec3) bool {
			if p.Y < y0-0.3 || p.Y > Y[k] || !b.shell(p, R[k], borobudurWall) {
				return false
			}
			along := p.Z
			if math.Abs(p.Z) > math.Abs(p.X) {
				along = p.X
			}
			if p.Y < y0+0.35 || p.Y > Y[k]-0.4 {
				return true // Base and cornice mouldings.
			}
			if k == 0 {
				return math.Abs(p.Y-(y0+Y[k])/2) < 0.2
			}
			panel := int(math.Floor((along + 1.1) / 2.2))
			h := Hash01(panel, k, int(math.Floor(p.X+p.Z)), 11)
			f := Hash01(int(math.Floor(p.X*2.6)), int(math.Floor(p.Y*2.6)), int(math.Floor(p.Z*2.6)), 12)
			return b.mod(along+1.1, 2.2) > 0.5 && f < 0.08+0.22*h
		}})
	}
	// Moss on damp stone, heavier low down.
	g.Paint(Foliage, Func{big, func(p Vec3) bool {
		if !b.redent(p.X, p.Z, R[0]+0.2) || p.Y < borobudurG-0.2 {
			return false
		}
		c := Hash01(int(math.Floor(p.X/2.3)), int(math.Floor(p.Y/1.7)), int(math.Floor(p.Z/2.3)), 21)
		f := Hash01(int(math.Floor(p.X*2.6)), int(math.Floor(p.Y*2.6)), int(math.Floor(p.Z*2.6)), 22)
		lim := 0.11 - 0.004*(p.Y-borobudurG)
		return c < lim && f < 0.6
	}})
}

// niches cuts Buddha niches into the balustrades and crowns them with small stupas.
func (b Borobudur) niches(g *Grid) {
	R, Y := borobudurR, borobudurY
	const d = borobudurNotch
	for k := 0; k < 5; k++ {
		r, y := R[k], Y[k]
		var spots []float64
		for v := 3.8; v <= r-3*d-0.6; v += 2.4 {
			spots = append(spots, v, -v)
		}
		for _, v := range [3]float64{r - 2.5*d, r - 1.5*d, r - 0.5*d} {
			spots = append(spots, v, -v)
		}
		for side := 0; side < 4; side++ {
			for _, v := range spots {
				e := b.edge(r, v)
				if math.Abs(v) <= r-3*d {
					// Niche: a dark recess with a seated Buddha.
					g.Carve(b.sideFunc(side, e-0.45, e+0.1, v-0.55, v+0.55, y+0.25, y+1.15, func(u, vv, yy float64) bool { return true }))
					g.Paint(Charcoal, b.sideFunc(side, e-1.0, e-0.3, v-0.55, v+0.55, y+0.2, y+1.2, func(u, vv, yy float64) bool { return true }))
					g.Use(Stone).AddAt(b.at(side, e-0.5, v, y+0.6))
				}
				base := b.at(side, e-0.5, v, y+borobudurBal)
				g.Use(Stone).Fill(Sphere(base.Add(V(0, 0.35, 0)), 0.55))
				g.Use(DarkStone).Fill(Capsule{A: base.Add(V(0, 0.6, 0)), B: base.Add(V(0, 1.25, 0)), Radius: 0.3})
			}
		}
	}
}

// climb is one flight of an axial stair: it is recessed rec into the wall at edge
// and projects proj in front of it, rising from y0 to y1.
type borobudurClimb struct {
	edge, y0, y1, rec, proj, hw float64
	gate                        bool
}

func (b Borobudur) stairs(g *Grid) {
	R, Y, CR, CY := borobudurR, borobudurY, borobudurCR, borobudurCY
	climbs := []borobudurClimb{{R[0], borobudurG, Y[0], 1.5, 1.4, borobudurHW, true}}
	for k := 1; k < 5; k++ {
		climbs = append(climbs, borobudurClimb{R[k], Y[k-1], Y[k], 1.5, 1.0, borobudurHW, true})
	}
	prev := Y[4]
	for i := 0; i < 3; i++ {
		climbs = append(climbs, borobudurClimb{CR[i], prev, CY[i], 0.9, 0.9, 1.2, false})
		prev = CY[i]
	}
	for side := 0; side < 4; side++ {
		for _, c := range climbs {
			c := c
			u0, u1 := c.edge-c.rec, c.edge+c.proj
			top := c.y1 + 3.0
			if c.gate {
				top = c.y1 + borobudurBal + 1.5
			}
			g.Carve(b.sideFunc(side, u0, u1+0.05, -c.hw, c.hw, c.y0+0.05, top, func(u, v, y float64) bool { return true }))
			if c.gate {
				b.gate(g, side, c.edge, c.y1, c.hw)
			}
			n := math.Max(2, math.Round((c.y1-c.y0)/0.42))
			stepTop := func(u float64) float64 {
				f := math.Max(0, math.Min(1, (u1-u)/(u1-u0)))
				return c.y0 + (c.y1-c.y0)*math.Min(1, (math.Floor(f*n)+1)/n)
			}
			g.Use(Stone).Fill(b.sideFunc(side, u0, u1, -c.hw, c.hw, c.y0-1.0, c.y1, func(u, v, y float64) bool {
				return y <= stepTop(u)
			}))
			// Cheek walls beside the projecting flight.
			g.Use(DarkStone).Fill(b.sideFunc(side, c.edge, u1, -c.hw-0.8, c.hw+0.8, c.y0-0.5, c.y1+1.0, func(u, v, y float64) bool {
				return math.Abs(v) > c.hw && y <= stepTop(u)+0.7
			}))
		}
	}
}

// gate is a kala-makara gateway arching over the top of a stair at a balustrade.
func (b Borobudur) gate(g *Grid, side int, edge, y, hw float64) {
	const spring = 1.2
	g.Use(Stone).Fill(b.sideFunc(side, edge-1.3, edge+0.3, -hw-1.1, hw+1.1, y, y+3.4, func(u, v, yy float64) bool { return true }))
	g.Use(DarkStone).Fill(b.sideFunc(side, edge-1.1, edge+0.1, -hw-0.6, hw+0.6, y+3.4, y+4.0, func(u, v, yy float64) bool { return true }))
	g.Use(Stone).Fill(b.sideFunc(side, edge-0.9, edge-0.1, -hw-0.1, hw+0.1, y+4.0, y+4.5, func(u, v, yy float64) bool { return true }))
	g.Carve(b.sideFunc(side, edge-1.4, edge+0.4, -hw, hw, y-0.01, y+spring+hw, func(u, v, yy float64) bool {
		r := yy - y - spring
		return r <= 0 || v*v+r*r <= hw*hw
	}))
	// Kala head above the arch.
	g.Paint(Charcoal, b.sideFunc(side, edge-0.1, edge+0.4, -0.45, 0.45, y+spring+hw+0.6, y+3.1, func(u, v, yy float64) bool { return true }))
	base := b.at(side, edge-0.5, 0, y+4.5)
	g.Use(DarkStone).Fill(Sphere(base.Add(V(0, 0.4, 0)), 0.65))
	g.Use(Stone).Fill(Capsule{A: base.Add(V(0, 0.8, 0)), B: base.Add(V(0, 1.6, 0)), Radius: 0.32})
}

// circles raises the three circular terraces.
func (b Borobudur) circles(g *Grid) {
	CR, CY := borobudurCR, borobudurCY
	prev := borobudurY[4]
	for i := 0; i < 3; i++ {
		g.Use(DarkStone).Fill(Cylinder{P(0, 0), CR[i], CR[i] - borobudurWall, prev - 1.0, CY[i]})
		inner := 6.4
		if i < 2 {
			inner = CR[i+1] - borobudurWall
		}
		g.Use(DarkStone).Fill(Cylinder{P(0, 0), CR[i], inner, CY[i] - 1.0, CY[i]})
		g.Paint(Stone, Func{Cylinder{P(0, 0), CR[i] - borobudurWall, inner, CY[i] - 1.0, CY[i]}.Bounds(), b.worn})
		g.Paint(Stone, Cylinder{P(0, 0), CR[i] + 0.1, CR[i] - borobudurWall, CY[i] - 0.35, CY[i]})
		prev = CY[i]
	}
}

// bell is a perforated bell stupa on a lotus cushion with a seated Buddha inside.
// Lower rings have diamond openings, the top ring square ones.
func (b Borobudur) bell(g *Grid, c Vec2, y float64, square bool) {
	const R, H = 1.42, 2.7
	yb := y + 0.45
	prof := func(py float64) float64 {
		u := (py - yb) / H
		if u < 0 || u > 1 {
			return -1
		}
		return R * math.Pow(1-math.Pow(u, 2.2), 0.5)
	}
	box := Box{V(c.X-R-0.4, y-0.4, c.Y-R-0.4), V(c.X+R+0.4, yb+H+1.8, c.Y+R+0.4)}
	g.Use(Stone).FillVolume(Cylinder{c, R + 0.05, 0, y - 0.3, yb})
	g.Use(Stone).FillVolume(Func{box, func(p Vec3) bool {
		r := prof(p.Y)
		if r < 0 {
			return false
		}
		d := math.Hypot(p.X-c.X, p.Z-c.Y)
		if d > r || d < r-0.75 {
			return false
		}
		u := (p.Y - yb) / H
		if u < 0.08 || u > 0.82 {
			return true
		}
		row := math.Floor((u - 0.08) / 0.37)
		fv := (u-0.08)/0.37 - row
		fa := b.mod(math.Atan2(p.Z-c.Y, p.X-c.X)*5/(2*math.Pi)+0.5*row, 1)
		if square {
			return !(math.Abs(fa-0.5) < 0.3 && math.Abs(fv-0.5) < 0.32)
		}
		return math.Abs(fa-0.5)*2+math.Abs(fv-0.5)*2 >= 0.8
	}})
	g.Paint(Charcoal, Func{box, func(p Vec3) bool {
		r := prof(p.Y)
		if p.Y < y-0.1 || p.Y > yb+H*0.85 {
			return false
		}
		if r < 0 {
			r = R
		}
		return math.Hypot(p.X-c.X, p.Z-c.Y) < r-0.5
	}})
	g.Use(Stone).FillVolume(Ellipsoid{c.At(yb + 0.4), V(0.45, 0.6, 0.45)})
	g.Use(DarkStone).Fill(BlockAt(c, 1.0, 1.0, yb+H-0.2, yb+H+0.45))
	g.Use(Stone).Fill(Capsule{A: c.At(yb + H + 0.45), B: c.At(yb + H + 1.4), Radius: 0.38})
}

// rings sets the 72 bell stupas in quadrant groups, leaving the stair axes clear.
func (b Borobudur) rings(g *Grid) {
	for i := 0; i < 3; i++ {
		R, n := borobudurRng[i], borobudurNum[i]/4
		gap := 3.2 / R
		step := (math.Pi/2 - 2*gap) / float64(n-1)
		for q := 0; q < 4; q++ {
			for j := 0; j < n; j++ {
				a := float64(q)*math.Pi/2 + gap + float64(j)*step
				b.bell(g, Polar(P(0, 0), R, a), borobudurCY[i], i == 2)
			}
		}
	}
}

// crown is the great central stupa: lotus base, dome, harmika and tiered spire.
func (b Borobudur) crown(g *Grid) {
	y := borobudurCY[2]
	c := P(0, 0)
	g.Use(DarkStone).Fill(Cylinder{c, 7.6, 6.2, y - 0.5, y + 0.5})
	g.Use(Stone).FillVolume(Cylinder{c, 7.2, 6.0, y + 0.5, y + 1.3})
	ctr := c.At(y + 1.3)
	g.Use(Stone).FillVolume(Intersect(
		Subtract(Ellipsoid{ctr, V(6.8, 5.6, 6.8)}, Ellipsoid{ctr, V(5.6, 4.4, 5.6)}),
		Block(V(-8, y+1.3, -8), V(8, y+8, 8)),
	))
	g.Paint(DarkStone, Cylinder{c, 8, 0, y + 1.3, y + 1.8})
	g.Paint(DarkStone, Func{Box{V(-8, y+1.8, -8), V(8, y+8, 8)}, func(p Vec3) bool {
		k := int(math.Floor(math.Atan2(p.Z, p.X) * 9))
		return Hash01(k, int(math.Floor(p.Y/1.4)), 41) < 0.25
	}})
	top := y + 1.3 + 5.6
	g.Use(DarkStone).Fill(BlockAt(c, 3.8, 3.8, top-1.0, top+0.7))
	g.Use(Stone).Fill(BlockAt(c, 4.4, 4.4, top+0.7, top+1.2))
	g.Use(Stone).FillVolume(Frustum{c, 1.35, 0.45, top + 1.2, top + 6.2})
	for i := 0; i < 4; i++ {
		yy := top + 1.7 + float64(i)*1.05
		g.Use(DarkStone).FillVolume(Cylinder{c, 1.55 - 0.18*float64(i), 0, yy, yy + 0.45})
	}
}

// palm is a coconut palm: a curving trunk and drooping fronds.
func (Borobudur) palm(g *Grid, base Vec3, seed int) {
	h := 5.5 + 2.5*Hash01(seed, 1)
	lean := Polar(P(0, 0), 0.6+1.0*Hash01(seed, 2), Hash01(seed, 3)*2*math.Pi)
	mid := base.Plan().Add(lean.Scale(0.3)).At(base.Y + h*0.55)
	top := base.Plan().Add(lean).At(base.Y + h)
	g.Use(Wood).Fill(Capsule{A: base.Add(V(0, -0.5, 0)), B: mid, Radius: 0.45}, Capsule{A: mid, B: top, Radius: 0.42})
	g.Use(Foliage).FillVolume(Sphere(top, 0.8))
	for k := 0; k < 7; k++ {
		a := (float64(k) + 0.6*Hash01(seed, k, 4)) * 2 * math.Pi / 7
		m := Polar(top.Plan(), 1.6, a).At(top.Y + 0.5)
		tip := Polar(top.Plan(), 3.1, a).At(top.Y - 0.9)
		col := Foliage
		if k%2 == 0 {
			col = Grass
		}
		g.Use(col).Fill(Capsule{A: top, B: m, Radius: 0.45}, Capsule{A: m, B: tip, Radius: 0.4})
	}
}

// tree is a broad shade tree.
func (Borobudur) tree(g *Grid, base Vec3, seed int) {
	h := 3.0 + 1.5*Hash01(seed, 7)
	r := 2.2 + 0.9*Hash01(seed, 8)
	g.Use(Wood).Fill(Capsule{A: base.Add(V(0, -0.5, 0)), B: base.Add(V(0, h, 0)), Radius: 0.5})
	g.Use(Foliage).FillVolume(Ellipsoid{base.Add(V(0, h+r*0.5, 0)), V(r, r*0.75, r)})
}

// grove scatters palms and shade trees over the lawn and slopes, off the paths.
func (b Borobudur) grove(g *Grid) {
	const s = 5.2
	n := int(math.Ceil(borobudurSite / s))
	for i := -n; i <= n; i++ {
		for j := -n; j <= n; j++ {
			x := (float64(i) + 0.25 + 0.5*Hash01(i, j, 31)) * s
			z := (float64(j) + 0.25 + 0.5*Hash01(i, j, 32)) * s
			m := math.Max(math.Abs(x), math.Abs(z))
			if m < borobudurR[0]+4.5 || math.Min(math.Abs(x), math.Abs(z)) < 8.5 {
				continue
			}
			h, ok := b.ground(x, z)
			if !ok || h < 0.9 || Hash01(i, j, 33) > 0.8 {
				continue
			}
			seed := (i+20)*97 + j + 20
			if Hash01(i, j, 34) < 0.75 {
				b.palm(g, V(x, h, z), seed)
			} else {
				b.tree(g, V(x, h, z), seed)
			}
		}
	}
}

func (b Borobudur) Build(g *Grid) Order {
	b.terrain(g)
	b.levels(g)
	b.circles(g)
	b.reliefs(g)
	b.niches(g)
	b.stairs(g)
	b.rings(g)
	b.crown(g)
	b.grove(g)
	return Sweep(P(0, 0), math.Pi/2)
}
