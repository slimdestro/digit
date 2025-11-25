package middleware

import (
	"context"
	"net"
	"net/http"
	"strings"
	"sync"

	"golang.org/x/time/rate"
)

type Tracer interface {
	StartSpan(ctx context.Context, name string) context.Context
	FinishSpan(ctx context.Context)
}

func Chain(final http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	if len(middlewares) == 0 {
		return final
	}
	h := final
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

func TelemetryMiddleware(t Tracer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if t != nil {
				ctx := t.StartSpan(r.Context(), r.Method+" "+r.URL.Path)
				r = r.WithContext(ctx)
				defer t.FinishSpan(ctx)
			}
			next.ServeHTTP(w, r)
		})
	}
}

var visitors sync.Map

func getLimiter(ip string, rps int, burst int) *rate.Limiter {
	v, ok := visitors.Load(ip)
	if ok {
		return v.(*rate.Limiter)
	}
	lim := rate.NewLimiter(rate.Limit(rps), burst)
	visitors.Store(ip, lim)
	return lim
}

func RateLimitMiddleware(rps int, burst int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := realIP(r)
			lim := getLimiter(ip, rps, burst)
			if !lim.Allow() {
				http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func SecurityMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("Referrer-Policy", "no-referrer")

			if r.Method == http.MethodPost || r.Method == http.MethodPut {
				r.Body = http.MaxBytesReader(w, r.Body, 1048576)
				ct := r.Header.Get("Content-Type")
				if !strings.Contains(ct, "application/json") {
					http.Error(w, "invalid content type", http.StatusUnsupportedMediaType)
					return
				}
			}

			q := r.URL.Query()
			for k, vals := range q {
				for i := range vals {
					vals[i] = sanitize(vals[i])
				}
				q[k] = vals
			}
			r.URL.RawQuery = q.Encode()

			next.ServeHTTP(w, r)
		})
	}
}

func APIKeyMiddleware(key string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("X-API-KEY") != key {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func sanitize(s string) string {
	s = strings.ReplaceAll(s, "<", "")
	s = strings.ReplaceAll(s, ">", "")
	s = strings.ReplaceAll(s, "javascript:", "")
	return s
}

func realIP(r *http.Request) string {
	ip := r.Header.Get("X-Forwarded-For")
	if ip != "" {
		parts := strings.Split(ip, ",")
		return strings.TrimSpace(parts[0])
	}
	ip = r.Header.Get("X-Real-IP")
	if ip != "" {
		return ip
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
