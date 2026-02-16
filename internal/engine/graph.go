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
