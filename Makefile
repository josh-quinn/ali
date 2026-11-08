VERSION ?= dev

.PHONY: build test vet fmt-check man lint vuln check hooks secrets
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
	shellcheck .githooks/pre-push
vuln:
	go run golang.org/x/vuln/cmd/govulncheck@v1.9.0 ./...
# scan the whole history (every branch) for secrets, as CI does
secrets:
	gitleaks git --log-opts="--all" --redact --no-banner -v .
# use the repository's git hooks (pre-push secret scan) in this clone
hooks:
	git config core.hooksPath .githooks
	@echo "git hooks enabled: pushes are scanned for secrets (needs gitleaks: brew install gitleaks)"
# everything CI checks, in one command
check: fmt-check vet lint test vuln secrets
