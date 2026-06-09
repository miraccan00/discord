# Lessons

## golangci-lint v2: verify config the way CI does, not just `run`

**Mistake:** Repeatedly claimed the `.golangci.yml` was "verified locally" using `golangci-lint run`,
but CI kept failing on `config verify`.

**Root cause:** `golangci-lint-action@v7` runs **`golangci-lint config verify`** (strict JSON-schema
validation) *before* `run`. `golangci-lint run` silently tolerates deprecated/misplaced keys; `config verify`
rejects them. So a clean `run` does **not** prove the config is valid for CI.

**Rule:** When validating a golangci-lint config locally, always run **`golangci-lint config verify`** first,
using the **exact same binary version CI installs** (here v2.12.2). Only then run `golangci-lint run`.

**v1 → v2 schema key migrations that bit me (verify lists ALL invalid keys at once — read the full list):**
- `linters.disable-all: true` → `linters.default: none`
- `revive.ignore-generated-header` → removed; use `linters.exclusions.generated`
- `issues.exclude-rules` / `issues.exclude` → `linters.exclusions.rules` (each rule may carry `linters`, `path`, `text`)
- `output.formats.colored-tab` → invalid; use `output.formats.text: { colors: true }`
- `output.show-statistics` → removed
- `gofmt`/`goimports` are **formatters** in v2, not linters → separate top-level `formatters:` block

**Also:** `go test -race` needs cgo (a C compiler). On this Windows box gcc is absent, so `-race` can't run
locally — verify test logic with plain `go test ./...` and let CI (which has gcc) run the race detector.
