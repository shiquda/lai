package tui

import (
	"strconv"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/shiquda/lai/internal/config"
)

type FieldEditor interface {
	Init(field *config.FieldMetadata, value string, width int) tea.Cmd
	Update(msg tea.Msg) (tea.Cmd, bool)
	SetWidth(width int)
	InputView() string
	Value() string
}

type TextFieldEditor struct {
	input     textinput.Model
	sensitive bool
}

type BoolFieldEditor struct {
	value bool
}

func NewFieldEditor(field *config.FieldMetadata) FieldEditor {
	if field != nil && field.Type == config.TypeBool {
		return &BoolFieldEditor{}
	}
	return &TextFieldEditor{}
}

func (e *TextFieldEditor) Init(field *config.FieldMetadata, value string, width int) tea.Cmd {
	e.input = textinput.New()
	e.input.Placeholder = "Enter configuration value..."
	e.input.SetValue(value)
	e.input.Width = width
	e.sensitive = field != nil && field.Sensitive
	if e.sensitive {
		e.input.EchoMode = textinput.EchoPassword
	} else {
		e.input.EchoMode = textinput.EchoNormal
	}
	return e.input.Focus()
}

func (e *TextFieldEditor) Update(msg tea.Msg) (tea.Cmd, bool) {
	var cmd tea.Cmd
	e.input, cmd = e.input.Update(msg)
	return cmd, true
}

func (e *TextFieldEditor) SetWidth(width int) {
	if width > 0 {
		e.input.Width = width
	}
}

func (e *TextFieldEditor) InputView() string {
	return inputFocusedStyle.Render(e.input.View())
}

func (e *TextFieldEditor) Value() string {
	return e.input.Value()
}

func (e *BoolFieldEditor) Init(field *config.FieldMetadata, value string, width int) tea.Cmd {
	if value == "" {
		e.value = false
		return nil
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		e.value = false
		return nil
	}
	e.value = parsed
	return nil
}

func (e *BoolFieldEditor) Update(msg tea.Msg) (tea.Cmd, bool) {
	return nil, false
}

func (e *BoolFieldEditor) SetWidth(width int) {}

func (e *BoolFieldEditor) InputView() string {
	trueOption := buttonStyle.Render("true")
	falseOption := buttonStyle.Render("false")
	if e.value {
		trueOption = buttonActiveStyle.Render("true")
	} else {
		falseOption = buttonActiveStyle.Render("false")
	}
	return lipgloss.JoinHorizontal(lipgloss.Left, trueOption, falseOption)
}

func (e *BoolFieldEditor) Value() string {
	return strconv.FormatBool(e.value)
}

func (e *BoolFieldEditor) Toggle() {
	e.value = !e.value
}

var _ FieldEditor = (*TextFieldEditor)(nil)
var _ FieldEditor = (*BoolFieldEditor)(nil)
