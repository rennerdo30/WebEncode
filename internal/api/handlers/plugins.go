package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rennerdo30/webencode/internal/plugin_manager"
	"github.com/rennerdo30/webencode/pkg/db/store"
	"github.com/rennerdo30/webencode/pkg/errors"
	"github.com/rennerdo30/webencode/pkg/logger"
)

type PluginsHandler struct {
	db     store.Querier
	pm     *plugin_manager.Manager
	logger *logger.Logger
}

func NewPluginsHandler(db store.Querier, pm *plugin_manager.Manager, l *logger.Logger) *PluginsHandler {
	return &PluginsHandler{db: db, pm: pm, logger: l}
}

func (h *PluginsHandler) Register(r chi.Router) {
	r.Get("/v1/plugins", h.ListPlugins)
	r.Get("/v1/plugins/{id}", h.GetPlugin)
	r.Put("/v1/plugins/{id}", h.UpdatePlugin)
	r.Post("/v1/plugins/{id}/enable", h.EnablePlugin)
	r.Post("/v1/plugins/{id}/disable", h.DisablePlugin)
}

// PluginStatus represents the runtime status of a plugin
type PluginStatus struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Config      map[string]interface{} `json:"config,omitempty"`
	IsEnabled   bool                   `json:"is_enabled"`
	Priority    int                    `json:"priority"`
	Health      string                 `json:"health"` // healthy, degraded, failed, disabled
	Version     string                 `json:"version,omitempty"`
	LastChecked string                 `json:"last_checked,omitempty"`
}

func (h *PluginsHandler) ListPlugins(w http.ResponseWriter, r *http.Request) {
	plugins := []PluginStatus{}

	// If we have a plugin manager, get plugins from it (auto-discovery)
	if h.pm != nil {
		for id, pluginType := range h.pm.Types {
			// Check if plugin exists in DB, if not auto-register it
			cfg, err := h.db.GetPluginConfig(r.Context(), id)
			if err != nil {
				// Plugin not in DB, auto-register it
				defaultConfig := []byte("{}")
				h.db.RegisterPluginConfig(r.Context(), store.RegisterPluginConfigParams{
					ID:         id,
					PluginType: pluginType,
					ConfigJson: defaultConfig,
				})

				plugins = append(plugins, PluginStatus{
					ID:        id,
					Type:      pluginType,
					Config:    map[string]interface{}{},
					IsEnabled: true,
					Priority:  0,
					Health:    "healthy", // Running in manager = healthy
				})
				continue
			}

			var configMap map[string]interface{}
			json.Unmarshal(cfg.ConfigJson, &configMap)

			plugins = append(plugins, PluginStatus{
				ID:        cfg.ID,
				Type:      cfg.PluginType,
				Config:    configMap,
				IsEnabled: cfg.IsEnabled.Bool,
				Priority:  int(cfg.Priority.Int32),
				Health:    "healthy", // Running in manager = healthy
			})
		}
	}

	// Also include any plugins configured in DB but not currently loaded
	configs, err := h.db.ListPluginConfigs(r.Context())
	if err != nil {
		h.logger.Error("Failed to list plugins", "error", err)
		errors.Response(w, r, errors.ErrInternal)
		return
	}

	// Add DB plugins that aren't already in the list
	existingIDs := make(map[string]bool)
	for _, p := range plugins {
		existingIDs[p.ID] = true
	}

	for _, cfg := range configs {
		if existingIDs[cfg.ID] {
			continue
		}

		var configMap map[string]interface{}
		json.Unmarshal(cfg.ConfigJson, &configMap)

		health := "disabled"
		if cfg.IsEnabled.Bool {
			health = "failed" // Enabled in DB but not loaded = failed
		}

		plugins = append(plugins, PluginStatus{
			ID:        cfg.ID,
			Type:      cfg.PluginType,
			Config:    configMap,
			IsEnabled: cfg.IsEnabled.Bool,
			Priority:  int(cfg.Priority.Int32),
			Health:    health,
		})
	}

	if plugins == nil {
		plugins = []PluginStatus{}
	}

	if err := json.NewEncoder(w).Encode(plugins); err != nil {
		h.logger.Error("Failed to encode response", "error", err)
	}
}

func (h *PluginsHandler) GetPlugin(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	cfg, err := h.db.GetPluginConfig(r.Context(), id)
	if err != nil {
		errors.Response(w, r, errors.ErrNotFound)
		return
	}

	var configMap map[string]interface{}
	json.Unmarshal(cfg.ConfigJson, &configMap)

	health := "disabled"
	if cfg.IsEnabled.Bool {
		health = "healthy"
		if h.pm != nil {
			if _, ok := h.pm.Types[cfg.ID]; !ok {
				health = "failed"
			}
		}
	}

	if err := json.NewEncoder(w).Encode(PluginStatus{
		ID:        cfg.ID,
		Type:      cfg.PluginType,
		Config:    configMap,
		IsEnabled: cfg.IsEnabled.Bool,
		Priority:  int(cfg.Priority.Int32),
		Health:    health,
	}); err != nil {
		h.logger.Error("Failed to encode response", "error", err)
	}
}

type UpdatePluginRequest struct {
	Config   map[string]interface{} `json:"config"`
	Priority int                    `json:"priority"`
}

func (h *PluginsHandler) UpdatePlugin(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req UpdatePluginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.Response(w, r, errors.ErrInvalidParams)
		return
	}

	configBytes, _ := json.Marshal(req.Config)

	err := h.db.UpdatePluginConfig(r.Context(), store.UpdatePluginConfigParams{
		ID:         id,
		ConfigJson: configBytes,
	})
	if err != nil {
		h.logger.Error("Failed to update plugin", "error", err)
		errors.Response(w, r, errors.ErrInternal)
		return
	}

	if err := json.NewEncoder(w).Encode(map[string]string{"status": "updated"}); err != nil {
		h.logger.Error("Failed to encode response", "error", err)
	}
}

func (h *PluginsHandler) EnablePlugin(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	err := h.db.EnablePlugin(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to enable plugin", "error", err)
		errors.Response(w, r, errors.ErrInternal)
		return
	}

	if err := json.NewEncoder(w).Encode(map[string]string{"status": "enabled"}); err != nil {
		h.logger.Error("Failed to encode response", "error", err)
	}
}

func (h *PluginsHandler) DisablePlugin(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	err := h.db.DisablePlugin(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to disable plugin", "error", err)
		errors.Response(w, r, errors.ErrInternal)
		return
	}

	if err := json.NewEncoder(w).Encode(map[string]string{"status": "disabled"}); err != nil {
		h.logger.Error("Failed to encode response", "error", err)
	}
}
