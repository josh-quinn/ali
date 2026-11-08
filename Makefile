VERSION ?= dev

.PHONY: build test vet fmt-check man lint vuln check
build:
	go build -ldflags "-X main.version=$(VERSION)" -o ali ./cmd/ali
test:
	go test -count=1 ./...
vet:
	go vet ./...
fmt-check:
	@test -z "$$(gofmt -l .)" || (gofmt -l .; exit 1)
man:
	go run ./internal/tools/gendocs -out man -version $(VERSION)
lint:
	golangci-lint run ./...
vuln:
	go run golang.org/x/vuln/cmd/govulncheck@v1.9.0 ./...
# everything CI checks, in one command
check: fmt-check vet lint test vuln
