# Research prompt — adjacent data developers track in random files

*Created 2026-06-16. A self-contained prompt to hand to a fresh research
session (plain, or under the `deep-research` skill for a cited report).
Goal: surface accomplishment-adjacent data — beyond projects, including
non-project / career-relational things (mentoring, kudos, network) — that
developers and other CLI-type knowledge workers keep in scattered ad-hoc
files because no tool serves it well locally. Findings feed the backlog's
"Adjacent data — is bragfile a personal work-log substrate?" item.*

---

You are researching ADJACENT DATA that developers and other CLI-oriented knowledge workers
track in scattered, ad-hoc files — to inform a local-first tool's roadmap.

## Background (context for aim — don't get anchored on it)

bragfile is a local-first Go CLI that captures "brags" — career-worthy accomplishments — into a
local SQLite DB (~/.bragfile), no cloud. Each entry has: title, description, tags, project, type,
impact, timestamps. It has a projects model (named projects with filesystem locations, cwd-aware
auto-fill) and a polymorphic tagging schema, so adding *new kinds of capture* is cheap. Its
philosophy: rule-based core, AI via piping (no built-in LLM), and a strong preference for PASSIVE
capture — the best data is captured as a byproduct of work (e.g. by an AI coding agent at the end
of a session) rather than by the user remembering to run a command. Its emerging thesis is
"agent-native accomplishment memory" feeding impact analysis and storytelling (turning records into
promo packets, reviews, blog posts, a mirror on one's trajectory).

## The research question

Beyond accomplishments and projects, **what other categories of work/career data do developers and
similar CLI-type knowledge workers repeatedly track in random, ad-hoc files** (notes.txt, TODO.md,
a `did.txt`, Apple Notes, a gist, an Obsidian vault, a spreadsheet) — *because no specialized tool
serves it well locally* — that would be valuable, low-friction, and a natural fit to live in a local
SQLite + CLI alongside brags?

Cast wide on personas and on NON-project, career/relational data — not just coding artifacts:
- IC software engineer; staff+/tech lead; engineering manager; SRE/on-call; **sales engineer /
  solutions engineer**; consultant/freelancer; researcher/data scientist.
- Explicitly include things that aren't project-scoped: mentoring, 1:1s, relationships/network,
  kudos received, learning, reading, ideas, decisions, "waiting on / blocked by," metrics/numbers.

## What to find for each candidate category

For every category you surface (expand well beyond the seeds below), report:
1. **What it is** + a concrete example of the data.
2. **Who tracks it** (which persona(s)) and **how often / what triggers** a capture.
3. **Where it lives today** — the actual random file / tool people use, with at least one real,
   citable source (a blog post, HN/Reddit thread, a public dotfiles or TIL repo, a PKM workflow
   write-up). Flag anything you can only assert without a source as speculative.
4. **Why it's painful / underserved** — why a random file, not a dedicated tool.
5. **Capture friction & structure** — could it be captured passively (by an agent, a hook, from
   git) or does it need human input? What minimal structure does it need?
6. **Fit for local-SQLite-CLI + agent capture** — is this a good fit or a stretch? Would it be a new
   entry "type," a new taggable object, or its own thing?
7. **Relation to accomplishment / impact / storytelling** — does it strengthen the brag/impact/story
   loop (e.g. metrics quantify impact; kudos are promo evidence; decisions explain the "why")?

## Seed categories (expand, validate, and add more — these are not exhaustive)

Decisions/ADRs · learnings/TILs · kudos & feedback received · metrics/before-after numbers ·
blockers & "waiting on" · open questions/follow-ups · goals/OKRs · people & relationships (mentoring,
1:1 notes, network/contacts, who-helped-whom) · time/effort signals · reading/references · code
snippets · incidents/postmortems · meeting notes · ideas/someday · wins & losses · interview/hiring
notes · "brag-document"-style work journals.

## Sources worth mining (not exhaustive)

The "brag document" discourse (Julia Evans et al.); "work journal" / "done list" / "did.txt" tools
and writeups; public TIL repos (e.g. jbranchaud/til) and "learning in public"; dotfiles repos and
plain-text-life / PKM (Obsidian, Logseq) dev workflows; HN/Reddit threads like "how do you track
your accomplishments / decisions / what you learned"; engineering-manager resources on 1:1 and
growth tracking; sales-engineering / solutions-engineering enablement and note-taking practices;
ADR tooling communities.

## Output

1. A ranked shortlist of the strongest candidate categories for bragfile, each with the 7 fields
   above, a one-line **passive-capturable? (Y/N/partial)** flag, and a **feeds impact/storytelling?**
   flag.
2. A short section on the NON-project, career/relational categories specifically (mentoring, kudos,
   network) — these are the ones the requester suspects are missing.
3. A closing recommendation: which 2–3 are the highest-leverage adjacencies and why, and which are
   traps (well-served by existing tools, or too high-friction to capture).

## Quality bar

Prefer evidence over speculation — ground categories in real artifacts people actually keep. Be
honest about which ideas are well-served by existing tools (calendars, task managers, CRMs) and
therefore NOT worth duplicating. The valuable findings are the underserved, scattered-in-random-files
data that a local CLI + passive agent capture is uniquely good at.
