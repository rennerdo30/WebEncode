package errors

import (
	"encoding/json"
	"net/http"
	"runtime/debug"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rennerdo30/webencode/pkg/appcontext"
	"github.com/rennerdo30/webencode/pkg/db/store"
)

// WebEncodeError is a standardized error wrapper
type WebEncodeError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	HTTPStatus int    `json:"-"`
}

func (e *WebEncodeError) Error() string {
	return e.Message
}

// Catalog
var (
	ErrJobNotFound   = &WebEncodeError{Code: "JOB_NOT_FOUND", Message: "Job not found", HTTPStatus: 404}
	ErrNotFound      = &WebEncodeError{Code: "NOT_FOUND", Message: "Resource not found", HTTPStatus: 404}
	ErrWorkerBusy    = &WebEncodeError{Code: "WORKER_BUSY", Message: "Worker is busy", HTTPStatus: 503}
	ErrUnauthorized  = &WebEncodeError{Code: "UNAUTHORIZED", Message: "Unauthorized", HTTPStatus: 401}
	ErrForbidden     = &WebEncodeError{Code: "FORBIDDEN", Message: "Access denied", HTTPStatus: 403}
	ErrInternal      = &WebEncodeError{Code: "INTERNAL_ERROR", Message: "Internal server error", HTTPStatus: 500}
	ErrInvalidParams = &WebEncodeError{Code: "INVALID_PARAMS", Message: "Invalid parameters", HTTPStatus: 400}
	ErrConflict      = &WebEncodeError{Code: "CONFLICT", Message: "Resource conflict", HTTPStatus: 409}
	ErrRateLimited   = &WebEncodeError{Code: "RATE_LIMITED", Message: "Too many requests", HTTPStatus: 429}
)

// Response writes the error as JSON to the response writer
func Response(w http.ResponseWriter, r *http.Request, err error) {
	var we *WebEncodeError
	if e, ok := err.(*WebEncodeError); ok {
		we = e
	} else {
		we = ErrInternal
		// Log original error here
		logInternalError(r, err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(we.HTTPStatus)
	json.NewEncoder(w).Encode(we)
}

func logInternalError(r *http.Request, originalErr error) {
	ctx := r.Context()
	l := appcontext.GetLogger(ctx)
	db := appcontext.GetQuerier(ctx)

	// Log to stdout
	l.Error("Internal Server Error", "error", originalErr, "path", r.URL.Path)

	if db == nil {
		l.Warn("Cannot persist error: querier not in context")
		return
	}

	// Persist to DB
	// Extract context info
	contextData := map[string]interface{}{
		"method": r.Method,
		"path":   r.URL.Path,
		"query":  r.URL.RawQuery,
		"ip":     r.RemoteAddr,
	}
	contextBytes, _ := json.Marshal(contextData)

	// Get stack trace for non-standard errors
	stack := string(debug.Stack())

	_, err := db.CreateErrorEvent(ctx, store.CreateErrorEventParams{
		SourceComponent: "backend:api",
		Column2:         store.ErrorSeverityError,
		Message:         originalErr.Error(),
		StackTrace:      pgtype.Text{String: stack, Valid: true},
		ContextData:     contextBytes,
	})
	if err != nil {
		l.Error("Failed to persist internal error", "error", err)
	}
}
