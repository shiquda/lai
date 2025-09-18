package tui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"

	"github.com/shiquda/lai/internal/config"
)

// KeyMap defines the key bindings for the interactive config
type KeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Left   key.Binding
	Right  key.Binding
	Enter  key.Binding
	Escape key.Binding
	Tab    key.Binding
	Back   key.Binding
	Save   key.Binding
	Quit   key.Binding
	Help   key.Binding
}

// DefaultKeyMap returns the default key mappings
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "move up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "move down"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "back"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "enter"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "confirm"),
		),
		Escape: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next item"),
		),
		Back: key.NewBinding(
			key.WithKeys("backspace"),
			key.WithHelp("backspace", "back"),
		),
		Save: key.NewBinding(
			key.WithKeys("ctrl+s"),
			key.WithHelp("ctrl+s", "save"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
	}
}

// ViewState represents the current view state
type ViewState int

const (
	ViewMainMenu ViewState = iota
	ViewSectionList
	ViewFieldEdit
	ViewConfirm
	ViewHelp
)

// ConfigItem represents an item in the configuration interface
type ConfigItem struct {
	Title       string
	Description string
	Key         string
	Value       string
	ItemType    string // "section", "field", "action"
	Level       int
	Required    bool
	Sensitive   bool
	Editable    bool
	Metadata    *config.FieldMetadata
	Section     *SectionDescriptor
	Provider    *ProviderDescriptor
}

// FilterState returns the title for list filtering
func (c ConfigItem) FilterValue() string {
	return c.Title
}

// ConfigModel represents the main model for the interactive config
type ConfigModel struct {
	// State management
	keyMap          KeyMap
	quitting        bool
	width           int
	height          int
	err             error
	currentState    UIState
	stateBeforeHelp UIState

	// (debug fields removed)

	// Configuration data
	globalConfig *config.GlobalConfig
	metadata     *config.ConfigMetadata
	navigator    *NavigationBuilder

	// Navigation
	breadcrumb []string

	// UI components
	list  list.Model
	items []ConfigItem

	// Edit state
	hasChanges bool

	// Messages
	statusMessage string
	statusType    string // "success", "error", "warning", "info"
}

// NewConfigModel creates a new interactive config model
func NewConfigModel() (*ConfigModel, error) {
	// Load current global configuration
	globalConfig, err := config.LoadGlobalConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load global config: %w", err)
	}

	// Get configuration metadata
	metadata := config.GetConfigMetadata()

	// Create list model - use minimal initial size, will be resized properly
	listModel := list.New([]list.Item{}, NewConfigDelegate(), 10, 10)
	listModel.Title = "Lai Configuration"
	listModel.SetShowStatusBar(false)
	listModel.SetFilteringEnabled(true)
	// Hide built-in help to avoid duplicate help blocks; we'll use custom footer only
	listModel.SetShowHelp(false)

	// Set basic styles without width constraints - let the list manage its own width
	listModel.Styles.Title = titleStyle
	listModel.Styles.PaginationStyle = textMutedStyle
	listModel.Styles.HelpStyle = helpStyle

	model := &ConfigModel{
		keyMap:       DefaultKeyMap(),
		globalConfig: globalConfig,
		metadata:     metadata,
		navigator:    NewNavigationBuilder(metadata, globalConfig),
		list:         listModel,
		width:        10, // Will be set properly by WindowSizeMsg
		height:       10, // Will be set properly by WindowSizeMsg
		// debug env feature flags removed
	}

	// Initialize main menu state
	model.setState(NewMainMenuState())

	return model, nil
}

// Init initializes the model
func (m *ConfigModel) Init() tea.Cmd {
	// Try to get real terminal size using multiple methods
	width, height := 120, 30 // Default fallback

	// First try with term package
	if w, h, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
		width, height = w, h
	}

	// Ensure minimum reasonable size
	if width < 80 {
		width = 120
	}
	if height < 20 {
		height = 30
	}

	// Enable Windows key support and alt screen
	return tea.Batch(
		tea.EnterAltScreen,
		tea.EnableMouseCellMotion,
		// Send initial window size message with detected dimensions
		func() tea.Msg {
			return tea.WindowSizeMsg{Width: width, Height: height}
		},
	)
}

func (m *ConfigModel) setState(state UIState) tea.Cmd {
	m.currentState = state
	if state == nil {
		return nil
	}
	return state.OnEnter(m)
}

func (m *ConfigModel) resetBreadcrumb(labels ...string) {
	m.breadcrumb = append([]string{}, labels...)
}

func (m *ConfigModel) breadcrumbTrail() []string {
	return append([]string{}, m.breadcrumb...)
}

// Update handles messages and updates the model
func (m *ConfigModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Set list width directly following the official documentation pattern
		// Leave some margin for the container
		listWidth := msg.Width - 2 // minimal margin
		if listWidth < 40 {
			listWidth = 40
		}
		m.list.SetWidth(listWidth)
		// Expand help panel to full list width so its border reaches the right edge
		hs := m.list.Styles.HelpStyle
		// helpStyle currently has RoundedBorder + Padding(1) so horizontal frame = 2 (border) + 2 (padding)
		contentWidth := listWidth - 4
		if contentWidth < 10 { // safety cap
			contentWidth = listWidth - 2
		}
		hs = hs.Width(contentWidth)
		m.list.Styles.HelpStyle = hs

		// Set height for better display
		listHeight := msg.Height - 10 // Account for header, footer, and padding
		if listHeight < 10 {
			listHeight = 10
		}
		m.list.SetHeight(listHeight)

		if handler, ok := m.currentState.(WindowAwareState); ok {
			handler.HandleWindowSize(m, msg)
		}
	case statusMsg:
		m.statusMessage = string(msg)
		m.statusType = "info"
	case tea.KeyMsg:
		if handled, cmd := m.handleGlobalKey(msg); handled {
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
			if len(cmds) > 0 {
				return m, tea.Batch(cmds...)
			}
			return m, nil
		}
	}

	if m.currentState != nil {
		nextState, cmd := m.currentState.Update(m, msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		if nextState != nil && nextState != m.currentState {
			if transitionCmd := m.setState(nextState); transitionCmd != nil {
				cmds = append(cmds, transitionCmd)
			}
		}
	}

	if len(cmds) > 0 {
		return m, tea.Batch(cmds...)
	}

	return m, nil
}

func (m *ConfigModel) handleGlobalKey(msg tea.KeyMsg) (bool, tea.Cmd) {
	noop := func() tea.Msg { return nil }

	switch {
	case key.Matches(msg, m.keyMap.Quit):
		if m.hasChanges {
			m.statusMessage = "Unsaved changes! Press Ctrl+S to save or press q again to force quit"
			m.statusType = "warning"
			return true, noop
		}
		m.quitting = true
		return true, tea.Quit

	case key.Matches(msg, m.keyMap.Help):
		if _, isHelp := m.currentState.(*HelpState); isHelp {
			if m.stateBeforeHelp != nil {
				cmd := m.setState(m.stateBeforeHelp)
				m.stateBeforeHelp = nil
				return true, cmd
			}
			return true, nil
		}
		m.stateBeforeHelp = m.currentState
		return true, m.setState(NewHelpState())

	case key.Matches(msg, m.keyMap.Save):
		return true, m.saveConfig()
	}

	return false, nil
}

func (m *ConfigModel) runAction(actionKey string) tea.Cmd {
	switch actionKey {
	case "save":
		return m.saveConfig()
	case "reset":
		return m.resetConfig()
	}
	return nil
}

func (m *ConfigModel) saveConfig() tea.Cmd {
	return func() tea.Msg {
		if err := config.SaveGlobalConfig(m.globalConfig); err != nil {
			return statusMsg(fmt.Sprintf("Save failed: %v", err))
		}

		m.hasChanges = false
		return statusMsg("Configuration saved")
	}
}

func (m *ConfigModel) resetConfig() tea.Cmd {
	return func() tea.Msg {
		defaultConfig := config.GetDefaultGlobalConfig()
		previousConfig := m.globalConfig

		m.globalConfig = defaultConfig
		m.navigator.UpdateConfig(m.globalConfig)

		if err := config.SaveGlobalConfig(m.globalConfig); err != nil {
			m.globalConfig = previousConfig
			m.navigator.UpdateConfig(m.globalConfig)
			return statusMsg(fmt.Sprintf("Reset failed: %v", err))
		}

		m.hasChanges = false

		if m.currentState != nil {
			if cmd := m.currentState.OnEnter(m); cmd != nil {
				// We intentionally execute the command immediately so any
				// synchronous side effects (e.g. list focus updates) run
				// before the status message is emitted. Future states that
				// return asynchronous commands should update this section
				// to enqueue the resulting message alongside the status.
				cmd()
			}
		}

		return statusMsg("Configuration reset to defaults")
	}
}

func (m *ConfigModel) getFieldValue(key string) (string, error) {
	return getFieldByPath(m.globalConfig, key)
}

func (m *ConfigModel) setFieldValue(key, value string) error {
	return setFieldByPath(m.globalConfig, key, value)
}

// View renders the current view
func (m *ConfigModel) View() string {
	if m.quitting {
		return "Configuration saved!\n"
	}

	var content string
	if m.currentState != nil {
		content = m.currentState.View(m)
	}

	// Header with breadcrumb
	header := m.renderHeader()

	// Footer with status and help
	footer := m.renderFooter()

	// Combine all parts and optionally debug info
	return lipgloss.JoinVertical(lipgloss.Left, header, content, footer)
}

// renderHeader renders the header with title and breadcrumb
func (m *ConfigModel) renderHeader() string {
	title := titleStyle.Render("🔧 Lai Interactive Configuration")

	breadcrumbText := ""
	if len(m.breadcrumb) > 0 {
		breadcrumbText = CreateBreadcrumb(m.breadcrumb)
	}

	statusText := ""
	if m.statusMessage != "" {
		statusText = StatusMessage(m.statusMessage, m.statusType)
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.NewStyle().Width(m.width).Render(title),
		lipgloss.NewStyle().Width(m.width).Render(breadcrumbText),
		lipgloss.NewStyle().Width(m.width).Render(statusText),
	)
}

// renderFooter renders the footer with help text
func (m *ConfigModel) renderFooter() string {
	var helpItems []string

	switch m.currentState.(type) {
	case *FieldEditState:
		helpItems = []string{
			"Enter: save", "Esc: cancel", "Ctrl+S: save config",
		}
	case *HelpState:
		helpItems = []string{
			"?: close help", "q: quit",
		}
	default:
		helpItems = []string{
			"↑/↓: navigate", "Enter: select", "Esc: back", "q: quit", "?: help",
			"/: filter",
		}
	}

	if m.hasChanges {
		helpItems = append(helpItems, "🟡 Unsaved changes")
	}

	helpText := textMutedStyle.Render(strings.Join(helpItems, " • "))
	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), true, false, false, false).
		BorderForeground(borderColor).
		Padding(1, 2).
		Width(m.width).
		Render(helpText)
}

// HasError checks if the model has an error
func (m *ConfigModel) HasError() bool {
	return m.err != nil
}

// GetError returns the current error
func (m *ConfigModel) GetError() error {
	return m.err
}

// Status message type
type statusMsg string

// Helper methods for navigation and state management will be implemented in the next file
// to keep this file manageable...

// (debug ruler removed)
