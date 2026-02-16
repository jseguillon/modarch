package kgraph

import (
	"encoding/json"
	"fmt"
	"strings"
)

func ApplyPatchFiles(resources []Resource, files []string) ([]Resource, error) {
	out := cloneResources(resources)
	for _, f := range files {
		ps, err := loadPatchSet(f)
		if err != nil {
			return nil, err
		}
		for _, op := range ps.Operations {
			var next []Resource
			for _, r := range out {
				if !matchesSelector(r, op.Selector) {
					next = append(next, r)
					continue
				}
				switch op.Action {
				case "remove":
					continue
				case "annotate":
					for k, v := range op.Values {
						r.Metadata.Annotations[k] = v
					}
				case "label":
					for k, v := range op.Values {
						r.Metadata.Labels[k] = v
					}
				default:
					return nil, fmt.Errorf("unsupported patch action %q", op.Action)
				}
				next = append(next, r)
			}
			out = next
		}
	}
	return out, nil
}

func loadPatchSet(path string) (PatchSet, error) {
	var ps PatchSet
	docs, err := readYAMLDocsAsMaps(path)
	if err != nil {
		return ps, err
	}
	if len(docs) == 0 {
		return ps, nil
	}
	raw := docs[0]
	if kind, _ := raw["kind"].(string); kind != "PatchSet" {
		return ps, fmt.Errorf("%s is not a PatchSet", path)
	}
	spec, ok := raw["spec"]
	if !ok {
		return ps, nil
	}
	bb, _ := json.Marshal(spec)
	err = json.Unmarshal(bb, &ps)
	return ps, err
}

func matchesSelector(r Resource, s Selector) bool {
	if s.Kind != "" && r.Kind != s.Kind {
		return false
	}
	if len(s.Kinds) > 0 {
		ok := false
		for _, k := range s.Kinds {
			if r.Kind == k {
				ok = true
				break
			}
		}
		if !ok {
			return false
		}
	}
	if s.Name != "" && r.Metadata.Name != s.Name {
		return false
	}
	if s.Namespace != "" && r.Metadata.Namespace != s.Namespace {
		return false
	}
	if s.LabelSelector != "" {
		parts := strings.SplitN(s.LabelSelector, "=", 2)
		if len(parts) != 2 {
			return false
		}
		if r.Metadata.Labels[strings.TrimSpace(parts[0])] != strings.TrimSpace(parts[1]) {
			return false
		}
	}
	return true
}

func cloneResources(resources []Resource) []Resource {
	out := make([]Resource, 0, len(resources))
	for _, r := range resources {
		cp := r
		cp.Metadata.Labels = map[string]string{}
		for k, v := range r.Metadata.Labels {
			cp.Metadata.Labels[k] = v
		}
		cp.Metadata.Annotations = map[string]string{}
		for k, v := range r.Metadata.Annotations {
			cp.Metadata.Annotations[k] = v
		}
		out = append(out, cp)
	}
	return out
}
