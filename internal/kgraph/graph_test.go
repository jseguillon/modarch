package kgraph

import (
	"path/filepath"
	"testing"
)

func svgSnapshotPath(name string) string {
	return filepath.Join("..", "..", "test_artifacts", "tests", name+".svg")
}

func TestBuildGraphFromFixture(t *testing.T) {
	resources, err := LoadResources("../../examples/testdata/mock.yaml")
	if err != nil {
		t.Fatal(err)
	}
	g, err := BuildGraph(resources)
	if err != nil {
		t.Fatal(err)
	}
	if len(g.Nodes) == 0 {
		t.Fatal("expected nodes")
	}
	if len(g.Edges) == 0 {
		t.Fatal("expected edges")
	}
	out := svgSnapshotPath("TestBuildGraphFromFixture")
	if err := WriteFile(out, RenderSVG(g)); err != nil {
		t.Fatal(err)
	}
}

func TestPatchRemovesService(t *testing.T) {
	resources, err := LoadResources("../../examples/testdata/mock.yaml")
	if err != nil {
		t.Fatal(err)
	}
	patched, err := ApplyPatchFiles(resources, []string{"../../patches/frame-003.yaml"})
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range patched {
		if r.Kind == "Service" {
			if r.Metadata.Annotations["kgraph.io/ignore"] != "true" {
				t.Fatalf("expected service ignore annotation")
			}
		}
	}
	g, err := BuildGraph(patched)
	if err != nil {
		t.Fatal(err)
	}
	out := svgSnapshotPath("TestPatchRemovesService")
	if err := WriteFile(out, RenderSVG(g)); err != nil {
		t.Fatal(err)
	}
}
