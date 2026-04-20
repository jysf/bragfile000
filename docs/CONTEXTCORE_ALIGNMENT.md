# ContextCore Alignment

This template is philosophically aligned with [ContextCore](https://github.com/neil-the-nowledgeable/contextcore)
but requires no ContextCore infrastructure. You run on markdown files today;
the vocabulary is chosen so you can migrate cleanly if you ever want to.

## What we borrow from ContextCore

### Semantic conventions (used as YAML front-matter keys)

| Namespace | Used in | Fields we use |
|---|---|---|
| `task.*` | `/projects/*/specs/SPEC-*.md` | `task.id`, `task.cycle`, `task.type`, `task.priority`, `task.complexity`, `task.blocked` |
| `project.*` | specs, stages | `project.id`, `project.stage` |
| `repo.*` | specs, stages, `.repo-context.yaml` | `repo.id` |
| `business.*` | `.repo-context.yaml` | `business.criticality`, `business.owner`, `business.contacts` |
| `insight.*` | `/decisions/DEC-*.md` | `insight.id`, `insight.type`, `insight.confidence`, `insight.audience` |
| `guidance.*` | `/guidance/constraints.yaml`, `questions.yaml` | structured constraint and question IDs |
| `agents.*` | `/projects/*/specs/SPEC-*.md` | `agents.architect`, `agents.implementer` — roles Claude plays in different sessions |

### Artifact model

- **Decisions as first-class, queryable, supersedable records** — each
  DEC-\* is a standalone file with stable ID, confidence, audience,
  and optional `supersedes:` pointer. No decision is ever deleted;
  supersession creates a new DEC.
- **Constraints as structured rules** — `guidance/constraints.yaml`
  gives each rule a stable ID, severity, path glob, and rationale.
  This mirrors ContextCore's `agentGuidance.constraints` CRD shape.
- **Tasks with a lifecycle** — specs have a `cycle` field tracked
  through Frame → Design → Build → Verify → Ship. ContextCore's
  `TaskTracker` tracks the same thing as OTel spans.
- **Handoffs as first-class objects** — in the claude-plus-agents
  variant, every architect-to-implementer delegation is a file with
  stable ID, from/to agents, and completion tracking. This maps to
  ContextCore's `HandoffManager`.

## What we deliberately don't do

- No actual OTel instrumentation. Nothing emits spans.
- No queryable observability (Tempo / TraceQL / PromQL).
- No Grafana dashboards.
- No Kubernetes CRDs.
- No MCP server wrappers.

You get 80% of the value from the vocabulary and artifact model alone,
for 0% of the infrastructure weight.

## If you want to graduate to full ContextCore

Each file in this template parses cleanly into OTel spans:

| Our file | ContextCore span |
|---|---|
| `decisions/DEC-*.md` | `InsightEmitter.emit_decision()` span |
| `projects/*/specs/SPEC-*.md` | `TaskTracker` span (task.* attributes) |
| (no handoffs in this variant — Implementation Context lives inside the spec) | — |
| `guidance/constraints.yaml` | `ProjectContext` CRD `agentGuidance.constraints` |
| `guidance/questions.yaml` | `ProjectContext` CRD `agentGuidance.questions` |
| `.repo-context.yaml` | `ProjectContext` CRD metadata |

A one-time Python or TypeScript script can walk the repo and emit
these as spans. Cost: a few hours, maybe less. The structural work is
already done.

## When to consider graduating

Triggers that mean the files approach is creaking:
- `/decisions/` has >30 entries and grepping feels slow
- You want real-time dashboards for project or portfolio health
- Compliance requires audit trails
- Multiple people or agents need to query state simultaneously
- You're already running the Grafana/Tempo/Loki stack for another reason

Until one of these is true, don't graduate. Each piece of infrastructure
you add is weight.
