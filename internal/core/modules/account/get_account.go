package account

import (
	"context"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/core/models"
)

func (s Service) GetAccountByID(ctx context.Context, ID uuid.UUID) (models.Account, error) {
	return s.repo.GetAccountByID(ctx, ID)
}

func (s Service) GetAccountByEmail(ctx context.Context, email string) (models.Account, error) {
	return s.repo.GetAccountByEmail(ctx, email)
}

func (s Service) GetAccountByUsername(ctx context.Context, username string) (models.Account, error) {
	return s.repo.GetAccountByUsername(ctx, username)
}

func (s Service) GetAccountEmail(ctx context.Context, ID uuid.UUID) (models.AccountEmail, error) {
	return s.repo.GetAccountEmail(ctx, ID)
}
