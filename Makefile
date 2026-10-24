BINARY := whatsmyip
IMAGE  := whatsmyip:dev
TRIVY  ?= aquasec/trivy:0.76.1

.PHONY: all build run test race vet fmt lint helm-lint docker-build scan clean help

all: build

build:
	CGO_ENABLED=0 go build -trimpath -ldflags '-s -w' -o bin/$(BINARY) .

run:
	go run .

test:
	go test ./...

race:
	go test -race ./...

vet:
	go vet ./...

fmt:
	gofmt -w .

lint:
	golangci-lint run ./...

helm-lint:
	helm lint charts/$(BINARY)
	helm template t charts/$(BINARY) > /dev/null

docker-build:
	docker build -t $(IMAGE) .

# Trivy, as in CI: the source tree (dependencies, secrets) and a fresh
# image. Fails on any vulnerability that has a fix.
scan: docker-build
	docker run --rm -v $(HOME)/.cache/trivy:/root/.cache/trivy -v $(CURDIR):/src:ro $(TRIVY) fs --quiet \
		--scanners vuln,secret --ignore-unfixed --exit-code 1 --skip-dirs bin /src
	docker run --rm -v $(HOME)/.cache/trivy:/root/.cache/trivy -v /var/run/docker.sock:/var/run/docker.sock $(TRIVY) image --quiet \
		--scanners vuln --ignore-unfixed --exit-code 1 $(IMAGE)

clean:
	rm -rf bin .helm-packages

help:
	@echo "build         Build bin/$(BINARY)"
	@echo "run           Run with go run"
	@echo "test / race   Run tests (race: with -race)"
	@echo "vet / fmt     go vet / gofmt -w"
	@echo "lint          golangci-lint"
	@echo "helm-lint     Lint and render the Helm chart"
	@echo "docker-build  Build $(IMAGE)"
	@echo "scan          Trivy scan of source and image"
	@echo "clean         Remove build output"
