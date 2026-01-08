package modules

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"test-go/pkg/response"

	"github.com/labstack/echo/v4"
)

type ServiceE struct {
	enabled       bool
	algorithm     string
	encryptionKey []byte
}

func NewServiceE(enabled bool, config map[string]interface{}) *ServiceE {
	// Extract configuration
	algorithm := "aes-256-gcm"
	key := ""

	if config != nil {
		if alg, ok := config["algorithm"].(string); ok && alg != "" {
			algorithm = alg
		}
		if k, ok := config["key"].(string); ok && k != "" {
			key = k
		}
	}

	// Ensure key is 32 bytes for AES-256
	// If key is shorter, pad it; if longer, truncate it
	keyBytes := []byte(key)
	if len(keyBytes) < 32 {
		// Pad with zeros
		paddedKey := make([]byte, 32)
		copy(paddedKey, keyBytes)
		keyBytes = paddedKey
	} else if len(keyBytes) > 32 {
		// Truncate to 32 bytes
		keyBytes = keyBytes[:32]
	}

	return &ServiceE{
		enabled:       enabled,
		algorithm:     algorithm,
		encryptionKey: keyBytes,
	}
}

func (s *ServiceE) Name() string  { return "Service E (Encryption)" }
func (s *ServiceE) Enabled() bool { return s.enabled }
func (s *ServiceE) Endpoints() []string {
	return []string{"/encryption/encrypt", "/encryption/decrypt", "/encryption/status", "/encryption/key-rotate"}
}

func (s *ServiceE) RegisterRoutes(g *echo.Group) {
	sub := g.Group("/encryption")

	// Encrypt endpoint
	sub.POST("/encrypt", s.EncryptData)

	// Decrypt endpoint
	sub.POST("/decrypt", s.DecryptData)

	// Status endpoint
	sub.GET("/status", s.GetStatus)

	// Key rotation endpoint
	sub.POST("/key-rotate", s.RotateKey)
}

// Request/Response structs
type EncryptRequest struct {
	Data        string `json:"data" validate:"required"`
	ContentType string `json:"content_type,omitempty"` // e.g., "application/json", "text/plain"
}

type EncryptResponse struct {
	EncryptedData string `json:"encrypted_data"`
	Algorithm     string `json:"algorithm"`
	Timestamp     int64  `json:"timestamp"`
	ContentType   string `json:"content_type,omitempty"`
}

type DecryptRequest struct {
	EncryptedData string `json:"encrypted_data" validate:"required"`
	ContentType   string `json:"content_type,omitempty"`
}

type DecryptResponse struct {
	DecryptedData string `json:"decrypted_data"`
	Algorithm     string `json:"algorithm"`
	Timestamp     int64  `json:"timestamp"`
	ContentType   string `json:"content_type,omitempty"`
}

type StatusResponse struct {
	Enabled      bool   `json:"enabled"`
	Algorithm    string `json:"algorithm"`
	CurrentKey   string `json:"current_key"`
	KeyLength    int    `json:"key_length"`
	RotateKeys   bool   `json:"rotate_keys"`
	LastRotation int64  `json:"last_rotation"`
}

type KeyRotateRequest struct {
	NewKey string `json:"new_key" validate:"required,min=16,max=64"`
}

// Encryption/Decryption functions
func (s *ServiceE) encrypt(data []byte) (string, error) {
	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %v", err)
	}

	// Create a new GCM instance
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %v", err)
	}

	// Create a nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %v", err)
	}

	// Encrypt the data
	encrypted := gcm.Seal(nonce, nonce, data, nil)

	// Return as base64 encoded string
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

func (s *ServiceE) decrypt(encryptedData string) ([]byte, error) {
	// Decode from base64
	data, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %v", err)
	}

	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %v", err)
	}

	// Create a new GCM instance
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %v", err)
	}

	// Extract the nonce and ciphertext
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, errors.New("encrypted data too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]

	// Decrypt the data
	decrypted, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %v", err)
	}

	return decrypted, nil
}

// Handlers
func (s *ServiceE) EncryptData(c echo.Context) error {
	var req EncryptRequest
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	// Validate content type
	contentType := req.ContentType
	if contentType == "" {
		contentType = "text/plain"
	}

	// Encrypt the data
	encrypted, err := s.encrypt([]byte(req.Data))
	if err != nil {
		return response.InternalServerError(c, fmt.Sprintf("Encryption failed: %v", err))
	}

	resp := EncryptResponse{
		EncryptedData: encrypted,
		Algorithm:     s.algorithm,
		Timestamp:     time.Now().Unix(),
		ContentType:   contentType,
	}

	return response.Success(c, resp, "Data encrypted successfully")
}

func (s *ServiceE) DecryptData(c echo.Context) error {
	var req DecryptRequest
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	// Validate content type
	contentType := req.ContentType
	if contentType == "" {
		contentType = "text/plain"
	}

	// Decrypt the data
	decrypted, err := s.decrypt(req.EncryptedData)
	if err != nil {
		return response.BadRequest(c, fmt.Sprintf("Decryption failed: %v", err))
	}

	resp := DecryptResponse{
		DecryptedData: string(decrypted),
		Algorithm:     s.algorithm,
		Timestamp:     time.Now().Unix(),
		ContentType:   contentType,
	}

	return response.Success(c, resp, "Data decrypted successfully")
}

func (s *ServiceE) GetStatus(c echo.Context) error {
	// Get current key info (show only first 8 chars for security)
	currentKeyPreview := fmt.Sprintf("%s...", hex.EncodeToString(s.encryptionKey[:4]))

	resp := StatusResponse{
		Enabled:      s.enabled,
		Algorithm:    s.algorithm,
		CurrentKey:   currentKeyPreview,
		KeyLength:    len(s.encryptionKey),
		RotateKeys:   false, // TODO: Implement key rotation
		LastRotation: time.Now().Unix(),
	}

	return response.Success(c, resp, "Encryption service status")
}

func (s *ServiceE) RotateKey(c echo.Context) error {
	var req KeyRotateRequest
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	// Validate new key length (must be at least 16 chars for security)
	if len(req.NewKey) < 16 {
		return response.BadRequest(c, "New key must be at least 16 characters long")
	}

	// Update the encryption key
	newKeyBytes := []byte(req.NewKey)
	if len(newKeyBytes) < 32 {
		// Pad with zeros
		paddedKey := make([]byte, 32)
		copy(paddedKey, newKeyBytes)
		s.encryptionKey = paddedKey
	} else if len(newKeyBytes) > 32 {
		// Truncate to 32 bytes
		s.encryptionKey = newKeyBytes[:32]
	} else {
		s.encryptionKey = newKeyBytes
	}

	// Update algorithm if needed (for future compatibility)
	if strings.Contains(req.NewKey, "-") {
		s.algorithm = "aes-256-gcm-custom"
	}

	return response.Success(c, map[string]string{
		"message":         "Encryption key rotated successfully",
		"new_key_preview": fmt.Sprintf("%s...", hex.EncodeToString(s.encryptionKey[:4])),
	}, "Key rotation successful")
}

// Middleware for automatic request/response encryption
func (s *ServiceE) EncryptionMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Skip encryption for encryption service endpoints
			if strings.HasPrefix(c.Request().URL.Path, "/api/v1/encryption") {
				return next(c)
			}

			// Skip encryption for health and other system endpoints
			if c.Request().URL.Path == "/health" || c.Request().URL.Path == "/restart" {
				return next(c)
			}

			// Only encrypt JSON content
			contentType := c.Request().Header.Get("Content-Type")
			if !strings.Contains(contentType, "application/json") {
				return next(c)
			}

			// For now, just pass through - full encryption middleware will be implemented separately
			return next(c)
		}
	}
}

// Helper function to encrypt JSON data
func (s *ServiceE) EncryptJSON(data interface{}) (string, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %v", err)
	}

	return s.encrypt(jsonData)
}

// Helper function to decrypt to JSON
func (s *ServiceE) DecryptJSON(encryptedData string, target interface{}) error {
	decrypted, err := s.decrypt(encryptedData)
	if err != nil {
		return fmt.Errorf("failed to decrypt: %v", err)
	}

	return json.Unmarshal(decrypted, target)
}
