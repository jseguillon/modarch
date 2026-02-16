package model

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

func ParseBundle(path string) (*Bundle, error) {
	cmd := exec.Command("ruby", "-ryaml", "-rjson", "-e", "puts JSON.generate(YAML.load_stream(File.read(ARGV[0])))", path)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ruby yaml decode failed: %w", err)
	}
	var docs []map[string]any
	if err := json.Unmarshal(out, &docs); err != nil {
		return nil, err
	}
	bundle := &Bundle{}
	for _, doc := range docs {
		if len(doc) == 0 {
			continue
		}
		kind := asString(doc["kind"])
		switch kind {
		case "Frame":
			f, err := parseFrame(doc)
			if err != nil {
				return nil, err
			}
			bundle.Frames = append(bundle.Frames, f)
		case "EdgeStyle":
			bundle.EdgeStyles = append(bundle.EdgeStyles, parseEdgeStyle(doc))
		case "Animation", "GraphStyle":
			continue
		default:
			bundle.Resources = append(bundle.Resources, parseResource(doc))
		}
	}
	return bundle, nil
}

func parseResource(doc map[string]any) Resource {
	meta, _ := doc["metadata"].(map[string]any)
	spec, _ := doc["spec"].(map[string]any)
	data, _ := doc["data"].(map[string]any)
	return Resource{APIVersion: asString(doc["apiVersion"]), Kind: asString(doc["kind"]), Namespace: asString(meta["namespace"]), Name: asString(meta["name"]), Labels: asStringMap(meta["labels"]), Annotations: asStringMap(meta["annotations"]), Spec: spec, Data: data, Raw: doc}
}

func parseFrame(doc map[string]any) (Frame, error) {
	meta, _ := doc["metadata"].(map[string]any)
	spec, _ := doc["spec"].(map[string]any)
	patchesAny, _ := spec["patches"].([]any)
	f := Frame{Name: asString(meta["name"])}
	for _, p := range patchesAny {
		raw, ok := p.(map[string]any)
		if !ok {
			return Frame{}, fmt.Errorf("invalid frame patch in %s", f.Name)
		}
		patch := FramePatch{Target: mapFromAny(raw["target"])}
		opsAny, _ := raw["ops"].([]any)
		for _, opAny := range opsAny {
			opMap := mapFromAny(opAny)
			patch.Ops = append(patch.Ops, PatchOp{Op: asString(opMap["op"]), Path: asString(opMap["path"]), Value: opMap["value"]})
		}
		f.Patches = append(f.Patches, patch)
	}
	return f, nil
}

func parseEdgeStyle(doc map[string]any) EdgeStyle {
	spec := mapFromAny(doc["spec"])
	sel := mapFromAny(spec["selector"])
	edge := mapFromAny(spec["edge"])
	return EdgeStyle{Type: asString(sel["type"]), Attrs: asStringMap(edge["attrs"])}
}

func asString(v any) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

func asStringMap(v any) map[string]string {
	out := map[string]string{}
	raw, ok := v.(map[string]any)
	if !ok {
		return out
	}
	for k, vv := range raw {
		out[k] = fmt.Sprintf("%v", vv)
	}
	return out
}

func mapFromAny(v any) map[string]any {
	m, _ := v.(map[string]any)
	if m == nil {
		return map[string]any{}
	}
	return m
}
