# Changelog

All notable changes to this project are documented here. Headings use `## [X.Y.Z] - YYYY-MM-DD`; the release workflow takes the notes from them.

## [Unreleased]

### Added

- Helm chart `charts/whatsmyip` with values for `config.include`, `config.exclude`, `config.whoisBaseURL` and `ingress.*`. Hardened security context and read-only root filesystem.
- CI workflow: gofmt, vet, tests with `-race`, golangci-lint, Trivy file system and image scans, Helm lint and template.
- Release workflow on `v*` tags: tests, Trivy-gated multi-arch image push (semver tags and `latest`), Helm chart push to `oci://ghcr.io/mnorrsken/charts`, GitHub release with notes from this file.
- Daily Trivy scan of `ghcr.io/mnorrsken/whatsmyip:latest` with SARIF upload to code scanning. Fails on fixable vulnerabilities.
- Dependabot for Go modules, GitHub Actions and Docker (weekly, grouped).
- Makefile targets: `build`, `run`, `test`, `race`, `vet`, `fmt`, `lint`, `helm-lint`, `docker-build`, `scan`, `clean`, `help`.
- Dark mode.

### Changed

- Image is built with `golang:1.27-alpine` and runs on `gcr.io/distroless/static-debian13:nonroot` as a non-root user. Templates are in `/app/templates`.
- Go dependencies updated; `go` directive is now 1.26.
- The copy button for the IP address is now a real `<button>`.
- The image is no longer pushed on every push to `main`. It is published only by the release workflow.

### Removed

- Tailwind CDN script. The page uses a small inline stylesheet and has no external requests for styling.
- Raw Kubernetes manifests in `deploy/`. Use the Helm chart.
- Workflow `docker-image.yml`.

### Security

- Container runs as non-root with a read-only root filesystem and all capabilities dropped (Helm chart).
- Images and dependencies are scanned with Trivy in CI, before release, and daily.
