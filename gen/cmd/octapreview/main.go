// Command octapreview renders .octa files to PNG on the CPU: no window, GPU or Godot.
// It draws the exposed faces of every cell with flat shading and sun shadows over a
// ground plane, from the studio's camera views, as one 2×2 sheet per structure.
//
//	go run ./cmd/octapreview castle                   # ../captures/preview/castle.png
//	go run ./cmd/octapreview -views hero -size 1600x1000 ../structures/arena.octa
//	go run ./cmd/octapreview -video ../captures/showcase.mp4 castle arena cube
//
// Sheet layout: hero, side on the top row; front, top below. Names without a slash or
// extension resolve to ../structures/<name>.octa.
//
// -video renders one MP4 (through ffmpeg) instead: each build orbits fast, slows while its
// last cells land and speeds up again; the builds swap at full speed under motion blur.
package main

import (
	"flag"
	"fmt"
	"image"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"godocta/gen/octa"
)

// Views match tools/capture.gd: orbit yaw and pitch in radians.
var views = map[string][2]float64{
	"hero":  {0.68, 0.48},
	"front": {0, 0.08},
	"side":  {1.5707963, 0.08},
	"top":   {0, 1.35},
}

func main() {
	out := flag.String("out", "../captures/preview", "output directory")
	viewList := flag.String("views", "hero,side,front,top", "comma-separated views: hero, front, side, top")
	size := flag.String("size", "1000x640", "size of each view in pixels")
	ss := flag.Int("ss", 2, "supersampling factor per axis")
	video := flag.String("video", "", "render one showcase MP4 of all the builds to this path")
	fps := flag.Int("fps", 30, "video frames per second")
	samples := flag.Int("samples", 8, "video samples per frame, for motion blur and antialiasing")
	shot := flag.Float64("shot", 4, "video seconds per build")
	swap := flag.Float64("swap", 0.3, "video seconds of the blurred swap around each cut between builds")
	turn := flag.Float64("turn", 150, "video orbit per build, in degrees")
	left := flag.Float64("left", 1, "video assembly seconds left when each shot starts")
	theme := flag.String("theme", "../scripts/rendering/world_theme.gd", "world themes, for video colors")
	flag.Parse()
	sizeSet := false
	flag.Visit(func(f *flag.Flag) { sizeSet = sizeSet || f.Name == "size" })
	if *video != "" && !sizeSet {
		*size = "1920x1080"
	}
	var w, h int
	if _, err := fmt.Sscanf(*size, "%dx%d", &w, &h); err != nil || w < 16 || h < 16 {
		fmt.Fprintln(os.Stderr, "bad -size:", *size)
		os.Exit(2)
	}
	names := strings.Split(*viewList, ",")
	for _, v := range names {
		if _, ok := views[v]; !ok {
			fmt.Fprintln(os.Stderr, "unknown view:", v)
			os.Exit(2)
		}
	}
	if flag.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "usage: octapreview [flags] name|file.octa ...")
		os.Exit(2)
	}
	if *video != "" {
		if *shot <= *swap || *fps < 1 || *samples < 1 {
			fmt.Fprintln(os.Stderr, "-shot must exceed -swap; -fps and -samples must be positive")
			os.Exit(2)
		}
		var structures []*octa.Structure
		var names []string
		for _, arg := range flag.Args() {
			s, err := octa.ReadFile(structurePath(arg))
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s: %v\n", arg, err)
				os.Exit(1)
			}
			structures = append(structures, s)
			names = append(names, strings.TrimSuffix(filepath.Base(structurePath(arg)), ".octa"))
		}
		o := videoOptions{w: w, h: h, fps: *fps, samples: *samples, shot: *shot, swap: *swap, left: *left,
			turn: *turn * math.Pi / 180, theme: *theme}
		if err := renderVideo(*video, structures, names, o); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		abs, _ := filepath.Abs(*video)
		fmt.Println("→", abs)
		return
	}
	if err := os.MkdirAll(*out, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	failed := false
	for _, arg := range flag.Args() {
		path := structurePath(arg)
		start := time.Now()
		s, err := octa.ReadFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", arg, err)
			failed = true
			continue
		}
		scene := newScene(s)
		cols := 2
		if len(names) == 1 {
			cols = 1
		}
		rows := (len(names) + cols - 1) / cols
		sheet := image.NewRGBA(image.Rect(0, 0, w*cols, h*rows))
		for i, v := range names {
			img := scene.render(views[v][0], views[v][1], w, h, *ss)
			at := image.Pt((i%cols)*w, (i/cols)*h)
			for y := 0; y < h; y++ {
				copy(sheet.Pix[sheet.PixOffset(at.X, at.Y+y):], img.Pix[img.PixOffset(0, y):img.PixOffset(w, y)])
			}
		}
		target := filepath.Join(*out, strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))+".png")
		if err := writePNG(target, sheet); err != nil {
			fmt.Fprintln(os.Stderr, err)
			failed = true
			continue
		}
		b := scene.bounds
		size := b.Max.Sub(b.Min)
		abs, _ := filepath.Abs(target)
		fmt.Printf("%s: %q %d cells, %d faces, %.0f×%.0f×%.0f units → %s (%s)\n", arg, s.Title, len(s.Cells),
			len(scene.faces), size.X, size.Y, size.Z, abs, time.Since(start).Round(time.Millisecond))
	}
	if failed {
		os.Exit(1)
	}
}

// structurePath resolves a bare build name to ../structures/<name>.octa.
func structurePath(arg string) string {
	if !strings.ContainsRune(arg, '/') && filepath.Ext(arg) == "" {
		return filepath.Join("..", "structures", arg+".octa")
	}
	return arg
}

func writePNG(path string, img image.Image) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	if err := png.Encode(f, img); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}
