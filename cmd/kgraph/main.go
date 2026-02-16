package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/example/modarch/internal/kgraph"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	switch os.Args[1] {
	case "render":
		if err := runRender(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
	case "animate":
		if err := runAnimate(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
	default:
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Println("kgraph <render|animate>")
}

func runRender(args []string) error {
	fs := flag.NewFlagSet("render", flag.ExitOnError)
	input := fs.String("i", "", "YAML input file or directory")
	output := fs.String("o", "out/graph.dot", "output DOT file")
	svg := fs.String("svg", "", "optional svg output")
	patch := fs.String("p", "", "optional patch file")
	png := fs.String("png", "", "optional png output (requires dot)")
	_ = fs.Parse(args)
	if *input == "" {
		return fmt.Errorf("-i is required")
	}
	resources, err := kgraph.LoadResources(*input)
	if err != nil {
		return err
	}
	if *patch != "" {
		resources, err = kgraph.ApplyPatchFiles(resources, []string{*patch})
		if err != nil {
			return err
		}
	}
	g, err := kgraph.BuildGraph(resources)
	if err != nil {
		return err
	}
	dot := kgraph.RenderDOT(g)
	if err := kgraph.WriteFile(*output, dot); err != nil {
		return err
	}
	if *svg != "" {
		if err := kgraph.WriteFile(*svg, kgraph.RenderSVG(g)); err != nil {
			return err
		}
	}
	if *png != "" {
		if err := kgraph.RenderDOTToPNG(*output, *png); err != nil {
			return fmt.Errorf("render png: %w", err)
		}
	}
	fmt.Println("wrote", *output)
	return nil
}

func runAnimate(args []string) error {
	fs := flag.NewFlagSet("animate", flag.ExitOnError)
	input := fs.String("i", "", "YAML input file or directory")
	animName := fs.String("name", "", "animation name")
	outDir := fs.String("out", "out/frames", "output frame directory")
	apng := fs.String("apng", "", "animated png output path")
	_ = fs.Parse(args)
	if *input == "" {
		return fmt.Errorf("-i is required")
	}
	resources, err := kgraph.LoadResources(*input)
	if err != nil {
		return err
	}
	animRes, err := kgraph.FindAnimation(resources, *animName)
	if err != nil {
		return err
	}
	anim, err := kgraph.DecodeSpec[kgraph.AnimationSpec](animRes.Raw)
	if err != nil {
		return err
	}
	basePatched, err := kgraph.ApplyPatchFiles(resources, anim.Base.Patches)
	if err != nil {
		return err
	}
	framePatches := kgraph.FramePatchMap(resources)
	var cumulative []string
	if err := kgraph.EnsureDir(*outDir); err != nil {
		return err
	}
	for i, ref := range anim.Frames.Refs {
		cumulative = append(cumulative, framePatches[ref.Name]...)
		frameResources, err := kgraph.ApplyPatchFiles(basePatched, cumulative)
		if err != nil {
			return err
		}
		g, err := kgraph.BuildGraph(frameResources)
		if err != nil {
			return err
		}
		dotPath := filepath.Join(*outDir, fmt.Sprintf("frame-%03d.dot", i+1))
		pngPath := filepath.Join(*outDir, fmt.Sprintf("frame-%03d.png", i+1))
		svgPath := filepath.Join(*outDir, fmt.Sprintf("frame-%03d.svg", i+1))
		if err := kgraph.WriteFile(dotPath, kgraph.RenderDOT(g)); err != nil {
			return err
		}
		if err := kgraph.WriteFile(svgPath, kgraph.RenderSVG(g)); err != nil {
			return err
		}
		_ = kgraph.RenderDOTToPNG(dotPath, pngPath)
		fmt.Println("wrote", dotPath)
	}
	if *apng != "" {
		pattern := filepath.Join(*outDir, "frame-%03d.png")
		if err := kgraph.BuildAPNG(pattern, *apng); err != nil {
			return fmt.Errorf("build apng: %w", err)
		}
		fmt.Println("wrote", *apng)
	}
	return nil
}
