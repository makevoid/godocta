// Command octagen writes scene structures as OCTA v1 files for the studio.
//
//	go run ./cmd/octagen -list
//	go run ./cmd/octagen                 # every scene into ../structures
//	go run ./cmd/octagen -out dir castle arena
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"godocta/gen/scenes"
)

func main() {
	out := flag.String("out", "../structures", "output directory")
	list := flag.Bool("list", false, "list scenes and exit")
	flag.Parse()
	if *list {
		for _, s := range scenes.All() {
			fmt.Printf("%-16s %s\n", s.Name(), s.Title())
		}
		return
	}
	selected := scenes.All()
	if flag.NArg() > 0 {
		selected = nil
		for _, name := range flag.Args() {
			s, ok := scenes.Get(name)
			if !ok {
				fmt.Fprintf(os.Stderr, "unknown scene %q (try -list)\n", name)
				os.Exit(2)
			}
			selected = append(selected, s)
		}
	}
	failed := false
	for _, s := range selected {
		start := time.Now()
		structure, err := scenes.Compile(s)
		if err == nil {
			err = structure.WriteFile(filepath.Join(*out, s.Name()+".octa"))
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", s.Name(), err)
			failed = true
			continue
		}
		size := structure.Bounds().Size()
		fmt.Printf("%-16s %7d cells  %5.1f × %4.1f × %5.1f units  %v\n",
			s.Name(), len(structure.Cells), size.X, size.Y, size.Z, time.Since(start).Round(time.Millisecond))
	}
	if failed {
		os.Exit(1)
	}
}
