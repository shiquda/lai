package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/shiquda/lai/internal/config"
)

type FieldEditContext struct {
	BaseBreadcrumb []string
	Section        *SectionDescriptor
	Provider       *ProviderDescriptor
}

type FieldEditState struct {
	field         *config.FieldMetadata
	originalValue string
	editor        FieldEditor
	context       FieldEditContext
}

func NewFieldEditState(field *config.FieldMetadata, value string, context FieldEditContext) *FieldEditState {
	return &FieldEditState{
		field:         field,
		originalValue: value,
		editor:        NewFieldEditor(field),
		context:       context,
	}
}

func (s *FieldEditState) Name() string {
	return "field_edit"
}

func (s *FieldEditState) OnEnter(m *ConfigModel) tea.Cmd {
	if s.field == nil {
		return nil
	}

	if s.field.Key != "" {
		if latestValue, err := m.getFieldValue(s.field.Key); err == nil {
			s.originalValue = latestValue
		}
	}

	breadcrumb := append([]string{}, s.context.BaseBreadcrumb...)
	breadcrumb = append(breadcrumb, fmt.Sprintf("Edit: %s", s.field.DisplayName))
	m.resetBreadcrumb(breadcrumb...)

	width := s.computeInputWidth(m)
	cmd := s.editor.Init(s.field, s.originalValue, width)
	s.editor.SetWidth(width)
	return cmd
}

func (s *FieldEditState) Update(m *ConfigModel, msg tea.Msg) (UIState, tea.Cmd) {
	if s.field == nil {
		return nil, nil
	}

	switch typed := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(typed, m.keyMap.Enter):
			return s.exitState(m), s.commit(m)
		case key.Matches(typed, m.keyMap.Escape):
			return s.exitState(m), nil
		case key.Matches(typed, m.keyMap.Save):
			commitCmd := s.commit(m)
			nextState := s.exitState(m)
			return nextState, tea.Batch(commitCmd, m.saveConfig())
		case key.Matches(typed, m.keyMap.Left), key.Matches(typed, m.keyMap.Right):
			if boolEditor, ok := s.editor.(*BoolFieldEditor); ok {
				boolEditor.Toggle()
				return nil, nil
			}
		}

		if cmd, handled := s.editor.Update(typed); handled {
			return nil, cmd
		}
	default:
		if cmd, handled := s.editor.Update(msg); handled {
			return nil, cmd
		}
	}

	return nil, nil
}

func (s *FieldEditState) View(m *ConfigModel) string {
	if s.field == nil {
		return "Error: No field selected for editing"
	}

	field := s.field
	fieldInfo := lipgloss.JoinVertical(lipgloss.Left,
		headerStyle.Render("Edit Configuration Field"),
		"",
		contentStyle.Render(fmt.Sprintf("Name: %s", field.DisplayName)),
		descriptionStyle.Render(fmt.Sprintf("Description: %s", field.Description)),
		contentStyle.Render(fmt.Sprintf("Type: %s", field.Type)),
	)

	if field.Required {
		fieldInfo += "\n" + warningStyle.Render("* Required field")
	}

	currentValueText := ""
	if s.originalValue != "" {
		displayValue := FormatFieldValue(s.originalValue, string(field.Type), field.Sensitive)
		currentValueText = contentStyle.Render(fmt.Sprintf("Current value: %s", displayValue))
	} else {
		currentValueText = textMutedStyle.Render("Current value: not set")
	}

	inputLabel := contentStyle.Render("New value:")
	panelWidth := s.computePanelWidth(m)
	panelContentWidth := panelWidth - 6
	if panelContentWidth < 20 {
		panelContentWidth = 20
	}

	var inputField string
	if _, ok := s.editor.(*BoolFieldEditor); ok {
		inputField = s.editor.InputView()
	} else {
		inputWidth := panelContentWidth - 5
		if inputWidth < 10 {
			inputWidth = 10
		}
		s.editor.SetWidth(inputWidth)
		inputField = s.editor.InputView()
		for lipgloss.Width(inputField) > panelContentWidth && inputWidth > 5 {
			inputWidth--
			s.editor.SetWidth(inputWidth)
			inputField = s.editor.InputView()
		}
	}

	examplesText := ""
	if len(field.Examples) > 0 {
		examples := strings.Join(field.Examples, ", ")
		examplesText = textMutedStyle.Render(fmt.Sprintf("Examples: %s", examples))
	}

	dynamicPanelStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(1, 2).
		Margin(1, 0).
		Width(panelWidth)

	return dynamicPanelStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			fieldInfo,
			"",
			currentValueText,
			"",
			inputLabel,
			inputField,
			"",
			examplesText,
		),
	)
}

func (s *FieldEditState) HandleWindowSize(m *ConfigModel, _ tea.WindowSizeMsg) {
	if _, ok := s.editor.(*BoolFieldEditor); ok {
		return
	}
	width := s.computeInputWidth(m)
	s.editor.SetWidth(width)
}

func (s *FieldEditState) computeInputWidth(m *ConfigModel) int {
	inputWidth := m.width - 12
	if inputWidth < 30 {
		inputWidth = 30
	} else if inputWidth > m.width-20 {
		inputWidth = m.width - 20
	}
	return inputWidth
}

func (s *FieldEditState) computePanelWidth(m *ConfigModel) int {
	panelWidth := m.width - 4
	if panelWidth < 50 {
		panelWidth = 50
	}
	return panelWidth
}

func (s *FieldEditState) exitState(m *ConfigModel) UIState {
	if s.context.Provider != nil && s.context.Section != nil {
		return NewProviderConfigState(*s.context.Provider, *s.context.Section)
	}
	if s.context.Section != nil {
		return NewSectionState(*s.context.Section)
	}
	return NewMainMenuState()
}

func (s *FieldEditState) commit(m *ConfigModel) tea.Cmd {
	field := s.field
	if field == nil {
		return nil
	}

	newValue := s.editor.Value()
	return func() tea.Msg {
		if err := field.ValidateFieldValue(newValue); err != nil {
			return statusMsg(fmt.Sprintf("Validation failed: %v", err))
		}

		if err := m.setFieldValue(field.Key, newValue); err != nil {
			return statusMsg(fmt.Sprintf("Setting failed: %v", err))
		}

		s.originalValue = newValue
		m.hasChanges = true
		return statusMsg(fmt.Sprintf("Updated %s", field.DisplayName))
	}
}

var _ UIState = (*FieldEditState)(nil)
var _ WindowAwareState = (*FieldEditState)(nil)
