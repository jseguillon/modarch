# modarch

`modarch` is a Go CLI that turns Kubernetes + `kgraph.io` YAML into architecture graph outputs:

- DOT graph files (`.dot`)
- SVG frame renders (`.svg`)
- PNG frame images (`.png`)
- Animated PNG timeline (`.apng`)

## Features

1. **Schema discovery from YAML bundles**
   - Parses multi-document YAML with Kubernetes resources and custom `kgraph.io` resources.
   - Detects architecture nodes from Deployments, Services, Pods, ConfigMaps, ExternalResource, Links, and Notes.
2. **Graph extraction engine**
   - Infers `envFrom` edges (Deployment -> ConfigMap).
   - Infers Service selector edges to Deployments/Pods.
   - Builds explicit link edges from `Link` objects.
   - Builds note attachment edges from `Note` objects.
3. **Style system**
   - Applies `EdgeStyle` resources by edge type.
4. **Frame patching for animation**
   - Applies frame patches to fold resources.
   - Supports edge-level patch overrides for color/label.
5. **Renderer outputs**
   - Emits DOT and SVG for each frame.
   - Emits PNG for each frame.
   - Composes APNG from the frame PNG files.

## Test data

Test data is included in `testdata/mock.yaml` and mirrors your supplied scenario, including:

- app stack (`Deployment`, `Service`, `ConfigMap`, `Pod`)
- external resources (`ExternalResource`)
- graph style and edge style resources
- links and notes
- 4 animation `Frame`s and a final `Animation` resource

## Dev stack

- **Language:** Go 1.24+
- **Data format:** YAML (parsed via Ruby stdlib `YAML` bridge)
- **Animation encoding:** custom APNG chunk encoder in Go
- **Testing:** `go test ./...`

## Usage

```bash
go run ./cmd/modarch --input testdata/mock.yaml --out outputs --iteration 4
```

Progressive animation exports:

```bash
go run ./cmd/modarch --input testdata/mock.yaml --out outputs/iter1 --iteration 1
go run ./cmd/modarch --input testdata/mock.yaml --out outputs/iter2 --iteration 2
go run ./cmd/modarch --input testdata/mock.yaml --out outputs/iter3 --iteration 3
go run ./cmd/modarch --input testdata/mock.yaml --out outputs/iter4 --iteration 4
```

Each output folder includes frame DOT/SVG/PNG files and `iteration-N.apng`.
