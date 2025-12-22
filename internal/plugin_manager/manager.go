package plugin_manager

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/pelletier/go-toml/v2"
	"github.com/rennerdo30/webencode/pkg/logger"
	"github.com/rennerdo30/webencode/pkg/pluginsdk"
)

type Manifest struct {
	Plugin struct {
		ID   string `toml:"id"`
		Type string `toml:"type"`
		Name string `toml:"name"`
	} `toml:"plugin"`
}

type Manager struct {
	logger    *logger.Logger
	pluginDir string
	clients   map[string]*plugin.Client

	// Registry of loaded plugins
	Auth      map[string]interface{}
	Storage   map[string]interface{}
	Encoder   map[string]interface{}
	Live      map[string]interface{}
	Publisher map[string]interface{}

	// Map of ID to Type for registration
	Types map[string]string
}

func New(l *logger.Logger, pluginDir string) *Manager {
	return &Manager{
		logger:    l,
		pluginDir: pluginDir,
		clients:   make(map[string]*plugin.Client),
		Auth:      make(map[string]interface{}),
		Storage:   make(map[string]interface{}),
		Encoder:   make(map[string]interface{}),
		Live:      make(map[string]interface{}),
		Publisher: make(map[string]interface{}),
		Types:     make(map[string]string),
	}
}

func (m *Manager) LoadAll() error {
	entries, err := os.ReadDir(m.pluginDir)
	if err != nil {
		return fmt.Errorf("failed to read plugin dir: %w", err)
	}

	for _, entry := range entries {
		var manifestPath string
		var binPath string
		var id string

		if entry.IsDir() {
			id = entry.Name()
			manifestPath = filepath.Join(m.pluginDir, id, "plugin.toml")
			binPath = filepath.Join(m.pluginDir, id, id)
		} else if strings.HasSuffix(entry.Name(), ".bin") {
			id = strings.TrimSuffix(entry.Name(), ".bin")
			manifestPath = filepath.Join(m.pluginDir, id+".toml")
			binPath = filepath.Join(m.pluginDir, entry.Name())
		} else {
			continue
		}

		if _, err := os.Stat(manifestPath); err != nil {
			continue
		}

		data, err := os.ReadFile(manifestPath)
		if err != nil {
			continue
		}

		var manifest Manifest
		if err := toml.Unmarshal(data, &manifest); err != nil {
			m.logger.Warn("Failed to decode manifest", "path", manifestPath, "error", err)
			continue
		}

		if manifest.Plugin.ID != "" {
			id = manifest.Plugin.ID
		}

		if _, err := os.Stat(binPath); err != nil {
			continue
		}

		if err := m.loadPlugin(id, binPath, manifest.Plugin.Type); err != nil {
			m.logger.Error("Failed to load plugin", "id", id, "error", err)
		}
	}
	return nil
}

func (m *Manager) loadPlugin(id, path, pluginType string) error {
	m.logger.Info("Loading plugin", "id", id, "path", path, "type", pluginType)

	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: pluginsdk.HandshakeConfig,
		Plugins: map[string]plugin.Plugin{
			"webencode": &pluginsdk.Plugin{},
		},
		Cmd:              exec.Command(path),
		Logger:           hclog.New(&hclog.LoggerOptions{Name: id, Level: hclog.Info}),
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
	})

	rpcClient, err := client.Client()
	if err != nil {
		client.Kill()
		return err
	}

	raw, err := rpcClient.Dispense("webencode")
	if err != nil {
		client.Kill()
		return err
	}

	res := raw.(*pluginsdk.PluginResult)
	m.clients[id] = client
	m.Types[id] = pluginType

	// Register implementations based on manifest type
	switch pluginType {
	case "storage":
		if res.Storage != nil {
			m.Storage[id] = res.Storage
		}
	case "auth":
		if res.Auth != nil {
			m.Auth[id] = res.Auth
		}
	case "encoder":
		if res.Encoder != nil {
			m.Encoder[id] = res.Encoder
		}
	case "live":
		if res.Live != nil {
			m.Live[id] = res.Live
		}
	case "publisher":
		if res.Publisher != nil {
			m.Publisher[id] = res.Publisher
		}
	default:
		m.logger.Warn("Unknown plugin type", "id", id, "type", pluginType)
	}

	return nil
}

func (m *Manager) Shutdown() {
	for _, client := range m.clients {
		client.Kill()
	}
}
