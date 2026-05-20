---
title: Why bragfile
tagline: The gap between doing good work and being able to talk about it later.
target: TBD
status: draft
---

# Why bragfile

Every engineer knows the quarterly scramble. Review season arrives, you
open a blank self-evaluation, and you try to reconstruct six months of
work from memory, git history, and Slack archaeology. The work was real.
The evidence is scattered. The memory has already decayed.

The problem was never that you didn't do good work. It's the gap between
*doing* the work and being able to *articulate* it later. Shipping a
feature takes days. Remembering, months later, that you shipped it — in
enough detail to write one good sentence about its impact — takes a
system. Most of us don't have one.

## Memory decays in exactly the wrong order

The routine work blurs together. The hard-won stuff — the gnarly
incident you debugged at 11pm, the migration you de-risked, the junior
engineer you unblocked three times in one week — those fade first,
because by the time anyone asks, you've moved on to the next fire.

And someone always asks. A manager in a 1:1: "what are you proud of this
quarter?" A promotion packet that wants concrete evidence, not vibes. A
resume update where "worked on backend systems" is doing far too much
work. A retro where you genuinely cannot recall what the first half of
the sprint contained.

In every one of those moments, the bottleneck is the same: you did the
work, and you can't get to it.

## Why the obvious fixes don't stick

- **A spreadsheet.** Works for about a week. Capturing a row means
  switching apps, finding the file, picking a cell. The friction wins
  and you quietly stop.
- **A notes app.** No structure, so it becomes a junk drawer. You can't
  filter it, can't summarize it, can't trust it's complete.
- **"I'll remember."** You won't. This is the default, and it's the
  reason the scramble exists.
- **Do it at review time.** Too late — the specifics are already gone.
  You're reconstructing, not recalling.

The common failure is friction. If logging an accomplishment costs more
than a few seconds, it competes with real work, and real work wins every
time. Any capture tool that isn't nearly free to use will be abandoned.

## What I actually wanted

Four things:

1. **Fast enough to use in the moment.** Capture in under ten seconds,
   from a terminal that's already open, without breaking focus.
2. **Local and private.** Career reflection is personal. It shouldn't
   live in someone else's cloud, behind someone else's login, subject to
   someone else's data policy.
3. **Structured enough to query later.** Not a wall of text — entries
   with a title, a description, tags, a project, an impact line. Enough
   shape to filter, search, and summarize.
4. **Mine.** No account, no sync, no vendor who can deprecate it.

## What bragfile is

A command-line tool. You capture an entry the moment it happens:

```bash
brag add --title "shipped the orders-service migration, cut p99 40%"
```

That's the whole interaction. The entry lands in a SQLite database in
your home directory — `~/.bragfile/db.sqlite` — and it's yours. No
cloud, no sync, no account.

Later, when someone asks, the work is right there:

```bash
brag list --since 90d
brag search "migration"
brag summary --range month
brag review --week
```

`brag review` is the one that closes the loop: it pulls your recent
entries, groups them, and hands you a digest you can read straight into
a self-eval or pipe into an AI session for a first-draft narrative. The
scramble becomes a query.

## The name

It's called `brag` because the entries are brag-worthy moments — the
things you'd actually mention if someone asked what you're proud of.
Not a journal, not a task tracker, not a humblebrag generator. A log of
the moments worth remembering, captured while you still remember them.

## An honest note

bragfile is a personal project. I built it because I wanted it, and I
shipped it properly — Homebrew install, a real test suite, a security
review, the works — mostly for the learning value of doing distribution
correctly end to end.

If the quarterly scramble is a problem you recognize, you might find
bragfile useful too. And if you'd rather use something else, the idea
still travels: any low-friction, structured capture habit beats relying
on memory. The tool matters less than closing the gap. bragfile is just
how I closed mine.
