package kgraph

import "fmt"

type Metadata struct {
	Name        string
	Namespace   string
	Labels      map[string]string
	Annotations map[string]string
}

type Resource struct {
	APIVersion string
	Kind       string
	Metadata   Metadata
	Raw        map[string]any
}

func (r Resource) Key() string {
	return fmt.Sprintf("%s|%s|%s|%s", r.APIVersion, r.Kind, r.Metadata.Namespace, r.Metadata.Name)
}

type Selector struct {
	Kinds         []string `yaml:"kinds"`
	Kind          string   `yaml:"kind"`
	Name          string   `yaml:"name"`
	Namespace     string   `yaml:"namespace"`
	LabelSelector string   `yaml:"labelSelector"`
	Type          string   `yaml:"type"`
	ToNamespaces  []string `yaml:"toNamespaces"`
}

type GraphStyle struct {
	Selector Selector `yaml:"selector"`
	Node     struct {
		Attrs         map[string]any `yaml:"attrs"`
		LabelTemplate string         `yaml:"labelTemplate"`
	} `yaml:"node"`
}

type EdgeStyle struct {
	Selector Selector `yaml:"selector"`
	Edge     struct {
		Attrs map[string]any `yaml:"attrs"`
	} `yaml:"edge"`
}

type LinkSpec struct {
	Type    string `yaml:"type"`
	FromRef Ref    `yaml:"fromRef"`
	ToRef   Ref    `yaml:"toRef"`
}

type Ref struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Namespace  string `yaml:"namespace"`
	Name       string `yaml:"name"`
}

type NoteSpec struct {
	Text      string `yaml:"text"`
	TargetRef Ref    `yaml:"targetRef"`
}

type FrameSpec struct {
	Patches []string `yaml:"patches"`
}

type AnimationSpec struct {
	Base struct {
		Inputs  []string `yaml:"inputs"`
		Patches []string `yaml:"patches"`
	} `yaml:"base"`
	Frames struct {
		Refs []Ref `yaml:"refs"`
	} `yaml:"frames"`
}

type PatchSet struct {
	Operations []PatchOperation `yaml:"operations"`
}

type PatchOperation struct {
	Action   string            `yaml:"action"`
	Selector Selector          `yaml:"selector"`
	Values   map[string]string `yaml:"values"`
}

type Node struct {
	ID    string
	Label string
	Attrs map[string]string
}

type Edge struct {
	From  string
	To    string
	Type  string
	Attrs map[string]string
}

type Graph struct {
	Nodes map[string]*Node
	Edges []Edge
}
