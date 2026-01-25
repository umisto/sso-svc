package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/netbill/auth-svc/internal/core/errx"
	"github.com/netbill/auth-svc/internal/core/models"
	"github.com/netbill/auth-svc/internal/core/modules/account"
	"github.com/netbill/auth-svc/internal/repository/pgdb"
)

func (r Repository) CreateAccount(ctx context.Context, params account.CreateAccountParams) (models.Account, error) {
	now := time.Now().UTC()
	accountID := uuid.New()

	acc, err := r.accountsQ(ctx).Insert(ctx, pgdb.InsertAccountParams{
		ID:       accountID,
		Username: params.Username,
		Role:     params.Role,
	})
	if err != nil {
		return models.Account{}, fmt.Errorf("failed to insert account, cause: %w", err)
	}

	emailRow := pgdb.AccountEmail{
		AccountID: pgtype.UUID{Bytes: [16]byte(accountID), Valid: true},
		Email:     pgtype.Text{String: params.Email, Valid: true},
		Verified:  pgtype.Bool{Bool: false, Valid: true},
		CreatedAt: pgtype.Timestamptz{Time: now, Valid: true},
		UpdatedAt: pgtype.Timestamptz{Time: now, Valid: true},
	}

	if _, err = r.emailsQ(ctx).Insert(ctx, emailRow); err != nil {
		return models.Account{}, fmt.Errorf("failed to insert account email, cause: %w", err)
	}

	passwordRow := pgdb.AccountPassword{
		AccountID: pgtype.UUID{Bytes: [16]byte(accountID), Valid: true},
		Hash:      pgtype.Text{String: params.PasswordHash, Valid: true},
		CreatedAt: pgtype.Timestamptz{Time: now, Valid: true},
		UpdatedAt: pgtype.Timestamptz{Time: now, Valid: true},
	}

	if _, err = r.passwordsQ(ctx).Insert(ctx, passwordRow); err != nil {
		return models.Account{}, fmt.Errorf("failed to insert account password, cause: %w", err)
	}

	return acc.ToModel(), nil
}

func (r Repository) GetAccountByID(ctx context.Context, accountID uuid.UUID) (models.Account, error) {
	acc, err := r.accountsQ(ctx).FilterID(accountID).Get(ctx)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return models.Account{}, errx.ErrorAccountNotFound.Raise(
			fmt.Errorf("account with id %s not found", accountID),
		)
	case err != nil:
		return models.Account{}, fmt.Errorf("failed to get account, cause: %w", err)
	}

	return acc.ToModel(), nil
}

func (r Repository) ExistsAccountByID(ctx context.Context, accountID uuid.UUID) (bool, error) {
	exist, err := r.accountsQ(ctx).FilterID(accountID).Exists(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to check account existence by id %s, cause: %w", accountID, err)
	}

	return exist, nil
}

func (r Repository) GetAccountByEmail(ctx context.Context, email string) (models.Account, error) {
	acc, err := r.accountsQ(ctx).FilterEmail(email).Get(ctx)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return models.Account{}, errx.ErrorAccountNotFound.Raise(
			fmt.Errorf("account with email %s not found", email),
		)
	case err != nil:
		return models.Account{}, fmt.Errorf("failed to get account by email, cause: %w", err)
	}

	return acc.ToModel(), nil
}

func (r Repository) ExistsAccountByEmail(ctx context.Context, email string) (bool, error) {
	exist, err := r.accountsQ(ctx).FilterEmail(email).Exists(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to check account existence by email %s, cause: %w", email, err)
	}

	return exist, nil
}

func (r Repository) GetAccountByUsername(ctx context.Context, username string) (models.Account, error) {
	acc, err := r.accountsQ(ctx).FilterUsername(username).Get(ctx)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return models.Account{}, errx.ErrorAccountNotFound.Raise(
			fmt.Errorf("account with username %s not found", username),
		)
	case err != nil:
		return models.Account{}, fmt.Errorf("failed to get account by username, cause: %w", err)
	}

	return acc.ToModel(), nil
}

func (r Repository) ExistsAccountByUsername(ctx context.Context, username string) (bool, error) {
	exist, err := r.accountsQ(ctx).FilterUsername(username).Exists(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to check account existence by username %s, cause: %w", username, err)
	}

	return exist, nil
}

func (r Repository) GetAccountEmail(ctx context.Context, accountID uuid.UUID) (models.AccountEmail, error) {
	acc, err := r.emailsQ(ctx).FilterAccountID(accountID).Get(ctx)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return models.AccountEmail{}, errx.ErrorAccountEmailNotFound.Raise(err)
	case err != nil:
		return models.AccountEmail{}, fmt.Errorf(
			"failed to get account email for account %s, cause: %w", accountID, err,
		)
	}

	return acc.ToModel(), nil
}

func (r Repository) GetAccountPassword(ctx context.Context, accountID uuid.UUID) (models.AccountPassword, error) {
	acc, err := r.passwordsQ(ctx).FilterAccountID(accountID).Get(ctx)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return models.AccountPassword{}, errx.ErrorAccountPasswordNorFound.Raise(
			fmt.Errorf("account password for account %s not found", accountID),
		)
	case err != nil:
		return models.AccountPassword{}, fmt.Errorf(
			"failed to get account password for account %s, cause: %w", accountID, err,
		)
	}

	return acc.ToModel(), nil
}

func (r Repository) UpdateAccountPassword(
	ctx context.Context,
	accountID uuid.UUID,
	passwordHash string,
) (models.AccountPassword, error) {
	acc, err := r.passwordsQ(ctx).
		FilterAccountID(accountID).
		UpdateHash(passwordHash).
		UpdateOne(ctx)
	if err != nil {
		return models.AccountPassword{}, fmt.Errorf(
			"failed to update account password for account %s, cause: %w", accountID, err,
		)
	}

	return acc.ToModel(), nil
}

func (r Repository) UpdateAccountUsername(
	ctx context.Context,
	accountID uuid.UUID,
	username string,
) (models.Account, error) {
	acc, err := r.accountsQ(ctx).
		FilterID(accountID).
		UpdateUsername(username).
		UpdateOne(ctx)
	if err != nil {
		return models.Account{}, fmt.Errorf(
			"failed to update account username for account %s, cause: %w", accountID, err,
		)
	}

	return acc.ToModel(), nil
}

func (r Repository) DeleteAccount(ctx context.Context, accountID uuid.UUID) error {
	err := r.accountsQ(ctx).FilterID(accountID).Delete(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete account %s, cause: %w", accountID, err)
	}

	return nil
}
