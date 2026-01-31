// Package jwt provides JWT token creation and validation utilities.
package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenType distinguishes between access and refresh tokens.
type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

// Claims represents the JWT claims structure.
type Claims struct {
	UserID    uint      `json:"user_id"`
	Email     string    `json:"email"`
	TokenType TokenType `json:"token_type"`
	jwt.RegisteredClaims
}

// Manager handles JWT token operations.
type Manager struct {
	secretKey       []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

// NewManager creates a new JWT manager with the given configuration.
func NewManager(secretKey string, accessTTL, refreshTTL time.Duration) *Manager {
	return &Manager{
		secretKey:       []byte(secretKey),
		accessTokenTTL:  accessTTL,
		refreshTokenTTL: refreshTTL,
	}
}

// GenerateAccessToken creates a new access token for the given user.
func (m *Manager) GenerateAccessToken(userID uint, email string) (string, error) {
	return m.generateToken(userID, email, AccessToken, m.accessTokenTTL)
}

// GenerateRefreshToken creates a new refresh token for the given user.
func (m *Manager) GenerateRefreshToken(userID uint, email string) (string, error) {
	return m.generateToken(userID, email, RefreshToken, m.refreshTokenTTL)
}

// generateToken creates a JWT token with the specified parameters.
func (m *Manager) generateToken(userID uint, email string, tokenType TokenType, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID:    userID,
		Email:     email,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secretKey)
}

// ValidateToken parses and validates a JWT token, returning its claims.
func (m *Manager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return m.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// GetAccessTokenTTL returns the access token time-to-live duration.
func (m *Manager) GetAccessTokenTTL() time.Duration {
	return m.accessTokenTTL
}

// GetRefreshTokenTTL returns the refresh token time-to-live duration.
func (m *Manager) GetRefreshTokenTTL() time.Duration {
	return m.refreshTokenTTL
}
