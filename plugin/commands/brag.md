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
`model:claude-opus-4-8`).

If the work has a proving artifact, add an evidence link as a tag so the claim
is checkable by anyone with the repo. Prefer `pr:<number>`; otherwise
`commit:<hash on the default branch>`. Do not record the branch hash you have
while working — a squash-merge replaces it, so it resolves for nobody
afterwards. Never a range. If you cannot confirm a ref, record none: a hash
that does not resolve reads as evidence and is not.

Present the JSON for my approval. Do not execute
`brag add --json` (or the brag_add MCP tool) until I confirm.
