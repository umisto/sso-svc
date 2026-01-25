package account

import (
	"context"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/core/models"
	"github.com/netbill/restkit/pagi"
)

func (s Service) GetOwnSession(ctx context.Context, initiator InitiatorData, sessionID uuid.UUID) (models.Session, error) {
	_, _, err := s.validateSession(ctx, initiator)
	if err != nil {
		return models.Session{}, err
	}

	session, err := s.repo.GetAccountSession(ctx, initiator.AccountID, sessionID)
	if err != nil {
		return models.Session{}, err
	}

	return session, nil
}

func (s Service) GetOwnSessions(
	ctx context.Context,
	initiator InitiatorData,
	limit, offset uint,
) (pagi.Page[[]models.Session], error) {
	_, _, err := s.validateSession(ctx, initiator)
	if err != nil {
		return pagi.Page[[]models.Session]{}, err
	}

	sessions, err := s.repo.GetSessionsForAccount(ctx, initiator.AccountID, limit, offset)
	if err != nil {
		return pagi.Page[[]models.Session]{}, err
	}

	return sessions, nil
}
