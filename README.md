# WhatsMyIP

A simple Go web app that shows the client's source IP, whois information and request headers such as X-Forwarded-* and Host.

## Run

Needs Go. From the project directory:

```sh
go run .      # or: make run
```

Open `http://localhost:8080`.

### Flags

| Flag | Default | Meaning |
| --- | --- | --- |
| `-port` | `8080` | Port to listen on |
| `-host` | `0.0.0.0` | IP address to listen on |
| `-include` | empty | Comma separated header names to show (empty shows all) |
| `-exclude` | empty | Comma separated header names to hide |
| `-help` | | Show help |

Environment: `WHOIS_BASE_URL` sets the whois API base URL (default: ip-api.com).

### Who sees details

The client IP is always shown. Whois info and request headers are shown only if the client IP is private, or the request has a `Remote-User` header (set by an authenticating proxy). Other requests see the IP only.

## Make targets

`build` (to `bin/`), `run`, `test`, `race`, `vet`, `fmt`, `lint`, `helm-lint`, `docker-build`, `scan` (Trivy), `clean`. Run `make help` for the list.

## Container image

```sh
docker run --rm -p 8080:8080 ghcr.io/mnorrsken/whatsmyip:latest
```

The image is distroless and runs as non-root.

## Helm chart

```sh
helm install whatsmyip oci://ghcr.io/mnorrsken/charts/whatsmyip --version X.Y.Z
```

No secrets are required. Example `values.yaml`:

```yaml
config:
  exclude: "cookie,authorization"
ingress:
  enabled: true
  className: traefik
  annotations:
    traefik.ingress.kubernetes.io/router.entrypoints: websecure
  hosts:
    - host: whatsmyip.example.com
      paths:
        - path: /
          pathType: Prefix
```

Install with `-f values.yaml`. See `charts/whatsmyip/values.yaml` for all keys.

## Releases

Push a `vX.Y.Z` tag. Add a matching `## [X.Y.Z] - YYYY-MM-DD` section to `CHANGELOG.md` first; the release workflow uses it for the GitHub release notes.
