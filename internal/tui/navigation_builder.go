package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/list"

	"github.com/shiquda/lai/internal/config"
)

// SectionDescriptor describes a top-level configuration section for navigation.
type SectionDescriptor struct {
	Key         string
	Title       string
	Description string
	Category    config.Category
}

// ProviderDescriptor contains metadata for provider navigation nodes.
type ProviderDescriptor struct {
	Name        string
	DisplayName string
	Description string
}

// NavigationBuilder builds navigation trees based on configuration metadata.
type NavigationBuilder struct {
	metadata  *config.ConfigMetadata
	config    *config.GlobalConfig
	sections  map[string]SectionDescriptor
	mainOrder []string
	providers map[string]ProviderDescriptor
}

// NewNavigationBuilder creates a new navigation builder instance.
func NewNavigationBuilder(metadata *config.ConfigMetadata, cfg *config.GlobalConfig) *NavigationBuilder {
	builder := &NavigationBuilder{
		metadata:  metadata,
		config:    cfg,
		sections:  make(map[string]SectionDescriptor),
		mainOrder: []string{"general", "defaults", "openai", "providers", "logging"},
		providers: make(map[string]ProviderDescriptor),
	}

	for _, descriptor := range defaultSectionDescriptors() {
		builder.sections[descriptor.Key] = descriptor
	}

	for _, section := range metadata.GetSectionsByCategory(config.CategoryProviders) {
		builder.providers[section.Name] = ProviderDescriptor{
			Name:        section.Name,
			DisplayName: section.DisplayName,
			Description: section.Description,
		}
	}

	return builder
}

// UpdateConfig swaps the backing configuration reference used when deriving
// navigation items. It should be called whenever the global configuration
// pointer changes so subsequent value lookups reflect the latest data.
func (b *NavigationBuilder) UpdateConfig(cfg *config.GlobalConfig) {
	b.config = cfg
}

// MainMenuItems returns the items for the main menu view.
func (b *NavigationBuilder) MainMenuItems() []ConfigItem {
	var items []ConfigItem
	for _, key := range b.mainOrder {
		descriptor, ok := b.sections[key]
		if !ok {
			continue
		}
		items = append(items, ConfigItem{
			Title:       descriptor.Title,
			Description: descriptor.Description,
			Key:         descriptor.Key,
			ItemType:    "section",
			Section:     cloneSectionDescriptor(descriptor),
		})
	}

	items = append(items,
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
	)

	return items
}

// SectionDescriptorByKey returns the descriptor for a given section key.
func (b *NavigationBuilder) SectionDescriptorByKey(key string) (SectionDescriptor, bool) {
	descriptor, ok := b.sections[key]
	return descriptor, ok
}

// SectionItems builds the list items for a non-provider section.
func (b *NavigationBuilder) SectionItems(descriptor SectionDescriptor) []ConfigItem {
	items := []ConfigItem{
		{
			Title:       "← Back to Main Menu",
			Description: "Return to main configuration menu",
			Key:         "back",
			ItemType:    "navigation",
		},
	}

	for _, section := range b.matchingSections(descriptor) {
		if len(section.Fields) > 1 {
			items = append(items, ConfigItem{
				Title:       fmt.Sprintf("📂 %s", section.DisplayName),
				Description: section.Description,
				Key:         section.Name,
				ItemType:    "header",
				Level:       section.Level,
				Section:     cloneSectionDescriptor(descriptor),
			})
		}

		for _, field := range section.Fields {
			fieldCopy := field
			currentValue := b.fieldValue(fieldCopy.Key)
			displayValue := FormatFieldValue(currentValue, string(fieldCopy.Type), fieldCopy.Sensitive)

			items = append(items, ConfigItem{
				Title:       fieldCopy.DisplayName,
				Description: fmt.Sprintf("%s (current: %s)", fieldCopy.Description, displayValue),
				Key:         fieldCopy.Key,
				Value:       currentValue,
				ItemType:    "field",
				Level:       fieldCopy.Level,
				Required:    fieldCopy.Required,
				Sensitive:   fieldCopy.Sensitive,
				Editable:    true,
				Metadata:    &fieldCopy,
				Section:     cloneSectionDescriptor(descriptor),
			})
		}
	}

	return items
}

// ProviderListItems returns items for the provider selection state.
func (b *NavigationBuilder) ProviderListItems() []ConfigItem {
	items := []ConfigItem{
		{
			Title:       "← Back to Main Menu",
			Description: "Return to main configuration menu",
			Key:         "back",
			ItemType:    "navigation",
		},
	}

	var providerKeys []string
	for key := range b.providers {
		providerKeys = append(providerKeys, key)
	}
	sort.SliceStable(providerKeys, func(i, j int) bool {
		return b.providers[providerKeys[i]].DisplayName < b.providers[providerKeys[j]].DisplayName
	})

	for _, key := range providerKeys {
		descriptor := b.providers[key]
		enabled := b.isProviderEnabled(descriptor.Name)
		statusIcon := "❌"
		statusText := "Disabled"
		if enabled {
			statusIcon = "✅"
			statusText = "Enabled"
		}

		providerCopy := descriptor
		items = append(items, ConfigItem{
			Title:       fmt.Sprintf("%s %s", statusIcon, descriptor.DisplayName),
			Description: fmt.Sprintf("%s - %s", descriptor.Description, statusText),
			Key:         fmt.Sprintf("provider_%s", descriptor.Name),
			ItemType:    "provider_channel",
			Level:       1,
			Provider:    &providerCopy,
		})
	}

	return items
}

// ProviderConfigItems returns configuration items for a provider.
func (b *NavigationBuilder) ProviderConfigItems(descriptor ProviderDescriptor) []ConfigItem {
	items := []ConfigItem{
		{
			Title:       "← Back to Channels",
			Description: "Return to notification channels list",
			Key:         "back_to_providers",
			ItemType:    "navigation",
			Provider:    cloneProviderDescriptor(descriptor),
		},
	}

	section := b.providerSection(descriptor.Name)
	if section == nil {
		return items
	}

	items = append(items, ConfigItem{
		Title:       fmt.Sprintf("📧 %s Configuration", section.DisplayName),
		Description: section.Description,
		Key:         "provider_header",
		ItemType:    "header",
		Level:       1,
		Provider:    cloneProviderDescriptor(descriptor),
	})

	for _, field := range section.Fields {
		fieldCopy := field
		currentValue := b.fieldValue(fieldCopy.Key)
		displayValue := FormatFieldValue(currentValue, string(fieldCopy.Type), fieldCopy.Sensitive)

		items = append(items, ConfigItem{
			Title:       fieldCopy.DisplayName,
			Description: fmt.Sprintf("%s (current: %s)", fieldCopy.Description, displayValue),
			Key:         fieldCopy.Key,
			Value:       currentValue,
			ItemType:    "field",
			Level:       fieldCopy.Level,
			Required:    fieldCopy.Required,
			Sensitive:   fieldCopy.Sensitive,
			Editable:    true,
			Metadata:    &fieldCopy,
			Provider:    cloneProviderDescriptor(descriptor),
		})
	}

	return items
}

// ProviderDescriptorByName retrieves a provider descriptor by name.
func (b *NavigationBuilder) ProviderDescriptorByName(name string) (ProviderDescriptor, bool) {
	descriptor, ok := b.providers[name]
	return descriptor, ok
}

// IsProviderEnabled reports whether the given provider is enabled in the current configuration.
func (b *NavigationBuilder) IsProviderEnabled(name string) bool {
	return b.isProviderEnabled(name)
}

func (b *NavigationBuilder) matchingSections(descriptor SectionDescriptor) []config.ConfigSection {
	var sections []config.ConfigSection
	for _, section := range b.metadata.Sections {
		if section.Category == descriptor.Category || section.Name == descriptor.Key || strings.HasPrefix(section.Name, descriptor.Key) {
			sections = append(sections, section)
		}
	}
	return sections
}

func (b *NavigationBuilder) providerSection(name string) *config.ConfigSection {
	for _, section := range b.metadata.GetSectionsByCategory(config.CategoryProviders) {
		if section.Name == name {
			sectionCopy := section
			return &sectionCopy
		}
	}
	return nil
}

func (b *NavigationBuilder) fieldValue(key string) string {
	value, err := getFieldByPath(b.config, key)
	if err != nil {
		return ""
	}
	return value
}

func (b *NavigationBuilder) isProviderEnabled(name string) bool {
	key := fmt.Sprintf("notifications.providers.%s.enabled", name)
	value, err := getFieldByPath(b.config, key)
	if err != nil {
		return false
	}
	return value == "true"
}

func cloneSectionDescriptor(descriptor SectionDescriptor) *SectionDescriptor {
	copy := descriptor
	return &copy
}

func cloneProviderDescriptor(descriptor ProviderDescriptor) *ProviderDescriptor {
	copy := descriptor
	return &copy
}

func configItemsToListItems(items []ConfigItem) []list.Item {
	listItems := make([]list.Item, len(items))
	for i := range items {
		listItems[i] = items[i]
	}
	return listItems
}

func defaultSectionDescriptors() []SectionDescriptor {
	return []SectionDescriptor{
		{
			Key:         "general",
			Title:       "🔧 General Settings",
			Description: "Basic configuration options",
			Category:    config.CategoryGeneral,
		},
		{
			Key:         "defaults",
			Title:       "⚙️ Default Configuration",
			Description: "Application default behavior settings",
			Category:    config.CategoryDefaults,
		},
		{
			Key:         "openai",
			Title:       "🤖 OpenAI Configuration",
			Description: "AI service related configuration",
			Category:    config.CategoryOpenAI,
		},
		{
			Key:         "providers",
			Title:       "📱 Notification Providers",
			Description: "Configure various notification channels",
			Category:    config.CategoryProviders,
		},
		{
			Key:         "logging",
			Title:       "📝 Logging Configuration",
			Description: "Application logging settings",
			Category:    config.CategoryLogging,
		},
	}
}
