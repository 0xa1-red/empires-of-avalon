package middleware

import (
	"net/http"
	"time"

	"golang.org/x/exp/slog"
)

type reqIDType int

const RequestIDKey reqIDType = 1

type CustomResponseWriter struct {
	w http.ResponseWriter

	status    int
	bytesSent int
}

func (crw CustomResponseWriter) Header() http.Header {
	return crw.w.Header()
}

func (crw *CustomResponseWriter) Write(d []byte) (int, error) {
	n, err := crw.w.Write(d)
	crw.bytesSent = n

	return n, err
}

func (crw *CustomResponseWriter) WriteHeader(statusCode int) {
	crw.status = statusCode
	crw.w.WriteHeader(statusCode)
}

func AvalonLogger(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		cw := &CustomResponseWriter{w: w, status: 200, bytesSent: 0}

		next.ServeHTTP(cw, r)

		fields := []interface{}{
			"remote_addr", r.RemoteAddr,
			"remote_user", r.URL.User.Username(),
			"time_local", time.Now().Format(time.RFC3339),
			"request_path", r.RequestURI,
			"request_method", r.Method,
			"request_proto", r.Proto,
			"status", cw.status,
			"bytes_sent", cw.bytesSent,
			"referer", r.Referer(),
			"user_agent", r.UserAgent(),
		}

		reqID := ""
		if val := r.Context().Value(RequestIDKey); val != nil {
			reqID = val.(string)
		}

		if reqID == "" {
			reqID = cw.Header().Get("X-Trace-Id")
		}

		if reqID != "" {
			fields = append(fields, "trace_id", reqID)
		}

		slog.Info("HTTP Request",
			fields...,
		)
	}

	return http.HandlerFunc(fn)
}
