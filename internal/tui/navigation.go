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

	// Special handling for providers section to show channel list first
	if sectionKey == "providers" {
		m.loadProviderChannels(items)
		return
	}

	// Find matching sections in metadata
	for _, section := range m.metadata.Sections {
		if strings.HasPrefix(section.Name, sectionKey) ||
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

// loadProviderChannels loads the list of available notification channels
func (m *ConfigModel) loadProviderChannels(items []list.Item) {
	// Get all provider sections
	providerSections := m.metadata.GetSectionsByCategory(config.CategoryProviders)
	
	for _, section := range providerSections {
		// Get the enabled status for this provider
		enabledKey := fmt.Sprintf("notifications.providers.%s.enabled", section.Name)
		enabled, _ := m.getFieldValue(enabledKey)
		
		// Determine status icon and text
		statusIcon := "❌"
		statusText := "Disabled"
		if enabled == "true" {
			statusIcon = "✅"
			statusText = "Enabled"
		}
		
		// Create channel item
		channelItem := ConfigItem{
			Title:       fmt.Sprintf("%s %s", statusIcon, section.DisplayName),
			Description: fmt.Sprintf("%s - %s", section.Description, statusText),
			Key:         fmt.Sprintf("provider_%s", section.Name),
			ItemType:    "provider_channel",
			Level:       1,
		}
		
		items = append(items, channelItem)
	}
	
	m.list.SetItems(items)
	m.items = make([]ConfigItem, len(items))
	for i, item := range items {
		m.items[i] = item.(ConfigItem)
	}
}

// loadProviderConfig loads configuration fields for a specific provider
func (m *ConfigModel) loadProviderConfig(providerName string) {
	var items []list.Item

	// Add back navigation item
	items = append(items, ConfigItem{
		Title:       "← Back to Channels",
		Description: "Return to notification channels list",
		Key:         "back_to_providers",
		ItemType:    "navigation",
	})

	// Find the provider section
	providerSections := m.metadata.GetSectionsByCategory(config.CategoryProviders)
	var targetSection *config.ConfigSection
	
	for _, section := range providerSections {
		if section.Name == providerName {
			targetSection = &section
			break
		}
	}
	
	if targetSection == nil {
		// Provider not found, return to channels
		m.loadProviderChannels([]list.Item{
			ConfigItem{
				Title:       "← Back to Main Menu",
				Description: "Return to main configuration menu",
				Key:         "back",
				ItemType:    "navigation",
			},
		})
		return
	}

	// Add provider header
	items = append(items, ConfigItem{
		Title:       fmt.Sprintf("📧 %s Configuration", targetSection.DisplayName),
		Description: targetSection.Description,
		Key:         "provider_header",
		ItemType:    "header",
		Level:       1,
	})

	// Add fields for this provider
	for _, field := range targetSection.Fields {
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
		} else if item.Key == "back_to_providers" {
			// Return to provider channels list
			m.loadProviderChannels([]list.Item{
				ConfigItem{
					Title:       "← Back to Main Menu",
					Description: "Return to main configuration menu",
					Key:         "back",
					ItemType:    "navigation",
				},
			})
			// Update breadcrumb
			if len(m.breadcrumb) > 1 {
				m.breadcrumb = m.breadcrumb[:len(m.breadcrumb)-1]
			}
		}

	case "provider_channel":
		// Extract provider name from key (format: "provider_telegram")
		if strings.HasPrefix(item.Key, "provider_") {
			providerName := strings.TrimPrefix(item.Key, "provider_")
			m.loadProviderConfig(providerName)
			m.breadcrumb = append(m.breadcrumb, item.Title)
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

// determineEditContext determines the current editing context based on breadcrumb
func (m *ConfigModel) determineEditContext() {
	// Reset context
	m.editContext = "main"
	m.editContextKey = ""

	if len(m.breadcrumb) >= 2 {
		parentTitle := m.breadcrumb[len(m.breadcrumb)-1]

		// Check if we're in a provider configuration page
		if strings.Contains(parentTitle, "Configuration") {
			m.editContext = "provider"
			m.editContextKey = m.getProviderNameFromBreadcrumb(parentTitle)
			return
		}

		// Check if we're in a regular section
		sectionKey := m.getSectionKeyFromTitle(parentTitle)
		if sectionKey != "" {
			m.editContext = "section"
			m.editContextKey = sectionKey
			return
		}

		// Check if parent is notification providers
		if parentTitle == "📱 Notification Providers" {
			m.editContext = "providers"
			m.editContextKey = "providers"
			return
		}
	}
}

// startFieldEdit starts editing a configuration field
func (m *ConfigModel) startFieldEdit(item *ConfigItem) {
	m.state = ViewFieldEdit
	m.editingField = item.Metadata
	m.editingValue = item.Value
	m.originalValue = item.Value

	// Determine and set edit context based on current breadcrumb
	m.determineEditContext()

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

		// Clear edit context before canceling
		m.editContext = "main"
		m.editContextKey = ""
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

	// Clear edit context
	defer func() {
		m.editContext = "main"
		m.editContextKey = ""
	}()

	// Remove edit breadcrumb
	if len(m.breadcrumb) > 0 {
		m.breadcrumb = m.breadcrumb[:len(m.breadcrumb)-1]
	}

	// Use the stored edit context to determine where to return
	switch m.editContext {
	case "provider":
		if m.editContextKey != "" {
			m.loadProviderConfig(m.editContextKey)
		}
	case "providers":
		m.loadProviderChannels([]list.Item{
			ConfigItem{
				Title:       "← Back to Main Menu",
				Description: "Return to main configuration menu",
				Key:         "back",
				ItemType:    "navigation",
			},
		})
	case "section":
		if m.editContextKey != "" {
			m.loadSectionFields(m.editContextKey)
		}
	default:
		// Fallback: try to determine from breadcrumb
		if len(m.breadcrumb) > 1 {
			parentTitle := m.breadcrumb[len(m.breadcrumb)-1]

			// If parent is "📱 Notification Providers", return to provider channels
			if parentTitle == "📱 Notification Providers" {
				m.loadProviderChannels([]list.Item{
					ConfigItem{
						Title:       "← Back to Main Menu",
						Description: "Return to main configuration menu",
						Key:         "back",
						ItemType:    "navigation",
					},
				})
				return
			}

			// If parent contains "Configuration", return to provider config
			if strings.Contains(parentTitle, "Configuration") {
				providerName := m.getProviderNameFromBreadcrumb(parentTitle)
				if providerName != "" {
					m.loadProviderConfig(providerName)
					return
				}
			}

			// Try regular section
			sectionKey := m.getSectionKeyFromTitle(parentTitle)
			if sectionKey != "" {
				m.loadSectionFields(sectionKey)
				return
			}
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

// getProviderNameFromBreadcrumb extracts provider name from breadcrumb title
func (m *ConfigModel) getProviderNameFromBreadcrumb(title string) string {
	// Extract provider name from titles like "✅ Telegram Configuration" or "❌ Email Configuration"
	
	// Known provider mappings
	providerMap := map[string]string{
		"Telegram Configuration":    "telegram",
		"Email Configuration":       "email",
		"Discord Bot Configuration":  "discord",
		"Discord Webhook Configuration": "discord_webhook",
		"Slack Configuration":       "slack",
	}
	
	// Remove status icons and extract clean title
	cleanTitle := strings.TrimSpace(title)
	for _, prefix := range []string{"✅ ", "❌ ", "📧 "} {
		cleanTitle = strings.TrimPrefix(cleanTitle, prefix)
	}
	
	// Check if we have a direct mapping
	if providerName, ok := providerMap[cleanTitle]; ok {
		return providerName
	}
	
	// Try to extract from patterns like "ProviderName Configuration"
	if strings.HasSuffix(cleanTitle, " Configuration") {
		providerName := strings.TrimSuffix(cleanTitle, " Configuration")
		// Convert to lowercase for consistency
		return strings.ToLower(providerName)
	}
	
	return ""
}
