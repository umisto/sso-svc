package account

import (
	"context"

	"github.com/netbill/auth-svc/internal/core/models"
)

func (s Service) UpdateUsername(ctx context.Context, initiator InitiatorData, newUsername string) (account models.Account, err error) {
	account, _, err = s.validateSession(ctx, initiator)
	if err != nil {
		return models.Account{}, err
	}

	if err = s.checkUsernameRequirements(ctx, newUsername); err != nil {
		return models.Account{}, err
	}

	err = s.repo.Transaction(ctx, func(txCtx context.Context) error {
		account, err = s.repo.UpdateAccountUsername(ctx, initiator.AccountID, newUsername)
		if err != nil {
			return err
		}

		if err = s.messenger.WriteAccountUsernameUpdated(ctx, account); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return models.Account{}, err
	}

	return account, nil
}
