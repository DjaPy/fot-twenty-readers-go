
LOGIN := "default"
TOKEN := "default"
CONTAINER_VERSION := "0.0.0"


# run fmt scanner
fmt lint:
    golangci-lint fmt scanner/...


# run build docker container
container build:
    docker buildx build --build-arg GO_LIB_GITLAB_USER={{LOGIN}} --build-arg GO_LIB_GITLAB_TOKEN={{TOKEN}} --platform linux/amd64 -f Dockerfile -t scanner-local:{{CONTAINER_VERSION}} .

check-dependensies:
    @echo "=== Run check dependencies ==="
    go mod verify
    go mod tidy
    go build ./...
    @echo "=== Check dependencies success ==="

lint:
    @echo "=== Run linters ==="
    go clean -testcache
    goimports -w .
    go vet ./...
    go fmt ./...
    golangci-lint run ./internal/... ./cmd/...
    @echo "=== Linters success ==="

test:
    go test -v ./...

test-coverage:
    @echo "=== Run tests ==="
    go test -coverprofile=coverage.out ./...
    go tool cover -func=coverage.out
    @echo "=== Test success ==="

all-check:
    set -e
    just check-dependensies
    just lint
    just test-coverage