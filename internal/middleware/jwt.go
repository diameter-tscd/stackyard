package middleware

import (
	"errors"
	"strings"
	"time"

	"stackyard/pkg/response"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

// JWTConfig holds JWT configuration
type JWTConfig struct {
	SigningKey     string
	TokenLookup    string
	AuthScheme     string
	Skipper        func(c echo.Context) bool
	TokenValidator func(token string) (jwt.Claims, error)
}

// DefaultJWTConfig returns default JWT configuration
func DefaultJWTConfig(signingKey string) JWTConfig {
	return JWTConfig{
		SigningKey:  signingKey,
		TokenLookup: "header:Authorization",
		AuthScheme:  "Bearer",
		Skipper:     nil,
	}
}

// JWTClaims represents JWT claims
type JWTClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// JWT returns JWT authentication middleware
func JWT(config ...JWTConfig) echo.MiddlewareFunc {
	var cfg JWTConfig
	if len(config) > 0 {
		cfg = config[0]
	} else {
		cfg = DefaultJWTConfig("")
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if cfg.Skipper != nil && cfg.Skipper(c) {
				return next(c)
			}

			token, err := extractToken(c, cfg)
			if err != nil {
				return response.Unauthorized(c, "Missing or invalid token")
			}

			claims, err := validateToken(token, cfg.SigningKey)
			if err != nil {
				return response.Unauthorized(c, "Invalid token")
			}

			c.Set("user_id", claims.UserID)
			c.Set("username", claims.Username)
			c.Set("email", claims.Email)
			c.Set("role", claims.Role)
			c.Set("claims", claims)

			return next(c)
		}
	}
}

// JWTWithConfig returns JWT middleware with custom config
func JWTWithConfig(signingKey string) echo.MiddlewareFunc {
	return JWT(DefaultJWTConfig(signingKey))
}

// JWTRequired returns JWT middleware that requires authentication
func JWTRequired(signingKey string) echo.MiddlewareFunc {
	return JWT(DefaultJWTConfig(signingKey))
}

// JWTOptional returns JWT middleware that allows optional authentication
func JWTOptional(signingKey string) echo.MiddlewareFunc {
	cfg := DefaultJWTConfig(signingKey)
	cfg.Skipper = func(c echo.Context) bool {
		auth := c.Request().Header.Get(echo.HeaderAuthorization)
		if auth == "" {
			return true
		}
		return false
	}
	return JWT(cfg)
}

// extractToken extracts token from request
func extractToken(c echo.Context, cfg JWTConfig) (string, error) {
	parts := strings.Split(cfg.TokenLookup, ":")
	if len(parts) != 2 {
		return "", errors.New("invalid token lookup format")
	}

	source := parts[0]
	key := parts[1]

	var token string
	switch source {
	case "header":
		auth := c.Request().Header.Get(key)
		if auth == "" {
			return "", errors.New("missing authorization header")
		}

		if cfg.AuthScheme != "" {
			parts := strings.Split(auth, " ")
			if len(parts) != 2 || parts[0] != cfg.AuthScheme {
				return "", errors.New("invalid authorization scheme")
			}
			token = parts[1]
		} else {
			token = auth
		}
	case "query":
		token = c.QueryParam(key)
	case "cookie":
		cookie, err := c.Cookie(key)
		if err != nil {
			return "", err
		}
		token = cookie.Value
	}

	if token == "" {
		return "", errors.New("token not found")
	}

	return token, nil
}

// validateToken validates JWT token and returns claims
func validateToken(tokenString, signingKey string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(signingKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// GenerateToken generates a new JWT token
func GenerateToken(userID, username, email, role, signingKey string, expiration time.Duration) (string, error) {
	claims := JWTClaims{
		UserID:   userID,
		Username: username,
		Email:    email,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(signingKey))
}

// GenerateTokenWithClaims generates a JWT token with custom claims
func GenerateTokenWithClaims(claims *JWTClaims, signingKey string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(signingKey))
}

// GetUserID extracts user ID from context
func GetUserID(c echo.Context) string {
	if userID := c.Get("user_id"); userID != nil {
		return userID.(string)
	}
	return ""
}

// GetUsername extracts username from context
func GetUsername(c echo.Context) string {
	if username := c.Get("username"); username != nil {
		return username.(string)
	}
	return ""
}

// GetUserEmail extracts user email from context
func GetUserEmail(c echo.Context) string {
	if email := c.Get("email"); email != nil {
		return email.(string)
	}
	return ""
}

// GetUserRole extracts user role from context
func GetUserRole(c echo.Context) string {
	if role := c.Get("role"); role != nil {
		return role.(string)
	}
	return ""
}

// RequireRole returns middleware that requires specific role
func RequireRole(roles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userRole := GetUserRole(c)

			for _, role := range roles {
				if userRole == role {
					return next(c)
				}
			}

			return response.Forbidden(c, "Insufficient permissions")
		}
	}
}

// RequireAdmin returns middleware that requires admin role
func RequireAdmin() echo.MiddlewareFunc {
	return RequireRole("admin")
}

// RequireUser returns middleware that requires user role
func RequireUser() echo.MiddlewareFunc {
	return RequireRole("user", "admin")
}
