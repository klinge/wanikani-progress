package api

import (
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
)

// AuthMiddleware creates an authentication middleware
func AuthMiddleware(token string, logger *logrus.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract Authorization header
			authHeader := r.Header.Get("Authorization")

			// Check if Authorization header is present
			if authHeader == "" {
				logger.WithFields(logrus.Fields{
					"path":   r.URL.Path,
					"method": r.Method,
					"remote": r.RemoteAddr,
				}).Warn("Authentication failed: missing Authorization header")

				writeAuthError(w, "Authentication required", "Authorization header with Bearer token is required")
				return
			}

			// Check if it's a Bearer token
			if !strings.HasPrefix(authHeader, "Bearer ") {
				logger.WithFields(logrus.Fields{
					"path":   r.URL.Path,
					"method": r.Method,
					"remote": r.RemoteAddr,
				}).Warn("Authentication failed: invalid Authorization header format")

				writeAuthError(w, "Authentication required", "Authorization header must use Bearer token format")
				return
			}

			// Extract token
			providedToken := strings.TrimPrefix(authHeader, "Bearer ")

			// Validate token
			if providedToken != token {
				logger.WithFields(logrus.Fields{
					"path":   r.URL.Path,
					"method": r.Method,
					"remote": r.RemoteAddr,
				}).Warn("Authentication failed: invalid token")

				writeAuthError(w, "Authentication required", "Invalid authentication token")
				return
			}

			// Token is valid, proceed to next handler
			next.ServeHTTP(w, r)
		})
	}
}

// CORSMiddleware adds CORS headers to allow cross-origin requests
func CORSMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Allow specific origins (localhost for development)
			allowedOrigins := []string{
				"http://localhost:3000",
				"http://localhost:3003",
				"http://127.0.0.1:3000",
				"http://127.0.0.1:3003",
				"https://wkstats.klin.ge",
			}

			// Check if origin is allowed
			allowed := false
			for _, allowedOrigin := range allowedOrigins {
				if origin == allowedOrigin {
					allowed = true
					break
				}
			}

			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Max-Age", "86400")
			}

			// Handle preflight OPTIONS request
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			// Continue to next handler
			next.ServeHTTP(w, r)
		})
	}
}

// writeAuthError writes an authentication error response
func writeAuthError(w http.ResponseWriter, message, detail string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)

	// Write JSON response directly
	w.Write([]byte(`{"error":{"code":"UNAUTHORIZED","message":"` + message + `","details":{"header":"` + detail + `"}}}`))
}
