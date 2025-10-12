package security

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/nicklaros/jalanrusak-be/core/ports/external"
)

// JWTTokenGenerator implements the TokenGenerator interface using JWT
type JWTTokenGenerator struct {
	secretKey      []byte
	accessTokenTTL time.Duration
}

// NewJWTTokenGenerator creates a new JWT token generator
func NewJWTTokenGenerator(secretKey string, accessTokenTTLHours int) external.TokenGenerator {
	return &JWTTokenGenerator{
		secretKey:      []byte(secretKey),
		accessTokenTTL: time.Duration(accessTokenTTLHours) * time.Hour,
	}
}

// Claims represents the JWT claims structure
type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// GenerateAccessToken creates a new JWT access token for the given user ID
func (g *JWTTokenGenerator) GenerateAccessToken(ctx context.Context, userID string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(g.accessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(g.secretKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// GenerateRefreshToken creates a new cryptographically secure refresh token
func (g *JWTTokenGenerator) GenerateRefreshToken(ctx context.Context) (string, error) {
	// Generate 32 random bytes
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Encode as base64 URL-safe string
	token := base64.URLEncoding.EncodeToString(b)
	return token, nil
}

// ValidateAccessToken validates an access token and returns the user ID
func (g *JWTTokenGenerator) ValidateAccessToken(ctx context.Context, tokenString string) (userID string, err error) {
	// Parse token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return g.secretKey, nil
	})

	if err != nil {
		return "", fmt.Errorf("failed to parse token: %w", err)
	}

	// Extract claims
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return "", fmt.Errorf("invalid token claims")
	}

	return claims.UserID, nil
}

// HashToken creates a SHA-256 hash of the token for secure storage
func (g *JWTTokenGenerator) HashToken(ctx context.Context, token string) (string, error) {
	hash := sha256.Sum256([]byte(token))
	return base64.URLEncoding.EncodeToString(hash[:]), nil
}
