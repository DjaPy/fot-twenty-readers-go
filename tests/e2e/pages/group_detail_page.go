package pages

import (
	"fmt"

	"github.com/playwright-community/playwright-go"
)

type GroupDetailPage struct {
	page    playwright.Page
	baseURL string
}

func NewGroupDetailPage(page playwright.Page, baseURL string) *GroupDetailPage {
	return &GroupDetailPage{
		page:    page,
		baseURL: baseURL,
	}
}

func (p *GroupDetailPage) NavigateToGroup(groupID string) error {
	_, err := p.page.Goto(fmt.Sprintf("%s/groups/%s", p.baseURL, groupID))
	if err != nil {
		return fmt.Errorf("navigate to group: %w", err)
	}
	return nil
}

func (p *GroupDetailPage) WaitForLoad() error {
	err := p.page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateNetworkidle,
	})
	if err != nil {
		return fmt.Errorf("wait for load: %w", err)
	}
	return nil
}

func (p *GroupDetailPage) GetGroupName() (string, error) {
	groupName, err := p.page.Locator("main h1").TextContent()
	if err != nil {
		return "", fmt.Errorf("get group name: %w", err)
	}
	return groupName, nil
}

func (p *GroupDetailPage) AddReader(username, telegramID, phone string) error {
	if err := p.page.Locator("input[name='username']").Fill(username); err != nil {
		return fmt.Errorf("fill username: %w", err)
	}

	if err := p.page.Locator("input[name='telegram_id']").Fill(telegramID); err != nil {
		return fmt.Errorf("fill telegram_id: %w", err)
	}

	if err := p.page.Locator("input[name='phone']").Fill(phone); err != nil {
		return fmt.Errorf("fill phone: %w", err)
	}

	initialCount, err := p.page.Locator(".reader-item").Count()
	if err != nil {
		return fmt.Errorf("get initial readers count: %w", err)
	}

	if err := p.page.Locator("button:has-text('Добавить чтеца')").Click(); err != nil {
		return fmt.Errorf("click add reader: %w", err)
	}

	expectedCount := initialCount + 1
	errWait := p.page.Locator(".reader-item").Nth(expectedCount - 1).WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(5000),
	})
	if errWait != nil {
		return fmt.Errorf("wait for load: %w", errWait)
	}
	return nil
}

func (p *GroupDetailPage) HasReader(username string) (bool, error) {
	locator := p.page.Locator(fmt.Sprintf("text=%s", username))
	count, err := locator.Count()
	if err != nil {
		return false, fmt.Errorf("count readers: %w", err)
	}
	return count > 0, nil
}

func (p *GroupDetailPage) GetReadersCount() (int, error) {
	locator := p.page.Locator(".reader-item")
	count, err := locator.Count()
	if err != nil {
		return 0, fmt.Errorf("get readers count: %w", err)
	}
	return count, nil

}

func (p *GroupDetailPage) DeleteReader(username string) error {
	readerLocator := p.page.Locator(fmt.Sprintf("text=%s", username)).First()

	p.page.Once("dialog", func(dialog playwright.Dialog) {
		_ = dialog.Accept()
	})

	if err := p.page.Locator(fmt.Sprintf("text=%s >> .. >> button:has-text('Удалить')", username)).Click(); err != nil {
		return fmt.Errorf("click delete: %w", err)
	}

	err := readerLocator.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateDetached,
		Timeout: playwright.Float(5000),
	})
	if err != nil {
		return fmt.Errorf("wait for load: %w", err)
	}
	return nil
}

func (p *GroupDetailPage) GenerateCalendar(year string) error {
	generateButton := p.page.Locator("button:has-text('Сгенерировать календарь')")

	if err := generateButton.Click(); err != nil {
		return fmt.Errorf("click generate: %w", err)
	}

	err := generateButton.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(10000),
	})
	if err != nil {
		return fmt.Errorf("wait for generate: %w", err)
	}
	return nil
}

func (p *GroupDetailPage) GetCurrentKathisma(readerNumber string) (string, error) {
	if err := p.page.Locator("#reader-number").Fill(readerNumber); err != nil {
		return "", fmt.Errorf("fill reader_number: %w", err)
	}

	if err := p.page.Locator("button:has-text('Узнать кафизму')").Click(); err != nil {
		return "", fmt.Errorf("click check kathisma: %w", err)
	}

	resultLocator := p.page.Locator("#kathisma-result")
	if err := resultLocator.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(5000),
	}); err != nil {
		return "", fmt.Errorf("wait for result: %w", err)
	}

	text, err := resultLocator.TextContent()
	if err != nil {
		return "", fmt.Errorf("get kathisma text: %w", err)
	}
	return text, nil
}

func (p *GroupDetailPage) ClickEditButton() error {
	err := p.page.Locator("button:has-text('Редактировать')").Click()
	if err != nil {
		return fmt.Errorf("click edit button: %w", err)
	}
	return nil
}

func (p *GroupDetailPage) EditGroup(name, startOffset string) error {
	if err := p.ClickEditButton(); err != nil {
		return fmt.Errorf("click edit button: %w", err)
	}

	editForm := p.page.Locator("#edit-group-form")
	if err := editForm.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(3000),
	}); err != nil {
		return fmt.Errorf("wait for edit form: %w", err)
	}

	if err := p.page.Locator("input#edit-name").Fill(name); err != nil {
		return fmt.Errorf("fill name: %w", err)
	}

	if err := p.page.Locator("input#edit-start-offset").Fill(startOffset); err != nil {
		return fmt.Errorf("fill start offset: %w", err)
	}

	if err := p.page.Locator("button:has-text('Сохранить изменения')").Click(); err != nil {
		return fmt.Errorf("click save: %w", err)
	}

	errWait := editForm.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateHidden,
		Timeout: playwright.Float(5000),
	})
	if errWait != nil {
		return fmt.Errorf("%w", errWait)
	}
	return nil
}

func (p *GroupDetailPage) GetStartOffset() (string, error) {
	startOffset, err := p.page.Locator("span:has-text('Стартовая кафизма:')").TextContent()
	if err != nil {
		return "", fmt.Errorf("get start offset: %w", err)
	}
	return startOffset, nil
}

func (p *GroupDetailPage) RegenerateCalendar(year string) error {
	p.page.Once("dialog", func(dialog playwright.Dialog) {
		_ = dialog.Accept()
	})

	regenerateButton := p.page.Locator("button:has-text('Перегенерировать')")

	if err := regenerateButton.Click(); err != nil {
		return fmt.Errorf("click regenerate: %w", err)
	}

	errWait := regenerateButton.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(10000),
	})
	if errWait != nil {
		return fmt.Errorf("wait for regenerate button: %w", errWait)
	}
	return nil
}
