// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package auth

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Unexported type to prevent context key collisions across packages.
type contextKey string

const (
	actorContextKey    contextKey = "actor"
	tenantIDContextKey contextKey = "tenant_id"
)

// CustomClaims extends standard JWT claims with multi-tenant and actor fields.
type CustomClaims struct {
	jwt.RegisteredClaims
	TenantID uuid.UUID `json:"tenant_id,omitempty"`
}

// JWTConfig holds the configuration and key pairs for JWT signing and verification.
type JWTConfig struct {
	privateKey ed25519.PrivateKey
	publicKey  ed25519.PublicKey
	issuer     string
	audience   string
	expiry     time.Duration
}

// NewJWTConfig generates a new Ed25519 key pair for signing and verification.
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

// NewJWTConfigFromHex loads key pairs from hex strings for persistent environments.
func NewJWTConfigFromHex(privateKeyHex, publicKeyHex, issuer, audience string, expiry time.Duration) (*JWTConfig, error) {
	privBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("failed to decode private key hex: %w", err)
	}
	pubBytes, err := hex.DecodeString(publicKeyHex)
	if err != nil {
		return nil, fmt.Errorf("failed to decode public key hex: %w", err)
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

// GenerateToken creates a tenant-scoped JWT for an actor.
func (c *JWTConfig) GenerateToken(actor string, tenantID uuid.UUID) (string, error) {
	now := time.Now().UTC()
	claims := CustomClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    c.issuer,
			Audience:  jwt.ClaimStrings{c.audience},
			Subject:   actor,
			ExpiresAt: jwt.NewNumericDate(now.Add(c.expiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
		TenantID: tenantID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	signedToken, err := token.SignedString(c.privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}
	return signedToken, nil
}

// GenerateTokenWithTenant is a convenience wrapper for callers that still pass tenant IDs as strings.
func (c *JWTConfig) GenerateTokenWithTenant(actor, tenantID string) (string, error) {
	parsedTenantID, err := uuid.Parse(tenantID)
	if err != nil {
		return "", fmt.Errorf("invalid tenant id: %w", err)
	}
	return c.GenerateToken(actor, parsedTenantID)
}

// GenerateRefreshToken creates a long-lived refresh token for session renewals.
func (c *JWTConfig) GenerateRefreshToken(actor string, tenantID uuid.UUID) (string, error) {
	now := time.Now().UTC()
	claims := CustomClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    c.issuer,
			Audience:  jwt.ClaimStrings{c.audience},
			Subject:   actor,
			ExpiresAt: jwt.NewNumericDate(now.Add(7 * 24 * time.Hour)), // 7 Days
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
		TenantID: tenantID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	signedToken, err := token.SignedString(c.privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign refresh token: %w", err)
	}
	return signedToken, nil
}

// GenerateRefreshTokenWithTenant is a convenience wrapper for callers that still pass tenant IDs as strings.
func (c *JWTConfig) GenerateRefreshTokenWithTenant(actor, tenantID string) (string, error) {
	parsedTenantID, err := uuid.Parse(tenantID)
	if err != nil {
		return "", fmt.Errorf("invalid tenant id: %w", err)
	}
	return c.GenerateRefreshToken(actor, parsedTenantID)
}

// ValidateToken parses, validates, and extracts claims from a JWT string.
func (c *JWTConfig) ValidateToken(tokenString string) (string, uuid.UUID, error) {
	claims := &CustomClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodEd25519); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return c.publicKey, nil
	})

	if err != nil {
		return "", uuid.Nil, fmt.Errorf("token validation failed: %w", err)
	}

	if !token.Valid {
		return "", uuid.Nil, errors.New("token is invalid")
	}

	if claims.Subject == "" {
		return "", uuid.Nil, errors.New("missing subject claim")
	}

	return claims.Subject, claims.TenantID, nil
}

// ValidateRefreshToken validates a refresh token and returns the actor and tenant.
func (c *JWTConfig) ValidateRefreshToken(tokenString string) (string, uuid.UUID, error) {
	return c.ValidateToken(tokenString)
}

// GetPublicKeyHex returns the public key encoded as a hex string.
func (c *JWTConfig) GetPublicKeyHex() string {
	return hex.EncodeToString(c.publicKey)
}

// GetPrivateKeyHex returns the private key encoded as a hex string.
func (c *JWTConfig) GetPrivateKeyHex() string {
	return hex.EncodeToString(c.privateKey)
}

// ContextWithActorAndTenant attaches actor and tenant values to the request context.
func ContextWithActorAndTenant(ctx context.Context, actor string, tenantID any) context.Context {
	ctx = context.WithValue(ctx, actorContextKey, actor)

	switch v := tenantID.(type) {
	case uuid.UUID:
		if v != uuid.Nil {
			ctx = context.WithValue(ctx, tenantIDContextKey, v)
		}
	case string:
		if parsed, err := uuid.Parse(strings.TrimSpace(v)); err == nil && parsed != uuid.Nil {
			ctx = context.WithValue(ctx, tenantIDContextKey, parsed)
		}
	}

	return ctx
}

// ActorFromContext retrieves the actor from the context.
func ActorFromContext(ctx context.Context) (string, bool) {
	actor, ok := ctx.Value(actorContextKey).(string)
	return actor, ok
}

// TenantIDFromContext retrieves the tenant UUID from the context.
func TenantIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	switch v := ctx.Value(tenantIDContextKey).(type) {
	case uuid.UUID:
		return v, true
	case string:
		if parsed, err := uuid.Parse(strings.TrimSpace(v)); err == nil {
			return parsed, true
		}
	}
	return uuid.Nil, false
}

// JWTMiddleware validates incoming Bearer tokens and enriches request context with actor and tenant claims.
func JWTMiddleware(config *JWTConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, `{"error":"missing authorization header"}`, http.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				http.Error(w, `{"error":"invalid authorization header format"}`, http.StatusUnauthorized)
				return
			}

			tokenString := parts[1]
			actor, tenantID, err := config.ValidateToken(tokenString)
			if err != nil {
				http.Error(w, `{"error":"invalid or expired token"}`, http.StatusUnauthorized)
				return
			}

			ctx := ContextWithActorAndTenant(r.Context(), actor, tenantID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
