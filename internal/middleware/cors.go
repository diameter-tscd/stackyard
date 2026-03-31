package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           int
}

// DefaultCORSConfig returns default CORS configuration
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{
			http.MethodGet,
			http.MethodHead,
			http.MethodPut,
			http.MethodPatch,
			http.MethodPost,
			http.MethodDelete,
		},
		AllowHeaders: []string{
			echo.HeaderOrigin,
			echo.HeaderContentType,
			echo.HeaderAccept,
			echo.HeaderAuthorization,
		},
		ExposeHeaders:    []string{},
		AllowCredentials: false,
		MaxAge:           86400,
	}
}

// CORS returns CORS middleware
func CORS(config ...CORSConfig) echo.MiddlewareFunc {
	var cfg CORSConfig
	if len(config) > 0 {
		cfg = config[0]
	} else {
		cfg = DefaultCORSConfig()
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			res := c.Response()

			origin := req.Header.Get(echo.HeaderOrigin)

			allowOrigin := ""
			for _, o := range cfg.AllowOrigins {
				if o == "*" {
					allowOrigin = "*"
					break
				}
				if o == origin {
					allowOrigin = origin
					break
				}
				if strings.HasPrefix(o, "*.") {
					domain := strings.TrimPrefix(o, "*")
					if strings.HasSuffix(origin, domain) {
						allowOrigin = origin
						break
					}
				}
			}

			if allowOrigin != "" {
				res.Header().Set(echo.HeaderAccessControlAllowOrigin, allowOrigin)
			}

			if req.Method == http.MethodOptions {
				res.Header().Set(echo.HeaderAccessControlAllowMethods, strings.Join(cfg.AllowMethods, ","))
				res.Header().Set(echo.HeaderAccessControlAllowHeaders, strings.Join(cfg.AllowHeaders, ","))

				if len(cfg.ExposeHeaders) > 0 {
					res.Header().Set(echo.HeaderAccessControlExposeHeaders, strings.Join(cfg.ExposeHeaders, ","))
				}

				if cfg.AllowCredentials {
					res.Header().Set(echo.HeaderAccessControlAllowCredentials, "true")
				}

				if cfg.MaxAge > 0 {
					res.Header().Set(echo.HeaderAccessControlMaxAge, string(rune(cfg.MaxAge)))
				}

				return c.NoContent(http.StatusNoContent)
			}

			if len(cfg.ExposeHeaders) > 0 {
				res.Header().Set(echo.HeaderAccessControlExposeHeaders, strings.Join(cfg.ExposeHeaders, ","))
			}

			if cfg.AllowCredentials {
				res.Header().Set(echo.HeaderAccessControlAllowCredentials, "true")
			}

			return next(c)
		}
	}
}

// CORSWithConfig returns CORS middleware with custom config
func CORSWithConfig(allowOrigins []string) echo.MiddlewareFunc {
	return CORS(CORSConfig{
		AllowOrigins: allowOrigins,
		AllowMethods: []string{
			http.MethodGet,
			http.MethodHead,
			http.MethodPut,
			http.MethodPatch,
			http.MethodPost,
			http.MethodDelete,
		},
		AllowHeaders: []string{
			echo.HeaderOrigin,
			echo.HeaderContentType,
			echo.HeaderAccept,
			echo.HeaderAuthorization,
		},
		ExposeHeaders:    []string{},
		AllowCredentials: false,
		MaxAge:           86400,
	})
}

// CORSAllowAll returns CORS middleware that allows all origins
func CORSAllowAll() echo.MiddlewareFunc {
	return CORS(DefaultCORSConfig())
}
