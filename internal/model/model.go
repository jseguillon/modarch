package model

import "fmt"

type Resource struct {
	APIVersion  string
	Kind        string
	Namespace   string
	Name        string
	Labels      map[string]string
	Annotations map[string]string
	Spec        map[string]any
	Data        map[string]any
	Raw         map[string]any
}

func (r Resource) ID() string {
	if r.Namespace == "" {
		return fmt.Sprintf("%s/%s", r.Kind, r.Name)
	}
	return fmt.Sprintf("%s/%s/%s", r.Kind, r.Namespace, r.Name)
}

type Ref struct {
	Kind      string
	Namespace string
	Name      string
}

func (r Ref) ID() string {
	if r.Namespace == "" {
		return fmt.Sprintf("%s/%s", r.Kind, r.Name)
	}
	return fmt.Sprintf("%s/%s/%s", r.Kind, r.Namespace, r.Name)
}

type Edge struct {
	From  Ref
	To    Ref
	Type  string
	Attrs map[string]string
}

type FramePatch struct {
	Target map[string]any `yaml:"target"`
	Ops    []PatchOp      `yaml:"ops"`
}

type PatchOp struct {
	Op    string `yaml:"op"`
	Path  string `yaml:"path"`
	Value any    `yaml:"value"`
}

type Frame struct {
	Name    string
	Patches []FramePatch
}

type EdgeStyle struct {
	Type  string
	Attrs map[string]string
}

type Bundle struct {
	Resources  []Resource
	Frames     []Frame
	EdgeStyles []EdgeStyle
}
