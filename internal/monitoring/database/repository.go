package database

import (
	"database/sql"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// GetUserSettings retrieves the user settings (creates default if not exists)
func GetUserSettings() (*UserSettings, error) {
	db := GetDB()

	var settings UserSettings
	err := db.QueryRow(`
		SELECT id, username, COALESCE(photo_path, ''), password_hash, created_at, updated_at 
		FROM user_settings 
		LIMIT 1
	`).Scan(&settings.ID, &settings.Username, &settings.PhotoPath, &settings.PasswordHash, &settings.CreatedAt, &settings.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil // No settings found
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user settings: %w", err)
	}

	return &settings, nil
}

// CreateDefaultUser creates a default user with the given password
func CreateDefaultUser(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	db := GetDB()
	_, err = db.Exec(`
		INSERT INTO user_settings (username, password_hash) 
		VALUES (?, ?)
	`, "Admin", string(hashedPassword))

	if err != nil {
		return fmt.Errorf("failed to create default user: %w", err)
	}

	return nil
}

// UpdateUsername updates the username
func UpdateUsername(username string) error {
	db := GetDB()
	_, err := db.Exec(`
		UPDATE user_settings 
		SET username = ?, updated_at = ? 
		WHERE id = 1
	`, username, time.Now())

	if err != nil {
		return fmt.Errorf("failed to update username: %w", err)
	}

	return nil
}

// UpdatePassword updates the password after verifying the current password
func UpdatePassword(currentPassword, newPassword string) error {
	settings, err := GetUserSettings()
	if err != nil {
		return err
	}
	if settings == nil {
		return fmt.Errorf("user not found")
	}

	// Verify current password
	err = bcrypt.CompareHashAndPassword([]byte(settings.PasswordHash), []byte(currentPassword))
	if err != nil {
		return fmt.Errorf("incorrect current password")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	// Update in database
	db := GetDB()
	_, err = db.Exec(`
		UPDATE user_settings 
		SET password_hash = ?, updated_at = ? 
		WHERE id = 1
	`, string(hashedPassword), time.Now())

	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// UpdatePhotoPath updates the photo path
func UpdatePhotoPath(photoPath string) error {
	db := GetDB()
	_, err := db.Exec(`
		UPDATE user_settings 
		SET photo_path = ?, updated_at = ? 
		WHERE id = 1
	`, photoPath, time.Now())

	if err != nil {
		return fmt.Errorf("failed to update photo path: %w", err)
	}

	return nil
}

// DeletePhoto removes the photo path
func DeletePhoto() error {
	return UpdatePhotoPath("")
}

// VerifyPassword checks if the password is correct
func VerifyPassword(password string) error {
	settings, err := GetUserSettings()
	if err != nil {
		return err
	}
	if settings == nil {
		return fmt.Errorf("user not found")
	}

	return bcrypt.CompareHashAndPassword([]byte(settings.PasswordHash), []byte(password))
}
