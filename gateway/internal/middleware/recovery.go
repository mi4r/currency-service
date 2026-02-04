package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/mi4r/currency-service/pkg/httputil"
)

// Recovery creates a panic recovery middleware.
func Recovery(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger.Error("panic recovered",
						slog.Any("error", err),
						slog.String("stack", string(debug.Stack())),
					)
					httputil.InternalError(w, "internal server error")
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
