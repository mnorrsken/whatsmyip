ARG BUILDPLATFORM
FROM --platform=$BUILDPLATFORM golang:1.27-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./

ARG TARGETOS TARGETARCH
# -s -w strips symbol table and DWARF; -trimpath drops build-host paths.
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags="-s -w" -o /out/whatsmyip .

FROM gcr.io/distroless/static-debian13:nonroot

# The binary reads templates/index.html relative to its working directory.
WORKDIR /app
COPY --from=builder /out/whatsmyip /usr/local/bin/whatsmyip
COPY templates /app/templates

ENV WHOIS_BASE_URL=http://ip-api.com/json

USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/whatsmyip"]
