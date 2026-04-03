package middleware

import (
	"time"

	"stackyrd/pkg/logger"

	"github.com/labstack/echo/v4"
)

// AuditConfig holds audit logging configuration
type AuditConfig struct {
	Logger           *logger.Logger
	Skipper          func(c echo.Context) bool
	LogRequestBody   bool
	LogHeaders       bool
	SensitiveHeaders []string
}

// DefaultAuditConfig returns default audit configuration
func DefaultAuditConfig(log *logger.Logger) AuditConfig {
	return AuditConfig{
		Logger:         log,
		LogRequestBody: false,
		LogHeaders:     false,
		SensitiveHeaders: []string{
			"Authorization",
			"Cookie",
			"X-Api-Key",
		},
	}
}

// AuditLog represents an audit log entry
type AuditLog struct {
	Timestamp   time.Time              `json:"timestamp"`
	Method      string                 `json:"method"`
	Path        string                 `json:"path"`
	Query       string                 `json:"query,omitempty"`
	StatusCode  int                    `json:"status_code"`
	Latency     time.Duration          `json:"latency"`
	UserID      string                 `json:"user_id,omitempty"`
	Username    string                 `json:"username,omitempty"`
	IP          string                 `json:"ip"`
	UserAgent   string                 `json:"user_agent"`
	RequestID   string                 `json:"request_id,omitempty"`
	RequestBody interface{}            `json:"request_body,omitempty"`
	Headers     map[string]string      `json:"headers,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Audit returns audit logging middleware
func Audit(config ...AuditConfig) echo.MiddlewareFunc {
	var cfg AuditConfig
	if len(config) > 0 {
		cfg = config[0]
	} else {
		cfg = DefaultAuditConfig(nil)
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if cfg.Skipper != nil && cfg.Skipper(c) {
				return next(c)
			}

			start := time.Now()
			req := c.Request()

			err := next(c)

			latency := time.Since(start)

			auditLog := AuditLog{
				Timestamp:  start,
				Method:     req.Method,
				Path:       req.URL.Path,
				Query:      req.URL.RawQuery,
				StatusCode: c.Response().Status,
				Latency:    latency,
				IP:         c.RealIP(),
				UserAgent:  req.UserAgent(),
				RequestID:  c.Response().Header().Get(echo.HeaderXRequestID),
			}

			if userID := c.Get("user_id"); userID != nil {
				auditLog.UserID = userID.(string)
			}

			if username := c.Get("username"); username != nil {
				auditLog.Username = username.(string)
			}

			if cfg.LogHeaders {
				headers := make(map[string]string)
				for key, values := range req.Header {
					skip := false
					for _, sensitive := range cfg.SensitiveHeaders {
						if key == sensitive {
							skip = true
							break
						}
					}
					if !skip && len(values) > 0 {
						headers[key] = values[0]
					}
				}
				auditLog.Headers = headers
			}

			if cfg.Logger != nil {
				logMsg := "Audit log"
				logFields := []interface{}{
					"method", auditLog.Method,
					"path", auditLog.Path,
					"status", auditLog.StatusCode,
					"latency", auditLog.Latency.String(),
					"ip", auditLog.IP,
				}

				if auditLog.UserID != "" {
					logFields = append(logFields, "user_id", auditLog.UserID)
				}

				if auditLog.Username != "" {
					logFields = append(logFields, "username", auditLog.Username)
				}

				if auditLog.StatusCode >= 400 {
					cfg.Logger.Warn(logMsg, logFields...)
				} else {
					cfg.Logger.Info(logMsg, logFields...)
				}
			}

			return err
		}
	}
}

// AuditWithConfig returns audit middleware with custom config
func AuditWithConfig(log *logger.Logger) echo.MiddlewareFunc {
	return Audit(DefaultAuditConfig(log))
}

// AuditSkipHealthCheck returns audit middleware that skips health check endpoints
func AuditSkipHealthCheck(log *logger.Logger) echo.MiddlewareFunc {
	cfg := DefaultAuditConfig(log)
	cfg.Skipper = func(c echo.Context) bool {
		return c.Request().URL.Path == "/health" || c.Request().URL.Path == "/health/infrastructure"
	}
	return Audit(cfg)
}
