---
# Maps to ContextCore insight.* semantic conventions.

insight:
  id: DEC-019
  type: decision
  confidence: 0.90
  audience:
    - developer
    - agent

agent:
  id: claude-sonnet-4-6
  session_id: null

# Decisions are repo-level, but it's useful to track which project
# caused them to be emitted.
project:
  id: PROJ-002
repo:
  id: bragfile

created_at: 2026-06-10
supersedes: null
superseded_by: null

tags:
  - projects
  - cli
  - filesystem
---

# DEC-019: `brag project here` resolves cwd by nearest-ancestor (longest-prefix) matching

## Decision

`brag project here` (and the Store method `ProjectForPath` it calls) resolves
the current working directory against registered `project_locations` using
**nearest-ancestor, longest-prefix matching**:

1. Normalize both the cwd and each stored path with `filepath.Clean` before
   comparison.
2. A registered path matches the cwd if and only if:
   - `cwd == filepath.Clean(loc)` (exact match), OR
   - `strings.HasPrefix(cwd, filepath.Clean(loc) + string(filepath.Separator))`
     (loc is an ancestor directory of cwd).
   The separator guard prevents `/home/user/work` from matching
   `/home/user/worker`.
3. When multiple registered paths match (possible when different projects
   register ancestor directories of the cwd), the **longest matching path
   wins** — the most specific (deepest) registered ancestor takes
   precedence. Equal-length ties cannot arise in practice because paths are
   globally unique (`UNIQUE(path)` in `project_locations`); two distinct paths
   of the same length that are both true ancestors of the same cwd would
   have to be identical, which `UNIQUE` forbids.
4. `filepath.EvalSymlinks` is **out of scope for v0.2.0** — symlink resolution
   adds cross-platform complexity and surface area for no documented user
   pain. Call it out if/when reported.

The resolver is exposed as `func (s *Store) ProjectForPath(cwd string) (*Project, error)`.
It returns `nil, nil` — **not** an error wrapping `ErrNotFound` — when no match
exists. "Not in any project" is a normal, expected state, not an error
condition; the CLI layer converts nil to the user-facing
`"not inside any registered project"` message (stderr, exit 1).

## Context

STAGE-007 adds `brag project here`: run it from inside a registered project's
directory and learn which project you're in. The central UX question is what
"inside" means: must the user be at the exact registered path, or can they
be in any subdirectory of it?

The primary use case is "I'm deep in a project's tree; what project am I
in?" — this is precisely the question the `here` command answers. Users
rarely work at a project's root; they're in `src/`, `test/`, `docs/`, or a
dozen levels deep. A resolver that demands the exact registered path would
be nearly useless in daily use.

SPEC-031 (the `here` command) ships this resolver; SPEC-032 (`brag add`
auto-fill) reuses `ProjectForPath` directly — so the choice of resolution
policy binds two specs and must be locked, not left to per-spec inference.

## Alternatives Considered

- **Option A: Exact match only** — `cwd` must equal a registered path
  exactly.
  - *Why rejected:* fails the primary use case ("I'm deep in the tree").
    The user who registered `/home/user/projects/bragfile` must be at
    that exact path, not in `/home/user/projects/bragfile/internal/cli/`,
    which is where they actually work. This makes `here` and the SPEC-032
    auto-fill nearly useless in practice.

- **Option B (chosen): Nearest-ancestor (longest-prefix)** — registered
  path must be an ancestor of cwd; when multiple ancestors match, the
  longest wins.
  - *Why selected:* maps to the mental model users already have from git
    (`git status` works from anywhere inside a repo). The longest-prefix
    tie-break is the right semantics for nested projects (project A at
    `/work`, project B at `/work/sub` — the user in `/work/sub/deep`
    belongs to B, the more specific registration). Separator-guarded
    prefix avoids the `/work` vs `/worker` false-positive. Paths are
    globally unique (SPEC-027 schema; `UNIQUE(path)`), so the tie-break
    case ("two different projects with equal-length matching ancestor
    paths") is actually impossible — included in the algorithm statement
    for clarity only.

- **Option C: Exact match OR nearest-ancestor via a flag** — add a
  `--exact` flag to opt into strict matching.
  - *Why rejected:* YAGNI. The nearest-ancestor model handles both cases
    (an exact match is just the degenerate case of ancestor-match where
    `len(cwd) == len(loc)`). A flag adds CLI surface, testing, and
    cognitive load for no incremental value. Option B subsumes Option A.

## Consequences

- **Positive:** the `here` command and `brag add` auto-fill (SPEC-032)
  work from anywhere inside a registered project's directory tree —
  consistent with the git mental model users bring to every terminal
  session.
- **Positive:** the SPEC-027 global-path-uniqueness guarantee (`UNIQUE(path)`)
  makes the resolution deterministic: a path can match at most one
  `project_locations` row, and the longest-match tie-break is
  well-defined.
- **Positive:** `filepath.Clean` normalization handles
  `/home/user/work/./src/../src` → `/home/user/work/src` transparently,
  so users typing slightly unclean paths still get a match.
- **Negative:** symlinks are not resolved. A user who has registered
  `/home/user/real-path` but is working under a symlinked alias
  `/home/user/alias → /home/user/real-path` will not get a match.
  Accepted for v0.2.0; add `filepath.EvalSymlinks` if reported.
- **Negative:** the stored path must be the canonical `filepath.Clean`
  form (no trailing slash, no `.` segments) to match correctly. SPEC-028's
  `brag project new --path` stores paths verbatim; `ProjectForPath`
  calls `filepath.Clean` on both sides, so minor untidiness at
  registration time is tolerated.

## Validation

Right if:
- `ProjectForPath("/home/user/work/src/deep")` returns the project
  registered at `/home/user/work` (nearest-ancestor test).
- `ProjectForPath("/home/user/worker")` returns nil when
  `/home/user/work` is registered (separator-guard test).
- `ProjectForPath("/a/b/sub/deep")` returns the project registered at
  `/a/b/sub` over the one at `/a/b` (longest-prefix test).
- `ProjectForPath("/unregistered/path")` returns nil, nil — no error.

Revisit if:
- Symlink-related misses are reported by real users → add
  `filepath.EvalSymlinks` on both sides, emit a new DEC superseding
  this one.
- A use case requires exact-match semantics (e.g., multiple closely
  nested projects where ancestor-match misfires) → add `--exact` flag
  or reconsider Option A/C.

## References

- Related specs: **SPEC-031** (emits this DEC; the `here` command +
  `ProjectForPath` Store method), **SPEC-032** (reuses `ProjectForPath`
  for `brag add --project` auto-fill).
- Related decisions: DEC-017 (`entries.project` soft match; the
  `project_locations.path` global-uniqueness guarantee that makes
  longest-prefix tie-break well-defined), DEC-002 (forward-only
  migrations — the schema this resolver queries).
- Related docs: `docs/data-model.md` (`project_locations` table; the
  `UNIQUE(path)` guarantee).
