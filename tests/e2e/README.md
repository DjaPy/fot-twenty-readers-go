# E2E Tests

End-to-end tests for the web interface using Playwright.

## Prerequisites

1. Install Playwright browsers:
```bash
go run github.com/playwright-community/playwright-go/cmd/playwright@latest install --with-deps chromium
```

2. Start the test environment:
```bash
cd deploy
docker-compose -f docker-compose.test.yml up -d
```

## Running Tests

### All E2E tests
```bash
go test -v ./tests/e2e/...
```

### Specific test
```bash
go test -v ./tests/e2e/scenarios -run TestSmokeGroupsPageWithAuth
```

### With visible browser (non-headless)
```bash
E2E_HEADLESS=false go test -v ./tests/e2e/...
```

## Environment Variables

- `E2E_BASE_URL` - Base URL of the application (default: `http://localhost:8080`)
- `E2E_USERNAME` - Basic Auth username (default: `admin`)
- `E2E_PASSWORD` - Basic Auth password (default: `admin`)
- `E2E_HEADLESS` - Run browser in headless mode (default: `true`)

## Test Structure

```
tests/e2e/
├── helpers/       - Test utilities and setup
├── pages/         - Page Object Models
└── scenarios/     - Test scenarios
```

## Writing Tests

Use the Page Object pattern:

```go
func TestCreateGroup(t *testing.T) {
    env := helpers.NewTestEnv()
    pw := helpers.SetupPlaywright(t)
    browser := helpers.LaunchBrowser(t, pw, true)

    authHelper := helpers.NewAuthHelper(env.Username, env.Password)
    context := authHelper.CreateAuthenticatedContext(t, browser)
    page := helpers.NewPage(t, context)

    groupsPage := pages.NewGroupsPage(page, env.BaseURL)
    err := groupsPage.Navigate()
    require.NoError(t, err)

    err = groupsPage.CreateGroup("Test Group", "1")
    require.NoError(t, err)

    hasGroup, err := groupsPage.HasGroupInList("Test Group")
    require.NoError(t, err)
    assert.True(t, hasGroup)
}
```

## Cleanup

Stop test environment:
```bash
cd deploy
docker-compose -f docker-compose.test.yml down -v
```