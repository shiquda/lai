package tui

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
)

// Color palette for the TUI
var (
	// Primary colors
	primaryColor   = lipgloss.Color("#0EA5E9")  // Sky blue
	secondaryColor = lipgloss.Color("#8B5CF6")  // Purple
	accentColor    = lipgloss.Color("#10B981")  // Emerald
	
	// UI colors
	backgroundColor = lipgloss.Color("#1E293B")  // Slate 800
	surfaceColor    = lipgloss.Color("#334155")  // Slate 700
	borderColor     = lipgloss.Color("#475569")  // Slate 600
	
	// Text colors
	textPrimary   = lipgloss.Color("#F8FAFC")  // Slate 50
	textSecondary = lipgloss.Color("#CBD5E1")  // Slate 300
	textMuted     = lipgloss.Color("#94A3B8")  // Slate 400
	
	// Status colors
	successColor = lipgloss.Color("#10B981")  // Emerald 500
	warningColor = lipgloss.Color("#F59E0B")  // Amber 500
	errorColor   = lipgloss.Color("#EF4444")  // Red 500
	infoColor    = lipgloss.Color("#3B82F6")  // Blue 500
)

// Common styles
var (
	// Base styles
	baseStyle = lipgloss.NewStyle().
		Foreground(textPrimary).
		Background(backgroundColor)

	// Title styles
	titleStyle = lipgloss.NewStyle().
		Foreground(primaryColor).
		Bold(true).
		Padding(0, 1)

	headerStyle = lipgloss.NewStyle().
		Foreground(textPrimary).
		Bold(true).
		Padding(0, 1)

	subtitleStyle = lipgloss.NewStyle().
		Foreground(textSecondary).
		Italic(true).
		Padding(0, 1)

	// Content styles - use minimal padding to allow full width
	contentStyle = lipgloss.NewStyle().
		Foreground(textPrimary).
		Padding(0, 1)

	descriptionStyle = lipgloss.NewStyle().
		Foreground(textMuted).
		Padding(0, 1)

	// Interactive styles
	focusedStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(primaryColor).
		Foreground(textPrimary).
		Padding(0, 1)

	selectedStyle = lipgloss.NewStyle().
		Background(primaryColor).
		Foreground(backgroundColor).
		Bold(true).
		Padding(0, 1)

	// Input styles - no fixed width, will be set dynamically
	inputStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Foreground(textPrimary).
		Padding(0, 1)

	inputFocusedStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(primaryColor).
		Foreground(textPrimary).
		Padding(0, 1)

	// Button styles
	buttonStyle = lipgloss.NewStyle().
		Background(surfaceColor).
		Foreground(textPrimary).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 2).
		MarginRight(1)

	buttonActiveStyle = lipgloss.NewStyle().
		Background(primaryColor).
		Foreground(backgroundColor).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(primaryColor).
		Padding(0, 2).
		MarginRight(1).
		Bold(true)

	// Status styles
	successStyle = lipgloss.NewStyle().
		Foreground(successColor).
		Bold(true)

	errorStyle = lipgloss.NewStyle().
		Foreground(errorColor).
		Bold(true)

	warningStyle = lipgloss.NewStyle().
		Foreground(warningColor).
		Bold(true)

	infoStyle = lipgloss.NewStyle().
		Foreground(infoColor).
		Bold(true)

	// Panel styles
	panelStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(1, 2).
		Margin(1, 0)

	activePanelStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(primaryColor).
		Padding(1, 2).
		Margin(1, 0)

	// List styles - use minimal padding to allow full width
	listItemStyle = lipgloss.NewStyle().
		Padding(0, 1).
		MarginBottom(0)

	listItemSelectedStyle = lipgloss.NewStyle().
		Background(primaryColor).
		Foreground(backgroundColor).
		Padding(0, 1).
		MarginBottom(0).
		Bold(true)

	// Help styles
	helpStyle = lipgloss.NewStyle().
		Foreground(textMuted).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(1).
		Margin(1, 0)
		
	// Text utility styles
	textPrimaryStyle = lipgloss.NewStyle().
		Foreground(textPrimary)
		
	textSecondaryStyle = lipgloss.NewStyle().
		Foreground(textSecondary)
		
	textMutedStyle = lipgloss.NewStyle().
		Foreground(textMuted)
		
	successTextStyle = lipgloss.NewStyle().
		Foreground(successColor)
		
	warningTextStyle = lipgloss.NewStyle().
		Foreground(warningColor)
		
	errorTextStyle = lipgloss.NewStyle().
		Foreground(errorColor)
)

// ApplyLevelIndent applies indentation based on configuration level
func ApplyLevelIndent(style lipgloss.Style, level int) lipgloss.Style {
	indent := level * 4 // 4 spaces per level
	return style.PaddingLeft(indent)
}

// StatusMessage creates styled status messages
func StatusMessage(message string, status string) string {
	var style lipgloss.Style
	var prefix string

	switch status {
	case "success":
		style = successStyle
		prefix = "✓ "
	case "error":
		style = errorStyle
		prefix = "✗ "
	case "warning":
		style = warningStyle
		prefix = "⚠ "
	case "info":
		style = infoStyle
		prefix = "ℹ "
	default:
		style = baseStyle
		prefix = ""
	}

	return style.Render(prefix + message)
}

// FormatFieldValue formats field values with appropriate styling
func FormatFieldValue(value string, fieldType string, sensitive bool) string {
	if sensitive && value != "" {
		// Show masked value for sensitive fields
		if len(value) <= 8 {
			return textMutedStyle.Render(lipgloss.PlaceHorizontal(len(value), lipgloss.Left, "●●●●●●●●"[:len(value)]))
		}
		return textMutedStyle.Render(value[:3] + "●●●●●●" + value[len(value)-3:])
	}

	switch fieldType {
	case "bool":
		if value == "true" {
			return successTextStyle.Render("✓ " + value)
		} else if value == "false" {
			return textMutedStyle.Render("✗ " + value)
		}
	case "secret":
		if value != "" {
			return textMutedStyle.Render("●●●●●●●●")
		}
		return textMutedStyle.Render("not set")
	}

	if value == "" {
		return textMutedStyle.Render("not set")
	}

	return textPrimaryStyle.Render(value)
}

// CreateProgressBar creates a simple progress bar
func CreateProgressBar(current, total int, width int) string {
	if total == 0 {
		return ""
	}
	
	progress := float64(current) / float64(total)
	filled := int(progress * float64(width))
	
	var bar []rune
	for i := 0; i < width; i++ {
		if i < filled {
			bar = append(bar, '█')
		} else {
			bar = append(bar, '░')
		}
	}
	
	progressText := lipgloss.JoinHorizontal(lipgloss.Left,
		lipgloss.NewStyle().Foreground(primaryColor).Render(string(bar[:filled])),
		textMutedStyle.Render(string(bar[filled:])),
	)
	
	return lipgloss.JoinHorizontal(lipgloss.Left,
		progressText,
		textSecondaryStyle.Render(" "),
		textSecondaryStyle.Render(lipgloss.PlaceHorizontal(8, lipgloss.Right, 
			fmt.Sprintf("%d/%d", current, total))),
	)
}

// CreateBreadcrumb creates navigation breadcrumb
func CreateBreadcrumb(sections []string) string {
	if len(sections) == 0 {
		return ""
	}
	
	var parts []string
	for i, section := range sections {
		if i == len(sections)-1 {
			// Last section is highlighted
			parts = append(parts, lipgloss.NewStyle().Foreground(primaryColor).Render(section))
		} else {
			parts = append(parts, textSecondaryStyle.Render(section))
		}
	}
	
	return textMutedStyle.Render("📍 ") + lipgloss.JoinHorizontal(lipgloss.Left,
		parts[0],
		textMutedStyle.Render(" → "),
		lipgloss.JoinHorizontal(lipgloss.Left, parts[1:]...))
}