# modarch / kgraph

A Go-based toolchain that reads Kubernetes YAML and `kgraph.io` config resources, builds a graph model, renders pure DOT (with HTML labels), and can produce per-frame PNG plus animated PNG from cumulative patch-based animation frames.

## Implemented features

- **YAML inventory loader**
  - Reads a file or directory of `*.yaml` / `*.yml` documents.
  - Accepts both native Kubernetes resources and tool-only resources (`GraphStyle`, `EdgeStyle`, `Link`, `Note`, `Frame`, `Animation`, `PatchSet`).

- **Graph derivation**
  - Nodes for resources.
  - Derived edges:
    - `Service -> Deployment|Pod` by selector matching.
    - `Deployment -> ConfigMap` for `envFrom.configMapRef`.
  - Explicit edges via `Link` resource.
  - Notes as nodes + optional `note` edge to target.

- **Styling via resources + annotations**
  - `GraphStyle` selectors by kind(s), name/namespace and simple `labelSelector` (`key=value`) modify node attrs and label template.
  - `EdgeStyle` selectors by edge `type` and destination namespaces.
  - Annotation overrides:
    - `kgraph.io/ignore=true` hides a resource.
    - `kgraph.io/node.attr.<attr>=<value>` overrides DOT node attrs.
    - `kgraph.io/node.labelTemplate=<go-template>` overrides label rendering.

- **Patch support (filter + mutate)**
  - `PatchSet` (`kgraph.io/v1`) with operations:
    - `remove`: filter resources out.
    - `annotate`: mutate metadata annotations.
    - `label`: mutate metadata labels.

- **Animation support**
  - `Animation` object with base patches + frame refs.
  - `Frame` objects reference patch files.
  - Per-frame patches are applied **cumulatively**.
  - Emits `frame-XXX.dot` and `frame-XXX.svg` for every frame, tries `frame-XXX.png` (Graphviz `dot`), and optional APNG (`ffmpeg`).

- **Template-first DOT labels (HTML-like)**
  - Supports HTML DOT labels (`<TABLE>...</TABLE>`) through `GraphStyle.node.labelTemplate` and annotation override template.

## CLI

```bash
kgraph render  -i <yaml-file|dir> [-p patch.yaml] -o out/graph.dot [-svg out/graph.svg] [-png out/graph.png]
kgraph animate -i <yaml-file|dir> [-name animationName] -out out/frames [-apng out/example.png]
```

## Test data

- Main fixture: `examples/testdata/mock.yaml`
- Frame patch sets: `patches/frame-001.yaml` ... `patches/frame-004.yaml`


## Test artifacts

- `go test ./...` writes SVG snapshots for each test case to `test_artifacts/tests/` for quick visual inspection.
- These SVG snapshots are intended to be committed so reviewers can inspect graph changes in PRs.

## Dev stack

- **Go 1.22** (module-based).
- **Graphviz `dot`** optional for PNG rendering.
- **ffmpeg** optional for APNG assembly.
- **Docker Compose** includes a Go dev container and a Graphviz container.

## Local workflow

```bash
make build
make test
make run-render
make run-animate
```

Outputs go to `out/`.
