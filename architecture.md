# Architecture

## System flow

1. **Parse bundle** (`internal/model`)
   - Read multi-document YAML.
   - Normalize resources, edge styles, and frames into typed models.
2. **Build base graph** (`internal/engine`)
   - Create nodes from parsed resources.
   - Derive edges from inferred + explicit relationships.
   - Apply edge styles by edge type.
3. **Apply animation frame patches** (`internal/engine`)
   - Mutate target node annotations for fold/collapse behavior.
   - Mutate target edge attributes for visual overrides.
4. **Render** (`internal/render`)
   - DOT export for graph tool interoperability.
   - SVG export for web-friendly frame visualization.
   - PNG export for image frame pipeline.
5. **Compose animation** (`internal/anim`)
   - Collect frame PNG images.
   - Encode into APNG timeline.

## Package layout

- `cmd/modarch`: CLI entrypoint and orchestration.
- `internal/model`: parsing and canonical models.
- `internal/engine`: graph derivation and frame patch execution.
- `internal/render`: DOT/SVG/PNG renderers.
- `internal/anim`: APNG encoding.
- `testdata`: sample YAML bundle.

## Feature mapping

- **Feature A: YAML architecture schema**
  - Implemented in parser models and discovery logic.
- **Feature B: Graph + style extraction**
  - Implemented in base graph generation and edge style merge.
- **Feature C: Animation frames**
  - Implemented with Frame patches and graph filtering.
- **Feature D: Deliverables**
  - DOT + SVG + PNG + APNG produced per run and per iteration.
