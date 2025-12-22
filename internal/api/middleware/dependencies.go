package middleware

import (
	"net/http"

	"github.com/rennerdo30/webencode/pkg/appcontext"
	"github.com/rennerdo30/webencode/pkg/db/store"
	"github.com/rennerdo30/webencode/pkg/logger"
)

// InjectDependencies injects the logger and database querier into the request context
func InjectDependencies(db store.Querier, l *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			ctx = appcontext.WithLogger(ctx, l)
			ctx = appcontext.WithQuerier(ctx, db)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
