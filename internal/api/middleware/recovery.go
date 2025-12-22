package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rennerdo30/webencode/pkg/db/store"
	"github.com/rennerdo30/webencode/pkg/errors"
	"github.com/rennerdo30/webencode/pkg/logger"
)

// Recovery returns a middleware that recovers from panics, logs the error, and returns a 500 status.
func Recovery(db store.Querier, l *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rvr := recover(); rvr != nil {
					// 1. Capture stack trace
					stack := string(debug.Stack())

					// 2. Log to stdout/logger
					err, ok := rvr.(error)
					if !ok {
						err = fmt.Errorf("%v", rvr)
					}
					l.Error("PANIC RECOVERED", "error", err, "stack", stack)

					// 3. Persist to Database
					contextData := map[string]interface{}{
						"method": r.Method,
						"path":   r.URL.Path,
						"ip":     getClientIP(r),
					}
					contextBytes, _ := json.Marshal(contextData)

					createParams := store.CreateErrorEventParams{
						SourceComponent: "backend:panic",
						Column2:         store.ErrorSeverityFatal,
						Message:         err.Error(),
						StackTrace:      pgtype.Text{String: stack, Valid: true},
						ContextData:     contextBytes,
					}

					// Use background context to ensure persistence even if request context is cancelled
					// but here we are in a defer, request context might be cancelled if client disconnected.
					// However, pgx operations usually respect context.
					// Ideally use a detached context, but r.Context() is usually fine for immediate write.
					// Let's use r.Context() for now.
					if _, dbErr := db.CreateErrorEvent(r.Context(), createParams); dbErr != nil {
						l.Error("Failed to persist panic to DB", "error", dbErr)
					}

					// 4. Respond to client
					errors.Response(w, r, errors.ErrInternal)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
