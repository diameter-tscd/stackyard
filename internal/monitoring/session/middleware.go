package session

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

const SessionCookieName = "gobp_session_id"

// Middleware creates Echo middleware for session validation
func Middleware(manager *Manager) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get session cookie
			cookie, err := c.Cookie(SessionCookieName)
			if err != nil {
				return c.Redirect(http.StatusFound, "/")
			}

			// Validate session
			session, exists := manager.Get(cookie.Value)
			if !exists {
				// Clear invalid cookie
				clearCookie(c)
				c.Response().Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
				return c.Redirect(http.StatusFound, "/")
			}

			// Store session in context for handlers
			c.Set("session", session)

			return next(c)
		}
	}
}

// SetCookie sets the session cookie
func SetCookie(c echo.Context, sessionID string, maxAge int) {
	cookie := new(http.Cookie)
	cookie.Name = SessionCookieName
	cookie.Value = sessionID
	cookie.Path = "/"
	cookie.HttpOnly = true
	cookie.Secure = false // Set to true in production with HTTPS
	cookie.SameSite = http.SameSiteLaxMode
	cookie.MaxAge = maxAge
	c.SetCookie(cookie)
}

// ClearCookie clears the session cookie
func clearCookie(c echo.Context) {
	cookie := new(http.Cookie)
	cookie.Name = SessionCookieName
	cookie.Value = "deleted"
	cookie.Path = "/"
	cookie.MaxAge = -1
	cookie.Expires = time.Unix(0, 0) // Set to epoch for better compatibility
	cookie.HttpOnly = true
	cookie.SameSite = http.SameSiteLaxMode
	c.SetCookie(cookie)
}

// ClearCookie exports the clear cookie function
func ClearCookie(c echo.Context) {
	clearCookie(c)
}
