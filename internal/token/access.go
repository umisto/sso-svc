package token

import (
	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/core/models"
	"github.com/netbill/restkit/auth"
)

func (s Service) GenerateAccess(account models.Account, sessionID uuid.UUID) (string, error) {
	return auth.GenerateAccountJWT(auth.GenerateAccountJwtRequest{
		Issuer:    s.iss,
		Audience:  []string{s.iss},
		AccountID: account.ID,
		SessionID: sessionID,
		Role:      account.Role,
		Ttl:       s.accessTTL,
	}, s.accessSK)
}

func (s Service) ParseAccessClaims(tokenStr string) (auth.AccountClaims, error) {
	return auth.VerifyAccountJWT(tokenStr, s.accessSK)
}
