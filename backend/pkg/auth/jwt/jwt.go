// Package jwt provides JWT token generation and validation
package jwt

import (
	"crypto/rsa"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims represents JWT custom claims
type Claims struct {
	Subject  string `json:"sub"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// Manager manages JWT tokens
type Manager struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	config     Config
}

// Config holds JWT configuration
type Config struct {
	Issuer          string
	AccessDuration  time.Duration
	RefreshDuration time.Duration
}

// NewManager creates a new JWT manager
func NewManager(privateKey *rsa.PrivateKey, publicKey *rsa.PublicKey, config Config) *Manager {
	return &Manager{
		privateKey: privateKey,
		publicKey:  publicKey,
		config:     config,
	}
}

// GenerateAccessToken generates an access token
func (m *Manager) GenerateAccessToken(userID, username string) (string, error) {
	now := time.Now()
	claims := Claims{
		Subject:  userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(m.config.AccessDuration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    m.config.Issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(m.privateKey)
}

// GenerateRefreshToken generates a refresh token
func (m *Manager) GenerateRefreshToken(userID string) (string, error) {
	tokenID := uuid.New().String()
	now := time.Now()

	claims := jwt.MapClaims{
		"sub": tokenID,
		"uid": userID,
		"typ": "refresh",
		"exp": now.Add(m.config.RefreshDuration).Unix(),
		"iat": now.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(m.privateKey)
}

// ValidateAccessToken validates an access token
func (m *Manager) ValidateAccessToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.publicKey, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

// ValidateRefreshToken validates a refresh token and returns tokenID and userID
func (m *Manager) ValidateRefreshToken(tokenString string) (string, string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.publicKey, nil
	})

	if err != nil {
		return "", "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", "", fmt.Errorf("invalid token")
	}

	tokenType, _ := claims["typ"].(string)
	if tokenType != "refresh" {
		return "", "", fmt.Errorf("not a refresh token")
	}

	tokenID, _ := claims["sub"].(string)
	userID, _ := claims["uid"].(string)

	return tokenID, userID, nil
}
