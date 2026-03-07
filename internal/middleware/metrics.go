package middleware

import (
	"net/http"
	"time"

	"nis-pipo/internal/metrics"
)

func Metrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := newResponseRecorder(w)
		start := time.Now()
		next.ServeHTTP(rec, r)
		metrics.RecordRequest(r, rec.status, time.Since(start))
	})
}
