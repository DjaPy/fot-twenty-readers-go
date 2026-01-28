package helpers

import (
	"os"
	"testing"

	"github.com/playwright-community/playwright-go"
)

type TestEnv struct {
	BaseURL  string
	Username string
	Password string
}

func NewTestEnv() *TestEnv {
	return &TestEnv{
		BaseURL:  getEnv("E2E_BASE_URL", "http://localhost:8080"),
		Username: getEnv("E2E_USERNAME", "admin"),
		Password: getEnv("E2E_PASSWORD", "admin"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func SetupPlaywright(t *testing.T) *playwright.Playwright {
	t.Helper()

	pw, err := playwright.Run()
	if err != nil {
		t.Fatalf("could not start Playwright: %v", err)
	}

	t.Cleanup(func() {
		if err := pw.Stop(); err != nil {
			t.Errorf("could not stop Playwright: %v", err)
		}
	})

	return pw
}

func LaunchBrowser(t *testing.T, pw *playwright.Playwright, headless bool) playwright.Browser {
	t.Helper()

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(headless),
	})
	if err != nil {
		t.Fatalf("could not launch browser: %v", err)
	}

	t.Cleanup(func() {
		if err := browser.Close(); err != nil {
			t.Errorf("could not close browser: %v", err)
		}
	})

	return browser
}
