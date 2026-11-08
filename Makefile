VERSION ?= dev

.PHONY: build test vet fmt-check man
build:
	go build -ldflags "-X main.version=$(VERSION)" -o ali ./cmd/ali
test:
	go test ./...
vet:
	go vet ./...
fmt-check:
	@test -z "$$(gofmt -l .)" || (gofmt -l .; exit 1)
man:
	go run ./internal/tools/gendocs -out man -version $(VERSION)
