package token

import (
	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/core/models"
	"github.com/netbill/restkit/auth"
)

func (s Service) GenerateRefresh(account models.Account, sessionID uuid.UUID) (string, error) {
	return auth.GenerateAccountJWT(auth.GenerateAccountJwtRequest{
		Issuer:    s.iss,
		Audience:  []string{s.iss},
		AccountID: account.ID,
		SessionID: sessionID,
		Role:      account.Role,
		Ttl:       s.refreshTTL,
	}, s.refreshSK)
}

func (s Service) ParseRefreshClaims(tokenStr string) (auth.AccountClaims, error) {
	return auth.VerifyAccountJWT(tokenStr, s.refreshSK)
}

func (s Service) HashRefresh(rawRefresh string) (string, error) {
	return hmacB64("refresh."+rawRefresh, s.refreshHK)
}
