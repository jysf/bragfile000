# Contributing to Bragfile

Bragfile is a personal-tool project — built primarily for the
author's own daily use. PRs are welcome but not actively recruited.
If you've found a bug, opened an issue, or want to suggest a small
improvement, you're in the right place.

## Development setup

Requires Go 1.26+ and `just` (optional but recommended).

```bash
git clone https://github.com/jysf/bragfile000.git
cd bragfile000
just install              # or: go install ./cmd/brag
just test                 # run the Go test suite
brag --version            # confirm install
```

## How this repo is built

This project uses a structured workflow where Claude (the AI assistant)
plays each role across separate sessions: writing specifications,
implementing them, and reviewing the result. The development process
is documented in [`docs/development.md`](docs/development.md);
[`AGENTS.md`](AGENTS.md) is the full conventions document.

If you're proposing a change, the simplest path is to:

1. Open an issue describing what you'd like to change and why.
2. Wait for confirmation that the direction makes sense.
3. Open a PR against `main`.

## Pull request conventions

- One change per PR.
- Branch naming: `feat/<slug>` for features, `fix/<slug>` for
  fixes, `chore/<slug>` for tooling/docs.
- Commit messages: short conventional-style subject
  (e.g. `feat(storage): add Entry type`); body optional.
- See [`AGENTS.md` §10](AGENTS.md) for full git/PR conventions.

## Tests

```bash
just test                 # Go test suite
just test-docs            # documentation-content assertions
gofmt -l .                # formatting check (must be empty)
go vet ./...              # static checks
```

## License

By contributing, you agree that your contributions will be
licensed under the project's [MIT License](LICENSE).
