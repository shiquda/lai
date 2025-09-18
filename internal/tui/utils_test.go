package tui

import (
	"testing"

	"github.com/shiquda/lai/internal/config"
	"github.com/stretchr/testify/require"
)

func TestSetProviderFieldValueUpdatesProviderName(t *testing.T) {
	cfg := config.GetDefaultGlobalConfig()

	err := setProviderFieldValue(cfg, "notifications.providers.discord.provider", "discord_webhook")
	require.NoError(t, err)

	provider, ok := cfg.Notifications.Providers["discord"]
	require.True(t, ok)
	require.Equal(t, "discord_webhook", provider.Provider)
}

func TestSetProviderFieldValueRejectsInvalidProviderName(t *testing.T) {
	cfg := config.GetDefaultGlobalConfig()

	err := setProviderFieldValue(cfg, "notifications.providers.discord.provider", "slack")
	require.Error(t, err)
}

func TestChannelIDsRoundTripThroughProviderFieldHelpers(t *testing.T) {
	cfg := config.GetDefaultGlobalConfig()

	input := "1234567890, 0987654321"
	err := setProviderFieldValue(cfg, "notifications.providers.discord.config.channel_ids", input)
	require.NoError(t, err)

	provider := cfg.Notifications.Providers["discord"]
	rawIDs, ok := provider.Config["channel_ids"].([]interface{})
	require.True(t, ok)
	require.ElementsMatch(t, []interface{}{"1234567890", "0987654321"}, rawIDs)

	value, err := getProviderFieldValue(cfg, "notifications.providers.discord.config.channel_ids")
	require.NoError(t, err)
	require.Equal(t, "1234567890,0987654321", value)
}

func TestGetProviderFieldValueReturnsProviderName(t *testing.T) {
	cfg := config.GetDefaultGlobalConfig()
	cfg.Notifications.Providers["discord"] = config.ServiceConfig{
		Enabled:  true,
		Provider: "discord",
		Config:   map[string]interface{}{"channel_ids": []interface{}{"42"}},
	}

	value, err := getProviderFieldValue(cfg, "notifications.providers.discord.provider")
	require.NoError(t, err)
	require.Equal(t, "discord", value)
}
