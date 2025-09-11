package tui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ConfigDelegate implements the list.ItemDelegate interface for ConfigItem
type ConfigDelegate struct{}

// NewConfigDelegate creates a new ConfigDelegate
func NewConfigDelegate() list.ItemDelegate {
	return ConfigDelegate{}
}

// Height returns the height of list items
func (d ConfigDelegate) Height() int {
	return 2
}

// Width returns the width of list items (if needed)
func (d ConfigDelegate) Width() int {
	// Let the list handle width dynamically
	return 0
}

// Spacing returns the spacing between list items
func (d ConfigDelegate) Spacing() int {
	return 1
}

// Update handles delegate updates
func (d ConfigDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}

// Render renders a list item
func (d ConfigDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(ConfigItem)
	if !ok {
		return
	}

	// Determine if this item is selected
	isSelected := index == m.Index()

	// Create the item display based on type and selection
	var renderedItem string

	switch item.ItemType {
	case "navigation":
		renderedItem = d.renderNavigationItem(item, isSelected)
	case "section":
		renderedItem = d.renderSectionItem(item, isSelected)
	case "header":
		renderedItem = d.renderHeaderItem(item, isSelected)
	case "field":
		renderedItem = d.renderFieldItem(item, isSelected)
	case "action":
		renderedItem = d.renderActionItem(item, isSelected)
	default:
		renderedItem = d.renderDefaultItem(item, isSelected)
	}

	// Pad to list width for full-row visual (avoid narrow perceived width)
	target := m.Width()
	if target > 0 {
		lines := strings.Split(renderedItem, "\n")
		for i, line := range lines {
			// Use lipgloss.Width to account for ANSI + wide chars
			if diff := target - lipgloss.Width(line); diff > 0 {
				lines[i] = line + strings.Repeat(" ", diff)
			}
		}
		renderedItem = strings.Join(lines, "\n")
	}

	fmt.Fprint(w, renderedItem)
}

// renderNavigationItem renders navigation items (like "back")
func (d ConfigDelegate) renderNavigationItem(item ConfigItem, isSelected bool) string {
	title := item.Title
	desc := item.Description

	if isSelected {
		title = selectedStyle.Render(title)
		desc = selectedStyle.Render(desc)
	} else {
		title = textSecondaryStyle.Render(title)
		desc = textMutedStyle.Render(desc)
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		title,
		desc,
	)
}

// renderSectionItem renders section items
func (d ConfigDelegate) renderSectionItem(item ConfigItem, isSelected bool) string {
	title := item.Title
	desc := item.Description

	if isSelected {
		title = selectedStyle.Render(title)
		desc = selectedStyle.Render(desc)
	} else {
		title = headerStyle.Render(title)
		desc = textSecondaryStyle.Render(desc)
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		title,
		desc,
	)
}

// renderHeaderItem renders section header items
func (d ConfigDelegate) renderHeaderItem(item ConfigItem, isSelected bool) string {
	title := item.Title
	desc := item.Description

	// Apply level indentation
	titleStyleToUse := headerStyle
	descStyleToUse := textMutedStyle

	if item.Level > 0 {
		titleStyleToUse = ApplyLevelIndent(titleStyleToUse, item.Level)
		descStyleToUse = ApplyLevelIndent(descStyleToUse, item.Level)
	}

	if isSelected {
		title = selectedStyle.Render(title)
		desc = selectedStyle.Render(desc)
	} else {
		title = titleStyleToUse.Render(title)
		desc = descStyleToUse.Render(desc)
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		title,
		desc,
	)
}

// renderFieldItem renders configuration field items
func (d ConfigDelegate) renderFieldItem(item ConfigItem, isSelected bool) string {
	title := item.Title
	desc := item.Description

	// Add required indicator
	if item.Required {
		title = title + textMutedStyle.Render(" *")
	}

	// Add sensitive indicator
	if item.Sensitive {
		title = title + textMutedStyle.Render(" 🔒")
	}

	// Apply level indentation
	titleStyleToUse := contentStyle
	descStyleToUse := descriptionStyle

	if item.Level > 0 {
		titleStyleToUse = ApplyLevelIndent(titleStyleToUse, item.Level)
		descStyleToUse = ApplyLevelIndent(descStyleToUse, item.Level)
	}

	if isSelected {
		// Highlight selected item
		title = selectedStyle.Render("▶ " + title)
		desc = selectedStyle.Render(desc)
	} else {
		title = titleStyleToUse.Render("  " + title)
		desc = descStyleToUse.Render(desc)
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		title,
		desc,
	)
}

// renderActionItem renders action items (save, reset, etc.)
func (d ConfigDelegate) renderActionItem(item ConfigItem, isSelected bool) string {
	title := item.Title
	desc := item.Description

	if isSelected {
		title = buttonActiveStyle.Render(title)
		desc = selectedStyle.Render(desc)
	} else {
		title = buttonStyle.Render(title)
		desc = textSecondaryStyle.Render(desc)
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		title,
		desc,
	)
}

// renderDefaultItem renders default/fallback items
func (d ConfigDelegate) renderDefaultItem(item ConfigItem, isSelected bool) string {
	title := item.Title
	desc := item.Description

	if isSelected {
		title = selectedStyle.Render(title)
		desc = selectedStyle.Render(desc)
	} else {
		title = contentStyle.Render(title)
		desc = descriptionStyle.Render(desc)
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		title,
		desc,
	)
}

// Additional helper methods for better item rendering
func (d ConfigDelegate) getItemIcon(itemType string) string {
	icons := map[string]string{
		"navigation": "←",
		"section":    "📂",
		"header":     "📋",
		"field":      "⚙️",
		"action":     "🔧",
	}

	if icon, ok := icons[itemType]; ok {
		return icon + " "
	}
	return ""
}

// formatFieldStatus returns status indicators for fields
func (d ConfigDelegate) formatFieldStatus(item ConfigItem) string {
	var indicators []string

	if item.Required {
		indicators = append(indicators, textMutedStyle.Render("required"))
	}

	if item.Sensitive {
		indicators = append(indicators, textMutedStyle.Render("sensitive"))
	}

	if item.Value == "" {
		indicators = append(indicators, warningTextStyle.Render("not set"))
	} else {
		indicators = append(indicators, successTextStyle.Render("configured"))
	}

	if len(indicators) > 0 {
		return " " + textMutedStyle.Render("[") + strings.Join(indicators, textMutedStyle.Render("|")) + textMutedStyle.Render("]")
	}

	return ""
}
