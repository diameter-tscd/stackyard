package middleware

import (
	"bytes"
	"encoding/base64"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

// Obfuscator returns a middleware that Base64 encodes the response body
// if the enabled flag is true.
func Obfuscator(enabled bool) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if !enabled {
				return next(c)
			}

			// Only obfuscate API endpoints, exclude streaming or static API paths
			path := c.Request().URL.Path
			excludedPrefixes := []string{"/api/logs", "/api/cpu", "/api/user/photos"}

			if !strings.HasPrefix(path, "/api/") {
				return next(c)
			}

			for _, prefix := range excludedPrefixes {
				if strings.HasPrefix(path, prefix) {
					return next(c)
				}
			}

			// Create a recorder
			resBody := new(bytes.Buffer)
			recorder := &ResponseRecorder{
				ResponseWriter: c.Response().Writer,
				Body:           resBody,
				StatusCode:     http.StatusOK, // Default
			}
			c.Response().Writer = recorder

			// Call next handler
			err := next(c)

			// If handler returned an error (and didn't handle it), Echo will handle it.
			// But since we swapped the writer, Echo might write to our recorder or call ErrorHandler.
			// If err != nil, we usually return it and let a global ErrorHandler deal with it.
			// Ideally we want to capture *that* output too.
			// Echo's default error handler writes to the response writer.
			if err != nil {
				c.Error(err) // This writes the error JSON to our recorder
			}

			// Now we have the response in recorder.Body
			// We check if we should obfuscate

			// If explicitly marked as event-stream, do not touch
			contentType := recorder.Header().Get("Content-Type")
			if strings.Contains(contentType, "text/event-stream") {
				recorder.FlushOriginal()
				return nil
			}

			// Only obfuscate if content-type is json
			if !strings.Contains(contentType, "application/json") {
				recorder.FlushOriginal()
				return nil
			}

			// Obfuscate
			data := recorder.Body.Bytes()
			if len(data) > 0 {
				encoded := base64.StdEncoding.EncodeToString(data)
				encodedBytes := []byte(encoded)

				// Set headers
				recorder.ResponseWriter.Header().Set("X-Obfuscated", "true")
				// Set new Content-Length to avoid mismatch or truncation
				recorder.ResponseWriter.Header().Set("Content-Length", strconv.Itoa(len(encodedBytes)))

				recorder.ResponseWriter.WriteHeader(recorder.StatusCode)
				recorder.ResponseWriter.Write(encodedBytes)
			} else {
				recorder.ResponseWriter.WriteHeader(recorder.StatusCode)
			}

			return nil
		}
	}
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
	r.ResponseWriter.Header().Del("Content-Length") // Prevent mismatch if buffer differs from original header
	r.ResponseWriter.WriteHeader(r.StatusCode)
	r.ResponseWriter.Write(r.Body.Bytes())
}
