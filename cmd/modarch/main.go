package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/example/modarch/internal/engine"
	"github.com/example/modarch/internal/model"
	"github.com/example/modarch/internal/render"
)

func main() {
	input := flag.String("input", "testdata/mock.yaml", "Path to YAML docs")
	outDir := flag.String("out", "outputs", "Output directory")
	flag.Parse()

	bundle, err := model.ParseBundle(*input)
	if err != nil {
		panic(err)
	}
	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		panic(err)
	}

	graph := engine.BuildBaseGraph(bundle)
	dotPath := filepath.Join(*outDir, "graph.dot")
	svgPath := filepath.Join(*outDir, "graph.svg")
	pngPath := filepath.Join(*outDir, "graph.png")

	if err := os.WriteFile(dotPath, []byte(render.ToDOT(graph)), 0o644); err != nil {
		panic(err)
	}
	if err := os.WriteFile(svgPath, []byte(render.ToSVG(graph, "architecture graph")), 0o644); err != nil {
		panic(err)
	}
	if err := render.WritePNG(graph, "architecture graph", pngPath); err != nil {
		panic(err)
	}

	fmt.Printf("generated %s and %s\n", dotPath, svgPath)
}
