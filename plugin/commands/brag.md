---
description: Draft a brag entry from this session
---

Review what shipped in this session. If a moment is brag-worthy per BRAG.md
(shipped feature, fixed significant bug, architectural decision, delivered
artifact), draft a single JSON object validating against
docs/brag-entry.schema.json: required `title` (action-verb, <=100 chars),
plus optional `description`, `project`, `type`, `tags` (comma-joined string
per DEC-004), and `impact` (concrete metric or named outcome). When the work
was agent-driven, include provenance as reserved tags `agent:<name>` and
`model:<id>` (lowercase, no spaces; e.g. `agent:claude-code`,
`model:claude-opus-4-8`). Present the JSON for my approval. Do not execute
`brag add --json` (or the brag_add MCP tool) until I confirm.
