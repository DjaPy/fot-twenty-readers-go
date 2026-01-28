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
	if err != nil {
		return fmt.Errorf("navigate to groups: %w", err)
	}
	return nil
}

func (p *GroupsPage) WaitForLoad() error {
	err := p.page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateNetworkidle,
	})
	if err != nil {
		return fmt.Errorf("wait for load: %w", err)
	}
	return nil
}

func (p *GroupsPage) GetTitle() (string, error) {
	title, err := p.page.Title()
	if err != nil {
		return "", fmt.Errorf("get page title: %w", err)
	}
	return title, nil
}

func (p *GroupsPage) FillGroupName(name string) error {
	err := p.page.Locator("input#name").Fill(name)
	if err != nil {
		return fmt.Errorf("get group name: %w", err)
	}
	return nil
}

func (p *GroupsPage) FillStartOffset(offset string) error {
	err := p.page.Locator("input#start_offset").Fill(offset)
	if err != nil {
		return fmt.Errorf("get group start_offset: %w", err)
	}
	return nil
}

func (p *GroupsPage) SubmitCreateGroup() error {
	err := p.page.Locator("button:has-text('Создать группу')").Click()
	if err != nil {
		return fmt.Errorf("create group: %w", err)
	}
	return nil
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

	groupLocator := p.page.Locator(fmt.Sprintf("h3:text('%s')", name))
	if err := groupLocator.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
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
		return false, fmt.Errorf("count groups: %w", err)
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
	err := detailsLink.Click()
	if err != nil {
		return fmt.Errorf("click group details: %w", err)
	}
	return nil
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

	errWait := groupCard.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateDetached,
		Timeout: playwright.Float(5000),
	})
	if errWait != nil {
		return fmt.Errorf("wait for group to appear: %w", errWait)
	}
	return nil
}

func (p *GroupsPage) GroupExists(groupName string) (bool, error) {
	locator := p.page.Locator(fmt.Sprintf("#groups-list h3:has-text('%s')", groupName))
	count, err := locator.Count()
	if err != nil {
		return false, fmt.Errorf("check group exists: %w", err)
	}
	return count > 0, nil
}
