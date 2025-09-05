package auth

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

// TokenAuthMiddleware validates bearer tokens
func TokenAuthMiddleware(token string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Skip authentication if no token is configured
			if token == "" {
				return next(c)
			}

			// Check for Authorization header
			auth := c.Request().Header.Get("Authorization")
			if auth == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "Missing Authorization header")
			}

			// Validate Bearer token format
			if !strings.HasPrefix(auth, "Bearer ") {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid Authorization header format")
			}

			// Extract token
			providedToken := strings.TrimPrefix(auth, "Bearer ")

			// Constant-time comparison to prevent timing attacks
			if subtle.ConstantTimeCompare([]byte(providedToken), []byte(token)) != 1 {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token")
			}

			return next(c)
		}
	}
}

// NoAuthMiddleware skips authentication for specific routes
func NoAuthMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Skip any authentication middleware
			return next(c)
		}
	}
}