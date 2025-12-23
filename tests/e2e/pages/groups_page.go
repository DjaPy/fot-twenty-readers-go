package pages

import (
	"fmt"

	"github.com/playwright-community/playwright-go"
)

type GroupsPage struct {
	page    playwright.Page
	baseURL string
}

func NewGroupsPage(page playwright.Page, baseURL string) *GroupsPage {
	return &GroupsPage{
		page:    page,
		baseURL: baseURL,
	}
}

func (p *GroupsPage) Navigate() error {
	_, err := p.page.Goto(fmt.Sprintf("%s/groups", p.baseURL))
	return err
}

func (p *GroupsPage) WaitForLoad() error {
	return p.page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateNetworkidle,
	})
}

func (p *GroupsPage) GetTitle() (string, error) {
	return p.page.Title()
}

func (p *GroupsPage) FillGroupName(name string) error {
	return p.page.Fill("input#name", name)
}

func (p *GroupsPage) FillStartOffset(offset string) error {
	return p.page.Fill("input#start_offset", offset)
}

func (p *GroupsPage) SubmitCreateGroup() error {
	return p.page.Click("button:has-text('Создать группу')")
}

func (p *GroupsPage) CreateGroup(name, startOffset string) error {
	if err := p.FillGroupName(name); err != nil {
		return fmt.Errorf("fill group name: %w", err)
	}

	if err := p.FillStartOffset(startOffset); err != nil {
		return fmt.Errorf("fill start offset: %w", err)
	}

	if err := p.SubmitCreateGroup(); err != nil {
		return fmt.Errorf("submit create group: %w", err)
	}

	if _, err := p.page.WaitForSelector(fmt.Sprintf("h3:text('%s')", name), playwright.PageWaitForSelectorOptions{
		Timeout: playwright.Float(5000),
	}); err != nil {
		return fmt.Errorf("wait for group to appear: %w", err)
	}

	return nil
}

func (p *GroupsPage) HasGroupInList(groupName string) (bool, error) {
	locator := p.page.Locator(fmt.Sprintf("text=%s", groupName))
	count, err := locator.Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (p *GroupsPage) ClickGroupDetails(groupName string) error {
	groupTitle := p.page.Locator(fmt.Sprintf("h3:has-text('%s')", groupName)).First()

	if err := groupTitle.WaitFor(playwright.LocatorWaitForOptions{
		Timeout: playwright.Float(5000),
	}); err != nil {
		return fmt.Errorf("wait for group title: %w", err)
	}

	detailsLink := p.page.Locator("a:has-text('Подробнее')").First()
	return detailsLink.Click()
}

func (p *GroupsPage) DeleteGroup(groupName string) error {
	groupCard := p.page.Locator(".p-4").Filter(playwright.LocatorFilterOptions{
		HasText: groupName,
	}).First()

	p.page.On("dialog", func(dialog playwright.Dialog) {
		_ = dialog.Accept()
	})

	deleteButton := groupCard.Locator("button:has-text('🗑️')")

	if err := deleteButton.Click(); err != nil {
		return fmt.Errorf("click delete: %w", err)
	}

	p.page.WaitForTimeout(1500)

	return nil
}

func (p *GroupsPage) GroupExists(groupName string) (bool, error) {
	locator := p.page.Locator(fmt.Sprintf("#groups-list h3:has-text('%s')", groupName))
	count, err := locator.Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
