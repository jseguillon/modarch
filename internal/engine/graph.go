package engine

import (
	"fmt"
	"slices"

	"github.com/example/modarch/internal/model"
)

type Graph struct {
	Nodes []model.Resource
	Edges []model.Edge
}

func BuildBaseGraph(bundle *model.Bundle) Graph {
	nodes := append([]model.Resource{}, bundle.Resources...)
	edges := []model.Edge{}
	for _, n := range nodes {
		switch n.Kind {
		case "Deployment":
			edges = append(edges, deploymentEnvFromEdges(n)...)
		case "Service":
			edges = append(edges, serviceToBackends(n, nodes)...)
		case "Link":
			edges = append(edges, explicitLink(n))
		case "Note":
			e, ok := noteEdge(n)
			if ok {
				edges = append(edges, e)
			}
		}
	}

	styleByType := map[string]map[string]string{}
	for _, s := range bundle.EdgeStyles {
		styleByType[s.Type] = s.Attrs
	}
	for i := range edges {
		if style, ok := styleByType[edges[i].Type]; ok {
			for k, v := range style {
				edges[i].Attrs[k] = v
			}
		}
	}
	return Graph{Nodes: nodes, Edges: edges}
}

func ApplyFrame(g Graph, f model.Frame) Graph {
	next := cloneGraph(g)
	for _, patch := range f.Patches {
		if patch.Target["kind"] == "Edge" {
			for i, edge := range next.Edges {
				if edgeMatchesPatch(edge, patch.Target) {
					for _, op := range patch.Ops {
						switch op.Path {
						case "/metadata/annotations/kgraph.io~1edgeColor":
							edge.Attrs["color"] = fmt.Sprintf("%v", op.Value)
						case "/metadata/annotations/kgraph.io~1edgeLabel":
							edge.Attrs["label"] = fmt.Sprintf("%v", op.Value)
						}
					}
					next.Edges[i] = edge
				}
			}
			continue
		}

		for i, node := range next.Nodes {
			if node.Kind == fmt.Sprintf("%v", patch.Target["kind"]) &&
				node.Name == fmt.Sprintf("%v", patch.Target["name"]) &&
				node.Namespace == fmt.Sprintf("%v", patch.Target["namespace"]) {
				if node.Annotations == nil {
					node.Annotations = map[string]string{}
				}
				for _, op := range patch.Ops {
					switch op.Path {
					case "/metadata/annotations/kgraph.io~1fold":
						node.Annotations["kgraph.io/fold"] = fmt.Sprintf("%v", op.Value)
					case "/metadata/annotations/kgraph.io~1foldMode":
						node.Annotations["kgraph.io/foldMode"] = fmt.Sprintf("%v", op.Value)
					}
				}
				next.Nodes[i] = node
			}
		}
	}
	next.Edges = filterEdges(next)
	return next
}

func cloneGraph(g Graph) Graph {
	nodes := make([]model.Resource, len(g.Nodes))
	for i, n := range g.Nodes {
		copied := n
		copied.Labels = cloneMap(n.Labels)
		copied.Annotations = cloneMap(n.Annotations)
		nodes[i] = copied
	}
	edges := make([]model.Edge, len(g.Edges))
	for i, e := range g.Edges {
		edges[i] = model.Edge{From: e.From, To: e.To, Type: e.Type, Attrs: cloneMap(e.Attrs)}
	}
	return Graph{Nodes: nodes, Edges: edges}
}

func cloneMap(in map[string]string) map[string]string {
	if in == nil {
		return map[string]string{}
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func filterEdges(g Graph) []model.Edge {
	folded := map[string]bool{}
	for _, n := range g.Nodes {
		if n.Annotations["kgraph.io/fold"] == "true" {
			folded[n.ID()] = true
		}
	}
	filtered := []model.Edge{}
	for _, e := range g.Edges {
		if folded[e.From.ID()] || folded[e.To.ID()] {
			continue
		}
		filtered = append(filtered, e)
	}
	return filtered
}

func deploymentEnvFromEdges(dep model.Resource) []model.Edge {
	out := []model.Edge{}
	template := toMap(dep.Spec["template"])
	spec := toMap(template["spec"])
	containers, _ := spec["containers"].([]any)
	for _, c := range containers {
		cm := toMap(c)
		envFrom, _ := cm["envFrom"].([]any)
		for _, item := range envFrom {
			cfg := toMap(toMap(item)["configMapRef"])
			name := fmt.Sprintf("%v", cfg["name"])
			out = append(out, model.Edge{
				From:  model.Ref{Kind: "Deployment", Namespace: dep.Namespace, Name: dep.Name},
				To:    model.Ref{Kind: "ConfigMap", Namespace: dep.Namespace, Name: name},
				Type:  "envFrom",
				Attrs: map[string]string{},
			})
		}
	}
	return out
}

func serviceToBackends(svc model.Resource, nodes []model.Resource) []model.Edge {
	out := []model.Edge{}
	sel := toMap(svc.Spec["selector"])
	app := fmt.Sprintf("%v", sel["app"])
	if app == "" {
		return out
	}
	for _, n := range nodes {
		if n.Namespace != svc.Namespace || !slices.Contains([]string{"Deployment", "Pod"}, n.Kind) {
			continue
		}
		if n.Labels["app"] == app {
			out = append(out, model.Edge{
				From:  model.Ref{Kind: "Service", Namespace: svc.Namespace, Name: svc.Name},
				To:    model.Ref{Kind: n.Kind, Namespace: n.Namespace, Name: n.Name},
				Type:  "selects",
				Attrs: map[string]string{"style": "dotted", "color": "#6b7280", "label": "selects"},
			})
		}
	}
	return out
}

func explicitLink(link model.Resource) model.Edge {
	from := toMap(link.Spec["fromRef"])
	to := toMap(link.Spec["toRef"])
	return model.Edge{
		From:  model.Ref{Kind: str(from["kind"]), Namespace: str(from["namespace"]), Name: str(from["name"])},
		To:    model.Ref{Kind: str(to["kind"]), Namespace: str(to["namespace"]), Name: str(to["name"])},
		Type:  str(link.Spec["type"]),
		Attrs: map[string]string{},
	}
}

func noteEdge(note model.Resource) (model.Edge, bool) {
	target := toMap(note.Spec["targetRef"])
	if len(target) == 0 {
		return model.Edge{}, false
	}
	return model.Edge{
		From:  model.Ref{Kind: "Note", Namespace: note.Namespace, Name: note.Name},
		To:    model.Ref{Kind: str(target["kind"]), Namespace: str(target["namespace"]), Name: str(target["name"])},
		Type:  "note",
		Attrs: map[string]string{},
	}, true
}

func edgeMatchesPatch(e model.Edge, target map[string]any) bool {
	from := toMap(target["from"])
	to := toMap(target["to"])
	return e.Type == fmt.Sprintf("%v", target["type"]) &&
		e.From.Kind == fmt.Sprintf("%v", from["kind"]) && e.From.Name == fmt.Sprintf("%v", from["name"]) && e.From.Namespace == fmt.Sprintf("%v", from["namespace"]) &&
		e.To.Kind == fmt.Sprintf("%v", to["kind"]) && e.To.Name == fmt.Sprintf("%v", to["name"]) && e.To.Namespace == fmt.Sprintf("%v", to["namespace"])
}

func toMap(v any) map[string]any {
	m, _ := v.(map[string]any)
	if m == nil {
		return map[string]any{}
	}
	return m
}

func str(v any) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}
