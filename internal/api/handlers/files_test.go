package handlers

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/rennerdo30/webencode/internal/plugin_manager"
	pb "github.com/rennerdo30/webencode/pkg/api/v1"
	"github.com/rennerdo30/webencode/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupFilesTest() (*FilesHandler, *MockStorageClient) {
	mockStorage := new(MockStorageClient)
	pm := &plugin_manager.Manager{
		Storage: map[string]interface{}{
			"mock-storage": mockStorage,
		},
	}
	logger := logger.New("test")
	handler := NewFilesHandler(pm, logger)
	return handler, mockStorage
}

func TestFilesHandler_GetCapabilities(t *testing.T) {
	handler, mockStorage := setupFilesTest()

	mockStorage.On("GetCapabilities", mock.Anything, mock.Anything).Return(&pb.StorageCapabilities{
		StorageType:    "s3",
		SupportsBrowse: true,
	}, nil)

	req := httptest.NewRequest("GET", "/v1/files/capabilities", nil)
	w := httptest.NewRecorder()

	handler.GetCapabilities(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp []StoragePluginInfo
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Len(t, resp, 1)
	assert.Equal(t, "mock-storage", resp[0].PluginID)
	assert.Equal(t, "s3", resp[0].StorageType)
}

func TestFilesHandler_GetRoots(t *testing.T) {
	handler, mockStorage := setupFilesTest()

	mockStorage.On("GetCapabilities", mock.Anything, mock.Anything).Return(&pb.StorageCapabilities{
		StorageType:    "s3",
		SupportsBrowse: true,
	}, nil)

	mockStorage.On("BrowseRoots", mock.Anything, mock.Anything).Return(&pb.BrowseRootsResponse{
		Roots: []*pb.BrowseEntry{
			{Name: "Root1", Path: "/root1"},
		},
	}, nil)

	req := httptest.NewRequest("GET", "/v1/files/roots", nil)
	w := httptest.NewRecorder()

	handler.GetRoots(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp []BrowseRoot
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Len(t, resp, 1)
	assert.Equal(t, "Root1", resp[0].Name)
}

func TestFilesHandler_Browse(t *testing.T) {
	handler, mockStorage := setupFilesTest()

	mockStorage.On("GetCapabilities", mock.Anything, mock.Anything).Return(&pb.StorageCapabilities{
		StorageType:    "s3",
		SupportsBrowse: true,
	}, nil)

	mockStorage.On("Browse", mock.Anything, mock.MatchedBy(func(req *pb.BrowseRequest) bool {
		return req.Path == "/some/path"
	})).Return(&pb.BrowseResponse{
		CurrentPath: "/some/path",
		Entries: []*pb.BrowseEntry{
			{Name: "file.txt", IsDirectory: false, Size: 100},
		},
	}, nil)

	req := httptest.NewRequest("GET", "/v1/files/browse?plugin=mock-storage&path=/some/path", nil)
	w := httptest.NewRecorder()

	handler.Browse(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp BrowseResponse
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Equal(t, "/some/path", resp.CurrentPath)
	assert.Len(t, resp.Entries, 1)
}

func TestFilesHandler_Browse_NoPlugin(t *testing.T) {
	handler, mockStorage := setupFilesTest()

	// Should try to find browsable plugin
	mockStorage.On("GetCapabilities", mock.Anything, mock.Anything).Return(&pb.StorageCapabilities{
		StorageType:    "s3",
		SupportsBrowse: true,
	}, nil)

	mockStorage.On("Browse", mock.Anything, mock.Anything).Return(&pb.BrowseResponse{
		CurrentPath: "/",
		Entries:     []*pb.BrowseEntry{},
	}, nil)

	req := httptest.NewRequest("GET", "/v1/files/browse?path=/", nil)
	w := httptest.NewRecorder()

	handler.Browse(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestFilesHandler_Upload(t *testing.T) {
	handler, mockStorage := setupFilesTest()

	// Setup multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "test.txt")
	part.Write([]byte("content"))
	writer.Close()

	// Setup Mocks
	mockStream := new(MockUploadStream)

	// Expect GetCapabilities for auto-discovery if plugin_id not provided
	mockStorage.On("GetCapabilities", mock.Anything, mock.Anything).Return(&pb.StorageCapabilities{
		StorageType:    "s3",
		SupportsUpload: true,
	}, nil)

	mockStorage.On("Upload", mock.Anything).Return(mockStream, nil)

	// Expect Metadata (first chunk)
	mockStream.On("Send", mock.MatchedBy(func(chunk *pb.FileChunk) bool {
		if content, ok := chunk.Content.(*pb.FileChunk_Metadata); ok {
			return content.Metadata.Path == "uploads/test.txt"
		}
		return false
	})).Return(nil)

	// Expect Data (second chunk)
	mockStream.On("Send", mock.MatchedBy(func(chunk *pb.FileChunk) bool {
		_, ok := chunk.Content.(*pb.FileChunk_Data)
		return ok
	})).Return(nil)

	// Expect CloseAndRecv
	mockStream.On("CloseAndRecv").Return(&pb.UploadSummary{
		Url: "http://s3/test.txt",
	}, nil)

	req := httptest.NewRequest("POST", "/v1/files/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	handler.Upload(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp UploadResponse
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Equal(t, "http://s3/test.txt", resp.URL)
}

func TestFilesHandler_GetUploadURL(t *testing.T) {
	handler, mockStorage := setupFilesTest()

	reqBody := UploadURLRequest{
		PluginID:  "mock-storage",
		ObjectKey: "test.mp4",
	}
	body, _ := json.Marshal(reqBody)

	mockStorage.On("GetUploadURL", mock.Anything, mock.Anything).Return(&pb.SignedUrlResponse{
		Url:       "http://upload/url",
		ExpiresAt: 1234567890,
	}, nil)

	req := httptest.NewRequest("POST", "/v1/files/upload-url", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.GetUploadURL(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestFilesHandler_Register(t *testing.T) {
	handler, _ := setupFilesTest()
	r := chi.NewRouter()
	handler.Register(r)
	assert.NotNil(t, r)
}
