package render

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"sort"
	"strings"

	"github.com/example/modarch/internal/engine"
	"github.com/example/modarch/internal/model"
)

type LayoutNode struct {
	Node model.Resource
	X    int
	Y    int
}

func ToDOT(g engine.Graph) string {
	b := &strings.Builder{}
	b.WriteString("digraph modarch {\n")
	b.WriteString("  rankdir=LR;\n")
	for _, n := range g.Nodes {
		if n.Annotations["kgraph.io/fold"] == "true" && n.Annotations["kgraph.io/foldMode"] == "collapse" {
			b.WriteString(fmt.Sprintf("  \"%s\" [shape=box style=filled fillcolor=\"#f3f4f6\" label=\"%s\\n(collapsed)\"];\n", n.ID(), n.ID()))
			continue
		}
		b.WriteString(fmt.Sprintf("  \"%s\" [shape=box label=\"%s\"];\n", n.ID(), n.ID()))
	}
	for _, e := range g.Edges {
		attrs := []string{}
		for k, v := range e.Attrs {
			attrs = append(attrs, fmt.Sprintf("%s=\"%s\"", k, v))
		}
		sort.Strings(attrs)
		b.WriteString(fmt.Sprintf("  \"%s\" -> \"%s\" [%s];\n", e.From.ID(), e.To.ID(), strings.Join(attrs, ",")))
	}
	b.WriteString("}\n")
	return b.String()
}

func ToSVG(g engine.Graph, title string) string {
	layout := layoutNodes(g)
	w, h := 1200, 700
	b := &strings.Builder{}
	b.WriteString(fmt.Sprintf("<svg xmlns='http://www.w3.org/2000/svg' width='%d' height='%d'>", w, h))
	b.WriteString("<rect width='100%' height='100%' fill='white'/>")
	b.WriteString(fmt.Sprintf("<text x='20' y='30' font-size='20' font-family='Arial'>%s</text>", title))
	for _, e := range g.Edges {
		from, ok1 := layout[e.From.ID()]
		to, ok2 := layout[e.To.ID()]
		if !ok1 || !ok2 {
			continue
		}
		color := pick(e.Attrs["color"], "#374151")
		b.WriteString(fmt.Sprintf("<line x1='%d' y1='%d' x2='%d' y2='%d' stroke='%s' stroke-width='2'/>", from.X+120, from.Y+30, to.X, to.Y+30, color))
		if label := e.Attrs["label"]; label != "" {
			mx, my := (from.X+to.X)/2, (from.Y+to.Y)/2
			b.WriteString(fmt.Sprintf("<text x='%d' y='%d' font-size='12' fill='%s'>%s</text>", mx, my, color, label))
		}
	}
	for _, n := range layout {
		fill := "#eef2ff"
		if n.Node.Kind == "ExternalResource" {
			fill = "#dcfce7"
		}
		if n.Node.Kind == "Note" {
			fill = "#ffedd5"
		}
		if n.Node.Annotations["kgraph.io/fold"] == "true" {
			fill = "#f3f4f6"
		}
		b.WriteString(fmt.Sprintf("<rect x='%d' y='%d' width='120' height='60' fill='%s' stroke='#111827'/>", n.X, n.Y, fill))
		b.WriteString(fmt.Sprintf("<text x='%d' y='%d' font-size='11' font-family='Arial'>%s</text>", n.X+6, n.Y+25, n.Node.Kind))
		b.WriteString(fmt.Sprintf("<text x='%d' y='%d' font-size='11' font-family='Arial'>%s</text>", n.X+6, n.Y+42, n.Node.Name))
	}
	b.WriteString("</svg>")
	return b.String()
}

func WritePNG(g engine.Graph, title, path string) error {
	layout := layoutNodes(g)
	img := image.NewRGBA(image.Rect(0, 0, 1200, 700))
	draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)
	for _, n := range layout {
		fill := color.RGBA{238, 242, 255, 255}
		if n.Node.Kind == "ExternalResource" {
			fill = color.RGBA{220, 252, 231, 255}
		}
		if n.Node.Kind == "Note" {
			fill = color.RGBA{255, 237, 213, 255}
		}
		r := image.Rect(n.X, n.Y, n.X+120, n.Y+60)
		draw.Draw(img, r, &image.Uniform{fill}, image.Point{}, draw.Src)
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func layoutNodes(g engine.Graph) map[string]LayoutNode {
	out := map[string]LayoutNode{}
	x, y, row := 40, 80, 0
	for _, n := range g.Nodes {
		out[n.ID()] = LayoutNode{Node: n, X: x, Y: y}
		y += 90
		row++
		if row == 6 {
			row = 0
			y = 80
			x += 180
		}
	}
	return out
}

func pick(v, d string) string {
	if v == "" {
		return d
	}
	return v
}
