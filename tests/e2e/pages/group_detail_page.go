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
	return err
}

func (p *GroupDetailPage) WaitForLoad() error {
	return p.page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateNetworkidle,
	})
}

func (p *GroupDetailPage) GetGroupName() (string, error) {
	return p.page.TextContent("main h1")
}

func (p *GroupDetailPage) AddReader(username, telegramID, phone string) error {
	if err := p.page.Fill("input[name='username']", username); err != nil {
		return fmt.Errorf("fill username: %w", err)
	}

	if err := p.page.Fill("input[name='telegram_id']", telegramID); err != nil {
		return fmt.Errorf("fill telegram_id: %w", err)
	}

	if err := p.page.Fill("input[name='phone']", phone); err != nil {
		return fmt.Errorf("fill phone: %w", err)
	}

	if err := p.page.Click("button:has-text('Добавить чтеца')"); err != nil {
		return fmt.Errorf("click add reader: %w", err)
	}

	p.page.WaitForTimeout(500)
	return nil
}

func (p *GroupDetailPage) HasReader(username string) (bool, error) {
	locator := p.page.Locator(fmt.Sprintf("text=%s", username))
	count, err := locator.Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (p *GroupDetailPage) GetReadersCount() (int, error) {
	locator := p.page.Locator(".reader-item")
	return locator.Count()
}

func (p *GroupDetailPage) DeleteReader(username string) error {
	if err := p.page.Click(fmt.Sprintf("text=%s >> .. >> button:has-text('Удалить')", username)); err != nil {
		return fmt.Errorf("click delete: %w", err)
	}

	p.page.Once("dialog", func(dialog playwright.Dialog) {
		_ = dialog.Accept()
	})

	p.page.WaitForTimeout(500)
	return nil
}

func (p *GroupDetailPage) GenerateCalendar(year string) error {
	generateButton := p.page.Locator("button:has-text('Сгенерировать календарь')")

	if err := generateButton.Click(); err != nil {
		return fmt.Errorf("click generate: %w", err)
	}

	p.page.WaitForTimeout(2000)
	return nil
}

func (p *GroupDetailPage) GetCurrentKathisma(readerNumber string) (string, error) {
	if err := p.page.Fill("#reader-number", readerNumber); err != nil {
		return "", fmt.Errorf("fill reader_number: %w", err)
	}

	if err := p.page.Click("button:has-text('Узнать кафизму')"); err != nil {
		return "", fmt.Errorf("click check kathisma: %w", err)
	}

	p.page.WaitForTimeout(500)

	return p.page.TextContent("#kathisma-result")
}

func (p *GroupDetailPage) ClickEditButton() error {
	return p.page.Click("button:has-text('Редактировать')")
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

	if err := p.page.Fill("input#edit-name", name); err != nil {
		return fmt.Errorf("fill name: %w", err)
	}

	if err := p.page.Fill("input#edit-start-offset", startOffset); err != nil {
		return fmt.Errorf("fill start offset: %w", err)
	}

	if err := p.page.Click("button:has-text('Сохранить изменения')"); err != nil {
		return fmt.Errorf("click save: %w", err)
	}

	p.page.WaitForTimeout(1000)
	return nil
}

func (p *GroupDetailPage) GetStartOffset() (string, error) {
	return p.page.TextContent("span:has-text('Стартовая кафизма:')")
}

func (p *GroupDetailPage) RegenerateCalendar(year string) error {
	p.page.Once("dialog", func(dialog playwright.Dialog) {
		_ = dialog.Accept()
	})

	regenerateButton := p.page.Locator("button:has-text('Перегенерировать')")

	if err := regenerateButton.Click(); err != nil {
		return fmt.Errorf("click regenerate: %w", err)
	}

	p.page.WaitForTimeout(2000)
	return nil
}
