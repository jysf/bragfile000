---
description: Draft a brag entry from this session
---

Review what shipped in this session. If a moment is brag-worthy per
[BRAG.md](https://github.com/jysf/bragfile000/blob/main/BRAG.md)
(shipped feature, fixed significant bug, architectural decision,
delivered artifact), draft a single JSON object validating against
[`docs/brag-entry.schema.json`](https://github.com/jysf/bragfile000/blob/main/docs/brag-entry.schema.json):
required `title` (action-verb, ≤100 chars), plus optional
`description`, `project`, `type`, `tags` (comma-joined string per
DEC-004), and `impact` (concrete metric or named outcome). Present
the JSON for my approval. Do not execute `brag add --json` until I
confirm.
