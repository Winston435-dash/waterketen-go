# Agent instructions — waterketen-go

Go rewrite of the Rotterdam waterketen TUI/CLI (Bubble Tea + Bubbles + Lip
Gloss). Reads the same managed store as the Python CLI:
`~/.local/share/waterketen`. Feature parity with the Python tool is the goal;
`fetch`/`bibliotheek`/`kaart`-export still live in Python (`MIGRATION.md`).

## Skills to load
- `waterketen-cli` — domain + this repo's architecture (data-ink ownership).
- `bubbletea-tui-development` — read `go.mod` first: the Charm v1/v2 split.
- `terminal-ui-preview` — verify TUI changes with real captures, not mocks.

## Checkpoint before new work (standing convention)
Before starting a new change here, if the tree is dirty, checkpoint it first:

```bash
git wip "<topic>"     # local alias: stages everything into a `wip:` commit
```

It is a no-op when the tree is clean, so it is safe to call before every task.
Build the new work on top of that commit; keep `wip:` messages for throwaway
states and give the real change a conventional-commit message before pushing.
Never carry a whole session's unreviewed work uncommitted into the next one.

## Verify (the authority — not the editor lint)
```bash
gofmt -l . && go build ./... && go vet ./... && go test -count=1 ./...
```
The editor's post-write lint on a Go module falsely reports `undefined: X`
(it resolves imports outside the module) — ignore it, trust the commands above.
For TUI changes, diff real captures (`tmux capture-pane -e` before/after): after
a pure refactor the map must stay **byte-identical**. Preview files go under
`~/.hermes/cache/scratch/`, never in the repo.

## Contracts (enforced by tests, see README.md)
Screens define no colours/borders/padding (they use `ui/components`);
`internal/core` imports no lipgloss/bubbletea; atlas data ink is core's
(`core.CellColour`) and `components.RenderAtlas` is the only core→lipgloss
point. `internal/ui/screens/contract_test.go` fails if a screen imports lipgloss.
