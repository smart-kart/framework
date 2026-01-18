package jwt

import (
	"errors"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/cozy-hub-app/framework/env"
)

var (
	// ErrInvalidToken is returned when token is invalid
	ErrInvalidToken = errors.New("invalid token")
	// ErrExpiredToken is returned when token has expired
	ErrExpiredToken = errors.New("token has expired")
	// ErrTokenNotValidYet is returned when token is not yet valid
	ErrTokenNotValidYet = errors.New("token not valid yet")
)

// JWTClaims defines the structure of JWT token claims
type JWTClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role,omitempty"` // "admin" for admin users, empty for regular users
	jwt.RegisteredClaims
}

// JWTManager handles JWT token generation and validation
type JWTManager struct {
	secretKey       string
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
	issuer          string
}

var (
	jwtManagerInstance *JWTManager
	jwtManagerOnce     sync.Once
)

// NewJWTManager creates a new JWT manager
func NewJWTManager(secretKey string, accessTokenTTL, refreshTokenTTL time.Duration, issuer string) *JWTManager {
	return &JWTManager{
		secretKey:       secretKey,
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
		issuer:          issuer,
	}
}

// GetJWTManager returns the singleton JWT manager instance
// This is initialized once from environment variables
func GetJWTManager() *JWTManager {
	jwtManagerOnce.Do(func() {
		// Get JWT config from environment
		secretKey := env.Get(env.JWTSecretKey)

		// CRITICAL SECURITY: JWT secret MUST be set and strong
		// A weak or missing secret allows anyone to forge authentication tokens
		if secretKey == "" {
			panic("FATAL: JWT_SECRET_KEY environment variable is not set. " +
				"This is a critical security requirement. " +
				"Generate a secure key with: openssl rand -base64 64")
		}

		// Enforce minimum key length (64 characters = ~384 bits of entropy)
		// Industry standard for HMAC-SHA256 is at least 256 bits
		if len(secretKey) < 64 {
			panic("FATAL: JWT_SECRET_KEY must be at least 64 characters long for security. " +
				"Current length: " + string(rune(len(secretKey))) + ". " +
				"Generate a secure key with: openssl rand -base64 64")
		}

		// Parse TTL durations
		accessTTL := parseDuration(env.Get(env.JWTAccessTokenTTL), 15*time.Minute)
		refreshTTL := parseDuration(env.Get(env.JWTRefreshTokenTTL), 168*time.Hour)

		issuer := env.Get(env.JWTIssuer)
		if issuer == "" {
			issuer = "cozy-hub-service"
		}

		jwtManagerInstance = NewJWTManager(secretKey, accessTTL, refreshTTL, issuer)
	})

	return jwtManagerInstance
}

// GenerateAccessToken generates an access token for regular users
func (m *JWTManager) GenerateAccessToken(userID, email string) (string, error) {
	return m.generateToken(userID, email, m.accessTokenTTL)
}

// GenerateRefreshToken generates a refresh token for regular users
func (m *JWTManager) GenerateRefreshToken(userID, email string) (string, error) {
	return m.generateToken(userID, email, m.refreshTokenTTL)
}

// generateToken is the internal token generation method
func (m *JWTManager) generateToken(userID, email string, ttl time.Duration) (string, error) {
	return m.generateTokenWithRole(userID, email, "", ttl)
}

// generateTokenWithRole generates a token with an optional role
func (m *JWTManager) generateTokenWithRole(userID, email, role string, ttl time.Duration) (string, error) {
	now := time.Now()
	expiresAt := now.Add(ttl)

	claims := JWTClaims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    m.issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.secretKey))
}

// GenerateAdminAccessToken generates an access token with admin role
func (m *JWTManager) GenerateAdminAccessToken(userID, email string) (string, error) {
	return m.generateTokenWithRole(userID, email, "admin", m.accessTokenTTL)
}

// GenerateAdminRefreshToken generates a refresh token with admin role
func (m *JWTManager) GenerateAdminRefreshToken(userID, email string) (string, error) {
	return m.generateTokenWithRole(userID, email, "admin", m.refreshTokenTTL)
}

// ValidateToken validates a JWT token and returns the claims
func (m *JWTManager) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&JWTClaims{},
		func(token *jwt.Token) (interface{}, error) {
			// Verify the signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, ErrInvalidToken
			}
			return []byte(m.secretKey), nil
		},
	)

	if err != nil {
		// Check for specific error types
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		if errors.Is(err, jwt.ErrTokenNotValidYet) {
			return nil, ErrTokenNotValidYet
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// RefreshAccessToken validates a refresh token and generates a new access token
func (m *JWTManager) RefreshAccessToken(refreshToken string) (string, error) {
	claims, err := m.ValidateToken(refreshToken)
	if err != nil {
		return "", err
	}

	return m.GenerateAccessToken(claims.UserID, claims.Email)
}

// GetUserIDFromToken extracts the user ID from a token without full validation
// Use this carefully - prefer ValidateToken for proper validation
func (m *JWTManager) GetUserIDFromToken(tokenString string) (string, error) {
	claims, err := m.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}
	return claims.UserID, nil
}

// GenerateTokenPair generates a pair of access and refresh tokens for regular users
func GenerateTokenPair(userID, email string) (accessToken string, refreshToken string, err error) {
	jwtManager := GetJWTManager()

	accessToken, err = jwtManager.GenerateAccessToken(userID, email)
	if err != nil {
		return "", "", err
	}

	refreshToken, err = jwtManager.GenerateRefreshToken(userID, email)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// GenerateAdminTokenPair generates a pair of access and refresh tokens with admin role
func GenerateAdminTokenPair(userID, email string) (accessToken string, refreshToken string, err error) {
	jwtManager := GetJWTManager()

	accessToken, err = jwtManager.GenerateAdminAccessToken(userID, email)
	if err != nil {
		return "", "", err
	}

	refreshToken, err = jwtManager.GenerateAdminRefreshToken(userID, email)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// parseDuration parses a duration string, returns defaultDuration if parsing fails
func parseDuration(durationStr string, defaultDuration time.Duration) time.Duration {
	if durationStr == "" {
		return defaultDuration
	}

	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		return defaultDuration
	}

	return duration
}
