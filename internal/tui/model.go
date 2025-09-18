package tui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
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
}

// FilterState returns the title for list filtering
func (c ConfigItem) FilterValue() string {
	return c.Title
}

// ConfigModel represents the main model for the interactive config
type ConfigModel struct {
	// State management
	state    ViewState
	keyMap   KeyMap
	quitting bool
	width    int
	height   int
	err      error

	// (debug fields removed)

	// Configuration data
	globalConfig *config.GlobalConfig
	metadata     *config.ConfigMetadata

	// Navigation
	breadcrumb  []string
	currentPath string

	// UI components
	list      list.Model
	textInput textinput.Model
	items     []ConfigItem

	// Edit state
	editingField     *config.FieldMetadata
	editingValue     string
	originalValue    string
	hasChanges       bool
	editContext      string // Track context: "provider", "section", or "main"
	editContextKey   string // Additional context key (e.g., provider name or section key)

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

	// Create text input model
	textInputModel := textinput.New()
	textInputModel.Placeholder = "Enter configuration value..."
	textInputModel.Focus()

	model := &ConfigModel{
		state:        ViewMainMenu,
		keyMap:       DefaultKeyMap(),
		globalConfig: globalConfig,
		metadata:     metadata,
		list:         listModel,
		textInput:    textInputModel,
		breadcrumb:   []string{"Main Menu"},
		width:        10, // Will be set properly by WindowSizeMsg
		height:       10, // Will be set properly by WindowSizeMsg
		// debug env feature flags removed
	}

	// Initialize main menu items
	model.initMainMenu()

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

// Update handles messages and updates the model
func (m *ConfigModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.state == ViewFieldEdit {
			return m.updateFieldEdit(msg)
		}

		// Handle navigation keys first, but let list handle up/down navigation
		if newModel, cmd := m.updateNavigation(msg); cmd != nil || newModel != m {
			return newModel, cmd
		}

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

		// Update text input width if needed
		if m.editingField != nil && m.editingField.Type != "bool" {
			inputWidth := m.width - 12
			if inputWidth < 30 {
				inputWidth = 30
			} else if inputWidth > m.width-20 {
				inputWidth = m.width - 20
			}
			m.textInput.Width = inputWidth
		}

		// Remove help style width adjustment
		// m.list.Styles.HelpStyle = hs
	case statusMsg:
		m.statusMessage = string(msg)
		m.statusType = "info"
	}

	// Update list component for navigation states
	if m.state == ViewMainMenu || m.state == ViewSectionList {
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	// Update text input component for non-key messages when in field edit mode
	if m.state == ViewFieldEdit {
		// Only update text input for non-key messages to avoid double handling
		if _, isKeyMsg := msg.(tea.KeyMsg); !isKeyMsg {
			var cmd tea.Cmd
			m.textInput, cmd = m.textInput.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	}

	if len(cmds) > 0 {
		return m, tea.Batch(cmds...)
	}

	return m, nil
}

// View renders the current view
func (m *ConfigModel) View() string {
	if m.quitting {
		return "Configuration saved!\n"
	}

	var content string

	// Main content based on current state
	switch m.state {
	case ViewMainMenu, ViewSectionList:
		// For list views, use dynamic width based on terminal size
		listContentStyle := lipgloss.NewStyle().
			Width(m.width - 4) // Account for minimal padding
		content = listContentStyle.Render(m.list.View())
	case ViewFieldEdit:
		content = m.renderFieldEdit()
	case ViewHelp:
		content = m.renderHelp()
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

	switch m.state {
	case ViewMainMenu, ViewSectionList:
		helpItems = []string{
			"↑/↓: navigate", "Enter: select", "Esc: back", "q: quit", "?: help",
		}
		helpItems = append(helpItems, "/: filter")
	case ViewFieldEdit:
		helpItems = []string{
			"Enter: save", "Esc: cancel", "Ctrl+S: save config",
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

// renderFieldEdit renders the field editing interface
func (m *ConfigModel) renderFieldEdit() string {
	if m.editingField == nil {
		return "Error: No field selected for editing"
	}

	field := m.editingField

	// Field information
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

	// Current value display
	currentValueText := ""
	if m.originalValue != "" {
		displayValue := FormatFieldValue(m.originalValue, string(field.Type), field.Sensitive)
		currentValueText = contentStyle.Render(fmt.Sprintf("Current value: %s", displayValue))
	} else {
		currentValueText = textMutedStyle.Render("Current value: not set")
	}

	// Input field
	inputLabel := contentStyle.Render("New value:")
	inputField := ""

	if field.Type == "bool" {
		// For boolean fields, show toggle options
		trueOption := buttonStyle.Render("true")
		falseOption := buttonStyle.Render("false")
		if m.textInput.Value() == "true" {
			trueOption = buttonActiveStyle.Render("true")
		} else if m.textInput.Value() == "false" {
			falseOption = buttonActiveStyle.Render("false")
		}
		inputField = lipgloss.JoinHorizontal(lipgloss.Left, trueOption, falseOption)
	} else {
		// Compute panel/content widths first so input aligns with panel borders.
		panelOuterWidth := m.width - 4 // must match calculation below for panelWidth
		if panelOuterWidth < 50 {
			panelOuterWidth = 50
		}
		// panel padding horizontal = 2(left)+2(right); borders = 2
		panelContentWidth := panelOuterWidth - 6
		if panelContentWidth < 20 {
			panelContentWidth = 20
		}
		// inputFocusedStyle has border + Padding(0,1) => theoretical extra width = 2(border)+2(padding)=4
		// But actual rendering adds 1 extra column (due to title/prefix styles or runewidth differences), so subtract 1 more for compensation
		inputInnerWidth := panelContentWidth - 5
		if inputInnerWidth < 10 {
			inputInnerWidth = 10
		}
		m.textInput.Width = inputInnerWidth
		if field.Sensitive {
			m.textInput.EchoMode = textinput.EchoPassword
		} else {
			m.textInput.EchoMode = textinput.EchoNormal
		}
		inputField = inputFocusedStyle.Render(m.textInput.View())
		// If still overflow (width exceeds panelContentWidth), decrement until it fits
		for lipgloss.Width(inputField) > panelContentWidth && m.textInput.Width > 5 {
			m.textInput.Width--
			inputField = inputFocusedStyle.Render(m.textInput.View())
		}
	}

	// Examples
	examplesText := ""
	if len(field.Examples) > 0 {
		examples := strings.Join(field.Examples, ", ")
		examplesText = textMutedStyle.Render(fmt.Sprintf("Examples: %s", examples))
	}

	// Discord-specific help
	discordHelp := ""
	if strings.Contains(field.Key, "discord") {
		if strings.Contains(field.Key, "bot_token") {
			discordHelp = textMutedStyle.Render(
				"💡 Tip: Get this from Discord Developer Portal → Your App → Bot → Token\n" +
				"Format should be like: MTE.MjA.XXXXX or similar",
			)
		} else if strings.Contains(field.Key, "webhook_url") {
			discordHelp = textMutedStyle.Render(
				"💡 Tip: Create webhook in Discord channel settings → Integrations\n" +
				"Format: https://discord.com/api/webhooks/123456789/...",
			)
		} else if strings.Contains(field.Key, "channel_ids") {
			discordHelp = textMutedStyle.Render(
				"💡 Tip: Right-click channel in Discord → Copy Channel ID\n" +
				"Format: numeric IDs like 123456789012345678",
			)
		}
	}

	// Create a dynamic panel style based on terminal width
	dynamicPanelStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(1, 2).
		Margin(1, 0)

	// Use full terminal width for panels (outer width already used above for input calc)
	panelWidth := m.width - 4
	if panelWidth < 50 {
		panelWidth = 50
	}

	dynamicPanelStyle = dynamicPanelStyle.Width(panelWidth)

	// Combine all content
	content := []string{
		fieldInfo,
		"",
		currentValueText,
		"",
		inputLabel,
		inputField,
		"",
		examplesText,
	}

	// Add Discord-specific help if available
	if discordHelp != "" {
		content = append(content, "", discordHelp)
	}

	return dynamicPanelStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left, content...),
	)
}

// renderHelp renders the help screen
func (m *ConfigModel) renderHelp() string {
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

### Discord Configuration Guide

#### Discord Bot Mode (Full-featured)
**Features:**
- Support for multiple channels
- Rich message formatting
- Interactive bot commands
- Full Discord API access

**Setup Steps:**
1. Go to Discord Developer Portal (https://discord.com/developers/applications)
2. Create a New Application
3. Go to "Bot" tab and click "Add Bot"
4. Enable Privileged Gateway Intents:
   - MESSAGE CONTENT INTENT
   - SERVER MEMBERS INTENT
5. Copy the Bot Token (format: MTE.MjA... or similar)
6. Invite bot to your server using OAuth2 URL Generator
7. Get Channel IDs from Discord (right-click channel → Copy ID)

**Bot Token Format:** Should match pattern: ABC123.def456.ghi789

#### Discord Webhook Mode (Simple)
**Features:**
- Simple setup
- Single channel support
- No bot required
- Basic message sending

**Setup Steps:**
1. Go to your Discord server channel
2. Click Channel Settings → Integrations
3. Click "Create Webhook"
4. Give it a name (e.g., "Lai Bot")
5. Copy the Webhook URL
6. Webhook URL format: https://discord.com/api/webhooks/ID/TOKEN

**Webhook URL Format:** Should match pattern: https://discord.com/api/webhooks/123456789/...

#### Which Mode to Choose?
- **Choose Discord Bot** if you need multiple channels or advanced features
- **Choose Discord Webhook** for simple single-channel notifications

#### Troubleshooting
- Invalid Bot Token: Check token format and ensure it's not expired
- Invalid Webhook URL: Verify the URL is complete and accessible
- Permission Issues: Ensure bot has "Send Messages" permission in channels
- Channel IDs: Must be numeric IDs, not channel names
`

	return helpStyle.Render(helpContent)
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
