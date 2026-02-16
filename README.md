# modarch

`modarch` is a Go CLI that turns Kubernetes + `kgraph.io` YAML into architecture graph outputs.

- DOT graph files (`.dot`)
- SVG graph renders (`.svg`)
- PNG graph image (`.png`, generated locally and ignored)

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
4. **Renderer outputs**
   - Emits DOT and SVG for the final architecture graph.
   - Emits PNG for local previewing.

## Test data

Test data is included in `testdata/mock.yaml` and includes:

- app stack (`Deployment`, `Service`, `ConfigMap`, `Pod`)
- external resources (`ExternalResource`)
- graph style and edge style resources
- links and notes

## Dev stack

- **Language:** Go 1.24+
- **Data format:** YAML (parsed via Ruby stdlib `YAML` bridge)
- **Testing:** `go test ./...`

## Usage

```bash
go run ./cmd/modarch --input testdata/mock.yaml --out outputs
```

Produces:

- `outputs/graph.dot`
- `outputs/graph.svg`
- `outputs/graph.png` (ignored by git)
