package tui

import tea "github.com/charmbracelet/bubbletea"

// UIState represents a state in the configuration TUI state machine.
type UIState interface {
	Name() string
	OnEnter(*ConfigModel) tea.Cmd
	Update(*ConfigModel, tea.Msg) (UIState, tea.Cmd)
	View(*ConfigModel) string
}

// WindowAwareState is implemented by states that need window size notifications.
type WindowAwareState interface {
	HandleWindowSize(*ConfigModel, tea.WindowSizeMsg)
}
