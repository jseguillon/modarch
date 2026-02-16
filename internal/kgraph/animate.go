package kgraph

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func FindAnimation(resources []Resource, name string) (*Resource, error) {
	for _, r := range resources {
		if r.Kind == "Animation" && (name == "" || r.Metadata.Name == name) {
			return &r, nil
		}
	}
	return nil, fmt.Errorf("animation %q not found", name)
}

func FramePatchMap(resources []Resource) map[string][]string {
	out := map[string][]string{}
	for _, r := range resources {
		if r.Kind != "Frame" {
			continue
		}
		s, _ := DecodeSpec[FrameSpec](r.Raw)
		out[r.Metadata.Name] = s.Patches
	}
	return out
}

func RenderDOTToPNG(dotPath, pngPath string) error {
	cmd := exec.Command("dot", "-Tpng", dotPath, "-o", pngPath)
	return cmd.Run()
}

func BuildAPNG(pattern, out string) error {
	cmd := exec.Command("ffmpeg", "-y", "-framerate", "1", "-i", pattern, "-plays", "0", "-f", "apng", out)
	return cmd.Run()
}

func EnsureDir(path string) error {
	return os.MkdirAll(path, 0o755)
}

func WriteFile(path, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644)
}
