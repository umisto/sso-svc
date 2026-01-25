package pgdb

import (
	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/core/models"
)

func (a *Account) ToModel() models.Account {
	var id uuid.UUID
	if a.ID.Valid {
		id = a.ID.Bytes
	}

	return models.Account{
		ID:        id,
		Username:  a.Username.String,
		Role:      a.Role.String,
		CreatedAt: a.CreatedAt.Time,
		UpdatedAt: a.UpdatedAt.Time,
	}
}

func (a *AccountPassword) ToModel() models.AccountPassword {
	var accountID uuid.UUID
	if a.AccountID.Valid {
		accountID = a.AccountID.Bytes
	}

	return models.AccountPassword{
		AccountID: accountID,
		Hash:      a.Hash.String,
		CreatedAt: a.CreatedAt.Time,
		UpdatedAt: a.UpdatedAt.Time,
	}
}

func (e *AccountEmail) ToModel() models.AccountEmail {
	var accountID uuid.UUID
	if e.AccountID.Valid {
		accountID = e.AccountID.Bytes
	}

	return models.AccountEmail{
		AccountID: accountID,
		Email:     e.Email.String,
		Verified:  e.Verified.Bool,
		CreatedAt: e.CreatedAt.Time,
		UpdatedAt: e.UpdatedAt.Time,
	}
}

func (s *Session) ToModel() models.Session {
	var id uuid.UUID
	if s.ID.Valid {
		id = s.ID.Bytes
	}

	var accountID uuid.UUID
	if s.AccountID.Valid {
		accountID = s.AccountID.Bytes
	}

	return models.Session{
		ID:        id,
		AccountID: accountID,
		LastUsed:  s.LastUsed.Time,
		CreatedAt: s.CreatedAt.Time,
	}
}
