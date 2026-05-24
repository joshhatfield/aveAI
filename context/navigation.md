# aveAI Context Index

## Domain
- [product-vision.md](core/domain/product-vision.md) — What aveAI is, why it exists, core principles

## Processes
- [architecture.md](core/processes/architecture.md) — System design, data flow, CLI lifecycle
- [data-model.md](core/processes/data-model.md) — Sort-key model, `.avdb` format, map.yaml schema
- [search-strategy.md](core/processes/search-strategy.md) — Inverted index, vector search, fallback logic

## Standards
- [code-quality.md](core/standards/code-quality.md) — Go conventions, project structure, testing patterns

## Plans
- [phase-1-foundation.md](plans/phase-1-foundation.md) — CLI scaffold, `.avdb` persistence, CRUD, map.yaml
- [phase-2-text-search.md](plans/phase-2-text-search.md) — Inverted index, ranking, query language
- [phase-3-vector-search.md](plans/phase-3-vector-search.md) — Embedding integration, similarity search

---

**Conventions:**
- Files use lowercase-kebab-case
- `map.yaml` at root of `.avdb` databases defines sort-key schema
- Phase plans are sequential — later phases depend on earlier ones
