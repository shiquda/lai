package tui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type MainMenuState struct{}

func NewMainMenuState() *MainMenuState {
	return &MainMenuState{}
}

func (s *MainMenuState) Name() string {
	return "main_menu"
}

func (s *MainMenuState) OnEnter(m *ConfigModel) tea.Cmd {
	m.resetBreadcrumb("Main Menu")
	items := m.navigator.MainMenuItems()
	m.items = make([]ConfigItem, len(items))
	copy(m.items, items)
	m.list.SetItems(configItemsToListItems(items))
	return nil
}

func (s *MainMenuState) Update(m *ConfigModel, msg tea.Msg) (UIState, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.Enter):
			return s.handleSelection(m)
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return nil, cmd
}

func (s *MainMenuState) View(m *ConfigModel) string {
	listContentStyle := lipgloss.NewStyle().Width(m.width - 4)
	return listContentStyle.Render(m.list.View())
}

func (s *MainMenuState) handleSelection(m *ConfigModel) (UIState, tea.Cmd) {
	selected := m.list.SelectedItem()
	if selected == nil {
		return nil, nil
	}

	item, ok := selected.(ConfigItem)
	if !ok {
		return nil, nil
	}

	switch item.ItemType {
	case "section":
		if item.Section == nil {
			return nil, nil
		}
		if item.Section.Key == "providers" {
			return NewProviderListState(*item.Section), nil
		}
		return NewSectionState(*item.Section), nil
	case "action":
		return nil, m.runAction(item.Key)
	}

	return nil, nil
}

var _ UIState = (*MainMenuState)(nil)
