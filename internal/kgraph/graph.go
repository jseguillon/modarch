package kgraph

import (
	"bytes"
	"fmt"
	"html"
	"sort"
	"strings"
	"text/template"
)

func BuildGraph(resources []Resource) (Graph, error) {
	g := Graph{Nodes: map[string]*Node{}}

	var gstyles []struct {
		Meta  Metadata
		Style GraphStyle
	}
	var estyles []EdgeStyle
	byKey := map[string]Resource{}
	for _, r := range resources {
		byKey[r.Key()] = r
		if r.Kind == "GraphStyle" {
			s, _ := DecodeSpec[GraphStyle](r.Raw)
			gstyles = append(gstyles, struct {
				Meta  Metadata
				Style GraphStyle
			}{Meta: r.Metadata, Style: s})
			continue
		}
		if r.Kind == "EdgeStyle" {
			s, _ := DecodeSpec[EdgeStyle](r.Raw)
			estyles = append(estyles, s)
			continue
		}
		if r.Kind == "Frame" || r.Kind == "Animation" || r.Kind == "PatchSet" {
			continue
		}
		if r.Metadata.Annotations["kgraph.io/ignore"] == "true" {
			continue
		}
		n := toNode(r)
		for _, st := range gstyles {
			if matchesSelector(r, st.Style.Selector) {
				for k, v := range st.Style.Node.Attrs {
					n.Attrs[k] = fmt.Sprintf("%v", v)
				}
				if st.Style.Node.LabelTemplate != "" {
					label, err := renderLabelTemplate(st.Style.Node.LabelTemplate, r)
					if err == nil {
						n.Label = label
					}
				}
			}
		}
		for k, v := range r.Metadata.Annotations {
			if strings.HasPrefix(k, "kgraph.io/node.attr.") {
				n.Attrs[strings.TrimPrefix(k, "kgraph.io/node.attr.")] = v
			}
		}
		if tpl, ok := r.Metadata.Annotations["kgraph.io/node.labelTemplate"]; ok {
			label, err := renderLabelTemplate(tpl, r)
			if err == nil {
				n.Label = label
			}
		}
		g.Nodes[n.ID] = n
	}

	for _, r := range resources {
		switch r.Kind {
		case "Service":
			addServiceEdges(&g, r, resources)
		case "Deployment":
			addEnvFromEdges(&g, r, resources)
		case "Link":
			spec, _ := DecodeSpec[LinkSpec](r.Raw)
			from := findNodeByRef(spec.FromRef, byKey)
			to := findNodeByRef(spec.ToRef, byKey)
			if from != "" && to != "" {
				g.Edges = append(g.Edges, Edge{From: from, To: to, Type: spec.Type, Attrs: map[string]string{}})
			}
		case "Note":
			spec, _ := DecodeSpec[NoteSpec](r.Raw)
			noteID := nodeID(r)
			if _, ok := g.Nodes[noteID]; !ok {
				continue
			}
			if spec.Text != "" {
				g.Nodes[noteID].Label = spec.Text
				g.Nodes[noteID].Attrs["shape"] = "note"
			}
			if spec.TargetRef.Name != "" {
				to := findNodeByRef(spec.TargetRef, byKey)
				if to != "" {
					g.Edges = append(g.Edges, Edge{From: noteID, To: to, Type: "note", Attrs: map[string]string{}})
				}
			}
		}
	}

	for i := range g.Edges {
		for _, es := range estyles {
			if es.Selector.Type != "" && es.Selector.Type != g.Edges[i].Type {
				continue
			}
			if len(es.Selector.ToNamespaces) > 0 {
				to := g.Nodes[g.Edges[i].To]
				if to == nil {
					continue
				}
				ns := to.Attrs["namespace"]
				match := false
				for _, allowed := range es.Selector.ToNamespaces {
					if ns == allowed {
						match = true
					}
				}
				if !match {
					continue
				}
			}
			for k, v := range es.Edge.Attrs {
				g.Edges[i].Attrs[k] = fmt.Sprintf("%v", v)
			}
		}
	}

	return g, nil
}

func toNode(r Resource) *Node {
	id := nodeID(r)
	label := fmt.Sprintf("%s\\n%s/%s", r.Kind, r.Metadata.Namespace, r.Metadata.Name)
	if r.Metadata.Namespace == "" {
		label = fmt.Sprintf("%s\\n%s", r.Kind, r.Metadata.Name)
	}
	return &Node{ID: id, Label: label, Attrs: map[string]string{"shape": "box", "namespace": r.Metadata.Namespace}}
}

func nodeID(r Resource) string {
	if r.Metadata.Namespace == "" {
		return fmt.Sprintf("%s_%s", r.Kind, r.Metadata.Name)
	}
	return fmt.Sprintf("%s_%s_%s", r.Kind, r.Metadata.Namespace, r.Metadata.Name)
}

func renderLabelTemplate(src string, r Resource) (string, error) {
	tpl, err := template.New("label").Option("missingkey=zero").Parse(src)
	if err != nil {
		return "", err
	}
	ctx := map[string]any{"kind": r.Kind, "name": r.Metadata.Name, "namespace": r.Metadata.Namespace}
	var b bytes.Buffer
	if err := tpl.Execute(&b, ctx); err != nil {
		return "", err
	}
	return b.String(), nil
}

func addServiceEdges(g *Graph, svc Resource, all []Resource) {
	spec, ok := svc.Raw["spec"].(map[string]any)
	if !ok {
		return
	}
	sel := toStringMap(spec["selector"])
	if len(sel) == 0 {
		return
	}
	for _, r := range all {
		if r.Kind != "Deployment" && r.Kind != "Pod" {
			continue
		}
		if r.Metadata.Namespace != svc.Metadata.Namespace {
			continue
		}
		labels := r.Metadata.Labels
		if r.Kind == "Deployment" {
			labels = deploymentPodLabels(r)
		}
		if matchesLabels(sel, labels) {
			g.Edges = append(g.Edges, Edge{From: nodeID(svc), To: nodeID(r), Type: "routesTo", Attrs: map[string]string{}})
		}
	}
}

func deploymentPodLabels(dep Resource) map[string]string {
	spec, ok := dep.Raw["spec"].(map[string]any)
	if !ok {
		return dep.Metadata.Labels
	}
	tpl, ok := spec["template"].(map[string]any)
	if !ok {
		return dep.Metadata.Labels
	}
	m, ok := tpl["metadata"].(map[string]any)
	if !ok {
		return dep.Metadata.Labels
	}
	return toStringMap(m["labels"])
}

func addEnvFromEdges(g *Graph, dep Resource, all []Resource) {
	spec, ok := dep.Raw["spec"].(map[string]any)
	if !ok {
		return
	}
	tpl, ok := spec["template"].(map[string]any)
	if !ok {
		return
	}
	pspec, ok := tpl["spec"].(map[string]any)
	if !ok {
		return
	}
	containers, ok := pspec["containers"].([]any)
	if !ok {
		return
	}
	for _, c := range containers {
		cm, ok := c.(map[string]any)
		if !ok {
			continue
		}
		envFrom, ok := cm["envFrom"].([]any)
		if !ok {
			continue
		}
		for _, ef := range envFrom {
			em, ok := ef.(map[string]any)
			if !ok {
				continue
			}
			cmref, ok := em["configMapRef"].(map[string]any)
			if !ok {
				continue
			}
			name, _ := cmref["name"].(string)
			for _, r := range all {
				if r.Kind == "ConfigMap" && r.Metadata.Name == name && r.Metadata.Namespace == dep.Metadata.Namespace {
					g.Edges = append(g.Edges, Edge{From: nodeID(dep), To: nodeID(r), Type: "envFrom", Attrs: map[string]string{}})
				}
			}
		}
	}
}

func matchesLabels(selector, labels map[string]string) bool {
	for k, v := range selector {
		if labels[k] != v {
			return false
		}
	}
	return true
}

func findNodeByRef(ref Ref, byKey map[string]Resource) string {
	for _, r := range byKey {
		if ref.Kind != "" && r.Kind != ref.Kind {
			continue
		}
		if ref.APIVersion != "" && r.APIVersion != ref.APIVersion {
			continue
		}
		if ref.Name != "" && r.Metadata.Name != ref.Name {
			continue
		}
		if ref.Namespace != "" && r.Metadata.Namespace != ref.Namespace {
			continue
		}
		return nodeID(r)
	}
	return ""
}

func RenderDOT(g Graph) string {
	var b strings.Builder
	b.WriteString("digraph G {\n")
	b.WriteString("  rankdir=LR;\n")
	b.WriteString("  graph [fontname=\"Inter\"];\n")
	b.WriteString("  node [fontname=\"Inter\"];\n")
	b.WriteString("  edge [fontname=\"Inter\"];\n")
	keys := make([]string, 0, len(g.Nodes))
	for k := range g.Nodes {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		n := g.Nodes[k]
		attrs := map[string]string{"label": n.Label}
		for ak, av := range n.Attrs {
			attrs[ak] = av
		}
		b.WriteString(fmt.Sprintf("  %q [%s];\n", n.ID, attrsDOT(attrs)))
	}
	for _, e := range g.Edges {
		attrs := map[string]string{}
		for k, v := range e.Attrs {
			attrs[k] = v
		}
		if _, ok := attrs["label"]; !ok {
			attrs["label"] = e.Type
		}
		b.WriteString(fmt.Sprintf("  %q -> %q [%s];\n", e.From, e.To, attrsDOT(attrs)))
	}
	b.WriteString("}\n")
	return b.String()
}

func attrsDOT(attrs map[string]string) string {
	keys := make([]string, 0, len(attrs))
	for k := range attrs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		v := strings.ReplaceAll(attrs[k], "\n", "\\n")
		if strings.Contains(v, "<") && strings.Contains(v, ">") {
			parts = append(parts, fmt.Sprintf("%s=%s", k, v))
			continue
		}
		parts = append(parts, fmt.Sprintf("%s=\"%s\"", k, strings.ReplaceAll(v, "\"", "\\\"")))
	}
	return strings.Join(parts, ",")
}

func RenderSVG(g Graph) string {
	keys := make([]string, 0, len(g.Nodes))
	for k := range g.Nodes {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	const (
		width      = 1280
		headerH    = 56
		nodeW      = 320
		nodeH      = 52
		nodeGapY   = 28
		nodeStartX = 60
		nodeStartY = 96
	)

	positions := map[string]struct{ x, y int }{}
	for i, id := range keys {
		y := nodeStartY + i*(nodeH+nodeGapY)
		positions[id] = struct{ x, y int }{x: nodeStartX, y: y}
	}

	height := nodeStartY + len(keys)*(nodeH+nodeGapY) + 80
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">`, width, height, width, height))
	b.WriteString(`<rect x="0" y="0" width="100%" height="100%" fill="#f8fafc"/>`)
	b.WriteString(`<text x="24" y="36" font-family="Inter,Arial,sans-serif" font-size="24" font-weight="700" fill="#0f172a">kgraph test render (SVG)</text>`)

	for _, e := range g.Edges {
		from, okFrom := positions[e.From]
		to, okTo := positions[e.To]
		if !okFrom || !okTo {
			continue
		}
		x1 := from.x + nodeW
		y1 := from.y + nodeH/2
		x2 := to.x
		y2 := to.y + nodeH/2
		color := e.Attrs["color"]
		if color == "" {
			color = "#64748b"
		}
		b.WriteString(fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="%s" stroke-width="2" marker-end="url(#arrow)"/>`, x1, y1, x2, y2, color))
		label := e.Attrs["label"]
		if label == "" {
			label = e.Type
		}
		mx := (x1 + x2) / 2
		my := (y1+y2)/2 - 6
		b.WriteString(fmt.Sprintf(`<text x="%d" y="%d" font-family="Inter,Arial,sans-serif" font-size="12" fill="#334155">%s</text>`, mx, my, html.EscapeString(label)))
	}

	b.WriteString(`<defs><marker id="arrow" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="7" markerHeight="7" orient="auto-start-reverse"><path d="M 0 0 L 10 5 L 0 10 z" fill="#64748b"/></marker></defs>`)

	for _, id := range keys {
		n := g.Nodes[id]
		pos := positions[id]
		color := n.Attrs["color"]
		if color == "" {
			color = "#1e293b"
		}
		fill := "#ffffff"
		b.WriteString(fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" rx="8" ry="8" fill="%s" stroke="%s" stroke-width="2"/>`, pos.x, pos.y, nodeW, nodeH, fill, color))
		label := n.Label
		label = strings.ReplaceAll(label, "\\n", " | ")
		label = strings.ReplaceAll(label, "\n", " | ")
		b.WriteString(fmt.Sprintf(`<text x="%d" y="%d" font-family="Inter,Arial,sans-serif" font-size="13" fill="#0f172a">%s</text>`, pos.x+12, pos.y+30, html.EscapeString(label)))
	}

	b.WriteString(`</svg>`)
	return b.String()
}
