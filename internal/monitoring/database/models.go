package database

import "time"

// UserSettings represents the user settings model
type UserSettings struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	PhotoPath    string    `json:"photo_path"`
	PasswordHash string    `json:"-"` // Never expose password hash in JSON
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
