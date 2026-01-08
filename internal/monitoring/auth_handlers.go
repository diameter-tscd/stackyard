package monitoring

import (
	"strings"
	"test-go/internal/monitoring/database"
	"test-go/internal/monitoring/session"
	"test-go/pkg/response"
	"time"

	"github.com/labstack/echo/v4"
)

// LoginRequest represents login credentials
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// handleLogin handles user login
func handleLogin(sessionManager *session.Manager) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req LoginRequest
		if err := c.Bind(&req); err != nil {
			return response.BadRequest(c, "Invalid request")
		}

		// Get user settings from database
		settings, err := database.GetUserSettings()
		if err != nil {
			return response.InternalServerError(c, "Internal server error")
		}

		if settings == nil {
			return response.Unauthorized(c, "Invalid username or password")
		}

		// Validate username matches database (case-insensitive)
		if !strings.EqualFold(req.Username, settings.Username) {
			return response.Unauthorized(c, "Invalid username or password")
		}

		// Validate password against database
		err = database.VerifyPassword(req.Password)
		if err != nil {
			return response.Unauthorized(c, "Invalid username or password")
		}

		// Create session using the actual username from database
		sess, err := sessionManager.Create(settings.Username)
		if err != nil {
			return response.InternalServerError(c, "Failed to create session")
		}

		// Set session cookie (24 hours)
		session.SetCookie(c, sess.ID, int(24*time.Hour.Seconds()))

		return response.Success(c, nil, "Login successful")
	}
}

// handleLogout handles user logout
func handleLogout(sessionManager *session.Manager) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Get session cookie
		cookie, err := c.Cookie(session.SessionCookieName)
		if err == nil {
			// Delete session from manager
			sessionManager.Delete(cookie.Value)
		}

		// Clear cookie
		session.ClearCookie(c)

		// Prevent caching of logout response
		c.Response().Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")

		return response.Success(c, nil, "Logged out successfully")
	}
}
