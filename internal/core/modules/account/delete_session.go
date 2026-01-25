package account

import (
	"context"

	"github.com/google/uuid"
)

func (s Service) Logout(ctx context.Context, initiator InitiatorData) error {
	err := s.repo.DeleteAccountSession(ctx, initiator.AccountID, initiator.SessionID)
	if err != nil {
		return err
	}

	return nil
}

func (s Service) DeleteOwnSession(ctx context.Context, initiator InitiatorData, sessionID uuid.UUID) error {
	_, _, err := s.validateSession(ctx, initiator)
	if err != nil {
		return err
	}

	err = s.repo.DeleteAccountSession(ctx, initiator.AccountID, sessionID)
	if err != nil {
		return err
	}

	return nil
}

func (s Service) DeleteOwnSessions(ctx context.Context, initiator InitiatorData) error {
	_, _, err := s.validateSession(ctx, initiator)
	if err != nil {
		return err
	}

	err = s.repo.DeleteSessionsForAccount(ctx, initiator.AccountID)
	if err != nil {
		return err
	}

	return nil
}
