package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/shiquda/lai/internal/config"
)

// Navigation and state management methods for ConfigModel

// initMainMenu initializes the main menu items
func (m *ConfigModel) initMainMenu() {
	items := []list.Item{
		ConfigItem{
			Title:       "🔧 General Settings",
			Description: "Basic configuration options",
			Key:         "general",
			ItemType:    "section",
		},
		ConfigItem{
			Title:       "⚙️ Default Configuration",
			Description: "Application default behavior settings",
			Key:         "defaults",
			ItemType:    "section",
		},
		ConfigItem{
			Title:       "🤖 OpenAI Configuration",
			Description: "AI service related configuration",
			Key:         "openai",
			ItemType:    "section",
		},
		ConfigItem{
			Title:       "📱 Notification Providers",
			Description: "Configure various notification channels",
			Key:         "providers",
			ItemType:    "section",
		},
		ConfigItem{
			Title:       "📝 Logging Configuration",
			Description: "Application logging settings",
			Key:         "logging",
			ItemType:    "section",
		},
		ConfigItem{
			Title:       "💾 Save Configuration",
			Description: "Save all changes to configuration file",
			Key:         "save",
			ItemType:    "action",
		},
		ConfigItem{
			Title:       "🔄 Reset Configuration",
			Description: "Restore default configuration settings",
			Key:         "reset",
			ItemType:    "action",
		},
	}

	m.list.SetItems(items)
	m.items = make([]ConfigItem, len(items))
	for i, item := range items {
		m.items[i] = item.(ConfigItem)
	}
}

// loadSectionFields loads fields for a specific section
func (m *ConfigModel) loadSectionFields(sectionKey string) {
	var items []list.Item

	// Add back navigation item
	items = append(items, ConfigItem{
		Title:       "← Back to Main Menu",
		Description: "Return to main configuration menu",
		Key:         "back",
		ItemType:    "navigation",
	})

	// Find matching sections in metadata
	for _, section := range m.metadata.Sections {
		if strings.HasPrefix(section.Name, sectionKey) ||
			(sectionKey == "providers" && section.Category == config.CategoryProviders) ||
			(sectionKey == "general" && section.Category == config.CategoryGeneral) ||
			(sectionKey == "defaults" && section.Category == config.CategoryDefaults) ||
			(sectionKey == "openai" && section.Category == config.CategoryOpenAI) ||
			(sectionKey == "logging" && section.Category == config.CategoryLogging) {

			// Add section header if it has multiple fields
			if len(section.Fields) > 1 {
				items = append(items, ConfigItem{
					Title:       fmt.Sprintf("📂 %s", section.DisplayName),
					Description: section.Description,
					Key:         section.Name,
					ItemType:    "header",
					Level:       section.Level,
				})
			}

			// Add fields
			for _, field := range section.Fields {
				// Get current value
				currentValue, _ := m.getFieldValue(field.Key)

				displayValue := FormatFieldValue(currentValue, string(field.Type), field.Sensitive)

				item := ConfigItem{
					Title:       field.DisplayName,
					Description: fmt.Sprintf("%s (current: %s)", field.Description, displayValue),
					Key:         field.Key,
					Value:       currentValue,
					ItemType:    "field",
					Level:       field.Level,
					Required:    field.Required,
					Sensitive:   field.Sensitive,
					Editable:    true,
					Metadata:    &field,
				}

				items = append(items, item)
			}
		}
	}

	m.list.SetItems(items)
	m.items = make([]ConfigItem, len(items))
	for i, item := range items {
		m.items[i] = item.(ConfigItem)
	}
}

// updateNavigation handles navigation key presses
func (m *ConfigModel) updateNavigation(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// no-op command to mark key as handled
	noop := func() tea.Msg { return nil }
	switch {
	case key.Matches(msg, m.keyMap.Quit):
		if m.hasChanges {
			m.statusMessage = "Unsaved changes! Press Ctrl+S to save or press q again to force quit"
			m.statusType = "warning"
			return m, noop
		}
		m.quitting = true
		return m, tea.Quit

	case key.Matches(msg, m.keyMap.Help):
		if m.state == ViewHelp {
			m.state = ViewMainMenu
		} else {
			m.state = ViewHelp
		}
		return m, noop

	case key.Matches(msg, m.keyMap.Save):
		return m, m.saveConfig()

	case key.Matches(msg, m.keyMap.Enter):
		model, cmd := m.handleEnterKey()
		if cmd == nil {
			cmd = noop
		}
		return model, cmd

	case key.Matches(msg, m.keyMap.Escape):
		if len(m.breadcrumb) > 1 {
			model, cmd := m.handleBackNavigation()
			if cmd == nil {
				cmd = noop
			}
			return model, cmd
		}
		return m, noop

	case key.Matches(msg, m.keyMap.Left) || key.Matches(msg, m.keyMap.Back):
		model, cmd := m.handleBackNavigation()
		if cmd == nil {
			cmd = noop
		}
		return model, cmd
	}

	// For all other keys (including up/down navigation), let the list component handle them
	// Return the same model without command to signal no action was taken
	return m, nil
}

// updateFieldEdit handles field editing
func (m *ConfigModel) updateFieldEdit(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keyMap.Enter):
		return m.saveFieldEdit()

	case key.Matches(msg, m.keyMap.Escape):
		m.cancelFieldEdit()
		return m, nil

	case key.Matches(msg, m.keyMap.Save):
		// Save field first, then save config
		cmd1 := m.saveFieldEditCmd()
		cmd2 := m.saveConfig()
		return m, tea.Batch(cmd1, cmd2)
	}

	// Handle boolean field toggle
	if m.editingField != nil && m.editingField.Type == config.TypeBool {
		if key.Matches(msg, m.keyMap.Left) || key.Matches(msg, m.keyMap.Right) {
			currentVal := m.textInput.Value()
			if currentVal != "true" && currentVal != "false" {
				m.textInput.SetValue("false")
			} else if currentVal == "true" {
				m.textInput.SetValue("false")
			} else {
				m.textInput.SetValue("true")
			}
			return m, nil
		}
	}

	// For regular text input, let the textInput component handle the key
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

// handleEnterKey handles Enter key press in navigation
func (m *ConfigModel) handleEnterKey() (tea.Model, tea.Cmd) {
	if m.state == ViewHelp {
		m.state = ViewMainMenu
		return m, nil
	}

	selectedItem := m.list.SelectedItem()
	if selectedItem == nil {
		return m, nil
	}

	item := selectedItem.(ConfigItem)

	switch item.ItemType {
	case "section":
		m.loadSectionFields(item.Key)
		m.state = ViewSectionList
		m.breadcrumb = append(m.breadcrumb, item.Title)

	case "field":
		if item.Editable {
			m.startFieldEdit(&item)
		}

	case "navigation":
		if item.Key == "back" {
			return m.handleBackNavigation()
		}

	case "action":
		return m.handleAction(item.Key)
	}

	return m, nil
}

// handleBackNavigation handles back navigation
func (m *ConfigModel) handleBackNavigation() (tea.Model, tea.Cmd) {
	if len(m.breadcrumb) > 1 {
		m.breadcrumb = m.breadcrumb[:len(m.breadcrumb)-1]
		if len(m.breadcrumb) == 1 {
			m.state = ViewMainMenu
			m.initMainMenu()
		} else {
			// Navigate to parent section
			m.state = ViewSectionList
		}
	}
	return m, nil
}

// handleAction handles action items
func (m *ConfigModel) handleAction(actionKey string) (tea.Model, tea.Cmd) {
	switch actionKey {
	case "save":
		return m, m.saveConfig()
	case "reset":
		return m, m.resetConfig()
	}
	return m, nil
}

// startFieldEdit starts editing a configuration field
func (m *ConfigModel) startFieldEdit(item *ConfigItem) {
	m.state = ViewFieldEdit
	m.editingField = item.Metadata
	m.editingValue = item.Value
	m.originalValue = item.Value

	// Set up text input
	m.textInput.SetValue(item.Value)
	m.textInput.Focus()

	// Update breadcrumb
	m.breadcrumb = append(m.breadcrumb, fmt.Sprintf("Edit: %s", item.Title))
}

// saveFieldEdit saves the current field edit
func (m *ConfigModel) saveFieldEdit() (tea.Model, tea.Cmd) {
	return m, m.saveFieldEditCmd()
}

// saveFieldEditCmd returns a command to save field edit
func (m *ConfigModel) saveFieldEditCmd() tea.Cmd {
	return func() tea.Msg {
		if m.editingField == nil {
			return statusMsg("No field is being edited")
		}

		// Store the field name before clearing it
		fieldDisplayName := m.editingField.DisplayName
		newValue := m.textInput.Value()

		// Validate the new value
		if err := m.editingField.ValidateFieldValue(newValue); err != nil {
			return statusMsg(fmt.Sprintf("Validation failed: %v", err))
		}

		// Set the field value
		if err := m.setFieldValue(m.editingField.Key, newValue); err != nil {
			return statusMsg(fmt.Sprintf("Setting failed: %v", err))
		}

		m.hasChanges = true
		m.cancelFieldEdit()

		return statusMsg(fmt.Sprintf("Updated %s", fieldDisplayName))
	}
}

// cancelFieldEdit cancels the current field edit
func (m *ConfigModel) cancelFieldEdit() {
	m.state = ViewSectionList
	m.editingField = nil
	m.editingValue = ""
	m.originalValue = ""

	// Remove edit breadcrumb
	if len(m.breadcrumb) > 0 {
		m.breadcrumb = m.breadcrumb[:len(m.breadcrumb)-1]
	}

	// Reload current section to update display values
	if len(m.breadcrumb) > 1 {
		// Extract section key from breadcrumb
		sectionTitle := m.breadcrumb[len(m.breadcrumb)-1]
		sectionKey := m.getSectionKeyFromTitle(sectionTitle)
		if sectionKey != "" {
			m.loadSectionFields(sectionKey)
		}
	}
}

// saveConfig saves the current configuration
func (m *ConfigModel) saveConfig() tea.Cmd {
	return func() tea.Msg {
		if err := config.SaveGlobalConfig(m.globalConfig); err != nil {
			return statusMsg(fmt.Sprintf("Save failed: %v", err))
		}

		m.hasChanges = false
		return statusMsg("Configuration saved")
	}
}

// resetConfig resets configuration to defaults
func (m *ConfigModel) resetConfig() tea.Cmd {
	return func() tea.Msg {
		defaultConfig := config.GetDefaultGlobalConfig()
		m.globalConfig = defaultConfig

		if err := config.SaveGlobalConfig(m.globalConfig); err != nil {
			return statusMsg(fmt.Sprintf("Reset failed: %v", err))
		}

		m.hasChanges = false
		return statusMsg("Configuration reset to defaults")
	}
}

// Helper methods for getting and setting field values
func (m *ConfigModel) getFieldValue(key string) (string, error) {
	return getFieldByPath(m.globalConfig, key)
}

func (m *ConfigModel) setFieldValue(key, value string) error {
	return setFieldByPath(m.globalConfig, key, value)
}

// getSectionKeyFromTitle extracts section key from title
func (m *ConfigModel) getSectionKeyFromTitle(title string) string {
	titleMap := map[string]string{
		"🔧 General Settings":       "general",
		"⚙️ Default Configuration": "defaults",
		"🤖 OpenAI Configuration":   "openai",
		"📱 Notification Providers": "providers",
		"📝 Logging Configuration":  "logging",
	}

	if key, ok := titleMap[title]; ok {
		return key
	}
	return ""
}
