package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rennerdo30/webencode/internal/plugin_manager"
	pb "github.com/rennerdo30/webencode/pkg/api/v1"
	"github.com/rennerdo30/webencode/pkg/logger"
)

// FilesHandler handles file browser requests via storage plugins
type FilesHandler struct {
	pm     *plugin_manager.Manager
	logger *logger.Logger
}

// NewFilesHandler creates a new files handler
func NewFilesHandler(pm *plugin_manager.Manager, l *logger.Logger) *FilesHandler {
	return &FilesHandler{
		pm:     pm,
		logger: l,
	}
}

// Register registers the files routes
func (h *FilesHandler) Register(r chi.Router) {
	r.Route("/v1/files", func(r chi.Router) {
		r.Get("/capabilities", h.GetCapabilities)
		r.Get("/roots", h.GetRoots)
		r.Get("/browse", h.Browse)
		r.Post("/upload-url", h.GetUploadURL)
		r.Post("/upload", h.Upload)
	})
}

// StoragePluginInfo represents info about a storage plugin's browse capabilities
type StoragePluginInfo struct {
	PluginID       string `json:"plugin_id"`
	StorageType    string `json:"storage_type"`
	SupportsBrowse bool   `json:"supports_browse"`
}

// GetCapabilities returns which storage plugins support file browsing
// GET /v1/files/capabilities
func (h *FilesHandler) GetCapabilities(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	plugins := []StoragePluginInfo{}

	for id, client := range h.pm.Storage {
		storageClient, ok := client.(pb.StorageServiceClient)
		if !ok {
			continue
		}

		caps, err := storageClient.GetCapabilities(ctx, &pb.Empty{})
		if err != nil {
			h.logger.Error("Failed to get capabilities", "plugin", id, "error", err)
			continue
		}

		plugins = append(plugins, StoragePluginInfo{
			PluginID:       id,
			StorageType:    caps.StorageType,
			SupportsBrowse: caps.SupportsBrowse,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(plugins); err != nil {
		h.logger.Error("Failed to encode response", "error", err)
	}
}

// BrowseRoot represents a root directory from a storage plugin
type BrowseRoot struct {
	PluginID    string `json:"plugin_id"`
	StorageType string `json:"storage_type"`
	Name        string `json:"name"`
	Path        string `json:"path"`
}

// GetRoots returns all available browse roots from all storage plugins
// GET /v1/files/roots
func (h *FilesHandler) GetRoots(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	roots := []BrowseRoot{}

	for id, client := range h.pm.Storage {
		storageClient, ok := client.(pb.StorageServiceClient)
		if !ok {
			continue
		}

		// First check if plugin supports browsing
		caps, err := storageClient.GetCapabilities(ctx, &pb.Empty{})
		if err != nil || !caps.SupportsBrowse {
			continue
		}

		// Get browse roots
		resp, err := storageClient.BrowseRoots(ctx, &pb.Empty{})
		if err != nil {
			h.logger.Error("Failed to get browse roots", "plugin", id, "error", err)
			continue
		}

		for _, root := range resp.Roots {
			roots = append(roots, BrowseRoot{
				PluginID:    id,
				StorageType: caps.StorageType,
				Name:        root.Name,
				Path:        root.Path,
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(roots); err != nil {
		h.logger.Error("Failed to encode response", "error", err)
	}
}

// BrowseEntry represents a file or directory entry
type BrowseEntry struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	IsDirectory bool   `json:"is_directory"`
	Size        int64  `json:"size"`
	ModTime     int64  `json:"mod_time"`
	Extension   string `json:"extension,omitempty"`
	IsVideo     bool   `json:"is_video"`
	IsAudio     bool   `json:"is_audio"`
	IsImage     bool   `json:"is_image"`
}

// BrowseResponse is the response for browse requests
type BrowseResponse struct {
	PluginID    string        `json:"plugin_id"`
	CurrentPath string        `json:"current_path"`
	ParentPath  string        `json:"parent_path"`
	Entries     []BrowseEntry `json:"entries"`
}

// Browse lists directory contents from a storage plugin
// GET /v1/files/browse?plugin=storage-fs&path=/some/path&media_only=true&show_hidden=false&search=query
func (h *FilesHandler) Browse(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	pluginID := r.URL.Query().Get("plugin")
	path := r.URL.Query().Get("path")
	mediaOnly := r.URL.Query().Get("media_only") == "true"
	showHidden := r.URL.Query().Get("show_hidden") == "true"
	searchQuery := r.URL.Query().Get("search")

	// If no plugin specified, try to find one that supports browsing
	if pluginID == "" {
		pluginID = h.findBrowsablePlugin(ctx)
		if pluginID == "" {
			http.Error(w, "No storage plugin with browse support found", http.StatusNotFound)
			return
		}
	}

	// Get the storage plugin
	client, ok := h.pm.Storage[pluginID]
	if !ok {
		http.Error(w, "Storage plugin not found: "+pluginID, http.StatusNotFound)
		return
	}

	storageClient, ok := client.(pb.StorageServiceClient)
	if !ok {
		http.Error(w, "Invalid storage plugin", http.StatusInternalServerError)
		return
	}

	// Check capabilities
	caps, err := storageClient.GetCapabilities(ctx, &pb.Empty{})
	if err != nil {
		h.logger.Error("Failed to get capabilities", "plugin", pluginID, "error", err)
		http.Error(w, "Failed to get plugin capabilities", http.StatusInternalServerError)
		return
	}

	if !caps.SupportsBrowse {
		http.Error(w, "Plugin does not support browsing", http.StatusBadRequest)
		return
	}

	// Browse the path
	resp, err := storageClient.Browse(ctx, &pb.BrowseRequest{
		Path:        path,
		ShowHidden:  showHidden,
		MediaOnly:   mediaOnly,
		SearchQuery: searchQuery,
	})
	if err != nil {
		h.logger.Error("Failed to browse", "plugin", pluginID, "path", path, "error", err)
		http.Error(w, "Failed to browse: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert to response format
	entries := make([]BrowseEntry, len(resp.Entries))
	for i, e := range resp.Entries {
		entries[i] = BrowseEntry{
			Name:        e.Name,
			Path:        e.Path,
			IsDirectory: e.IsDirectory,
			Size:        e.Size,
			ModTime:     e.ModTime,
			Extension:   e.Extension,
			IsVideo:     e.IsVideo,
			IsAudio:     e.IsAudio,
			IsImage:     e.IsImage,
		}
	}

	response := BrowseResponse{
		PluginID:    pluginID,
		CurrentPath: resp.CurrentPath,
		ParentPath:  resp.ParentPath,
		Entries:     entries,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response", "error", err)
	}
}

// findBrowsablePlugin finds the first storage plugin that supports browsing
func (h *FilesHandler) findBrowsablePlugin(ctx context.Context) string {
	for id, client := range h.pm.Storage {
		storageClient, ok := client.(pb.StorageServiceClient)
		if !ok {
			continue
		}

		caps, err := storageClient.GetCapabilities(ctx, &pb.Empty{})
		if err == nil && caps.SupportsBrowse {
			return id
		}
	}
	return ""
}

// UploadURLRequest is the request body for getting a pre-signed upload URL
type UploadURLRequest struct {
	PluginID      string `json:"plugin_id,omitempty"`
	Bucket        string `json:"bucket,omitempty"`
	ObjectKey     string `json:"object_key"`
	ContentType   string `json:"content_type,omitempty"`
	ExpirySeconds int64  `json:"expiry_seconds,omitempty"`
}

// UploadURLResponse contains the pre-signed URL for uploads
type UploadURLResponse struct {
	URL       string            `json:"url"`
	Headers   map[string]string `json:"headers,omitempty"`
	ExpiresAt int64             `json:"expires_at"`
	ObjectKey string            `json:"object_key"`
	PluginID  string            `json:"plugin_id"`
}

// GetUploadURL returns a pre-signed URL for direct browser uploads
// POST /v1/files/upload-url
func (h *FilesHandler) GetUploadURL(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req UploadURLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.ObjectKey == "" {
		http.Error(w, "object_key is required", http.StatusBadRequest)
		return
	}

	// Find a storage plugin that supports uploads
	pluginID := req.PluginID
	if pluginID == "" {
		pluginID = h.findUploadablePlugin(ctx)
		if pluginID == "" {
			http.Error(w, "No storage plugin with upload support found", http.StatusNotFound)
			return
		}
	}

	// Get the storage plugin
	client, ok := h.pm.Storage[pluginID]
	if !ok {
		http.Error(w, "Storage plugin not found: "+pluginID, http.StatusNotFound)
		return
	}

	storageClient, ok := client.(pb.StorageServiceClient)
	if !ok {
		http.Error(w, "Invalid storage plugin", http.StatusInternalServerError)
		return
	}

	// Default expiry of 1 hour
	expiry := req.ExpirySeconds
	if expiry <= 0 {
		expiry = 3600
	}

	// Default content type
	contentType := req.ContentType
	if contentType == "" {
		contentType = "video/mp4"
	}

	// Get upload URL from plugin
	resp, err := storageClient.GetUploadURL(ctx, &pb.SignedUrlRequest{
		Bucket:        req.Bucket,
		ObjectKey:     req.ObjectKey,
		ExpirySeconds: expiry,
		ContentType:   contentType,
		Method:        "PUT",
	})
	if err != nil {
		h.logger.Error("Failed to get upload URL", "plugin", pluginID, "error", err)
		http.Error(w, "Failed to get upload URL: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := UploadURLResponse{
		URL:       resp.Url,
		Headers:   resp.Headers,
		ExpiresAt: resp.ExpiresAt,
		ObjectKey: req.ObjectKey,
		PluginID:  pluginID,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response", "error", err)
	}
}

// UploadResponse contains information about the uploaded file
type UploadResponse struct {
	URL       string `json:"url"`
	PluginID  string `json:"plugin_id"`
	ObjectKey string `json:"object_key"`
	Size      int64  `json:"size"`
}

// Upload handles direct file uploads via multipart form
// POST /v1/files/upload
func (h *FilesHandler) Upload(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Limit upload size to 10GB
	r.Body = http.MaxBytesReader(w, r.Body, 10<<30)

	// Parse multipart form (32MB memory buffer, rest to disk)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		h.logger.Error("Failed to parse multipart form", "error", err)
		http.Error(w, "Failed to parse upload: "+err.Error(), http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "No file provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Get optional parameters
	pluginID := r.FormValue("plugin_id")
	bucket := r.FormValue("bucket")
	objectKey := r.FormValue("object_key")

	// Generate object key if not provided
	if objectKey == "" {
		objectKey = "uploads/" + header.Filename
	}

	// Find a storage plugin that supports uploads
	if pluginID == "" {
		pluginID = h.findUploadablePlugin(ctx)
		if pluginID == "" {
			http.Error(w, "No storage plugin with upload support found", http.StatusNotFound)
			return
		}
	}

	// Get the storage plugin
	client, ok := h.pm.Storage[pluginID]
	if !ok {
		http.Error(w, "Storage plugin not found: "+pluginID, http.StatusNotFound)
		return
	}

	storageClient, ok := client.(pb.StorageServiceClient)
	if !ok {
		http.Error(w, "Invalid storage plugin", http.StatusInternalServerError)
		return
	}

	// Start streaming upload
	stream, err := storageClient.Upload(ctx)
	if err != nil {
		h.logger.Error("Failed to start upload stream", "plugin", pluginID, "error", err)
		http.Error(w, "Failed to start upload", http.StatusInternalServerError)
		return
	}

	// Send metadata first
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	err = stream.Send(&pb.FileChunk{
		Content: &pb.FileChunk_Metadata{
			Metadata: &pb.FileMetadata{
				Bucket:      bucket,
				Path:        objectKey,
				Size:        header.Size,
				ContentType: contentType,
			},
		},
	})
	if err != nil {
		h.logger.Error("Failed to send upload metadata", "error", err)
		http.Error(w, "Failed to upload", http.StatusInternalServerError)
		return
	}

	// Stream file data in chunks
	buf := make([]byte, 64*1024) // 64KB chunks
	totalSent := int64(0)

	for {
		n, readErr := file.Read(buf)
		if n > 0 {
			err = stream.Send(&pb.FileChunk{
				Content: &pb.FileChunk_Data{
					Data: buf[:n],
				},
			})
			if err != nil {
				h.logger.Error("Failed to send chunk", "error", err)
				http.Error(w, "Failed to upload chunk", http.StatusInternalServerError)
				return
			}
			totalSent += int64(n)
		}
		if readErr != nil {
			break
		}
	}

	// Close and get response
	summary, err := stream.CloseAndRecv()
	if err != nil {
		h.logger.Error("Failed to complete upload", "plugin", pluginID, "error", err)
		http.Error(w, "Failed to complete upload: "+err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Info("File uploaded successfully",
		"plugin", pluginID,
		"object_key", objectKey,
		"size", totalSent,
		"url", summary.Url,
	)

	response := UploadResponse{
		URL:       summary.Url,
		PluginID:  pluginID,
		ObjectKey: objectKey,
		Size:      totalSent,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response", "error", err)
	}
}

// findUploadablePlugin finds the first storage plugin that supports uploads
func (h *FilesHandler) findUploadablePlugin(ctx context.Context) string {
	type pluginInfo struct {
		id          string
		storageType string
	}
	var candidates []pluginInfo

	for id, client := range h.pm.Storage {
		storageClient, ok := client.(pb.StorageServiceClient)
		if !ok {
			continue
		}

		caps, err := storageClient.GetCapabilities(ctx, &pb.Empty{})
		if err == nil && caps.SupportsUpload {
			candidates = append(candidates, pluginInfo{id: id, storageType: caps.StorageType})
		}
	}

	// Prefer S3/object storage for distributed worker access
	// Filesystem returns file:// URLs that aren't accessible to workers in containers
	for _, c := range candidates {
		if c.storageType == "s3" || c.storageType == "object" {
			return c.id
		}
	}

	// Fallback to any (including filesystem for local dev)
	if len(candidates) > 0 {
		return candidates[0].id
	}

	return ""
}
