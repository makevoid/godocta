package scenes

import (
	"math"

	. "godocta/gen/octa"
)

// PirateGalleon is a 17th-century galleon running before the wind: a curved hollow
// hull with tumblehome, charcoal wales, a gold stripe and two rows of gun ports; a
// tall stern castle with glazed stern and quarter galleries, a forecastle with a
// beakhead and bowsprit; three masts with fighting tops, crow's nests, billowing
// square sails, a lateen mizzen, rigging and a black flag. y = 0 is the keel just
// below the waterline (about y = 2). Bow points to +X.
type PirateGalleon struct {
	Length, Beam, Depth float64 // Hull length, beam and main-deck height. Defaults 112 × 34 × 17.
}

func init() { Register(PirateGalleon{}) }

func (PirateGalleon) Name() string  { return "pirate_galleon" }
func (PirateGalleon) Title() string { return "Pirate Galleon" }

// sheer is the rise of decks and bands toward both ends.
func (s PirateGalleon) sheer(x float64) float64 {
	u := x / (s.Length / 2)
	return 3.0 * u * u
}

// stern is the transom's X at height y; it rakes aft as it rises.
func (s PirateGalleon) stern(y float64) float64 {
	return -s.Length/2 - 0.25*math.Max(y-3, 0)
}

// bow is the stem's X at height y; it curves forward as it rises.
func (s PirateGalleon) bow(y float64) float64 {
	t := math.Min(math.Max(y/(s.Depth+6), 0), 1)
	return s.Length/2 - 7 + 13*math.Pow(t, 1.6)
}

// halfWidth is the hull's outer half-beam at (x, y), negative outside its length.
func (s PirateGalleon) halfWidth(x, y float64) float64 {
	xs, xb := s.stern(y), s.bow(y)
	if x < xs || x > xb {
		return -1
	}
	ym := 0.4 * s.Depth // Widest just above the waterline, then tumblehome.
	sec := 1.0
	if y < ym {
		t := (ym - y) / ym
		sec = 1 - 0.22*t*t
	} else {
		t := (y - ym) / (s.Depth + 14 - ym)
		sec = 1 - 0.26*t*t
	}
	plan := 1.0
	if fore := 0.1 * s.Length; x > fore {
		t := (x - fore) / (xb - fore)
		plan = math.Pow(math.Max(1-t*t, 0), 0.55)
	}
	if aft := -0.22 * s.Length; x < aft {
		t := (aft - x) / (aft - xs)
		plan = 1 - 0.32*t*t
		if y < ym { // Finer run toward the stern post below the waterline.
			plan *= 1 - 0.5*t*(ym-y)/ym
		}
	}
	return s.Beam / 2 * sec * plan
}

// innerWidth is the half-width of the hollow interior: the skin is about 1.3 thick
// both across and along the hull.
func (s PirateGalleon) innerWidth(x, y float64) float64 {
	const skin = 1.3
	a, b := s.halfWidth(x-skin, y), s.halfWidth(x+skin, y)
	if a < 0 || b < 0 {
		return -1
	}
	return math.Min(a, b) - skin
}

// Castle break positions along X.
func (s PirateGalleon) royalX() float64   { return -0.42 * s.Length }
func (s PirateGalleon) poopX() float64    { return -0.29 * s.Length }
func (s PirateGalleon) quarterX() float64 { return -0.13 * s.Length }
func (s PirateGalleon) foreX() float64    { return 0.25 * s.Length }

// deck is the walking surface at x (before sheer): poop, quarterdeck, waist or forecastle.
func (s PirateGalleon) deck(x float64) float64 {
	D := s.Depth
	switch {
	case x < s.royalX():
		return D + 11.6 + s.sheer(x)
	case x < s.poopX():
		return D + 8.2 + s.sheer(x)
	case x < s.quarterX():
		return D + 4.2 + s.sheer(x)
	case x < s.foreX():
		return D + s.sheer(x)
	}
	return D + 4 + s.sheer(x)
}

// rail is the top of the hull side at x.
func (s PirateGalleon) rail(x float64) float64 {
	D := s.Depth
	switch {
	case x < s.royalX():
		return D + 13.4 + s.sheer(x)
	case x < s.poopX():
		return D + 10 + s.sheer(x)
	case x < s.quarterX():
		return D + 6 + s.sheer(x)
	case x < s.foreX():
		return D + 3.4 + s.sheer(x)
	}
	return D + 5.8 + s.sheer(x)
}

// band selects the hull skin between heights y0 and y1 (above the sheer line),
// protruding by out beyond the hull and reaching in to depth.
func (s PirateGalleon) band(y0, y1, out, depth float64) Func {
	B := s.Beam
	return Func{Box{V(s.stern(y1+3)-1, y0, -B/2-out-1), V(s.bow(y1+3)+1, y1+3, B/2+out+1)}, func(p Vec3) bool {
		h := p.Y - s.sheer(p.X)
		if h < y0 || h > y1 || p.Y > s.rail(p.X)+0.2 {
			return false
		}
		w := s.halfWidth(p.X, p.Y)
		if w < 0 && p.X > 0 { // Wrap around the stem.
			w = s.halfWidth(s.bow(p.Y)-0.01, p.Y)
			if p.X > s.bow(p.Y)+out {
				return false
			}
		}
		z := math.Abs(p.Z)
		return w >= 0 && z <= w+out && z >= w-depth
	}}
}

func (s PirateGalleon) Build(g *Grid) Order {
	if s.Length == 0 {
		s.Length, s.Beam, s.Depth = 112, 34, 17
	}
	L, B, D := s.Length, s.Beam, s.Depth
	top := D + 17.0
	hullBox := Box{V(s.stern(top)-1, 0, -B/2-1), V(s.bow(top)+1, top, B/2+1)}

	// Hull: a curved shell, open above, filled by volume for a smooth silhouette.
	g.Use(Wood).FillVolume(Func{hullBox, func(p Vec3) bool {
		return p.Y <= s.rail(p.X) && math.Abs(p.Z) <= s.halfWidth(p.X, p.Y)
	}})
	g.Carve(Func{hullBox.Grow(1), func(p Vec3) bool {
		return math.Abs(p.Z) <= s.innerWidth(p.X, p.Y)
	}})

	// Decks: waist, quarterdeck, poop and forecastle, each clipped to the hull.
	deckSolid := func(x0, x1 float64) Func {
		return Func{Box{V(x0, D-1, -B/2), V(x1, top, B/2)}, func(p Vec3) bool {
			return math.Abs(p.Y-s.deck(p.X)) <= 0.5 && math.Abs(p.Z) <= s.halfWidth(p.X, p.Y)-0.4
		}}
	}
	g.Use(Sand).Fill(
		deckSolid(s.quarterX(), s.foreX()),
		deckSolid(s.poopX(), s.quarterX()),
		deckSolid(s.royalX(), s.poopX()),
		deckSolid(-L, s.royalX()),
		deckSolid(s.foreX(), L),
	)

	// Bulkheads at the castle breaks, with doors, and balustrades along the deck edges.
	g.Use(Wood)
	bulkhead := func(x, lower, upper float64, facing float64) {
		x0, x1 := x-1, x
		if facing < 0 {
			x0, x1 = x, x+1
		}
		g.Fill(Func{Box{V(x0, lower, -B/2), V(x1, upper+2.4, B/2)}, func(p Vec3) bool {
			return math.Abs(p.Z) <= s.halfWidth(p.X, p.Y)-0.4
		}})
		// Balustrade above the upper deck.
		g.Carve(Func{Box{V(x0, upper+0.5, -B/2), V(x1, upper+1.6, B/2)}, func(p Vec3) bool {
			return math.Mod(math.Abs(p.Z)+0.5, 2.2) > 1.1 && math.Abs(p.Z) < s.halfWidth(p.X, p.Y)-1.6
		}})
		for _, z := range []float64{-B / 5, 0, B / 5} {
			g.Carve(Arch{Sill: V(x-0.5*facing, lower+0.5, z), HalfWidth: 1.3, Spring: 1.6, Depth: 2.4, Through: AxisX})
		}
		g.Paint(Gold, Func{Box{V(x0, upper+1.6, -B/2), V(x1, upper+2.4, B/2)}, func(p Vec3) bool { return true }})
	}
	qx, px, fx, rx := s.quarterX(), s.poopX(), s.foreX(), s.royalX()
	bulkhead(rx, s.deck(rx+1)-0.5, s.deck(rx-1), 1)
	bulkhead(qx, s.deck(qx+1)-0.5, s.deck(qx-1), 1)
	bulkhead(px, s.deck(px+1)-0.5, s.deck(px-1), 1)
	bulkhead(fx, s.deck(fx-1)-0.5, s.deck(fx+1), -1)

	// Wales and gold stripe, following the sheer.
	g.Use(Charcoal).Fill(s.band(0.225*D, 0.3*D, 0.5, 0.8), s.band(0.625*D, 0.7*D, 0.5, 0.8))
	g.Use(Gold).Fill(s.band(0.825*D, 0.875*D, 0.3, 0.8))
	// Castle upper works: red panels between wooden pilasters over a gold molding,
	// and gilded rail caps outlining the stepped castles.
	castle := func(x float64) bool { return x < s.quarterX() || x > s.foreX() }
	g.Paint(Red, Func{hullBox, func(p Vec3) bool {
		h := p.Y - s.sheer(p.X)
		return castle(p.X) && h >= D+3.4 && p.Y <= s.rail(p.X)-0.9 &&
			math.Abs(p.Z) > s.innerWidth(p.X, p.Y)-0.2 && math.Mod(p.X+200, 4.4) > 1.1
	}})
	g.Paint(Gold, Func{hullBox, func(p Vec3) bool {
		h := p.Y - s.sheer(p.X)
		return castle(p.X) && h >= D+2.7 && h <= D+3.4 && math.Abs(p.Z) > s.innerWidth(p.X, p.Y)-0.2
	}})
	g.Paint(Gold, Func{hullBox, func(p Vec3) bool {
		return p.Y > s.rail(p.X)-0.6 && math.Abs(p.Z) > s.innerWidth(p.X, p.Y)-0.2
	}})
	// Deck planking seams.
	g.Paint(Wood, Func{hullBox, func(p Vec3) bool {
		return math.Abs(p.Y-s.deck(p.X)) <= 0.5 && math.Abs(p.Z) < s.innerWidth(p.X, p.Y) &&
			math.Mod(p.Z+100.2, 3.2) < 0.45
	}})

	// Gun ports: two rows of dark recesses with cannon muzzles.
	type port struct{ y0, y1, x0, x1, off float64 }
	for _, row := range []port{{0.41 * D, 0.525 * D, -0.36 * L, 0.32 * L, 0}, {D + 0.8, D + 2.4, -0.42 * L, 0.38 * L, 3.1}} {
		for x := row.x0 + row.off; x <= row.x1; x += 6.2 {
			xc := x
			h0, h1 := row.y0+s.sheer(xc), row.y1+s.sheer(xc)
			ym := (h0 + h1) / 2
			w := s.halfWidth(xc, ym)
			portBox := Box{V(xc-0.9, h0, -B/2-1), V(xc+0.9, h1, B/2+1)}
			g.Carve(Func{portBox, func(p Vec3) bool { return math.Abs(p.Z) >= s.halfWidth(p.X, p.Y)-0.4 }})
			g.Paint(Charcoal, Func{portBox, func(p Vec3) bool { return math.Abs(p.Z) >= s.halfWidth(p.X, p.Y)-1.6 }})
			inner := w - 0.8
			waist := row.y0 > D && xc > s.quarterX()+1.5 && xc < s.foreX()-1.5
			if waist { // Carriage guns on the open waist deck.
				inner = s.innerWidth(xc, ym) - 2.4
			}
			for _, side := range []float64{-1, 1} {
				g.With(DarkStone, func() {
					g.Fill(Capsule{V(xc, ym, side*inner), V(xc, ym, side*(w+1.4)), 0.5})
				})
				if waist {
					iw := s.innerWidth(xc, ym)
					dy := s.deck(xc) + 0.5
					g.With(Wood, func() {
						g.Fill(Block(V(xc-1.1, dy, side*(iw-0.4)), V(xc+1.1, ym-0.4, side*(iw-2.6))))
					})
				}
			}
		}
	}

	s.buildStern(g)
	s.buildBow(g)
	s.buildDeckGear(g)
	s.buildRig(g)
	return Along(P(1, 0))
}

// buildStern adds the glazed transom, stern gallery, crest, lanterns and quarter galleries.
func (s PirateGalleon) buildStern(g *Grid) {
	L, B, D := s.Length, s.Beam, s.Depth
	sh := s.sheer(-L / 2)
	face := func(y0, y1, depth float64, f func(p Vec3) bool) Func {
		return Func{Box{V(s.stern(y1)-3, y0, -B/2), V(s.stern(y0)+depth, y1, B/2)}, func(p Vec3) bool {
			return p.X <= s.stern(p.Y)+depth && p.X >= s.stern(p.Y)-0.4 && f(p)
		}}
	}
	// Three rows of glazed lights between gold moldings.
	for _, row := range [][2]float64{{D + 1.0, D + 3.6}, {D + 5.2, D + 7.8}, {D + 9.2, D + 11.8}} {
		y0, y1 := row[0]+sh, row[1]+sh
		g.Paint(Glass, face(y0, y1, 1.4, func(p Vec3) bool {
			w := s.halfWidth(s.stern(p.Y)+0.5, p.Y)
			return math.Abs(p.Z) < w-2 && math.Mod(p.Z+31, 3) < 2.0
		}))
	}
	for _, y := range []float64{D + 0.2, D + 4.2, D + 8.4, D + 12.4} {
		g.Paint(Gold, face(y+sh, y+0.7+sh, 1.4, func(p Vec3) bool { return true }))
	}
	// Stern gallery: a balcony with a gold balustrade before the great cabin.
	s.sternGallery(g, D+4.4+sh)
	// Taffrail crest: a gold arch above the poop rail with a red center panel.
	ry := D + 13.4 + sh
	g.Use(Gold).Fill(face(ry-0.5, ry+4, 1.2, func(p Vec3) bool {
		dz, dy := p.Z/(0.36*B), (p.Y-ry+0.5)/4.5
		return dz*dz+dy*dy <= 1
	}))
	g.Paint(Red, face(ry, ry+3, 1.2, func(p Vec3) bool {
		dz, dy := p.Z/(0.36*B-2), (p.Y-ry+0.5)/3.2
		return dz*dz+dy*dy <= 1
	}))
	// Three stern lanterns.
	for _, z := range []float64{-0.27 * B, 0, 0.27 * B} {
		x := s.stern(ry) + 0.6
		y := ry + 1.5
		if z == 0 {
			y = ry + 4
		}
		g.Use(Gold).Fill(Cylinder{Center: P(x, z), Radius: 0.5, Y0: y - 1.5, Y1: y})
		g.Use(Glass).Fill(Ellipsoid{V(x, y+1.3, z), V(1.0, 1.4, 1.0)})
		g.Use(Gold).Fill(Cone(P(x, z), 1.1, y+2.4, 1.2))
	}
	s.quarterGalleries(g)
}

// sternGallery is an open balcony across the transom at height gy.
func (s PirateGalleon) sternGallery(g *Grid, gy float64) {
	B := s.Beam
	g.Use(Wood).Fill(Func{Box{V(s.stern(gy)-2.6, gy-0.5, -B/2), V(s.stern(gy)+0.5, gy+0.3, B/2)}, func(p Vec3) bool {
		return math.Abs(p.Z) <= s.halfWidth(s.stern(p.Y)+0.5, p.Y)-1.2
	}})
	g.Use(Gold).Fill(Func{Box{V(s.stern(gy)-2.6, gy+0.3, -B/2), V(s.stern(gy)-1.6, gy+1.5, B/2)}, func(p Vec3) bool {
		z := math.Abs(p.Z)
		if z > s.halfWidth(s.stern(gy)+0.5, gy)-1.2 {
			return false
		}
		return p.Y > gy+0.9 || math.Mod(z+0.4, 1.8) < 0.8
	}})
}

// quarterGalleries are bulging, glazed, gold-domed turrets on both stern quarters.
func (s PirateGalleon) quarterGalleries(g *Grid) {
	L, B, D := s.Length, s.Beam, s.Depth
	// Quarter galleries on both sides: bulging, glazed, gold-domed.
	x0, half := -0.41*L, 0.065*L
	for _, side := range []float64{-1, 1} {
		gal := func(bulge, y0, y1 float64) Func {
			return Func{Box{V(x0-half, y0, -B), V(x0+half, y1, B)}, func(p Vec3) bool {
				t := (p.X - x0) / half
				if math.Abs(t) > 1 || p.Z*side < 0 {
					return false
				}
				w := s.halfWidth(p.X, p.Y)
				z := math.Abs(p.Z)
				return z >= w-1 && z <= w+bulge*math.Sqrt(1-t*t)
			}}
		}
		yb := D + 1 + s.sheer(x0)
		g.Use(Wood).FillVolume(gal(2.6, yb, yb+10))
		g.Paint(Gold, gal(2.6, yb+4.4, yb+5.2), gal(2.6, yb, yb+0.6), gal(2.6, yb+9.4, yb+10))
		g.Paint(Glass, Func{Box{V(x0-half, yb+1.4, -B), V(x0+half, yb+8.6, B)}, func(p Vec3) bool {
			if p.Z*side < 0 || math.Abs(p.Z) < s.halfWidth(p.X, p.Y)+0.6 {
				return false
			}
			h := p.Y - yb
			return (h < 3.8 || h > 5.8) && math.Mod(p.X-x0+30.6, 2.4) < 1.5
		}})
		zc := side * (s.halfWidth(x0, yb+10) + 0.8)
		g.Use(Gold).FillVolume(Ellipsoid{V(x0, yb+10, zc), V(half*0.8, 2.4, 2.4)})
		g.Fill(Cone(P(x0, zc), 0.8, yb+12, 1.6))
		g.Fill(Frustum{P(x0, zc), 0.4, 1.8, yb - 2.2, yb})
	}
}

// buildBow adds the cutwater, beakhead, figurehead, head rails, bowsprit and spritsail.
func (s PirateGalleon) buildBow(g *Grid) {
	D := s.Depth
	xb := s.bow(D)
	tip := xb + 11
	by := D - 1 + s.sheer(xb)
	// Cutwater: a curved knee from the waterline stem to the beak.
	g.Use(Wood).Fill(Func{Box{V(s.bow(2)-1, 2, -0.7), V(tip, by+1, 0.7)}, func(p Vec3) bool {
		t := math.Max(p.Y-2, 0) / (by - 2)
		return p.X >= s.bow(p.Y)-1.5 && p.X <= s.bow(p.Y)+(tip-s.bow(by))*t*t
	}})
	// Beakhead platform narrowing to the tip.
	g.Fill(Func{Box{V(xb-3, by-1, -5), V(tip, by+2, 5)}, func(p Vec3) bool {
		t := (p.X - (xb - 3)) / (tip - xb + 3)
		y := by + 0.12*(p.X-xb)
		return math.Abs(p.Y-y) <= 0.6 && math.Abs(p.Z) <= 4.2*(1-t)+0.6
	}})
	// Figurehead and head rails.
	g.Use(Gold).Fill(Ellipsoid{V(tip-0.5, by-1.2, 0), V(2.2, 1.8, 1.2)})
	g.Fill(Sphere(V(tip+1.2, by-0.4, 0), 1.0))
	for _, side := range []float64{-1, 1} {
		a := V(xb-2, by+3.5, side*3.4)
		g.Fill(Capsule{a, V(tip-1, by+1, side*0.6), 0.5})
		g.Fill(Capsule{V(xb-2, by+1, side*3.8), V(tip-1, by+1, side*0.6), 0.5})
	}
	// Bowsprit rising forward at about 28°.
	start, dir := s.bowspritStart(), s.bowspritDir()
	n := 0.36 * s.Length
	g.Use(Wood).Fill(Capsule{start, start.Add(dir.Scale(n * 0.55)), 0.9})
	g.Fill(Capsule{start.Add(dir.Scale(n * 0.5)), start.Add(dir.Scale(n)), 0.6})
	// Spritsail hanging below its yard.
	yard := start.Add(dir.Scale(n * 0.62))
	pivot := P(yard.X, 0)
	g.Fill(RotateY(Capsule{V(yard.X, yard.Y-0.8, -12), V(yard.X, yard.Y-0.8, 12), 0.5}, pivot, pirateGalleonBrace))
	g.Use(White).Fill(RotateY(s.squareSail(yard.X+0.3, yard.Y-1.3, yard.Y-12, 11, 12, 1.8), pivot, pirateGalleonBrace))
	// Sprit topmast standing at the bowsprit's end with its small square sail.
	end := start.Add(dir.Scale(n))
	g.Use(Wood).Fill(Frustum{P(end.X-0.5, 0), 0.6, 0.5, end.Y - 1, end.Y + 10})
	sy := end.Y + 9
	tp := P(end.X-0.5, 0)
	g.Fill(RotateY(Capsule{V(end.X, sy, -6), V(end.X, sy, 6), 0.5}, tp, pirateGalleonBrace))
	g.Use(White).Fill(RotateY(s.squareSail(end.X+0.2, sy-0.6, end.Y+1.2, 5.2, 6.5, 1.2), tp, pirateGalleonBrace))
}

func (s PirateGalleon) bowspritStart() Vec3 {
	x := s.bow(s.Depth+5) - 6
	return V(x, s.Depth+5.6+s.sheer(x), 0)
}

func (s PirateGalleon) bowspritDir() Vec3 {
	a := 28 * math.Pi / 180
	return V(math.Cos(a), math.Sin(a), 0)
}

// squareSail is a sail hanging from a yard at x between heights top and bottom, with
// half-widths wTop and wBottom, billowing forward (+X) by up to bulge.
func (s PirateGalleon) squareSail(x, top, bottom, wTop, wBottom, bulge float64) Func {
	w := math.Max(wTop, wBottom)
	return Func{Box{V(x-1, bottom-1, -w), V(x+bulge+1, top, w)}, func(p Vec3) bool {
		v := (top - p.Y) / (top - bottom) // 0 at the yard, 1 at the foot.
		half := wTop + (wBottom-wTop)*v
		u := p.Z / half
		if math.Abs(u) > 1 {
			return false
		}
		foot := bottom + 1.4*(1-u*u) // The foot arches up in the middle.
		if p.Y < foot || v < 0 {
			return false
		}
		b := bulge * (1 - 0.7*u*u) * math.Sin(math.Pi*0.5*math.Min(v*1.6, 1))
		return math.Abs(p.X-x-b) <= 0.6
	}}
}

// buildDeckGear adds a hatch grating, a longboat, a capstan, barrels and anchors.
func (s PirateGalleon) buildDeckGear(g *Grid) {
	L, D := s.Length, s.Depth
	// Anchors hanging from catheads on both bows.
	ax := 0.36 * L
	ay := s.rail(ax) - 0.4
	for _, side := range []float64{-1, 1} {
		w := s.halfWidth(ax, ay)
		z := side * (w + 2.2)
		g.Use(Wood).Fill(Block(V(ax-0.6, ay-0.6, side*(w-1)), V(ax+0.6, ay+0.6, side*(w+3))))
		g.Use(Charcoal).Fill(
			Capsule{V(ax, ay-1, z), V(ax, ay-8, z), 0.55},
			Capsule{V(ax-2.6, ay-1.8, z), V(ax+2.6, ay-1.8, z), 0.5},
			Capsule{V(ax, ay-8, z), V(ax-2.6, ay-6.2, z), 0.5},
			Capsule{V(ax, ay-8, z), V(ax+2.6, ay-6.2, z), 0.5},
		)
	}
	// Hatch grating in the waist.
	hx := -0.07 * L
	hy := s.deck(hx) + 0.5
	g.Use(Charcoal).Fill(BlockAt(P(hx, 0), 6, 6, hy, hy+0.8))
	g.Paint(Wood, Func{Box{V(hx-2.4, hy, -2.4), V(hx+2.4, hy+0.8, 2.4)}, func(p Vec3) bool {
		return math.Mod(p.X+100, 1.6) < 0.8 != (math.Mod(p.Z+100, 1.6) < 0.8)
	}})
	// Longboat resting on the waist deck.
	bx := 0.15 * L
	by := s.deck(bx) + 2.6
	g.Use(Wood).FillVolume(Subtract(
		Ellipsoid{V(bx, by, 0), V(7, 2.6, 2.8)},
		Ellipsoid{V(bx, by, 0), V(6, 1.7, 1.9)},
		Block(V(bx-8, by, -4), V(bx+8, by+3, 4)),
	))
	g.Paint(Red, Func{Box{V(bx-8, by-0.7, -4), V(bx+8, by, 4)}, func(p Vec3) bool { return true }})
	// Capstan on the quarterdeck.
	cx := -0.2 * L
	cy := s.deck(cx) + 0.5
	g.Use(Wood).Fill(Cylinder{Center: P(cx, 0), Radius: 1.4, Y0: cy, Y1: cy + 2.2})
	g.Use(Charcoal).Fill(Cylinder{Center: P(cx, 0), Radius: 2.0, Y0: cy + 1.6, Y1: cy + 2.4})
	// Barrels along the waist bulwarks.
	for i, x := range []float64{-0.1 * L, -0.04 * L, 0.06 * L, 0.21 * L} {
		y := s.deck(x) + 0.5
		z := 0.3 * s.halfWidth(x, D+1)
		if i%2 == 1 {
			z = -z
		}
		g.Use(Wood).Fill(Cylinder{Center: P(x, z), Radius: 1.1, Y0: y, Y1: y + 2.2})
		g.Paint(Charcoal, Cylinder{Center: P(x, z), Radius: 1.2, Y0: y + 1.6, Y1: y + 1.8})
	}
}

// pirateGalleonBrace turns the yards off square, as on a broad reach.
const pirateGalleonBrace = 0.38

// pirateGalleonMast describes one mast's stations.
type pirateGalleonMast struct {
	x, deck, top1, top2, tip float64
	rTop                     float64
	yards                    [3]float64 // Course, topsail and topgallant yard half-lengths.
}

// buildRig adds masts, tops, yards, sails, rigging, the flag and pennants.
func (s PirateGalleon) buildRig(g *Grid) {
	L := s.Length
	mast := func(x, height, f1, f2, rTop, yard float64) pirateGalleonMast {
		d := s.deck(x)
		return pirateGalleonMast{x, d, d + f1*height, d + f2*height, d + height, rTop,
			[3]float64{0.21 * L * yard, 0.155 * L * yard, 0.095 * L * yard}}
	}
	fore := mast(0.28*L, 0.62*L, 0.42, 0.75, 3.4, 0.86)
	main := mast(0.02*L, 0.74*L, 0.42, 0.75, 4.0, 1)
	mizzen := mast(-0.22*L, 0.42*L, 0.56, 0.8, 2.6, 0.55)
	for _, m := range []pirateGalleonMast{fore, main, mizzen} {
		s.buildMast(g, m)
	}
	for _, m := range []pirateGalleonMast{fore, main} {
		yc, yt, yg := m.top1-2.2, m.top2-2, m.tip-3.5
		braced := func(solid Solid) Solid { return RotateY(solid, P(m.x, 0), pirateGalleonBrace) }
		g.Use(Wood).Fill(
			braced(Capsule{V(m.x+1.4, yc, -m.yards[0]), V(m.x+1.4, yc, m.yards[0]), 0.6}),
			braced(Capsule{V(m.x+1.3, yt, -m.yards[1]), V(m.x+1.3, yt, m.yards[1]), 0.55}),
			braced(Capsule{V(m.x+1.2, yg, -m.yards[2]), V(m.x+1.2, yg, m.yards[2]), 0.5}),
		)
		courseFoot := s.rail(m.x) + 4.5
		g.Use(White).Fill(
			braced(s.squareSail(m.x+1.4, yc-0.6, courseFoot, m.yards[0]-1, m.yards[0]-0.5, 0.13*(yc-courseFoot))),
			braced(s.squareSail(m.x+1.3, yt-0.6, yc+0.8, m.yards[1]-1, m.yards[0]-1.5, 0.12*(yt-yc))),
			braced(s.squareSail(m.x+1.2, yg-0.6, yt+0.8, m.yards[2]-0.8, m.yards[1]-1.5, 0.12*(yg-yt))),
		)
	}
	// Mizzen: lateen sail on a slanted yard, and a small square topsail.
	{
		m := mizzen
		f := V(m.x+14, m.top1-12, 0)
		u := V(m.x-15, m.top1+4, 0)
		c := V(m.x-15, s.rail(m.x-15)+3, 0)
		dir := u.Sub(f).Scale(1 / u.Sub(f).Length())
		g.Use(Wood).Fill(Capsule{f.Sub(dir.Scale(1.5)), u.Add(dir.Scale(1.5)), 0.55})
		g.Use(White).Fill(s.lateen(f.Add(V(0, -0.9, 0)), u.Add(V(0, -0.9, 0)), c, 2.6))
		yt := m.tip - 3
		pivot := P(m.x, 0)
		g.Use(Wood).Fill(RotateY(Capsule{V(m.x+1.2, yt, -m.yards[1]), V(m.x+1.2, yt, m.yards[1]), 0.5}, pivot, pirateGalleonBrace))
		g.Use(White).Fill(RotateY(s.squareSail(m.x+1.2, yt-0.6, m.top1+2, m.yards[1]-0.8, m.yards[0]-0.8, 1.8), pivot, pirateGalleonBrace))
	}
	s.buildRigging(g, fore, main, mizzen)
	// Black flag with skull and crossbones streaming forward from the main truck.
	g.Use(Wood).Fill(Cylinder{Center: P(main.x, 0), Radius: 0.5, Y0: main.tip, Y1: main.tip + 9})
	fx0, fy0, fw, fh := main.x+0.4, main.tip+1.5, 12.0, 7.0
	wave := func(x float64) float64 { t := (x - fx0) / fw; return 0.9 * math.Sin(2*math.Pi*t*1.3) * t }
	g.Use(Charcoal).Fill(Func{Box{V(fx0, fy0, -2), V(fx0+fw, fy0+fh, 2)}, func(p Vec3) bool {
		return math.Abs(p.Z-wave(p.X)) <= 0.5
	}})
	flagArea := func(f func(x, y float64) bool) Func {
		return Func{Box{V(fx0, fy0, -2), V(fx0+fw, fy0+fh, 2)}, func(p Vec3) bool { return f(p.X-fx0, p.Y-fy0) }}
	}
	g.Paint(White, flagArea(func(x, y float64) bool {
		skull := math.Hypot((x-6)/1.6, (y-4.6)/1.5) <= 1
		jaw := math.Abs(x-6) <= 0.9 && y >= 2.9 && y <= 3.6
		bone := func(x0, y0, x1, y1 float64) bool {
			return Capsule{V(x0, y0, 0), V(x1, y1, 0), 0.45}.Contains(V(x, y, 0))
		}
		return skull || jaw || bone(3.2, 1.0, 8.8, 3.0) || bone(3.2, 3.0, 8.8, 1.0)
	}))
	g.Paint(Charcoal, flagArea(func(x, y float64) bool {
		return math.Hypot(x-5.35, y-4.7) <= 0.45 || math.Hypot(x-6.65, y-4.7) <= 0.45
	}))
	// Red pennants on the fore and mizzen trucks.
	for _, m := range []pirateGalleonMast{fore, mizzen} {
		g.Use(Wood).Fill(Cylinder{Center: P(m.x, 0), Radius: 0.5, Y0: m.tip, Y1: m.tip + 3})
		x0, y0 := m.x+0.4, m.tip+2.2
		g.Use(Red).Fill(Func{Box{V(x0, y0-1.5, -2), V(x0+10, y0+1.5, 2)}, func(p Vec3) bool {
			t := (p.X - x0) / 10
			c := y0 - 0.4*t
			return math.Abs(p.Y-c) <= 1.2*(1-t)+0.45 && math.Abs(p.Z-0.7*math.Sin(t*5)*t) <= 0.5
		}})
	}
}

// lateen is a triangular sail F-U-C in the XY plane, bulging toward +Z.
func (s PirateGalleon) lateen(f, u, c Vec3, bulge float64) Func {
	box := BoxOf(f, u).Union(BoxOf(c, c))
	box.Min.Z, box.Max.Z = -1, bulge+1
	area := (u.X-f.X)*(c.Y-f.Y) - (c.X-f.X)*(u.Y-f.Y)
	return Func{box, func(p Vec3) bool {
		l1 := ((u.X-p.X)*(c.Y-p.Y) - (c.X-p.X)*(u.Y-p.Y)) / area
		l2 := ((c.X-p.X)*(f.Y-p.Y) - (f.X-p.X)*(c.Y-p.Y)) / area
		l3 := 1 - l1 - l2
		if l1 < 0 || l2 < 0 || l3 < 0 {
			return false
		}
		return math.Abs(p.Z-bulge*27*l1*l2*l3) <= 0.55
	}}
}

// buildMast adds a tapering mast in three pieces, a fighting top and a crow's nest.
func (s PirateGalleon) buildMast(g *Grid, m pirateGalleonMast) {
	c := P(m.x, 0)
	g.Use(Wood).Fill(
		Frustum{c, 1.4, 1.0, m.deck - 1, m.top1 + 2},
		Frustum{c, 0.9, 0.7, m.top1 - 2, m.top2 + 2},
		Frustum{c, 0.6, 0.5, m.top2 - 1, m.tip},
	)
	// Fighting top: a round platform with a solid rim.
	g.Fill(Cylinder{Center: c, Radius: m.rTop, Y0: m.top1 - 0.5, Y1: m.top1 + 0.4})
	g.Fill(Cylinder{Center: c, Radius: m.rTop, Inner: m.rTop - 1, Y0: m.top1 + 0.4, Y1: m.top1 + 1.8})
	g.Paint(Red, Cylinder{Center: c, Radius: m.rTop, Inner: m.rTop - 1, Y0: m.top1 + 0.8, Y1: m.top1 + 1.4})
	// Crow's nest on the topmast.
	r := 0.55 * m.rTop
	g.Fill(Cylinder{Center: c, Radius: r, Y0: m.top2 - 0.4, Y1: m.top2 + 0.4})
	g.Fill(Cylinder{Center: c, Radius: r, Inner: r - 1, Y0: m.top2 + 0.4, Y1: m.top2 + 1.8})
}

// buildRigging adds shrouds from the tops to the hull sides and fore-and-aft stays.
func (s PirateGalleon) buildRigging(g *Grid, fore, main, mizzen pirateGalleonMast) {
	g.Use(Charcoal)
	line := func(a, b Vec3) { g.Fill(Capsule{a, b, 0.5}) }
	for _, m := range []pirateGalleonMast{fore, main, mizzen} {
		for _, side := range []float64{-1, 1} {
			for i, dx := range []float64{-6, -3.2, -0.4} {
				x := m.x + dx
				y := s.rail(x) - 0.4
				w := s.halfWidth(x, y) + 0.6
				line(V(m.x-0.8+float64(i)*0.6, m.top1, side*(m.rTop-0.5)), V(x, y, side*w))
			}
			// Channel: a ledge on the hull side where the shrouds fasten.
			y := s.rail(m.x) - 1.2
			w := s.halfWidth(m.x, y)
			g.With(Wood, func() {
				g.Fill(Block(V(m.x-7, y-0.4, side*(w-0.5)), V(m.x+1, y+0.4, side*(w+1.2))))
			})
			r := 0.55*m.rTop - 0.4
			for _, dx := range []float64{-1.2, 0.2} {
				line(V(m.x+dx*0.3, m.top2, side*r), V(m.x+dx, m.top1+1.8, side*(m.rTop-0.5)))
			}
		}
	}
	start, dir := s.bowspritStart(), s.bowspritDir()
	n := 0.36 * s.Length
	line(V(fore.x, fore.top1, 0), start.Add(dir.Scale(n*0.42)))
	line(V(fore.x, fore.top2, 0), start.Add(dir.Scale(n*0.96)))
	line(V(main.x, main.top1, 0), V(fore.x-1, fore.deck+2, 0))
	line(V(main.x, main.top2, 0), V(fore.x, fore.top1+1, 0))
	line(V(main.x, main.tip-1, 0), V(fore.x, fore.top2+1, 0))
	line(V(mizzen.x, mizzen.top1, 0), V(main.x-1, main.deck+6, 0))
	line(V(mizzen.x, mizzen.tip-1, 0), V(main.x, main.top1+1, 0))
}
