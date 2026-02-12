# Coding Agent Instructions

## Go Version & Tooling

- Go 1.24+ with `gofumpt` formatting.
- Lint with `golangci-lint` v2. Enabled linters: `bodyclose`, `godox`, `nakedret`, `predeclared`, `unconvert`.
- Do not use naked returns, predeclared identifier shadowing, or unnecessary type conversions.

## Style

- Every file starts with the copyright header:

```go
// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.
```

- Every exported symbol has a doc comment starting with its name: `// Run executes the command.`
- Every package has a doc comment: `// Package foo does bar.`
- Group imports: stdlib, then third-party, then internal (`github.com/microsoft/go-infra/...`), separated by blank lines.
- Use `context.Context` as the first parameter when a function needs one.
- Prefer closures for short callbacks instead of defining new types.

## Error Handling

- Wrap errors with `fmt.Errorf("context: %w", err)` to preserve the error chain.
- Check errors immediately: `if err != nil { return ..., fmt.Errorf("failed to ...: %w", err) }`.
- Use `errors.Join` when combining multiple independent errors.
- Never panic in non-test code.

## Testing

- Use table-driven tests with `t.Run` subtests:

```go
tests := []struct {
    name string
    // fields
}{
    // cases
}
for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        // assert
    })
}
```

- Use the `goldentest` package for snapshot comparison when output is large or complex.
- Report failures with `t.Errorf` (continue) or `t.Fatalf` (stop).
- Run tests: `go test ./...` from the module root. The `telemetry/` directory is a separate module.

## Project Structure

- Reusable packages live at the repository root (`gitcmd`, `subcmd`, `patch`, `executil`, …).
- CLI tools live under `cmd/`. Each uses the `subcmd` package for subcommands.
- Internal implementation details go in `internal/`.
- Keep one logical concern per package.
