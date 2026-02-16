package engine

import (
	"testing"

	"github.com/example/modarch/internal/model"
)

func TestBuildBaseGraph(t *testing.T) {
	bundle, err := model.ParseBundle("../../testdata/mock.yaml")
	if err != nil {
		t.Fatal(err)
	}
	g := BuildBaseGraph(bundle)
	if len(g.Nodes) == 0 {
		t.Fatalf("expected nodes")
	}
	if len(g.Edges) == 0 {
		t.Fatalf("expected edges")
	}
}

func TestExplicitLinkKeepsEmptyNamespace(t *testing.T) {
	bundle, err := model.ParseBundle("../../testdata/mock.yaml")
	if err != nil {
		t.Fatal(err)
	}
	g := BuildBaseGraph(bundle)
	for _, e := range g.Edges {
		if e.Type == "routesTo" {
			if e.From.Namespace != "" {
				t.Fatalf("expected empty namespace for external resource link, got %q", e.From.Namespace)
			}
			if e.From.ID() != "ExternalResource/public-lb" {
				t.Fatalf("unexpected routesTo from id: %s", e.From.ID())
			}
			return
		}
	}
	t.Fatalf("expected routesTo edge")
}
