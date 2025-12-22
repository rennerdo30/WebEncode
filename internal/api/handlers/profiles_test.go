package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"net/http"
	"net/http/httptest"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rennerdo30/webencode/pkg/db/store"
	"github.com/rennerdo30/webencode/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestProfilesHandler_ListProfiles(t *testing.T) {
	mockDB := new(MockStore)
	handler := NewProfilesHandler(mockDB, logger.New("test"))

	mockDB.On("ListEncodingProfiles", mock.Anything).Return([]store.EncodingProfile{
		{
			ID:          "custom-1",
			Name:        "Custom Profile",
			Description: pgtype.Text{String: "Custom Desc", Valid: true},
			VideoCodec:  "h264",
			AudioCodec:  pgtype.Text{String: "aac", Valid: true},
			Width:       pgtype.Int4{Int32: 1920, Valid: true},
			Height:      pgtype.Int4{Int32: 1080, Valid: true},
			BitrateKbps: pgtype.Int4{Int32: 5000, Valid: true},
			Preset:      pgtype.Text{String: "medium", Valid: true},
			Container:   pgtype.Text{String: "mp4", Valid: true},
			ConfigJson:  []byte(`{}`),
			IsSystem:    pgtype.Bool{Bool: false, Valid: true},
		},
	}, nil)

	req := httptest.NewRequest("GET", "/v1/profiles", nil)
	w := httptest.NewRecorder()

	handler.ListProfiles(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var profiles []ProfileResponse
	json.NewDecoder(w.Body).Decode(&profiles)

	// Should contain system profiles + 1 custom
	assert.True(t, len(profiles) > 1)

	foundCustom := false
	for _, p := range profiles {
		if p.ID == "custom-1" {
			foundCustom = true
			assert.Equal(t, "Custom Profile", p.Name)
			assert.Equal(t, 1920, p.Width)
			assert.False(t, p.IsSystem)
		}
	}
	assert.True(t, foundCustom)
}

func TestProfilesHandler_GetProfile_System(t *testing.T) {
	mockDB := new(MockStore)
	handler := NewProfilesHandler(mockDB, logger.New("test"))

	r := chi.NewRouter()
	r.Get("/v1/profiles/{id}", handler.GetProfile)

	req := httptest.NewRequest("GET", "/v1/profiles/1080p_h264", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var p ProfileResponse
	json.NewDecoder(w.Body).Decode(&p)
	assert.Equal(t, "1080p_h264", p.ID)
	assert.True(t, p.IsSystem)
}

func TestProfilesHandler_GetProfile_Custom(t *testing.T) {
	mockDB := new(MockStore)
	handler := NewProfilesHandler(mockDB, logger.New("test"))

	r := chi.NewRouter()
	r.Get("/v1/profiles/{id}", handler.GetProfile)

	req := httptest.NewRequest("GET", "/v1/profiles/custom-1", nil)
	w := httptest.NewRecorder()

	mockDB.On("GetEncodingProfile", mock.Anything, "custom-1").Return(store.EncodingProfile{
		ID:         "custom-1",
		Name:       "Custom",
		VideoCodec: "h264",
		IsSystem:   pgtype.Bool{Bool: false, Valid: true},
	}, nil)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var p ProfileResponse
	json.NewDecoder(w.Body).Decode(&p)
	assert.Equal(t, "custom-1", p.ID)
	assert.False(t, p.IsSystem)
}

func TestProfilesHandler_GetProfile_NotFound(t *testing.T) {
	mockDB := new(MockStore)
	handler := NewProfilesHandler(mockDB, logger.New("test"))

	r := chi.NewRouter()
	r.Get("/v1/profiles/{id}", handler.GetProfile)

	req := httptest.NewRequest("GET", "/v1/profiles/missing", nil)
	w := httptest.NewRecorder()

	mockDB.On("GetEncodingProfile", mock.Anything, "missing").Return(store.EncodingProfile{}, errors.New("not found"))

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestProfilesHandler_CreateProfile(t *testing.T) {
	mockDB := new(MockStore)
	handler := NewProfilesHandler(mockDB, logger.New("test"))

	reqBody := ProfileRequest{
		ID:         "new-profile",
		Name:       "New Profile",
		VideoCodec: "h264",
		Width:      1280,
		Height:     720,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/v1/profiles", bytes.NewReader(body))
	w := httptest.NewRecorder()

	mockDB.On("CreateEncodingProfile", mock.Anything, mock.MatchedBy(func(arg store.CreateEncodingProfileParams) bool {
		return arg.ID == "new-profile" && arg.Width.Int32 == 1280
	})).Return(store.EncodingProfile{
		ID: "new-profile",
	}, nil)

	handler.CreateProfile(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestProfilesHandler_UpdateProfile(t *testing.T) {
	mockDB := new(MockStore)
	handler := NewProfilesHandler(mockDB, logger.New("test"))

	r := chi.NewRouter()
	r.Put("/v1/profiles/{id}", handler.UpdateProfile)

	reqBody := ProfileRequest{
		Name:       "Updated Profile",
		VideoCodec: "h265",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", "/v1/profiles/custom-1", bytes.NewReader(body))
	w := httptest.NewRecorder()

	mockDB.On("UpdateEncodingProfile", mock.Anything, mock.MatchedBy(func(arg store.UpdateEncodingProfileParams) bool {
		return arg.ID == "custom-1" && arg.VideoCodec == "h265"
	})).Return(nil)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestProfilesHandler_DeleteProfile_System(t *testing.T) {
	mockDB := new(MockStore)
	handler := NewProfilesHandler(mockDB, logger.New("test"))

	r := chi.NewRouter()
	r.Delete("/v1/profiles/{id}", handler.DeleteProfile)

	// Try checking 1080p_h264, assuming it's a system profile
	req := httptest.NewRequest("DELETE", "/v1/profiles/1080p_h264", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code) // 403
}

func TestProfilesHandler_DeleteProfile_Custom(t *testing.T) {
	mockDB := new(MockStore)
	handler := NewProfilesHandler(mockDB, logger.New("test"))

	r := chi.NewRouter()
	r.Delete("/v1/profiles/{id}", handler.DeleteProfile)

	req := httptest.NewRequest("DELETE", "/v1/profiles/custom-1", nil)
	w := httptest.NewRecorder()

	mockDB.On("DeleteEncodingProfile", mock.Anything, "custom-1").Return(nil)

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestProfilesHandler_Router(t *testing.T) {
	mockDB := new(MockStore)
	handler := NewProfilesHandler(mockDB, logger.New("test"))
	r := chi.NewRouter()
	handler.Register(r)
	assert.NotNil(t, r)
}
