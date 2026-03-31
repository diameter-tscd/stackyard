package middleware

import (
	"github.com/labstack/echo/v4"
)

// SecurityConfig holds security headers configuration
type SecurityConfig struct {
	ContentSecurityPolicy string
	XContentTypeOptions   string
	XFrameOptions         string
	XXSSProtection        string
	ReferrerPolicy        string
	PermissionsPolicy     string
	HSTSMaxAge            int
	HSTSIncludeSubdomains bool
	HSTSPreload           bool
}

// DefaultSecurityConfig returns default security headers configuration
func DefaultSecurityConfig() SecurityConfig {
	return SecurityConfig{
		ContentSecurityPolicy: "default-src 'self'",
		XContentTypeOptions:   "nosniff",
		XFrameOptions:         "DENY",
		XXSSProtection:        "1; mode=block",
		ReferrerPolicy:        "strict-origin-when-cross-origin",
		PermissionsPolicy:     "camera=(), microphone=(), geolocation=()",
		HSTSMaxAge:            31536000,
		HSTSIncludeSubdomains: true,
		HSTSPreload:           false,
	}
}

// Security returns security headers middleware
func Security(config ...SecurityConfig) echo.MiddlewareFunc {
	var cfg SecurityConfig
	if len(config) > 0 {
		cfg = config[0]
	} else {
		cfg = DefaultSecurityConfig()
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			res := c.Response()

			if cfg.ContentSecurityPolicy != "" {
				res.Header().Set("Content-Security-Policy", cfg.ContentSecurityPolicy)
			}

			if cfg.XContentTypeOptions != "" {
				res.Header().Set("X-Content-Type-Options", cfg.XContentTypeOptions)
			}

			if cfg.XFrameOptions != "" {
				res.Header().Set("X-Frame-Options", cfg.XFrameOptions)
			}

			if cfg.XXSSProtection != "" {
				res.Header().Set("X-XSS-Protection", cfg.XXSSProtection)
			}

			if cfg.ReferrerPolicy != "" {
				res.Header().Set("Referrer-Policy", cfg.ReferrerPolicy)
			}

			if cfg.PermissionsPolicy != "" {
				res.Header().Set("Permissions-Policy", cfg.PermissionsPolicy)
			}

			if cfg.HSTSMaxAge > 0 {
				hsts := "max-age=" + string(rune(cfg.HSTSMaxAge))
				if cfg.HSTSIncludeSubdomains {
					hsts += "; includeSubDomains"
				}
				if cfg.HSTSPreload {
					hsts += "; preload"
				}
				res.Header().Set("Strict-Transport-Security", hsts)
			}

			return next(c)
		}
	}
}

// SecurityWithConfig returns security headers middleware with custom config
func SecurityWithConfig(csp string) echo.MiddlewareFunc {
	return Security(SecurityConfig{
		ContentSecurityPolicy: csp,
		XContentTypeOptions:   "nosniff",
		XFrameOptions:         "DENY",
		XXSSProtection:        "1; mode=block",
		ReferrerPolicy:        "strict-origin-when-cross-origin",
		PermissionsPolicy:     "camera=(), microphone=(), geolocation=()",
		HSTSMaxAge:            31536000,
		HSTSIncludeSubdomains: true,
		HSTSPreload:           false,
	})
}

// SecurityPermissive returns security headers middleware with permissive settings
func SecurityPermissive() echo.MiddlewareFunc {
	return Security(SecurityConfig{
		ContentSecurityPolicy: "default-src 'self' 'unsafe-inline' 'unsafe-eval'",
		XContentTypeOptions:   "nosniff",
		XFrameOptions:         "SAMEORIGIN",
		XXSSProtection:        "1; mode=block",
		ReferrerPolicy:        "strict-origin-when-cross-origin",
		PermissionsPolicy:     "camera=(), microphone=(), geolocation=()",
		HSTSMaxAge:            31536000,
		HSTSIncludeSubdomains: true,
		HSTSPreload:           false,
	})
}
