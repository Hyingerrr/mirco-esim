package middleware

import (
	"net/http"
	"time"
)

func Monitor(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		h.ServeHTTP(w, r)
		duration := time.Since(start)
		serverReqQPS.Inc(r.Method, r.URL.Path)
		serverReqDuration.Observe(duration.Seconds(), r.Method, r.URL.Path)
	}
}
