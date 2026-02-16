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

func TestApplyFrameEdgePatch(t *testing.T) {
	bundle, err := model.ParseBundle("../../testdata/mock.yaml")
	if err != nil {
		t.Fatal(err)
	}
	g := BuildBaseGraph(bundle)
	patched := ApplyFrame(g, bundle.Frames[3])
	found := false
	for _, e := range patched.Edges {
		if e.Type == "envFrom" && e.Attrs["label"] == "envFrom (patched)" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected patched envFrom edge")
	}
}
