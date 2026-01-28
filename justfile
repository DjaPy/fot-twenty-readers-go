
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
    golangci-lint run ./internal/... ./cmd/... ./tests/...
    djlint internal/kathismas/ports/templates/ --profile golang --extension gohtml --reformat
    djlint internal/kathismas/ports/templates/ --profile golang --extension gohtml --check
    @echo "=== Linters success ==="

test:
    go test -v ./...

test-unit:
    go test -v -race ./internal/...

test-coverage:
    @echo "=== Run tests ==="
    go test -coverprofile=coverage.out -atomic ./internal/...
    go tool cover -func=coverage.out
    @echo "=== Test success ==="

test-e2e:
    @echo "=== Run E2E tests ==="
    E2E_BASE_URL=http://localhost:8080 E2E_USERNAME=admin E2E_PASSWORD=admin go test -v ./tests/e2e/...
    @echo "=== E2E tests success ==="

test-e2e-env-start:
    @echo "=== Starting E2E test environment ==="
    cd deploy && docker-compose -f docker-compose.test.yml up -d
    @echo "Waiting for services to be healthy..."
    sleep 5
    @echo "Test environment ready at http://localhost:8080"
    @echo "Username: admin, Password: admin"

test-e2e-env-stop:
    @echo "=== Stopping E2E test environment ==="
    cd deploy && docker-compose -f docker-compose.test.yml down -v
    @echo "=== Test environment stopped ==="

install-playwright:
    @echo "=== Installing Playwright browsers ==="
    go run github.com/playwright-community/playwright-go/cmd/playwright@latest install --with-deps chromium
    @echo "=== Playwright installed ==="

generate-mocks:
    @echo "=== Generate mocks ==="
    moq -out internal/kathismas/domain/mocks/repository_reader_group_mock.go -pkg mocks internal/kathismas/domain RepositoryReaderGroup
    @echo "=== Mocks generated ==="

all-check:
    set -e
    just check-dependensies
    just lint