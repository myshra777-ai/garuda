package auth

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTConfig holds the configuration for JWT signing and validation.
type JWTConfig struct {
	privateKey ed25519.PrivateKey
	publicKey  ed25519.PublicKey
	issuer     string
	audience   string
	expiry     time.Duration
}

// NewJWTConfig generates a new Ed25519 key pair and returns a config.
func NewJWTConfig(issuer, audience string, expiry time.Duration) (*JWTConfig, error) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate Ed25519 key pair: %w", err)
	}
	return &JWTConfig{
		privateKey: priv,
		publicKey:  pub,
		issuer:     issuer,
		audience:   audience,
		expiry:     expiry,
	}, nil
}

// NewJWTConfigFromHex loads a key pair from hex strings (for persistence).
func NewJWTConfigFromHex(privateKeyHex, publicKeyHex, issuer, audience string, expiry time.Duration) (*JWTConfig, error) {
	privBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("failed to decode private key: %w", err)
	}
	pubBytes, err := hex.DecodeString(publicKeyHex)
	if err != nil {
		return nil, fmt.Errorf("failed to decode public key: %w", err)
	}
	if len(privBytes) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid private key length: expected %d, got %d", ed25519.PrivateKeySize, len(privBytes))
	}
	if len(pubBytes) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("invalid public key length: expected %d, got %d", ed25519.PublicKeySize, len(pubBytes))
	}
	return &JWTConfig{
		privateKey: ed25519.PrivateKey(privBytes),
		publicKey:  ed25519.PublicKey(pubBytes),
		issuer:     issuer,
		audience:   audience,
		expiry:     expiry,
	}, nil
}

// GenerateToken creates a JWT for a given actor.
func (c *JWTConfig) GenerateToken(actor string) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   actor,
		Issuer:    c.issuer,
		Audience:  jwt.ClaimStrings{c.audience},
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(c.expiry)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	return token.SignedString(c.privateKey)
}

// GenerateRefreshToken creates a long-lived token for session extension.
func (c *JWTConfig) GenerateRefreshToken(actor string) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   actor,
		Issuer:    c.issuer,
		Audience:  jwt.ClaimStrings{c.audience},
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)), // 7 days
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	return token.SignedString(c.privateKey)
}

// ValidateToken validates a JWT and returns the actor.
func (c *JWTConfig) ValidateToken(tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodEd25519); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return c.publicKey, nil
	})
	if err != nil {
		return "", fmt.Errorf("token validation failed: %w", err)
	}
	if !token.Valid {
		return "", errors.New("token is invalid")
	}
	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return "", errors.New("invalid claims structure")
	}
	return claims.Subject, nil
}

// ValidateRefreshToken validates a refresh token and returns the actor.
func (c *JWTConfig) ValidateRefreshToken(tokenString string) (string, error) {
	return c.ValidateToken(tokenString)
}

// GetPublicKeyHex returns the public key as a hex string (for sharing).
func (c *JWTConfig) GetPublicKeyHex() string {
	return hex.EncodeToString(c.publicKey)
}

// GetPrivateKeyHex returns the private key as a hex string (for persistence).
func (c *JWTConfig) GetPrivateKeyHex() string {
	return hex.EncodeToString(c.privateKey)
}

// JWTMiddleware extracts and validates the JWT from the Authorization header.
// JWTMiddleware extracts and validates the JWT from the Authorization header.
func JWTMiddleware(config *JWTConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "missing authorization header", http.StatusUnauthorized)
				return
			}
			// Expect "Bearer <token>"
			if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
				http.Error(w, "invalid authorization header format", http.StatusUnauthorized)
				return
			}
			token := authHeader[7:]
			actor, err := config.ValidateToken(token) // <-- Changed from VerifyToken
			if err != nil {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}
			// Set actor in context
			ctx := r.Context()
			ctx = context.WithValue(ctx, "actor", actor)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
