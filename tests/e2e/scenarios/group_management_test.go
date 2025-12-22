package scenarios

import (
	"testing"

	"github.com/DjaPy/fot-twenty-readers-go/tests/e2e/helpers"
	"github.com/DjaPy/fot-twenty-readers-go/tests/e2e/pages"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeleteGroup(t *testing.T) {
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

	groupName := "Test Group To Delete"
	err = groupsPage.FillGroupName(groupName)
	require.NoError(t, err)

	err = groupsPage.FillStartOffset("5")
	require.NoError(t, err)

	err = groupsPage.SubmitCreateGroup()
	require.NoError(t, err)

	err = groupsPage.WaitForLoad()
	require.NoError(t, err)

	exists, err := groupsPage.GroupExists(groupName)
	require.NoError(t, err)
	assert.True(t, exists, "group should exist after creation")

	err = groupsPage.DeleteGroup(groupName)
	require.NoError(t, err)

	err = groupsPage.WaitForLoad()
	require.NoError(t, err)

	exists, err = groupsPage.GroupExists(groupName)
	require.NoError(t, err)
	assert.False(t, exists, "group should not exist after deletion")

	t.Log("Successfully deleted group")
}

func TestEditGroup(t *testing.T) {
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

	originalName := "Test Group Original"
	err = groupsPage.FillGroupName(originalName)
	require.NoError(t, err)

	err = groupsPage.FillStartOffset("3")
	require.NoError(t, err)

	err = groupsPage.SubmitCreateGroup()
	require.NoError(t, err)

	err = groupsPage.WaitForLoad()
	require.NoError(t, err)

	err = groupsPage.ClickGroupDetails(originalName)
	require.NoError(t, err)

	detailPage := pages.NewGroupDetailPage(page, env.BaseURL)
	err = detailPage.WaitForLoad()
	require.NoError(t, err)

	name, err := detailPage.GetGroupName()
	require.NoError(t, err)
	assert.Contains(t, name, originalName, "original name should be displayed")

	newName := "Test Group Updated"
	err = detailPage.EditGroup(newName, "7")
	require.NoError(t, err)

	err = detailPage.WaitForLoad()
	require.NoError(t, err)

	updatedName, err := detailPage.GetGroupName()
	require.NoError(t, err)
	assert.Contains(t, updatedName, newName, "updated name should be displayed")

	t.Log("Successfully edited group")
}

func TestRegenerateCalendar(t *testing.T) {
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

	groupName := "Test Group For Regeneration"
	err = groupsPage.FillGroupName(groupName)
	require.NoError(t, err)

	err = groupsPage.FillStartOffset("1")
	require.NoError(t, err)

	err = groupsPage.SubmitCreateGroup()
	require.NoError(t, err)

	err = groupsPage.WaitForLoad()
	require.NoError(t, err)

	err = groupsPage.ClickGroupDetails(groupName)
	require.NoError(t, err)

	detailPage := pages.NewGroupDetailPage(page, env.BaseURL)
	err = detailPage.WaitForLoad()
	require.NoError(t, err)

	err = detailPage.GenerateCalendar("2025")
	require.NoError(t, err)

	err = detailPage.WaitForLoad()
	require.NoError(t, err)

	err = detailPage.RegenerateCalendar("2025")
	require.NoError(t, err)

	err = detailPage.WaitForLoad()
	require.NoError(t, err)

	t.Log("Successfully regenerated calendar")
}
