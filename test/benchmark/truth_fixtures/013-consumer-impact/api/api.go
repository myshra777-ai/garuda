package api

import (
	"context"
	"github.com/myshra777-ai/garuda/test/benchmark/truth_fixtures/013-consumer-impact/service"
)

type UserHandler struct {
	svc *service.UserService
}

func (h *UserHandler) HandleRegisterUser(ctx context.Context, id, email string) error {
	return h.svc.RegisterUser(ctx, id, email)
}
