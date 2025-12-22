package plugin_manager

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pelletier/go-toml/v2"
	"github.com/rennerdo30/webencode/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	l := logger.New("test")
	m := New(l, "/tmp/plugins")

	assert.NotNil(t, m)
	assert.NotNil(t, m.logger)
	assert.Equal(t, "/tmp/plugins", m.pluginDir)
	assert.NotNil(t, m.clients)
	assert.NotNil(t, m.Auth)
	assert.NotNil(t, m.Storage)
	assert.NotNil(t, m.Encoder)
	assert.NotNil(t, m.Live)
	assert.NotNil(t, m.Publisher)
	assert.NotNil(t, m.Types)
}

func TestManager_EmptyMaps(t *testing.T) {
	l := logger.New("test")
	m := New(l, "/tmp/plugins")

	assert.Len(t, m.clients, 0)
	assert.Len(t, m.Auth, 0)
	assert.Len(t, m.Storage, 0)
	assert.Len(t, m.Encoder, 0)
	assert.Len(t, m.Live, 0)
	assert.Len(t, m.Publisher, 0)
	assert.Len(t, m.Types, 0)
}

func TestManager_LoadAll_NonExistentDir(t *testing.T) {
	l := logger.New("test")
	m := New(l, "/nonexistent/plugin/dir")

	err := m.LoadAll()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read plugin dir")
}

func TestManager_LoadAll_EmptyDir(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "plugin-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	l := logger.New("test")
	m := New(l, tmpDir)

	err = m.LoadAll()
	assert.NoError(t, err)
	assert.Len(t, m.clients, 0)
}

func TestManager_LoadAll_NoManifest(t *testing.T) {
	// Create a temporary directory with a plugin subdirectory but no manifest
	tmpDir, err := os.MkdirTemp("", "plugin-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create plugin directory without manifest
	pluginDir := filepath.Join(tmpDir, "test-plugin")
	err = os.MkdirAll(pluginDir, 0755)
	require.NoError(t, err)

	l := logger.New("test")
	m := New(l, tmpDir)

	err = m.LoadAll()
	assert.NoError(t, err)
	assert.Len(t, m.clients, 0) // No plugins loaded because no manifest
}

func TestManager_LoadAll_InvalidManifest(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "plugin-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create plugin directory with invalid manifest
	pluginDir := filepath.Join(tmpDir, "test-plugin")
	err = os.MkdirAll(pluginDir, 0755)
	require.NoError(t, err)

	// Write invalid TOML
	manifestPath := filepath.Join(pluginDir, "plugin.toml")
	err = os.WriteFile(manifestPath, []byte("invalid { toml content"), 0644)
	require.NoError(t, err)

	l := logger.New("test")
	m := New(l, tmpDir)

	err = m.LoadAll()
	assert.NoError(t, err)
	assert.Len(t, m.clients, 0) // No plugins loaded because invalid manifest
}

func TestManager_LoadAll_NoBinary(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "plugin-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create plugin directory with valid manifest but no binary
	pluginDir := filepath.Join(tmpDir, "test-plugin")
	err = os.MkdirAll(pluginDir, 0755)
	require.NoError(t, err)

	// Write valid manifest
	manifest := Manifest{}
	manifest.Plugin.ID = "test-plugin"
	manifest.Plugin.Type = "storage"
	manifest.Plugin.Name = "Test Plugin"

	data, err := toml.Marshal(manifest)
	require.NoError(t, err)

	manifestPath := filepath.Join(pluginDir, "plugin.toml")
	err = os.WriteFile(manifestPath, data, 0644)
	require.NoError(t, err)

	l := logger.New("test")
	m := New(l, tmpDir)

	err = m.LoadAll()
	assert.NoError(t, err)
	assert.Len(t, m.clients, 0) // No plugins loaded because no binary
}

func TestManager_LoadAll_BinFile_NoManifest(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "plugin-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create a .bin file without corresponding manifest
	binPath := filepath.Join(tmpDir, "test-plugin.bin")
	err = os.WriteFile(binPath, []byte("fake binary"), 0755)
	require.NoError(t, err)

	l := logger.New("test")
	m := New(l, tmpDir)

	err = m.LoadAll()
	assert.NoError(t, err)
	assert.Len(t, m.clients, 0) // No plugins loaded because no manifest
}

func TestManager_Shutdown_Empty(t *testing.T) {
	l := logger.New("test")
	m := New(l, "/tmp/plugins")

	// Should not panic with no clients
	m.Shutdown()
	assert.Len(t, m.clients, 0)
}

func TestManifest_Unmarshal(t *testing.T) {
	tomlData := `
[plugin]
id = "my-plugin"
type = "storage"
name = "My Storage Plugin"
`
	var manifest Manifest
	err := toml.Unmarshal([]byte(tomlData), &manifest)

	assert.NoError(t, err)
	assert.Equal(t, "my-plugin", manifest.Plugin.ID)
	assert.Equal(t, "storage", manifest.Plugin.Type)
	assert.Equal(t, "My Storage Plugin", manifest.Plugin.Name)
}

func TestManifest_AllTypes(t *testing.T) {
	types := []string{"storage", "auth", "encoder", "live", "publisher"}

	for _, pluginType := range types {
		t.Run(pluginType, func(t *testing.T) {
			manifest := Manifest{}
			manifest.Plugin.ID = "test-" + pluginType
			manifest.Plugin.Type = pluginType
			manifest.Plugin.Name = "Test " + pluginType

			data, err := toml.Marshal(manifest)
			assert.NoError(t, err)

			var parsed Manifest
			err = toml.Unmarshal(data, &parsed)
			assert.NoError(t, err)
			assert.Equal(t, pluginType, parsed.Plugin.Type)
		})
	}
}

func TestManager_PluginTypeRegistration(t *testing.T) {
	l := logger.New("test")
	m := New(l, "/tmp/plugins")

	// Manually set types to verify map operations
	m.Types["plugin-1"] = "storage"
	m.Types["plugin-2"] = "auth"
	m.Types["plugin-3"] = "encoder"

	assert.Equal(t, "storage", m.Types["plugin-1"])
	assert.Equal(t, "auth", m.Types["plugin-2"])
	assert.Equal(t, "encoder", m.Types["plugin-3"])
	assert.Equal(t, "", m.Types["nonexistent"])
}

func TestManager_LoadAll_SkipsRegularFiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "plugin-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create regular files that should be skipped
	files := []string{"readme.txt", "config.json", "script.sh"}
	for _, f := range files {
		err := os.WriteFile(filepath.Join(tmpDir, f), []byte("content"), 0644)
		require.NoError(t, err)
	}

	l := logger.New("test")
	m := New(l, tmpDir)

	err = m.LoadAll()
	assert.NoError(t, err)
	assert.Len(t, m.clients, 0)
}

func TestManager_LoadAll_ManifestIDOverride(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "plugin-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create plugin directory with manifest that overrides ID
	pluginDir := filepath.Join(tmpDir, "directory-name")
	err = os.MkdirAll(pluginDir, 0755)
	require.NoError(t, err)

	// Manifest has different ID than directory name
	manifest := Manifest{}
	manifest.Plugin.ID = "manifest-id"
	manifest.Plugin.Type = "storage"
	manifest.Plugin.Name = "Test Plugin"

	data, err := toml.Marshal(manifest)
	require.NoError(t, err)

	manifestPath := filepath.Join(pluginDir, "plugin.toml")
	err = os.WriteFile(manifestPath, data, 0644)
	require.NoError(t, err)

	// The plugin won't load because there's no binary, but we test
	// that the code path handles manifest ID override
	l := logger.New("test")
	m := New(l, tmpDir)

	err = m.LoadAll()
	assert.NoError(t, err)
}

func TestManifest_EmptyFields(t *testing.T) {
	tomlData := `[plugin]`

	var manifest Manifest
	err := toml.Unmarshal([]byte(tomlData), &manifest)

	assert.NoError(t, err)
	assert.Equal(t, "", manifest.Plugin.ID)
	assert.Equal(t, "", manifest.Plugin.Type)
	assert.Equal(t, "", manifest.Plugin.Name)
}

func TestManifest_Marshal(t *testing.T) {
	manifest := Manifest{}
	manifest.Plugin.ID = "test-id"
	manifest.Plugin.Type = "encoder"
	manifest.Plugin.Name = "Test Encoder"

	data, err := toml.Marshal(manifest)
	assert.NoError(t, err)
	assert.Contains(t, string(data), "test-id")
	assert.Contains(t, string(data), "encoder")
	assert.Contains(t, string(data), "Test Encoder")
}
