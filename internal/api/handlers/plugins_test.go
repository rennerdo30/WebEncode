package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rennerdo30/webencode/internal/plugin_manager"
	"github.com/rennerdo30/webencode/pkg/db/store"
	"github.com/rennerdo30/webencode/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPluginsHandler_ListPlugins(t *testing.T) {
	mockDB := new(MockStore)

	// Create mock Plugin Manager with one discovered plugin
	pm := &plugin_manager.Manager{
		Types: map[string]string{
			"discovered-plugin": "storage",
		},
	}

	handler := NewPluginsHandler(mockDB, pm, logger.New("test"))

	// Mock DB behavior
	// 1. GetPluginConfig called for discovered plugin -> return error (not registered yet)
	mockDB.On("GetPluginConfig", mock.Anything, "discovered-plugin").Return(store.PluginConfig{}, errors.New("not found"))

	// 2. RegisterPluginConfig called for discovered plugin
	mockDB.On("RegisterPluginConfig", mock.Anything, mock.MatchedBy(func(arg store.RegisterPluginConfigParams) bool {
		return arg.ID == "discovered-plugin" && arg.PluginType == "storage"
	})).Return(store.PluginConfig{}, nil)

	// 3. ListPluginConfigs for other plugins
	mockDB.On("ListPluginConfigs", mock.Anything).Return([]store.PluginConfig{
		{
			ID:         "db-plugin",
			PluginType: "auth",
			ConfigJson: []byte(`{}`),
			IsEnabled:  pgtype.Bool{Bool: true, Valid: true},
			Priority:   pgtype.Int4{Int32: 10, Valid: true},
		},
	}, nil)

	req := httptest.NewRequest("GET", "/v1/plugins", nil)
	w := httptest.NewRecorder()

	handler.ListPlugins(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var plugins []PluginStatus
	json.NewDecoder(w.Body).Decode(&plugins)

	assert.True(t, len(plugins) >= 2)

	// Check discovered plugin
	var discovered *PluginStatus
	for _, p := range plugins {
		if p.ID == "discovered-plugin" {
			discovered = &p
			break
		}
	}
	assert.NotNil(t, discovered)
	assert.Equal(t, "healthy", discovered.Health)

	// Check DB plugin (should be failed because not in PM)
	var dbPlugin *PluginStatus
	for _, p := range plugins {
		if p.ID == "db-plugin" {
			dbPlugin = &p
			break
		}
	}
	assert.NotNil(t, dbPlugin)
	assert.Equal(t, "failed", dbPlugin.Health)
}

func TestPluginsHandler_GetPlugin(t *testing.T) {
	mockDB := new(MockStore)
	// Empty PM
	pm := &plugin_manager.Manager{
		Types: map[string]string{},
	}
	handler := NewPluginsHandler(mockDB, pm, logger.New("test"))

	r := chi.NewRouter()
	r.Get("/v1/plugins/{id}", handler.GetPlugin)

	req := httptest.NewRequest("GET", "/v1/plugins/my-plugin", nil)
	w := httptest.NewRecorder()

	mockDB.On("GetPluginConfig", mock.Anything, "my-plugin").Return(store.PluginConfig{
		ID:         "my-plugin",
		PluginType: "auth",
		ConfigJson: []byte(`{"foo":"bar"}`),
		IsEnabled:  pgtype.Bool{Bool: true, Valid: true},
	}, nil)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var p PluginStatus
	json.NewDecoder(w.Body).Decode(&p)

	assert.Equal(t, "my-plugin", p.ID)
	assert.Equal(t, "failed", p.Health) // Not in PM map
	assert.Equal(t, "bar", p.Config["foo"])
}

func TestPluginsHandler_UpdatePlugin(t *testing.T) {
	mockDB := new(MockStore)
	handler := NewPluginsHandler(mockDB, nil, logger.New("test"))

	r := chi.NewRouter()
	r.Put("/v1/plugins/{id}", handler.UpdatePlugin)

	reqBody := UpdatePluginRequest{
		Config:   map[string]interface{}{"key": "val"},
		Priority: 5,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", "/v1/plugins/p1", bytes.NewReader(body))
	w := httptest.NewRecorder()

	mockDB.On("UpdatePluginConfig", mock.Anything, mock.MatchedBy(func(arg store.UpdatePluginConfigParams) bool {
		return strings.Contains(string(arg.ConfigJson), "val")
	})).Return(nil)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPluginsHandler_EnablePlugin(t *testing.T) {
	mockDB := new(MockStore)
	handler := NewPluginsHandler(mockDB, nil, logger.New("test"))

	r := chi.NewRouter()
	r.Post("/v1/plugins/{id}/enable", handler.EnablePlugin)

	req := httptest.NewRequest("POST", "/v1/plugins/p1/enable", nil)
	w := httptest.NewRecorder()

	mockDB.On("EnablePlugin", mock.Anything, "p1").Return(nil)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPluginsHandler_DisablePlugin(t *testing.T) {
	mockDB := new(MockStore)
	handler := NewPluginsHandler(mockDB, nil, logger.New("test"))

	r := chi.NewRouter()
	r.Post("/v1/plugins/{id}/disable", handler.DisablePlugin)

	req := httptest.NewRequest("POST", "/v1/plugins/p1/disable", nil)
	w := httptest.NewRecorder()

	mockDB.On("DisablePlugin", mock.Anything, "p1").Return(nil)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPluginsHandler_Register(t *testing.T) {
	mockDB := new(MockStore)
	handler := NewPluginsHandler(mockDB, nil, logger.New("test"))
	r := chi.NewRouter()
	handler.Register(r)
	assert.NotNil(t, r)
}
