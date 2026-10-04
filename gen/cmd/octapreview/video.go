package main

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"godocta/gen/octa"
)

// videoOptions configures a showcase: one orbiting shot per build. Each orbit starts
// fast, slows to a drift mid-shot and speeds up again, so consecutive shots meet at full
// speed and the build swaps under motion blur while the spin carries on.
type videoOptions struct {
	w, h, fps, samples int
	workers            int     // Frames rendered at once.
	shot, swap, left   float64 // Seconds: shot length, blurred swap around each cut, assembly left.
	turn               float64 // Orbit per shot, in radians.
	theme              string  // world_theme.gd path, for sky, ground and light colors.
}

// orbitSpeed is the orbit's relative speed at shot progress u: a narrow burst at each cut
// (u = 0 and 1) over a slow drift. orbitTable integrates it, normalized to 0…1.
func orbitSpeed(u float64) float64 {
	return 0.1 + 0.9*math.Pow((1+math.Cos(2*math.Pi*u))/2, 4)
}

const orbitSteps = 4096

var orbitTable = func() []float64 {
	t := make([]float64, orbitSteps+1)
	for i := 1; i <= orbitSteps; i++ {
		t[i] = t[i-1] + orbitSpeed((float64(i)-0.5)/orbitSteps)
	}
	for i := range t {
		t[i] /= t[orbitSteps]
	}
	return t
}()

// flyer is a cell still in flight when its shot starts.
type flyer struct {
	center     octa.Vec3
	albedo     [3]float64
	rank, seed float64
}

// clip is one build in the showcase: its settled faces plus the cells still in flight,
// animated like shaders/cell_common.gdshaderinc with the AssemblyTimeline constants.
type clip struct {
	sc                      *scene
	static                  []face
	flying                  []flyer
	units                   []unitFace
	start, release, settle  float64 // Assembly clock at the shot's start, release window, settle.
	flight                  float64
	focus                   octa.Vec3
	yaw0, yaw1, pitch, dist float64
	hold                    float64 // Seconds before the assembly clock runs.
}

func newClip(s *octa.Structure, o videoOptions, look look) *clip {
	n := len(s.Cells)
	duration := math.Min(math.Max(4*math.Cbrt(float64(n)/1729), 4), 24)
	settle := 1.2 * math.Pow(duration/4, 0.25)
	c := &clip{sc: newScene(s), units: unitFaces(), release: duration - settle, settle: settle}
	c.sc.look = look
	c.sc.edges = 0.32
	c.start = math.Max(0, duration-o.left)
	// The last cells land as the orbit slows to its mid-shot drift.
	c.hold = math.Max(0, o.shot/2-o.left)
	rank := func(i int) float64 { return float64(i) / math.Max(float64(n-1), 1) }
	landed := func(i int) bool { return c.delay(rank(i))+settle <= c.start }
	c.static = exposedFaces(s, landed)
	colors := albedos(s)
	span := float64(s.Span)
	for i, cell := range s.Cells {
		if !landed(i) {
			seed := math.Mod(float64(i)*0.61803398875, 1)
			c.flying = append(c.flying, flyer{cell.Center().Scale(span), colors[i], rank(i), seed})
		}
	}
	size := c.sc.bounds.Size()
	c.flight = math.Min(math.Max(math.Pow(math.Max(size.X, math.Max(size.Y, size.Z))/10, 0.25), 1), 2)
	// Aim at the center of mass, so heavy bases and tall spires both stay in frame.
	for i := range s.Cells {
		c.focus = c.focus.Add(s.Cells[i].Center().Scale(span))
	}
	c.focus = c.focus.Scale(1 / float64(max(n, 1)))
	c.pitch = 0.24
	c.yaw0, c.yaw1 = 0.68-o.turn/2, 0.68+o.turn/2
	aspect := float64(o.w) / float64(o.h)
	for k := 0; k <= 16; k++ {
		yaw := c.yaw0 + (c.yaw1-c.yaw0)*float64(k)/16
		c.dist = math.Max(c.dist, c.sc.fit(c.focus, orbit(yaw, c.pitch), aspect))
	}
	c.sc.faces = nil // The settled faces and shadow map replace them.
	return c
}

func (c *clip) delay(rank float64) float64 {
	return c.release * (1 - math.Sqrt(1-math.Min(math.Max(rank, 0), 1)))
}

// at returns the camera and the moving faces tau seconds into the shot.
func (c *clip) at(tau, length float64) (view, []face) {
	yaw := c.yaw0 + (c.yaw1-c.yaw0)*sweep(tau/length)
	v := view{c.focus, orbit(yaw, c.pitch), c.dist * 0.82}
	clock := c.start + math.Max(0, tau-c.hold)
	span := c.sc.span
	var faces []face
	for _, f := range c.flying {
		t := math.Min(math.Max((clock-c.delay(f.rank))/c.settle, 0), 1)
		if t <= 0 {
			continue
		}
		remaining := math.Pow(1-t, 3)
		scale := smoothstep(0, 0.28, t)
		angle := (f.seed - 0.5) * 2.8 * remaining
		cs, sn := math.Cos(angle), math.Sin(angle)
		turn := func(p octa.Vec3) octa.Vec3 { return octa.V(cs*p.X+sn*p.Z, p.Y, -sn*p.X+cs*p.Z) }
		a := f.seed * 2 * math.Pi
		offset := octa.V(math.Sin(a)*3.5, 5+f.seed*2, math.Cos(a)*3.5).Scale(remaining * c.flight * span)
		glow := math.Sin(t*math.Pi) * 0.2
		for _, u := range c.units {
			verts := make([]octa.Vec3, len(u.verts))
			for k, p := range u.verts {
				verts[k] = f.center.Add(offset).Add(turn(p).Scale(scale * span))
			}
			faces = append(faces, face{verts, turn(u.normal), f.albedo, glow})
		}
	}
	return v, faces
}

// frame accumulates the shot at tau into blur, averaged over a half-frame shutter with
// sub-pixel jitter, so the orbit and falling cells blur, and into sharp without the
// shutter, for the vignette that keeps the screen edges crisp. Faster orbits take more
// samples, so the blur stays smooth rather than ghosted.
func (c *clip) frame(tau, length float64, o videoOptions, blur, sharp []float32, buf *depthBuffers) {
	shutter := 0.5 / float64(o.fps)
	render := func(acc []float32, shutter float64, samples int) {
		clear(acc)
		for k := 0; k < samples; k++ {
			t := tau + ((float64(k)+0.5)/float64(samples)-0.5)*shutter
			v, moving := c.at(t, length)
			c.sc.renderLinear([][]face{c.static, moving}, v, o.w, o.h, halton(k+1, 2)-0.5, halton(k+1, 3)-0.5,
				1/float64(samples), acc, buf)
		}
	}
	u := math.Mod(tau/length+1, 1)
	speed := math.Abs(c.yaw1-c.yaw0) / length * orbitSpeed(u) / orbitMean
	needed := int(math.Ceil(speed * shutter / (0.15 * math.Pi / 180)))
	render(blur, shutter, max(o.samples, needed))
	if needed <= o.samples/2 {
		copy(sharp, blur) // Too slow to blur visibly.
		return
	}
	render(sharp, 0, o.samples)
}

// orbitMean is the average of orbitSpeed over a shot.
var orbitMean = func() float64 {
	sum := 0.0
	for i := 0; i < orbitSteps; i++ {
		sum += orbitSpeed((float64(i) + 0.5) / orbitSteps)
	}
	return sum / orbitSteps
}()

// vignette is 1 in a circle around the frame center, fading to 0 towards the edges.
func vignette(w, h int) []float32 {
	m := make([]float32, w*h)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r := math.Hypot(float64(x)+0.5-float64(w)/2, float64(y)+0.5-float64(h)/2) / (float64(h) / 2)
			m[y*w+x] = float32(1 - smoothstep(0.4, 1.05, r))
		}
	}
	return m
}

// hblur box-blurs img along x over blur pixels into out, clamping at the frame edges.
func hblur(img, out []float32, w, h int, blur float64) {
	if blur <= 1 {
		copy(out, img)
		return
	}
	prefix := make([]float64, w+1)
	for y := 0; y < h; y++ {
		for ch := 0; ch < 3; ch++ {
			px := func(x int) float64 { return float64(img[(y*w+x)*3+ch]) }
			for x := 0; x < w; x++ {
				prefix[x+1] = prefix[x] + px(x)
			}
			integral := func(t float64) float64 {
				t = math.Min(math.Max(t, 0), float64(w))
				i := int(t)
				if i >= w {
					return prefix[w]
				}
				return prefix[i] + (t-float64(i))*px(i)
			}
			for x := 0; x < w; x++ {
				c := float64(x) + 0.5
				lo, hi := math.Max(c-blur/2, 0), math.Min(c+blur/2, float64(w))
				out[(y*w+x)*3+ch] = float32((integral(hi) - integral(lo)) / (hi - lo))
			}
		}
	}
}

// renderVideo renders the structures as one showcase and encodes it with ffmpeg. Each
// shot is encoded to its own part under <path>.parts/ first, so an interrupted render
// resumes from the last finished part; the parts are then joined without re-encoding.
func renderVideo(path string, structures []*octa.Structure, names []string, o videoOptions) error {
	partsDir := path + ".parts"
	if err := os.MkdirAll(partsDir, 0o755); err != nil {
		return err
	}
	clips := newClipCache(structures, names, o)
	shots := len(structures)
	shotFrames := func(i int) int { return int(math.Round(float64(i) * o.shot * float64(o.fps))) }
	frames := shotFrames(shots)
	start := time.Now()
	var list strings.Builder
	for p := range shots {
		part := filepath.Join(partsDir, fmt.Sprintf("part_%03d.mp4", p))
		abs, _ := filepath.Abs(part)
		fmt.Fprintf(&list, "file '%s'\n", abs)
		if _, err := os.Stat(part); err == nil {
			fmt.Printf("%s: done\n", names[p])
			continue
		}
		if p+1 < shots {
			clips.load(p + 1) // Prefetch: this part ends in a swap into the next build.
		}
		tmp := filepath.Join(partsDir, fmt.Sprintf("part_%03d.tmp.mp4", p))
		if err := renderPart(tmp, clips, shotFrames(p), shotFrames(p+1), shots, o, func(f int) {
			fmt.Printf("\r%s: frame %d/%d (%s)  ", names[p], f+1, frames, time.Since(start).Round(time.Second))
		}); err != nil {
			return err
		}
		fmt.Println()
		if err := os.Rename(tmp, part); err != nil {
			return err
		}
		if p > 0 {
			clips.release(p - 1) // Later parts reach back at most one build.
		}
	}
	listPath := filepath.Join(partsDir, "parts.txt")
	if err := os.WriteFile(listPath, []byte(list.String()), 0o644); err != nil {
		return err
	}
	join := exec.Command("ffmpeg", "-loglevel", "error", "-y", "-f", "concat", "-safe", "0", "-i", listPath,
		"-c", "copy", "-movflags", "+faststart", path)
	join.Stderr = os.Stderr
	if err := join.Run(); err != nil {
		return err
	}
	return os.RemoveAll(partsDir)
}

// renderPart renders frames [f0, f1) on o.workers goroutines and encodes them to path.
func renderPart(path string, clips *clipCache, f0, f1, shots int, o videoOptions, progress func(int)) error {
	ffmpeg := exec.Command("ffmpeg", "-loglevel", "error", "-y", "-f", "rawvideo", "-pix_fmt", "rgb24",
		"-s", fmt.Sprintf("%dx%d", o.w, o.h), "-r", strconv.Itoa(o.fps), "-i", "-",
		"-c:v", "libx264", "-preset", "slow", "-crf", "16", "-pix_fmt", "yuv420p", "-threads", "2", path)
	ffmpeg.Stderr = os.Stderr
	pipe, err := ffmpeg.StdinPipe()
	if err != nil {
		return err
	}
	if err := ffmpeg.Start(); err != nil {
		return err
	}
	n := f1 - f0
	jobs := make(chan int)
	done := make([]chan []byte, n)
	for i := range done {
		done[i] = make(chan []byte, 1)
	}
	// At most two frames per worker wait to be written, bounding memory.
	window := make(chan struct{}, 2*o.workers)
	mask := vignette(o.w, o.h)
	var wg sync.WaitGroup
	for range o.workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			pixels := o.w * o.h * 3
			blur, sharp := make([]float32, pixels), make([]float32, pixels)
			mixBlur, mixSharp := make([]float32, pixels), make([]float32, pixels)
			out := make([]float32, pixels)
			buf := &depthBuffers{}
			for f := range jobs {
				p := planFrame(f, shots, o)
				if p.prev < 0 {
					clips.get(p.next).frame(p.tau, o.shot, o, mixBlur, mixSharp, buf)
					copy(out, mixBlur)
				} else {
					// Around a cut, cross-fade the two builds at peak speed under extra
					// sideways blur, which swells and fades with the swap.
					d := p.tau
					u := d / (o.swap / 2) // -1 … 1 across the swap.
					mix := smoothstep(-0.4, 0.4, u)
					clear(mixBlur)
					clear(mixSharp)
					add := func(c *clip, tau, weight float64) {
						if weight <= 0 {
							return
						}
						c.frame(tau, o.shot, o, blur, sharp, buf)
						for k := range blur {
							mixBlur[k] += blur[k] * float32(weight)
							mixSharp[k] += sharp[k] * float32(weight)
						}
					}
					add(clips.get(p.prev), o.shot+d, 1-mix)
					add(clips.get(p.next), d, mix)
					hblur(mixBlur, out, o.w, o.h, 0.045*float64(o.w)*(1+math.Cos(math.Pi*u))/2)
				}
				// Blur only inside the vignette: the screen edges stay sharp.
				rgb := make([]byte, pixels)
				for k := range out {
					m := mask[k/3]
					rgb[k] = linearToSRGB(float64(mixSharp[k] + m*(out[k]-mixSharp[k])))
				}
				done[f-f0] <- rgb
			}
		}()
	}
	go func() {
		for f := f0; f < f1; f++ {
			window <- struct{}{}
			for _, i := range planFrame(f, shots, o).uses() {
				clips.load(i)
				clips.get(i)
			}
			jobs <- f
		}
		close(jobs)
	}()
	var writeErr error
	for k := range n {
		rgb := <-done[k]
		done[k] = nil
		<-window
		if writeErr == nil {
			_, writeErr = pipe.Write(rgb)
		}
		if (k+1)%10 == 0 || k+1 == n {
			progress(f0 + k)
		}
	}
	wg.Wait()
	pipe.Close()
	if err := ffmpeg.Wait(); err != nil {
		return err
	}
	return writeErr
}

// framePlan says which clips a frame shows: next alone at tau, or, within a swap, prev
// cross-fading into next with tau the signed seconds from their cut.
type framePlan struct {
	prev, next int
	tau        float64
}

func (p framePlan) uses() []int {
	if p.prev < 0 {
		return []int{p.next}
	}
	return []int{p.prev, p.next}
}

func planFrame(f, shots int, o videoOptions) framePlan {
	g := (float64(f) + 0.5) / float64(o.fps)
	i := min(int(g/o.shot), shots-1)
	tau := g - float64(i)*o.shot
	if i > 0 && tau < o.swap/2 {
		return framePlan{i - 1, i, tau}
	}
	if i+1 < shots && o.shot-tau < o.swap/2 {
		return framePlan{i, i + 1, tau - o.shot}
	}
	return framePlan{-1, i, tau}
}

// clipCache builds clips on demand in the background, since each holds a few hundred
// megabytes of faces and shadow map.
type clipCache struct {
	structures []*octa.Structure
	names      []string
	o          videoOptions
	mu         sync.Mutex
	started    []bool
	ready      []chan struct{}
	clips      []*clip
}

func newClipCache(structures []*octa.Structure, names []string, o videoOptions) *clipCache {
	c := &clipCache{structures: structures, names: names, o: o, started: make([]bool, len(structures)),
		ready: make([]chan struct{}, len(structures)), clips: make([]*clip, len(structures))}
	for i := range c.ready {
		c.ready[i] = make(chan struct{})
	}
	return c
}

// load starts building clip i unless it already has been.
func (c *clipCache) load(i int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.started[i] {
		return
	}
	c.started[i] = true
	go func() {
		look, _ := themeLook(c.o.theme, c.names[i])
		built := newClip(c.structures[i], c.o, look)
		c.mu.Lock()
		c.clips[i] = built
		c.mu.Unlock()
		close(c.ready[i])
	}()
}

// get waits for clip i, which load must have started.
func (c *clipCache) get(i int) *clip {
	<-c.ready[i]
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.clips[i]
}

func (c *clipCache) release(i int) {
	c.mu.Lock()
	c.clips[i] = nil
	c.mu.Unlock()
}

// themeLook reads a world's sky, ground and light colors from world_theme.gd, blending the
// warm late-afternoon light halfway towards the neutral preview light.
func themeLook(path, name string) (look, bool) {
	src, err := os.ReadFile(path)
	if err != nil {
		return neutralLook, false
	}
	block := func(pattern string) string {
		m := regexp.MustCompile(`(?s)` + pattern + `\{(.*?)\n\t*\}`).FindSubmatch(src)
		if m == nil {
			return ""
		}
		return string(m[1])
	}
	shared := block(`const LATE_AFTERNOON := `)
	world := block(`"` + regexp.QuoteMeta(name) + `": _late_afternoon\(`)
	if world == "" {
		return neutralLook, false
	}
	color := func(text, key string, fallback [3]float64) [3]float64 {
		m := regexp.MustCompile(`"` + key + `": Color\("([0-9a-fA-F]{6})"\)`).FindStringSubmatch(text)
		if m == nil {
			return fallback
		}
		v, _ := strconv.ParseUint(m[1], 16, 32)
		return [3]float64{srgbToLinear(uint8(v >> 16)), srgbToLinear(uint8(v >> 8)), srgbToLinear(uint8(v))}
	}
	light := func(key string, neutral [3]float64) [3]float64 {
		c := color(shared, key, neutral)
		peak := math.Max(c[0], math.Max(c[1], c[2]))
		return lerp3(neutral, [3]float64{c[0] / peak, c[1] / peak, c[2] / peak}, 0.35)
	}
	n := neutralLook
	return look{
		zenith:  color(world, "zenith", n.zenith),
		horizon: color(world, "horizon", n.horizon),
		ground:  color(world, "ground", n.ground),
		sun:     light("sun", n.sun),
		ambient: light("ambient", n.ambient),
	}, true
}

func smoothstep(a, b, x float64) float64 {
	t := math.Min(math.Max((x-a)/(b-a), 0), 1)
	return t * t * (3 - 2*t)
}

// sweep maps shot progress 0…1 to orbit progress 0…1, fast at both ends and slow in the
// middle; its slope matches at the ends, so consecutive shots spin at the same speed.
func sweep(u float64) float64 {
	x := math.Min(math.Max(u, 0), 1) * orbitSteps
	i := min(int(x), orbitSteps-1)
	return orbitTable[i] + (x-float64(i))*(orbitTable[i+1]-orbitTable[i])
}

// halton is the radical inverse of i in base b, for low-discrepancy jitter.
func halton(i, b int) float64 {
	r, f := 0.0, 1.0
	for ; i > 0; i /= b {
		f /= float64(b)
		r += f * float64(i%b)
	}
	return r
}
