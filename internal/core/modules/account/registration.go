package account

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/core/errx"
	"github.com/netbill/auth-svc/internal/core/models"
	"github.com/netbill/restkit/tokens/roles"
	"golang.org/x/crypto/bcrypt"
)

type RegistrationParams struct {
	Email    string
	Username string
	Password string
	Role     string
}

func (s Service) Registration(
	ctx context.Context,
	params RegistrationParams,
) (models.Account, error) {
	check, err := s.repo.ExistsAccountByEmail(ctx, params.Email)
	if err != nil {
		return models.Account{}, err
	}
	if check {
		return models.Account{}, errx.ErrorEmailAlreadyExist.Raise(
			fmt.Errorf("account with email %s already exists", params.Email),
		)
	}

	check, err = s.repo.ExistsAccountByUsername(ctx, params.Username)
	if err != nil {
		return models.Account{}, err
	}
	if check {
		return models.Account{}, errx.ErrorUsernameAlreadyTaken.Raise(
			fmt.Errorf("account with username %s already exists", params.Username),
		)
	}

	err = roles.ValidateUserSystemRole(params.Role)
	if err != nil {
		return models.Account{}, err
	}

	err = s.checkPasswordRequirements(params.Password)
	if err != nil {
		return models.Account{}, err
	}

	err = s.checkUsernameRequirements(ctx, params.Username)
	if err != nil {
		return models.Account{}, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(params.Password), bcrypt.DefaultCost)
	if err != nil {
		return models.Account{}, err
	}

	var account models.Account
	err = s.repo.Transaction(ctx, func(ctx context.Context) error {
		account, err = s.repo.CreateAccount(ctx, CreateAccountParams{
			Role:         params.Role,
			Username:     params.Username,
			Email:        params.Email,
			PasswordHash: string(hash),
		})
		if err != nil {
			return err
		}

		if err = s.messenger.WriteAccountCreated(ctx, account); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return models.Account{}, err
	}

	return account, nil
}

func (s Service) RegistrationByAdmin(
	ctx context.Context,
	initiatorID uuid.UUID,
	params RegistrationParams,
) (models.Account, error) {
	initiator, err := s.repo.GetAccountByID(ctx, initiatorID)
	if err != nil {
		return models.Account{}, err
	}

	if initiator.Role != roles.SystemAdmin {
		return models.Account{}, errx.ErrorNotEnoughRights.Raise(
			fmt.Errorf("account %s has insufficient permissions to register admin accounts", initiatorID),
		)
	}

	account, err := s.Registration(ctx, params)
	if err != nil {
		return models.Account{}, err
	}

	err = s.checkUsernameRequirements(ctx, params.Username)
	if err != nil {
		return models.Account{}, err
	}

	err = s.messenger.WriteAccountCreated(ctx, account)
	if err != nil {
		return models.Account{}, err
	}

	return account, nil
}
