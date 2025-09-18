package tui

import (
	"os"
	"strconv"
	"testing"

	"github.com/shiquda/lai/internal/config"
	"github.com/shiquda/lai/internal/testutils"
	"github.com/stretchr/testify/require"
)

func TestNavigationBuilderUpdateConfig(t *testing.T) {
	metadata := config.GetConfigMetadata()
	initialConfig := config.GetDefaultGlobalConfig()

	builder := NewNavigationBuilder(metadata, initialConfig)

	defaultsKey := "defaults.line_threshold"
	initialValue := builder.fieldValue(defaultsKey)
	require.Equal(t, strconv.Itoa(initialConfig.Defaults.LineThreshold), initialValue)

	updatedConfig := config.GetDefaultGlobalConfig()
	updatedConfig.Defaults.LineThreshold = 42

	builder.UpdateConfig(updatedConfig)
	require.Same(t, updatedConfig, builder.config)
	require.Equal(t, "42", builder.fieldValue(defaultsKey))
}

func TestConfigModelResetConfigRefreshesState(t *testing.T) {
	tempDir, cleanup := testutils.CreateTempDir(t)
	defer cleanup()

	originalHome, hasHome := os.LookupEnv("HOME")
	require.NoError(t, os.Setenv("HOME", tempDir))
	defer func() {
		if hasHome {
			require.NoError(t, os.Setenv("HOME", originalHome))
		} else {
			require.NoError(t, os.Unsetenv("HOME"))
		}
	}()

	model, err := NewConfigModel()
	require.NoError(t, err)

	defaultsDescriptor, ok := model.navigator.SectionDescriptorByKey("defaults")
	require.True(t, ok)

	require.NoError(t, model.setFieldValue("defaults.line_threshold", "25"))
	model.hasChanges = true

	model.setState(NewSectionState(defaultsDescriptor))

	beforeReset := findItemByKey(t, model.items, "defaults.line_threshold")
	require.Equal(t, "25", beforeReset.Value)

	resetCmd := model.resetConfig()
	msg := resetCmd()
	require.Equal(t, statusMsg("Configuration reset to defaults"), msg)

	require.False(t, model.hasChanges)
	require.Same(t, model.globalConfig, model.navigator.config)

	afterReset := findItemByKey(t, model.items, "defaults.line_threshold")
	expectedDefault := strconv.Itoa(config.GetDefaultGlobalConfig().Defaults.LineThreshold)
	require.Equal(t, expectedDefault, afterReset.Value)
}

func findItemByKey(t *testing.T, items []ConfigItem, key string) ConfigItem {
	t.Helper()

	for _, item := range items {
		if item.Key == key {
			return item
		}
	}

	t.Fatalf("item with key %s not found", key)
	return ConfigItem{}
}
