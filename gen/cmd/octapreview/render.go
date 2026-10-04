package main

import (
	"image"
	"math"
	"sort"

	"godocta/gen/octa"
)

// face is one exposed polygon of a cell, in world units.
type face struct {
	verts  []octa.Vec3
	normal octa.Vec3
	albedo [3]float64 // Linear RGB.
	glow   float64    // Added fraction of the albedo, for cells landing.
}

type scene struct {
	faces  []face
	bounds octa.Box
	span   float64
	ground float64
	sun    octa.Vec3
	shadow *shadowMap
	look   look
	edges  float64 // Darkening of cell outlines, 0 to 1.
}

// look holds the sky, ground and light colors, all linear RGB.
type look struct {
	zenith, horizon, ground [3]float64
	sun, ambient            [3]float64
}

var neutralLook = look{
	zenith:  [3]float64{srgbToLinear(0x8f), srgbToLinear(0xa9), srgbToLinear(0xc8)},
	horizon: [3]float64{srgbToLinear(0xe4), srgbToLinear(0xe2), srgbToLinear(0xdc)},
	ground:  [3]float64{srgbToLinear(0x9a), srgbToLinear(0x9c), srgbToLinear(0x8e)},
	sun:     [3]float64{1.0, 0.96, 0.9},
	ambient: [3]float64{0.95, 1.0, 1.08},
}

// unitFace is a truncated octahedron face around the origin for span 1, with the lattice
// offset of the neighbor that shares it.
type unitFace struct {
	verts    []octa.Vec3
	normal   octa.Vec3
	neighbor octa.Cell
}

// unitFaces builds the 6 squares and 8 hexagons. Vertices are the permutations of
// (0, ±1/4, ±1/2); squares face lattice offsets (±2,0,0)…, hexagons (±1,±1,±1).
func unitFaces() []unitFace {
	var verts []octa.Vec3
	for _, p := range [][3]int{{0, 1, 2}, {0, 2, 1}, {1, 0, 2}, {1, 2, 0}, {2, 0, 1}, {2, 1, 0}} {
		for _, sa := range []float64{-1, 1} {
			for _, sb := range []float64{-1, 1} {
				var c [3]float64
				c[p[0]] = 0
				c[p[1]] = sa * 0.25
				c[p[2]] = sb * 0.5
				verts = append(verts, octa.V(c[0], c[1], c[2]))
			}
		}
	}
	var offsets []octa.Cell
	for axis := 0; axis < 3; axis++ {
		for _, s := range []int32{-2, 2} {
			c := [3]int32{}
			c[axis] = s
			offsets = append(offsets, octa.Cell{X: c[0], Y: c[1], Z: c[2]})
		}
	}
	for _, x := range []int32{-1, 1} {
		for _, y := range []int32{-1, 1} {
			for _, z := range []int32{-1, 1} {
				offsets = append(offsets, octa.Cell{X: x, Y: y, Z: z})
			}
		}
	}
	faces := make([]unitFace, 0, 14)
	for _, o := range offsets {
		n := norm(octa.V(float64(o.X), float64(o.Y), float64(o.Z)))
		best := math.Inf(-1)
		for _, v := range verts {
			best = math.Max(best, v.Dot(n))
		}
		var poly []octa.Vec3
		var center octa.Vec3
		for _, v := range verts {
			if v.Dot(n) > best-1e-9 {
				poly = append(poly, v)
				center = center.Add(v)
			}
		}
		center = center.Scale(1 / float64(len(poly)))
		u := norm(poly[0].Sub(center))
		w := cross(n, u)
		sort.Slice(poly, func(i, j int) bool {
			a, b := poly[i].Sub(center), poly[j].Sub(center)
			return math.Atan2(a.Dot(w), a.Dot(u)) < math.Atan2(b.Dot(w), b.Dot(u))
		})
		faces = append(faces, unitFace{poly, n, o})
	}
	return faces
}

func srgbToLinear(c uint8) float64 {
	v := float64(c) / 255
	if v <= 0.04045 {
		return v / 12.92
	}
	return math.Pow((v+0.055)/1.055, 2.4)
}

func linearToSRGB(v float64) uint8 {
	v = math.Max(0, math.Min(1, v))
	if v <= 0.0031308 {
		v *= 12.92
	} else {
		v = 1.055*math.Pow(v, 1/2.4) - 0.055
	}
	return uint8(v*255 + 0.5)
}

// albedos returns each cell's linear color.
func albedos(s *octa.Structure) [][3]float64 {
	var palette [][3]float64
	for _, p := range s.Palette {
		palette = append(palette, [3]float64{srgbToLinear(p.R), srgbToLinear(p.G), srgbToLinear(p.B)})
	}
	base := [3]float64{srgbToLinear(0xb9), srgbToLinear(0xb3), srgbToLinear(0xa4)}
	out := make([][3]float64, len(s.Cells))
	for i := range s.Cells {
		out[i] = base
		if len(s.Colors) > 0 && int(s.Colors[i]) < len(palette) {
			out[i] = palette[s.Colors[i]]
		}
	}
	return out
}

// exposedFaces returns the faces of the kept cells (all cells when keep is nil) that no
// other kept cell covers.
func exposedFaces(s *octa.Structure, keep func(int) bool) []face {
	span := float64(s.Span)
	occupied := make(map[octa.Cell]struct{}, len(s.Cells))
	for i, c := range s.Cells {
		if keep == nil || keep(i) {
			occupied[c] = struct{}{}
		}
	}
	colors := albedos(s)
	units := unitFaces()
	var faces []face
	for i, c := range s.Cells {
		if keep != nil && !keep(i) {
			continue
		}
		center := c.Center().Scale(span)
		for _, u := range units {
			if _, hidden := occupied[c.Add(u.neighbor)]; hidden {
				continue
			}
			verts := make([]octa.Vec3, len(u.verts))
			for k, v := range u.verts {
				verts[k] = center.Add(v.Scale(span))
			}
			faces = append(faces, face{verts: verts, normal: u.normal, albedo: colors[i]})
		}
	}
	return faces
}

func newScene(s *octa.Structure) *scene {
	sc := &scene{span: float64(s.Span), bounds: s.Bounds(), look: neutralLook}
	sc.faces = exposedFaces(s, nil)
	sc.ground = sc.bounds.Min.Y
	// A fixed high sun from the hero camera's right, so every view shares one lighting.
	el, az := 0.85, 1.25
	sc.sun = octa.V(math.Sin(az)*math.Cos(el), math.Sin(el), math.Cos(az)*math.Cos(el))
	sc.shadow = newShadowMap(sc, 3072)
	return sc
}

// shadowMap is an orthographic depth map along the sun direction.
type shadowMap struct {
	right, up, dir octa.Vec3
	minX, minY     float64
	scale          float64
	size           int
	depth          []float32 // Largest height toward the sun; -Inf where empty.
	bias           float64
}

func newShadowMap(sc *scene, size int) *shadowMap {
	m := &shadowMap{dir: sc.sun, size: size, bias: sc.span * 0.3}
	m.right = norm(cross(octa.V(0, 1, 0), m.dir))
	m.up = cross(m.dir, m.right)
	minX, minY, maxX, maxY := math.Inf(1), math.Inf(1), math.Inf(-1), math.Inf(-1)
	for _, f := range sc.faces {
		for _, v := range f.verts {
			x, y := v.Dot(m.right), v.Dot(m.up)
			minX, maxX = math.Min(minX, x), math.Max(maxX, x)
			minY, maxY = math.Min(minY, y), math.Max(maxY, y)
		}
	}
	m.depth = make([]float32, size*size)
	for i := range m.depth {
		m.depth[i] = float32(math.Inf(-1))
	}
	if len(sc.faces) == 0 {
		return m
	}
	m.minX, m.minY = minX, minY
	m.scale = float64(size-1) / math.Max(maxX-minX, maxY-minY)
	pts := make([][3]float64, 0, 6)
	for _, f := range sc.faces {
		if f.normal.Dot(m.dir) <= 0 {
			continue // Back faces never occlude closer than the front faces of a cell.
		}
		pts = pts[:0]
		for _, v := range f.verts {
			pts = append(pts, [3]float64{(v.Dot(m.right) - minX) * m.scale, (v.Dot(m.up) - minY) * m.scale, v.Dot(m.dir)})
		}
		for k := 1; k+1 < len(pts); k++ {
			raster(pts[0], pts[k], pts[k+1], size, size, func(i int, d float64) {
				if float32(d) > m.depth[i] {
					m.depth[i] = float32(d)
				}
			})
		}
	}
	return m
}

// lit returns the unshadowed fraction at p with a 3×3 percentage-closer filter.
func (m *shadowMap) lit(p octa.Vec3) float64 {
	if m.scale == 0 {
		return 1
	}
	x := (p.Dot(m.right) - m.minX) * m.scale
	y := (p.Dot(m.up) - m.minY) * m.scale
	d := float32(p.Dot(m.dir) + m.bias)
	litCount := 0
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			ix, iy := int(x+0.5)+dx, int(y+0.5)+dy
			if ix < 0 || iy < 0 || ix >= m.size || iy >= m.size || m.depth[iy*m.size+ix] <= d {
				litCount++
			}
		}
	}
	return float64(litCount) / 9
}

// raster fills a triangle given in pixel coordinates (x, y, depth), sampling pixel
// centers, and calls plot with the pixel index and the affinely interpolated depth.
func raster(a, b, c [3]float64, w, h int, plot func(int, float64)) {
	area := (b[0]-a[0])*(c[1]-a[1]) - (b[1]-a[1])*(c[0]-a[0])
	if area == 0 {
		return
	}
	x0 := int(math.Max(0, math.Floor(math.Min(a[0], math.Min(b[0], c[0])))))
	x1 := int(math.Min(float64(w-1), math.Ceil(math.Max(a[0], math.Max(b[0], c[0])))))
	y0 := int(math.Max(0, math.Floor(math.Min(a[1], math.Min(b[1], c[1])))))
	y1 := int(math.Min(float64(h-1), math.Ceil(math.Max(a[1], math.Max(b[1], c[1])))))
	inv := 1 / area
	for y := y0; y <= y1; y++ {
		py := float64(y) + 0.5
		for x := x0; x <= x1; x++ {
			px := float64(x) + 0.5
			w0 := ((b[0]-px)*(c[1]-py) - (b[1]-py)*(c[0]-px)) * inv
			w1 := ((c[0]-px)*(a[1]-py) - (c[1]-py)*(a[0]-px)) * inv
			w2 := 1 - w0 - w1
			if w0 < 0 || w1 < 0 || w2 < 0 {
				continue
			}
			plot(y*w+x, w0*a[2]+w1*b[2]+w2*c[2])
		}
	}
}

// shade lights an albedo with a hemisphere ambient term plus the sun.
func (sc *scene) shade(albedo [3]float64, n, p octa.Vec3) [3]float64 {
	ndl := math.Max(0, n.Dot(sc.sun))
	sun := 0.0
	if ndl > 0 {
		sun = ndl * sc.shadow.lit(p)
	}
	amb := 0.34 + 0.16*n.Y
	a, l := sc.look.ambient, sc.look.sun
	return [3]float64{
		albedo[0] * (amb*a[0] + sun*l[0]),
		albedo[1] * (amb*a[1] + sun*l[1]),
		albedo[2] * (amb*a[2] + sun*l[2]),
	}
}

// view is a camera at dist from focus along the unit direction toCam, looking back at focus.
type view struct {
	focus, toCam octa.Vec3
	dist         float64
}

// orbit is the unit direction from the focus to a camera at yaw and pitch, in radians.
func orbit(yaw, pitch float64) octa.Vec3 {
	return octa.V(math.Sin(yaw)*math.Cos(pitch), math.Sin(pitch), math.Cos(yaw)*math.Cos(pitch))
}

// basis returns the camera's forward, right and up axes.
func (v view) basis() (fwd, right, up octa.Vec3) {
	fwd = v.toCam.Scale(-1)
	right = norm(cross(fwd, octa.V(0, 1, 0)))
	return fwd, right, cross(right, fwd)
}

var tanHalfFOV = math.Tan(45.0 / 2 * math.Pi / 180)

// fit is the camera distance that keeps every exposed face within 92% of the frustum.
func (sc *scene) fit(focus, toCam octa.Vec3, aspect float64) float64 {
	_, right, up := view{focus, toCam, 0}.basis()
	tanX := tanHalfFOV * aspect
	dist := 1.0
	for _, f := range sc.faces {
		c := f.verts[0].Sub(focus)
		z := c.Dot(toCam)
		dist = math.Max(dist, z+math.Abs(c.Dot(right))/(tanX*0.92))
		dist = math.Max(dist, z+math.Abs(c.Dot(up))/(tanHalfFOV*0.92))
	}
	return dist
}

func (sc *scene) render(yaw, pitch float64, w, h, ss int) *image.RGBA {
	W, H := w*ss, h*ss
	b := sc.bounds
	focus := b.Min.Add(b.Max).Scale(0.5)
	toCam := orbit(yaw, pitch)
	v := view{focus, toCam, sc.fit(focus, toCam, float64(W)/float64(H))}
	full := make([]float32, W*H*3)
	sc.renderLinear([][]face{sc.faces}, v, W, H, 0, 0, 1, full, nil)
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	n := float64(ss * ss)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			var a [3]float64
			for sy := 0; sy < ss; sy++ {
				for sx := 0; sx < ss; sx++ {
					i := ((y*ss+sy)*W + x*ss + sx) * 3
					a[0] += float64(full[i])
					a[1] += float64(full[i+1])
					a[2] += float64(full[i+2])
				}
			}
			o := (y*w + x) * 4
			img.Pix[o] = linearToSRGB(a[0] / n)
			img.Pix[o+1] = linearToSRGB(a[1] / n)
			img.Pix[o+2] = linearToSRGB(a[2] / n)
			img.Pix[o+3] = 255
		}
	}
	return img
}

// depthBuffers are per-render scratch buffers, reusable across renders of one size.
type depthBuffers struct {
	invz []float32 // 1/view depth; 0 is empty.
	ids  []int32   // Face index across the concatenated layers.
	all  []*face   // The concatenated layers.
}

// renderLinear draws the layers of faces from the view at W×H, offsetting pixel samples by
// (jx, jy) pixels, and adds weight × linear RGB into acc (3 floats per pixel). buf may be nil.
func (sc *scene) renderLinear(layers [][]face, v view, W, H int, jx, jy, weight float64, acc []float32, buf *depthBuffers) {
	fwd, right, up := v.basis()
	cam := v.focus.Add(v.toCam.Scale(v.dist))
	focal := float64(H) / 2 / tanHalfFOV
	if buf == nil {
		buf = &depthBuffers{}
	}
	if len(buf.invz) != W*H {
		buf.invz, buf.ids = make([]float32, W*H), make([]int32, W*H)
	} else {
		clear(buf.invz)
	}
	invz, ids := buf.invz, buf.ids
	cx, cy := float64(W)/2-jx, float64(H)/2-jy
	pts := make([][3]float64, 0, 6)
	all := buf.all[:0]
	for _, layer := range layers {
		for fi := range layer {
			all = append(all, &layer[fi])
		}
	}
	buf.all = all
	for fi, f := range all {
		if f.normal.Dot(cam.Sub(f.verts[0])) <= 0 {
			continue
		}
		pts = pts[:0]
		behind := false
		for _, p := range f.verts {
			r := p.Sub(cam)
			z := r.Dot(fwd)
			if z < 1e-3 {
				behind = true
				break
			}
			pts = append(pts, [3]float64{cx + r.Dot(right)*focal/z, cy - r.Dot(up)*focal/z, 1 / z})
		}
		if behind {
			continue
		}
		id := int32(fi)
		for k := 1; k+1 < len(pts); k++ {
			raster(pts[0], pts[k], pts[k+1], W, H, func(i int, d float64) {
				if float32(d) > invz[i] {
					invz[i] = float32(d)
					ids[i] = id
				}
			})
		}
	}
	for y := 0; y < H; y++ {
		for x := 0; x < W; x++ {
			i := y*W + x
			dir := fwd.Add(right.Scale((float64(x) + 0.5 - cx) / focal)).Add(up.Scale((cy - float64(y) - 0.5) / focal))
			var col [3]float64
			if invz[i] > 0 {
				f := all[ids[i]]
				p := cam.Add(dir.Scale(1 / float64(invz[i])))
				col = sc.shade(f.albedo, f.normal, p)
				for c := range col {
					col[c] = col[c]*sc.outline(f.verts, p) + f.albedo[c]*f.glow
				}
			} else {
				col = sc.background(cam, dir, v.dist)
			}
			acc[i*3] += float32(col[0] * weight)
			acc[i*3+1] += float32(col[1] * weight)
			acc[i*3+2] += float32(col[2] * weight)
		}
	}
}

// outline is the shading factor that darkens p near the edges of its face polygon.
func (sc *scene) outline(verts []octa.Vec3, p octa.Vec3) float64 {
	if sc.edges == 0 {
		return 1
	}
	d := math.Inf(1)
	for k, a := range verts {
		e := verts[(k+1)%len(verts)].Sub(a)
		d = math.Min(d, cross(p.Sub(a), e).Length()/e.Length())
	}
	t := math.Min(1, d/(sc.span*0.07))
	return 1 - sc.edges*(1-t*t*(3-2*t))
}

// background is the sky gradient, or the shadowed ground plane fading into the haze.
func (sc *scene) background(cam, dir octa.Vec3, dist float64) [3]float64 {
	d := norm(dir)
	if d.Y < 0 && cam.Y > sc.ground {
		t := (sc.ground - cam.Y) / dir.Y
		p := cam.Add(dir.Scale(t))
		g := sc.shade(sc.look.ground, octa.V(0, 1, 0), p)
		fade := math.Exp(-math.Max(0, p.Sub(cam).Length()-dist) / (dist * 1.5))
		return lerp3(sc.look.horizon, g, fade)
	}
	return lerp3(sc.look.horizon, sc.look.zenith, math.Sqrt(math.Max(0, d.Y)))
}

func lerp3(a, b [3]float64, t float64) [3]float64 {
	return [3]float64{a[0] + (b[0]-a[0])*t, a[1] + (b[1]-a[1])*t, a[2] + (b[2]-a[2])*t}
}

func cross(a, b octa.Vec3) octa.Vec3 {
	return octa.V(a.Y*b.Z-a.Z*b.Y, a.Z*b.X-a.X*b.Z, a.X*b.Y-a.Y*b.X)
}

func norm(v octa.Vec3) octa.Vec3 { return v.Scale(1 / v.Length()) }
