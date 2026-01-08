package monitoring

// User Settings Handlers

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"test-go/internal/monitoring/database"
	"test-go/pkg/response"
	"time"

	"github.com/labstack/echo/v4"
)

// getUserSettings returns the current user settings
func (h *Handler) getUserSettings(c echo.Context) error {
	settings, err := database.GetUserSettings()
	if err != nil {
		return response.InternalServerError(c, err.Error())
	}
	if settings == nil {
		return response.Success(c, map[string]string{
			"username":   "Admin",
			"photo_path": "",
		})
	}

	return response.Success(c, map[string]string{
		"username":   settings.Username,
		"photo_path": settings.PhotoPath,
	})
}

// updateUserSettings updates the username
func (h *Handler) updateUserSettings(c echo.Context) error {
	type Request struct {
		Username string `json:"username"`
	}

	var req Request
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "Invalid request")
	}

	if req.Username == "" {
		return response.BadRequest(c, "Username cannot be empty")
	}

	if err := database.UpdateUsername(req.Username); err != nil {
		return response.InternalServerError(c, err.Error())
	}

	return response.Success(c, nil, "Username updated successfully")
}

// changePassword changes the user password
func (h *Handler) changePassword(c echo.Context) error {
	type Request struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}

	var req Request
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "Invalid request")
	}

	if req.CurrentPassword == "" || req.NewPassword == "" {
		return response.BadRequest(c, "Both current and new password are required")
	}

	if len(req.NewPassword) < 4 {
		return response.BadRequest(c, "New password must be at least 4 characters")
	}

	if err := database.UpdatePassword(req.CurrentPassword, req.NewPassword); err != nil {
		if strings.Contains(err.Error(), "incorrect") {
			return response.Unauthorized(c, "Current password is incorrect")
		}
		return response.InternalServerError(c, err.Error())
	}

	return response.Success(c, nil, "Password changed successfully")
}

// uploadPhoto handles profile photo upload
func (h *Handler) uploadPhoto(c echo.Context) error {
	// Get file from request
	file, err := c.FormFile("photo")
	if err != nil {
		return response.BadRequest(c, "No file uploaded")
	}

	// Check file size (2MB default)
	maxSize := int64(h.config.Monitoring.MaxPhotoSizeMB) * 1024 * 1024
	if maxSize == 0 {
		maxSize = 2 * 1024 * 1024 // Default 2MB
	}
	if file.Size > maxSize {
		return response.BadRequest(c, fmt.Sprintf("File size exceeds %dMB limit", h.config.Monitoring.MaxPhotoSizeMB))
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".gif" {
		return response.BadRequest(c, "Only JPG, PNG, and GIF files are allowed")
	}

	// Open uploaded file
	src, err := file.Open()
	if err != nil {
		return response.InternalServerError(c, "Failed to read file")
	}
	defer src.Close()

	// Create unique filename
	filename := fmt.Sprintf("user_%d%s", time.Now().Unix(), ext)

	uploadDir := h.config.Monitoring.UploadDir
	if uploadDir == "" {
		uploadDir = "web/monitoring/uploads"
	}
	profilesDir := filepath.Join(uploadDir, "profiles")

	// Ensure directory exists
	if err := os.MkdirAll(profilesDir, 0755); err != nil {
		return response.InternalServerError(c, "Failed to create upload directory")
	}

	// Create destination file
	dstPath := filepath.Join(profilesDir, filename)
	dst, err := os.Create(dstPath)
	if err != nil {
		return response.InternalServerError(c, "Failed to save file")
	}
	defer dst.Close()

	// Copy file
	if _, err = io.Copy(dst, src); err != nil {
		return response.InternalServerError(c, "Failed to save file")
	}

	// Delete old photo if exists
	settings, _ := database.GetUserSettings()
	if settings != nil && settings.PhotoPath != "" {
		oldPath := filepath.Join(profilesDir, filepath.Base(settings.PhotoPath))
		os.Remove(oldPath) // Ignore error
	}

	// Update database
	photoPath := filename
	if err := database.UpdatePhotoPath(photoPath); err != nil {
		return response.InternalServerError(c, "Failed to update database")
	}

	return response.Success(c, map[string]string{
		"message":    "Photo uploaded successfully",
		"photo_path": photoPath,
	})
}

// deleteUserPhoto deletes the user's profile photo
func (h *Handler) deleteUserPhoto(c echo.Context) error {
	settings, err := database.GetUserSettings()
	if err != nil {
		return response.InternalServerError(c, err.Error())
	}

	if settings != nil && settings.PhotoPath != "" {
		// Delete file from disk
		uploadDir := h.config.Monitoring.UploadDir
		if uploadDir == "" {
			uploadDir = "web/monitoring/uploads"
		}
		photoPath := filepath.Join(uploadDir, "profiles", filepath.Base(settings.PhotoPath))
		os.Remove(photoPath) // Ignore error if file doesn't exist
	}

	// Update database
	if err := database.DeletePhoto(); err != nil {
		return response.InternalServerError(c, err.Error())
	}

	return response.Success(c, nil, "Photo deleted successfully")
}
