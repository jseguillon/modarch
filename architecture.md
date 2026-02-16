# Architecture

## System flow

1. **Parse bundle** (`internal/model`)
   - Read multi-document YAML.
   - Normalize resources and edge styles into typed models.
2. **Build graph** (`internal/engine`)
   - Create nodes from parsed resources.
   - Derive edges from inferred + explicit relationships.
   - Apply edge styles by edge type.
3. **Render** (`internal/render`)
   - DOT export for graph tool interoperability.
   - SVG export for web-friendly visualization.
   - PNG export for image preview pipeline.

## Package layout

- `cmd/modarch`: CLI entrypoint and orchestration.
- `internal/model`: parsing and canonical models.
- `internal/engine`: graph derivation.
- `internal/render`: DOT/SVG/PNG renderers.
- `testdata`: sample YAML bundle.

## Feature mapping

- **Feature A: YAML architecture schema**
  - Implemented in parser models and discovery logic.
- **Feature B: Graph + style extraction**
  - Implemented in base graph generation and edge style merge.
- **Feature C: Deliverables**
  - DOT + SVG + PNG produced per run.
