package kgraph

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func LoadResources(input string) ([]Resource, error) {
	info, err := os.Stat(input)
	if err != nil {
		return nil, err
	}
	var files []string
	if info.IsDir() {
		err = filepath.WalkDir(input, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			if strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml") {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	} else {
		files = []string{input}
	}

	var out []Resource
	for _, f := range files {
		docs, err := readYAMLDocsAsMaps(f)
		if err != nil {
			return nil, err
		}
		for _, raw := range docs {
			if len(raw) == 0 {
				continue
			}
			res := Resource{Raw: raw}
			if v, ok := raw["apiVersion"].(string); ok {
				res.APIVersion = v
			}
			if v, ok := raw["kind"].(string); ok {
				res.Kind = v
			}
			if m, ok := raw["metadata"].(map[string]any); ok {
				res.Metadata.Name, _ = m["name"].(string)
				res.Metadata.Namespace, _ = m["namespace"].(string)
				res.Metadata.Labels = toStringMap(m["labels"])
				res.Metadata.Annotations = toStringMap(m["annotations"])
			}
			if res.Metadata.Labels == nil {
				res.Metadata.Labels = map[string]string{}
			}
			if res.Metadata.Annotations == nil {
				res.Metadata.Annotations = map[string]string{}
			}
			out = append(out, res)
		}
	}
	return out, nil
}

func readYAMLDocsAsMaps(path string) ([]map[string]any, error) {
	cmd := exec.Command("ruby", "-ryaml", "-rjson", "-e", `docs = YAML.load_stream(File.read(ARGV[0]))
puts JSON.generate(docs.compact)
`, path)
	b, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("parse yaml %s: %w", path, err)
	}
	var docs []map[string]any
	if err := json.Unmarshal(b, &docs); err != nil {
		return nil, err
	}
	return docs, nil
}

func toStringMap(v any) map[string]string {
	m, ok := v.(map[string]any)
	if !ok {
		return map[string]string{}
	}
	out := map[string]string{}
	for k, v := range m {
		out[k] = fmt.Sprintf("%v", v)
	}
	return out
}

func DecodeSpec[T any](raw map[string]any) (T, error) {
	var zero T
	spec, ok := raw["spec"]
	if !ok {
		return zero, nil
	}
	b, err := json.Marshal(spec)
	if err != nil {
		return zero, err
	}
	if err := json.Unmarshal(b, &zero); err != nil {
		return zero, err
	}
	return zero, nil
}
