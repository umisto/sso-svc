package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/netbill/auth-svc/internal/core/errx"
	"github.com/netbill/auth-svc/internal/core/models"
	"github.com/netbill/auth-svc/internal/repository/pgdb"
	"github.com/netbill/restkit/pagi"
)

func (r Repository) CreateSession(ctx context.Context, sessionID, accountID uuid.UUID, hashToken string) (models.Session, error) {
	row, err := r.sessionsQ(ctx).Insert(ctx, pgdb.InsertSessionParams{
		ID:        sessionID,
		AccountID: accountID,
		HashToken: hashToken,
	})
	if err != nil {
		return models.Session{}, fmt.Errorf("failed to insert session, cause: %w", err)
	}

	return row.ToModel(), nil
}

func (r Repository) GetSession(ctx context.Context, sessionID uuid.UUID) (models.Session, error) {
	row, err := r.sessionsQ(ctx).FilterID(sessionID).Get(ctx)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return models.Session{}, errx.ErrorSessionNotFound.Raise(
			fmt.Errorf("session with id %s not found", sessionID),
		)
	case err != nil:
		return models.Session{}, err
	}

	return row.ToModel(), nil
}

func (r Repository) GetAccountSession(ctx context.Context, userID, sessionID uuid.UUID) (models.Session, error) {
	row, err := r.sessionsQ(ctx).
		FilterID(sessionID).
		FilterAccountID(userID).
		Get(ctx)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return models.Session{}, errx.ErrorSessionNotFound.Raise(
			fmt.Errorf("failed to get session with id %s for account %s, cause: %w", sessionID, userID, err),
		)
	case err != nil:
		return models.Session{}, err
	}

	return row.ToModel(), nil
}

func (r Repository) GetSessionsForAccount(ctx context.Context, userID uuid.UUID, limit, offset uint) (pagi.Page[[]models.Session], error) {
	rows, err := r.sessionsQ(ctx).
		FilterAccountID(userID).
		OrderCreatedAt(false).
		Page(limit, offset).
		Select(ctx)
	if err != nil {
		return pagi.Page[[]models.Session]{}, fmt.Errorf(
			"failed to get sessions for account %s, cause: %w", userID, err,
		)
	}

	total, err := r.sessionsQ(ctx).
		FilterAccountID(userID).
		Count(ctx)
	if err != nil {
		return pagi.Page[[]models.Session]{}, fmt.Errorf(
			"failed to count sessions for account %s, cause: %w", userID, err,
		)
	}

	collection := make([]models.Session, 0, len(rows))
	for _, s := range rows {
		collection = append(collection, s.ToModel())
	}

	return pagi.Page[[]models.Session]{
		Data:  collection,
		Page:  uint(offset/limit) + 1,
		Size:  uint(len(collection)),
		Total: total,
	}, nil
}

func (r Repository) GetSessionToken(ctx context.Context, sessionID uuid.UUID) (string, error) {
	row, err := r.sessionsQ(ctx).FilterID(sessionID).Get(ctx)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return "", errx.ErrorSessionNotFound.Raise(
			fmt.Errorf("session with id %s not found", sessionID),
		)
	case err != nil:
		return "", fmt.Errorf("failed to get session token for session %s, cause: %w", sessionID, err)
	}

	return row.HashToken.String, nil
}

func (r Repository) UpdateSessionToken(ctx context.Context, sessionID uuid.UUID, token string) (models.Session, error) {
	sess, err := r.sessionsQ(ctx).
		FilterID(sessionID).
		UpdateToken(token).
		Update(ctx)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return models.Session{}, errx.ErrorSessionNotFound.Raise(
			fmt.Errorf("session with id %s not found", sessionID),
		)
	case err != nil:
		return models.Session{}, fmt.Errorf("failed to update session token for session %s, cause: %w", sessionID, err)
	}

	if len(sess) != 1 {
		return models.Session{}, fmt.Errorf("expected 1 session, got %d", len(sess))
	}
	return sess[0].ToModel(), nil
}

func (r Repository) DeleteSession(ctx context.Context, sessionID uuid.UUID) error {
	err := r.sessionsQ(ctx).FilterID(sessionID).Delete(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete session with id %s, cause: %w", sessionID, err)
	}

	return nil
}

func (r Repository) DeleteSessionsForAccount(ctx context.Context, userID uuid.UUID) error {
	err := r.sessionsQ(ctx).FilterAccountID(userID).Delete(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete sessions for account %s, cause: %w", userID, err)
	}

	return nil
}

func (r Repository) DeleteAccountSession(ctx context.Context, userID, sessionID uuid.UUID) error {
	err := r.sessionsQ(ctx).
		FilterID(sessionID).
		FilterAccountID(userID).
		Delete(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete session with id %s, cause: %w", sessionID, err)
	}

	return nil
}
