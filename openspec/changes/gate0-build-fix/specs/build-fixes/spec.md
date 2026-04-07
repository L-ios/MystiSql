## ADDED Requirements

### Requirement: Default build succeeds without frontend assets
The system SHALL compile with `go build ./...` without requiring `web/dist/` directory or any frontend build toolchain.

#### Scenario: Build without webembed tag
- **WHEN** developer runs `go build ./...` without any build tags
- **THEN** compilation succeeds with zero errors and the resulting binary does not embed WebUI assets

#### Scenario: NewHandler returns nil in stub mode
- **WHEN** the binary is built without `-tags webembed`
- **THEN** `web.NewHandler()` returns `(nil, nil)` and the server gracefully handles the nil handler

### Requirement: WebUI embed enabled via build tag
The system SHALL support embedding WebUI assets when built with `-tags webembed`.

#### Scenario: Build with webembed tag and dist present
- **WHEN** developer runs `go build -tags webembed ./...` and `web/dist/` exists with built frontend assets
- **THEN** `web.NewHandler()` returns a functional Handler serving the SPA

#### Scenario: Build with webembed tag but dist missing
- **WHEN** developer runs `go build -tags webembed ./...` and `web/dist/` does not exist
- **THEN** compilation fails with a clear `//go:embed` error indicating the dist directory is missing

### Requirement: go vet passes cleanly
The system SHALL pass `go vet ./...` with zero warnings.

#### Scenario: Vet after build tag refactor
- **WHEN** developer runs `go vet ./...`
- **THEN** the command exits with code 0 and produces no output
