package gateway

import (
	"context"
	"errors"

	"example.com/corp/auth"
)

type GatewayProxy struct {
	validator *auth.TokenValidator
}

func NewGatewayProxy(v *auth.TokenValidator) *GatewayProxy {
	return &GatewayProxy{validator: v}
}

// AuthenticateRequest handles incoming HTTP auth headers across module boundaries.
func (g *GatewayProxy) AuthenticateRequest(ctx context.Context, header string) error {
	if !g.validator.ValidateToken(header) {
		return errors.New("unauthorized token")
	}
	return nil
}
