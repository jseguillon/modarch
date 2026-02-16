# Architecture

## 1) Pipeline

1. **Load** YAML docs (`LoadResources`).
2. **Transform** by applying zero or more `PatchSet` files (`ApplyPatchFiles`).
3. **Build graph** (`BuildGraph`):
   - collect styles
   - create nodes
   - create derived and explicit edges
   - apply edge styles
4. **Render** pure DOT (`RenderDOT`).
5. **Animate** (`animate` command): resolve `Animation` + `Frame` resources and apply frame patches cumulatively.

## 2) Data model

- `Resource`: generic normalized object (`apiVersion`, `kind`, metadata, raw map).
- `Graph`: `Nodes` map + `Edges` list.
- `Selector`: shared matching logic for style and patch operations.
- `PatchSet`: transform operations (`remove`, `annotate`, `label`).

## 3) Styling precedence

1. Built-in node defaults.
2. Matching `GraphStyle` resources.
3. Per-resource annotations (highest).

Edge styles are merged by matching `EdgeStyle` resources.

## 4) Animation semantics

For frame index `N`:

- start from base resources
- apply base patches
- apply frame patches for all frames `0..N`
- render frame DOT + SVG (+ optional PNG)

This guarantees deterministic, cumulative progression.

## 5) Extensibility points

- Add selector operators (`in`, `notin`, regex).
- Add patch actions (`set`, JSON6902).
- Add graph rules (Ingress, Secret, PVC/PV).
- Add output adapters while keeping DOT as canonical model.
