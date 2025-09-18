package tui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type SectionState struct {
	descriptor SectionDescriptor
}

func NewSectionState(descriptor SectionDescriptor) *SectionState {
	return &SectionState{descriptor: descriptor}
}

func (s *SectionState) Name() string {
	return "section"
}

func (s *SectionState) OnEnter(m *ConfigModel) tea.Cmd {
	m.resetBreadcrumb("Main Menu", s.descriptor.Title)
	items := m.navigator.SectionItems(s.descriptor)
	m.items = make([]ConfigItem, len(items))
	copy(m.items, items)
	m.list.SetItems(configItemsToListItems(items))
	return nil
}

func (s *SectionState) Update(m *ConfigModel, msg tea.Msg) (UIState, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.Enter):
			return s.handleSelection(m)
		case key.Matches(msg, m.keyMap.Escape), key.Matches(msg, m.keyMap.Left), key.Matches(msg, m.keyMap.Back):
			return NewMainMenuState(), nil
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return nil, cmd
}

func (s *SectionState) View(m *ConfigModel) string {
	listContentStyle := lipgloss.NewStyle().Width(m.width - 4)
	return listContentStyle.Render(m.list.View())
}

func (s *SectionState) handleSelection(m *ConfigModel) (UIState, tea.Cmd) {
	selected := m.list.SelectedItem()
	if selected == nil {
		return nil, nil
	}

	item, ok := selected.(ConfigItem)
	if !ok {
		return nil, nil
	}

	switch item.ItemType {
	case "navigation":
		if item.Key == "back" {
			return NewMainMenuState(), nil
		}
	case "field":
		if item.Metadata == nil {
			return nil, nil
		}
		context := FieldEditContext{
			BaseBreadcrumb: m.breadcrumbTrail(),
			Section:        cloneSectionDescriptor(s.descriptor),
		}
		return NewFieldEditState(item.Metadata, item.Value, context), nil
	}

	return nil, nil
}

var _ UIState = (*SectionState)(nil)
