package slogcorrelation

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/cyrusaf/ctxlog"
	"github.com/google/uuid"
)

func NewHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx = ctxlog.WithAttrs(ctx, slog.String("correlation_id", uuid.NewString()))
		r = r.WithContext(ctx)

		defer func() {
			if err := recover(); err != nil {
				slog.ErrorContext(ctx, "panic", "err", err, "stack", string(debug.Stack()))
				panic(err)
			}
		}()

		h.ServeHTTP(w, r)
	})
}
