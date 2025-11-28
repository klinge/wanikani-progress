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

// writeAuthError writes an authentication error response
func writeAuthError(w http.ResponseWriter, message, detail string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	
	// Write JSON response directly
	w.Write([]byte(`{"error":{"code":"UNAUTHORIZED","message":"` + message + `","details":{"header":"` + detail + `"}}}`))
}
