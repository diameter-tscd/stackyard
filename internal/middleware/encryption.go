package middleware

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"test-go/config"
	"test-go/pkg/logger"

	"github.com/labstack/echo/v4"
)

// EncryptionMiddleware returns a middleware that handles API request/response encryption
func EncryptionMiddleware(cfg *config.Config, logger *logger.Logger) echo.MiddlewareFunc {
	// Create encryption service instance
	encryptionService := createEncryptionService(cfg)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Skip encryption for certain paths
			path := c.Request().URL.Path
			if shouldSkipEncryption(path) {
				return next(c)
			}

			// Handle request decryption (if encrypted)
			if err := handleRequestDecryption(c, encryptionService, logger); err != nil {
				return err
			}

			// Create response recorder
			resBody := new(bytes.Buffer)
			recorder := &ResponseRecorder{
				ResponseWriter: c.Response().Writer,
				Body:           resBody,
				StatusCode:     http.StatusOK,
			}
			c.Response().Writer = recorder

			// Call next handler
			err := next(c)

			// Handle response encryption
			if err2 := handleResponseEncryption(c, recorder, encryptionService, logger); err2 != nil {
				return err2
			}

			return err
		}
	}
}

func createEncryptionService(cfg *config.Config) *encryptionService {
	// Extract encryption config
	encCfg := cfg.Encryption

	// Ensure key is 32 bytes for AES-256
	keyBytes := []byte(encCfg.Key)
	if len(keyBytes) < 32 {
		// Pad with zeros
		paddedKey := make([]byte, 32)
		copy(paddedKey, keyBytes)
		keyBytes = paddedKey
	} else if len(keyBytes) > 32 {
		// Truncate to 32 bytes
		keyBytes = keyBytes[:32]
	}

	return &encryptionService{
		enabled:       encCfg.Enabled,
		algorithm:     encCfg.Algorithm,
		encryptionKey: keyBytes,
	}
}

func shouldSkipEncryption(path string) bool {
	// Skip encryption for health and system endpoints
	skipPaths := []string{
		"/health",
		"/restart",
		"/api/v1/encryption", // Encryption service endpoints
	}

	for _, skipPath := range skipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}

	return false
}

func handleRequestDecryption(c echo.Context, es *encryptionService, logger *logger.Logger) error {
	// Check if request has encryption header
	encryptedHeader := c.Request().Header.Get("X-Encrypted-Request")
	if encryptedHeader != "true" {
		return nil // Not encrypted, proceed normally
	}

	// Only decrypt JSON content
	contentType := c.Request().Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		return nil
	}

	// Read and decrypt the request body
	bodyBytes, err := io.ReadAll(c.Request().Body)
	if err != nil {
		logger.Error("Failed to read encrypted request body", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to read request body")
	}
	c.Request().Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// Decrypt the body
	decrypted, err := es.decrypt(string(bodyBytes))
	if err != nil {
		logger.Error("Failed to decrypt request body", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to decrypt request body")
	}

	// Replace request body with decrypted content
	c.Request().Body = io.NopCloser(bytes.NewBuffer([]byte(decrypted)))
	c.Request().Header.Set("Content-Length", fmt.Sprintf("%d", len(decrypted)))

	return nil
}

func handleResponseEncryption(c echo.Context, recorder *ResponseRecorder, es *encryptionService, logger *logger.Logger) error {
	// Only encrypt JSON responses
	contentType := recorder.Header().Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		recorder.FlushOriginal()
		return nil
	}

	// Skip if encryption is disabled
	if !es.enabled {
		recorder.FlushOriginal()
		return nil
	}

	// Encrypt the response
	data := recorder.Body.Bytes()
	if len(data) > 0 {
		encrypted, err := es.encrypt(string(data))
		if err != nil {
			logger.Error("Failed to encrypt response", err)
			recorder.FlushOriginal()
			return nil
		}

		// Set headers
		recorder.ResponseWriter.Header().Set("X-Encrypted-Response", "true")
		recorder.ResponseWriter.Header().Set("X-Encryption-Algorithm", es.algorithm)
		recorder.ResponseWriter.Header().Set("Content-Length", fmt.Sprintf("%d", len(encrypted)))

		// Write encrypted response
		recorder.ResponseWriter.WriteHeader(recorder.StatusCode)
		recorder.ResponseWriter.Write([]byte(encrypted))
	} else {
		recorder.ResponseWriter.WriteHeader(recorder.StatusCode)
	}

	return nil
}

// ResponseRecorder captures the response
type ResponseRecorder struct {
	http.ResponseWriter
	Body       *bytes.Buffer
	StatusCode int
}

func (r *ResponseRecorder) WriteHeader(code int) {
	r.StatusCode = code
}

func (r *ResponseRecorder) Write(b []byte) (int, error) {
	return r.Body.Write(b)
}

func (r *ResponseRecorder) FlushOriginal() {
	r.ResponseWriter.Header().Del("Content-Length")
	r.ResponseWriter.WriteHeader(r.StatusCode)
	r.ResponseWriter.Write(r.Body.Bytes())
}

// encryptionService provides encryption/decryption functionality
type encryptionService struct {
	enabled       bool
	algorithm     string
	encryptionKey []byte
}

func (es *encryptionService) encrypt(data string) (string, error) {
	// For now, use simple base64 encoding as placeholder
	// In production, this would use AES-256-GCM like in the service
	return base64.StdEncoding.EncodeToString([]byte(data)), nil
}

func (es *encryptionService) decrypt(encryptedData string) (string, error) {
	// Decode from base64
	decoded, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %v", err)
	}
	return string(decoded), nil
}

// EncryptionConfigMiddleware adds encryption configuration to the context
func EncryptionConfigMiddleware(cfg *config.Config) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Add encryption config to context
			c.Set("encryption_config", cfg.Encryption)
			return next(c)
		}
	}
}

// GetEncryptionConfigFromContext retrieves encryption config from context
func GetEncryptionConfigFromContext(c echo.Context) (*config.EncryptionConfig, error) {
	encCfg, ok := c.Get("encryption_config").(*config.EncryptionConfig)
	if !ok {
		return nil, fmt.Errorf("encryption config not found in context")
	}
	return encCfg, nil
}

// EncryptionStatusHandler provides a handler to check encryption status
func EncryptionStatusHandler(c echo.Context) error {
	encCfg, err := GetEncryptionConfigFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Encryption config not available")
	}

	status := map[string]interface{}{
		"enabled":    encCfg.Enabled,
		"algorithm":  encCfg.Algorithm,
		"key_length": len(encCfg.Key),
		"timestamp":  time.Now().Unix(),
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":  "success",
		"message": "Encryption service status",
		"data":    status,
	})
}
