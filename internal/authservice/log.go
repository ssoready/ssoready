package authservice

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
)

func logHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		body, err := io.ReadAll(r.Body)
		if err != nil {
			panic(fmt.Errorf("read body: %w", err))
		}

		slog.InfoContext(r.Context(), "http_request", "method", r.Method, "path", r.URL.Path, "request_body", string(body))

		// rewrite the response to be a recorded one, and the request to have the original body
		recorder := httptest.NewRecorder()
		r.Body = io.NopCloser(bytes.NewBuffer(body))
		h.ServeHTTP(recorder, r)

		slog.InfoContext(r.Context(), "http_response", "method", r.Method, "path", r.URL.Path, "request_body", string(body), "status", recorder.Code, "response_headers", recorder.Header(), "response_body", recorder.Body.String())

		// write out recorded response to w
		for k, v := range recorder.Header() {
			w.Header()[k] = v
		}
		w.WriteHeader(recorder.Code)
		if _, err := recorder.Body.WriteTo(w); err != nil {
			panic(fmt.Errorf("write body: %w", err))
		}
	})
}
