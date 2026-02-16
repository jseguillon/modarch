package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/example/modarch/internal/anim"
	"github.com/example/modarch/internal/engine"
	"github.com/example/modarch/internal/model"
	"github.com/example/modarch/internal/render"
)

func main() {
	input := flag.String("input", "testdata/mock.yaml", "Path to YAML docs")
	outDir := flag.String("out", "outputs", "Output directory")
	iteration := flag.Int("iteration", 4, "How many frames to include")
	flag.Parse()

	bundle, err := model.ParseBundle(*input)
	if err != nil {
		panic(err)
	}
	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		panic(err)
	}

	base := engine.BuildBaseGraph(bundle)
	framePaths := []string{}
	maxFrame := min(*iteration, len(bundle.Frames))
	for i := 0; i < maxFrame; i++ {
		g := engine.ApplyFrame(base, bundle.Frames[i])
		dotPath := filepath.Join(*outDir, fmt.Sprintf("frame-%03d.dot", i+1))
		svgPath := filepath.Join(*outDir, fmt.Sprintf("frame-%03d.svg", i+1))
		pngPath := filepath.Join(*outDir, fmt.Sprintf("frame-%03d.png", i+1))
		if err := os.WriteFile(dotPath, []byte(render.ToDOT(g)), 0o644); err != nil {
			panic(err)
		}
		if err := os.WriteFile(svgPath, []byte(render.ToSVG(g, bundle.Frames[i].Name)), 0o644); err != nil {
			panic(err)
		}
		if err := render.WritePNG(g, bundle.Frames[i].Name, pngPath); err != nil {
			panic(err)
		}
		framePaths = append(framePaths, pngPath)
	}
	if len(framePaths) > 0 {
		apngPath := filepath.Join(*outDir, fmt.Sprintf("iteration-%d.apng", maxFrame))
		if err := anim.WriteAPNG(apngPath, framePaths, 35); err != nil {
			panic(err)
		}
		fmt.Printf("generated %s with %d frames\n", apngPath, len(framePaths))
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
