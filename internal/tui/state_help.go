package tui

import tea "github.com/charmbracelet/bubbletea"

type HelpState struct{}

func NewHelpState() *HelpState {
	return &HelpState{}
}

func (s *HelpState) Name() string {
	return "help"
}

func (s *HelpState) OnEnter(m *ConfigModel) tea.Cmd {
	base := m.breadcrumbTrail()
	base = append(base, "Help")
	m.resetBreadcrumb(base...)
	return nil
}

func (s *HelpState) Update(m *ConfigModel, msg tea.Msg) (UIState, tea.Cmd) {
	return nil, nil
}

func (s *HelpState) View(m *ConfigModel) string {
	helpContent := `
## Lai Interactive Configuration Help

### Navigation
- ↑/↓ or k/j: Move up/down
- →/l: Enter selected item
- ←/h: Go back
- Enter: Confirm selection
- Esc: Cancel current operation

### Editing Configuration
- Press Enter on configuration item to edit
- Enter new value and press Enter to confirm
- Press Esc to cancel editing

### Saving Configuration
- Ctrl+S: Save all changes
- Warning shown when exiting with unsaved changes

### Other
- q: Quit program
- ?: Show/hide help

### Configuration Types
- String: Direct text input
- Number: Integer input
- Boolean: Select true or false
- Duration: e.g., 30s, 5m, 1h
- List: Comma-separated values
- Secret: Input will be hidden
`
	return helpStyle.Render(helpContent)
}

var _ UIState = (*HelpState)(nil)
