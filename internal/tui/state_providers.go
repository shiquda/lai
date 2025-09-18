package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ProviderListState struct {
	descriptor SectionDescriptor
}

func NewProviderListState(descriptor SectionDescriptor) *ProviderListState {
	return &ProviderListState{descriptor: descriptor}
}

func (s *ProviderListState) Name() string {
	return "provider_list"
}

func (s *ProviderListState) OnEnter(m *ConfigModel) tea.Cmd {
	m.resetBreadcrumb("Main Menu", s.descriptor.Title)
	items := m.navigator.ProviderListItems()
	m.items = make([]ConfigItem, len(items))
	copy(m.items, items)
	m.list.SetItems(configItemsToListItems(items))
	return nil
}

func (s *ProviderListState) Update(m *ConfigModel, msg tea.Msg) (UIState, tea.Cmd) {
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

func (s *ProviderListState) View(m *ConfigModel) string {
	listContentStyle := lipgloss.NewStyle().Width(m.width - 4)
	return listContentStyle.Render(m.list.View())
}

func (s *ProviderListState) handleSelection(m *ConfigModel) (UIState, tea.Cmd) {
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
	case "provider_channel":
		if item.Provider == nil {
			return nil, nil
		}
		return NewProviderConfigState(*item.Provider, s.descriptor), nil
	}

	return nil, nil
}

var _ UIState = (*ProviderListState)(nil)

type ProviderConfigState struct {
	provider ProviderDescriptor
	parent   SectionDescriptor
}

func NewProviderConfigState(provider ProviderDescriptor, parent SectionDescriptor) *ProviderConfigState {
	return &ProviderConfigState{provider: provider, parent: parent}
}

func (s *ProviderConfigState) Name() string {
	return "provider_config"
}

func (s *ProviderConfigState) OnEnter(m *ConfigModel) tea.Cmd {
	enabled := m.navigator.IsProviderEnabled(s.provider.Name)
	statusIcon := "❌"
	if enabled {
		statusIcon = "✅"
	}
	providerTitle := fmt.Sprintf("%s %s Configuration", statusIcon, s.provider.DisplayName)
	m.resetBreadcrumb("Main Menu", s.parent.Title, providerTitle)

	items := m.navigator.ProviderConfigItems(s.provider)
	m.items = make([]ConfigItem, len(items))
	copy(m.items, items)
	m.list.SetItems(configItemsToListItems(items))
	return nil
}

func (s *ProviderConfigState) Update(m *ConfigModel, msg tea.Msg) (UIState, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.Enter):
			return s.handleSelection(m)
		case key.Matches(msg, m.keyMap.Escape), key.Matches(msg, m.keyMap.Left), key.Matches(msg, m.keyMap.Back):
			return NewProviderListState(s.parent), nil
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return nil, cmd
}

func (s *ProviderConfigState) View(m *ConfigModel) string {
	listContentStyle := lipgloss.NewStyle().Width(m.width - 4)
	return listContentStyle.Render(m.list.View())
}

func (s *ProviderConfigState) handleSelection(m *ConfigModel) (UIState, tea.Cmd) {
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
		if item.Key == "back_to_providers" {
			return NewProviderListState(s.parent), nil
		}
	case "field":
		if item.Metadata == nil {
			return nil, nil
		}
		context := FieldEditContext{
			BaseBreadcrumb: m.breadcrumbTrail(),
			Provider:       cloneProviderDescriptor(s.provider),
			Section:        cloneSectionDescriptor(s.parent),
		}
		return NewFieldEditState(item.Metadata, item.Value, context), nil
	}

	return nil, nil
}

var _ UIState = (*ProviderConfigState)(nil)
