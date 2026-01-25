package account

import (
	"context"
	"fmt"

	"github.com/netbill/auth-svc/internal/core/errx"
	"github.com/netbill/auth-svc/internal/core/models"
)

func (s Service) Refresh(ctx context.Context, oldRefreshToken string) (models.TokensPair, error) {
	tokenData, err := s.jwt.ParseRefreshClaims(oldRefreshToken)
	if err != nil {
		return models.TokensPair{}, err
	}

	account, err := s.GetAccountByID(ctx, tokenData.AccountID)
	if err != nil {
		return models.TokensPair{}, err
	}

	token, err := s.repo.GetSessionToken(ctx, tokenData.SessionID)
	if err != nil {
		return models.TokensPair{}, err
	}
	if token == "" {
		return models.TokensPair{}, err
	}

	refreshHash, err := s.jwt.HashRefresh(token)
	if err != nil {
		return models.TokensPair{}, err
	}
	if refreshHash != oldRefreshToken {
		return models.TokensPair{}, errx.ErrorSessionTokenMismatch.Raise(
			fmt.Errorf(
				"refresh token does not match for session %s and account %s, cause: %w",
				tokenData.SessionID, tokenData.AccountID, err,
			),
		)
	}

	refresh, err := s.jwt.GenerateRefresh(account, tokenData.SessionID)
	if err != nil {
		return models.TokensPair{}, err
	}

	refreshNewHash, err := s.jwt.HashRefresh(refresh)
	if err != nil {
		return models.TokensPair{}, err
	}

	access, err := s.jwt.GenerateAccess(account, tokenData.SessionID)
	if err != nil {
		return models.TokensPair{}, err
	}

	_, err = s.repo.UpdateSessionToken(ctx, tokenData.SessionID, refreshNewHash)
	if err != nil {
		return models.TokensPair{}, err
	}

	return models.TokensPair{
		SessionID: tokenData.SessionID,
		Refresh:   refresh,
		Access:    access,
	}, nil
}
