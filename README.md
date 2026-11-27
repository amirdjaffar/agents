# agents

Builds the rules section of a repo's `AGENTS.md` from parts you choose.

## Usage

```
go run . path/to/repo
```

The repo defaults to the current directory. Every part is listed with its state, and the panel below shows what the focused part writes.

| Key | Does |
| --- | --- |
| ↑↓, tab | Move, jump to the next group |
| ←→, space | Cycle the state |
| `r`, `o`, `u` | Set rule, opposite or unset |
| enter | Cycle the state of the focused part |
| `w` | Write the section |
| `q` | Quit without writing |

The section goes between markers in the repo's `AGENTS.md`, which is created if needed, and a `CLAUDE.md` containing `@AGENTS.md` is added. Text outside the markers belongs to the repo: a one-line description, layout, commands, invariants, and the changes it allows as autonomous. The opening marker records the choices, so the next run loads them.

## Parts

Each part has a rule and its opposite. A repo states either, or leaves the part unset and says nothing. Supervision and Planning pick one part, or none.

| Group | Parts |
| --- | --- |
| Supervision | paired, reviewed, delegated, autonomous |
| Code | understanding, simple, speculation, comments, tests, generated |
| Modularity | concern, files, functions, layers, duplication |
| Libraries | stdlib, ask, pinned, maintained |
| Writing | concise, docs-with-code |
| Git | atomic, squash, subject, conventional, linear, signed, attribution |
| Controls | scope, ask, checks, destructive, secrets, network, summary |
| Planning | none, light, milestones, kanban, rfc |

Supervision sets how much an agent may do before a human checks its work:

| | Human involvement | Agent may |
| --- | --- | --- |
| Paired | Follows every step | Edit one step at a time |
| Reviewed | Reviews each change before commit | Edit freely; commit and push when asked |
| Delegated | Reviews the PR | Commit and push its own branch, open PRs |
| Autonomous | Reviews after the fact | Merge its own PRs for changes the repo lists |

A repo can put specific paths under stricter supervision in its own text, e.g. security-critical code as paired.

## Layout

| File | Contents |
| --- | --- |
| `catalog.yaml` | Every group and part |
| `catalog.go` | Loading and validating the catalog |
| `section.go` | Selections, the marker, rendering and splicing the section |
| `tui.go` | The interactive list |
| `main.go` | Reading and writing the repo's files |

To add a part, give it an `id`, `summary`, `rule`, `opposite` and optionally a `default` in `catalog.yaml`, then run `go test ./...`.

## Why

- **`AGENTS.md`** is the file most coding agents read. `CLAUDE.md` imports it, so Claude Code reads the same rules.
- **Copied, not linked:** an agent only sees the repo it works in.
- **Rule and opposite:** an unstated rule leaves the agent to guess. Stating the opposite makes the choice explicit.
- **Short:** agents follow a short file more reliably than a long one.
- **Human-approved commit messages:** approving the message forces someone to understand the change before it lands.
- **No AI attribution:** it says nothing about the change. The author is whoever approved it.

## License

MIT
