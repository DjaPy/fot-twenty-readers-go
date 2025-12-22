package scenarios

import (
	"testing"

	"github.com/DjaPy/fot-twenty-readers-go/tests/e2e/helpers"
	"github.com/DjaPy/fot-twenty-readers-go/tests/e2e/pages"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSmokeGroupsPageWithAuth(t *testing.T) {
	env := helpers.NewTestEnv()
	pw := helpers.SetupPlaywright(t)
	browser := helpers.LaunchBrowser(t, pw, true)

	authHelper := helpers.NewAuthHelper(env.Username, env.Password)
	context := authHelper.CreateAuthenticatedContext(t, browser)
	page := helpers.NewPage(t, context)

	groupsPage := pages.NewGroupsPage(page, env.BaseURL)
	err := groupsPage.Navigate()
	require.NoError(t, err)

	err = groupsPage.WaitForLoad()
	require.NoError(t, err)

	title, err := groupsPage.GetTitle()
	require.NoError(t, err)
	assert.NotEmpty(t, title)

	t.Logf("Successfully accessed groups page with title: %s", title)
}

func TestSmokeGroupsPageWithoutAuth(t *testing.T) {
	env := helpers.NewTestEnv()
	pw := helpers.SetupPlaywright(t)
	browser := helpers.LaunchBrowser(t, pw, true)

	context, err := browser.NewContext()
	require.NoError(t, err)
	defer context.Close()

	page, err := context.NewPage()
	require.NoError(t, err)
	defer page.Close()

	groupsPage := pages.NewGroupsPage(page, env.BaseURL)
	err = groupsPage.Navigate()

	if err != nil {
		t.Logf("Navigation failed (expected): %v", err)
	}
}
