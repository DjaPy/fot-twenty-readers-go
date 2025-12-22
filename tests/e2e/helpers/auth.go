package helpers

import (
	"testing"

	"github.com/playwright-community/playwright-go"
)

type AuthHelper struct {
	username string
	password string
}

func NewAuthHelper(username, password string) *AuthHelper {
	return &AuthHelper{
		username: username,
		password: password,
	}
}

func (a *AuthHelper) CreateAuthenticatedContext(t *testing.T, browser playwright.Browser) playwright.BrowserContext {
	t.Helper()

	context, err := browser.NewContext(playwright.BrowserNewContextOptions{
		HttpCredentials: &playwright.HttpCredentials{
			Username: a.username,
			Password: a.password,
		},
	})
	if err != nil {
		t.Fatalf("could not create context: %v", err)
	}

	t.Cleanup(func() {
		if err := context.Close(); err != nil {
			t.Errorf("could not close context: %v", err)
		}
	})

	return context
}

func NewPage(t *testing.T, context playwright.BrowserContext) playwright.Page {
	t.Helper()

	page, err := context.NewPage()
	if err != nil {
		t.Fatalf("could not create page: %v", err)
	}

	t.Cleanup(func() {
		if err := page.Close(); err != nil {
			t.Errorf("could not close page: %v", err)
		}
	})

	return page
}
