package controller

import (
	"net/http"

	"github.com/netbill/ape"
	"github.com/netbill/ape/problems"
	"github.com/netbill/auth-svc/internal/core/modules/auth"
	"github.com/netbill/auth-svc/internal/rest"
)

func (s *Service) Logout(w http.ResponseWriter, r *http.Request) {
	initiator, err := rest.AccountData(r.Context())
	if err != nil {
		s.log.WithError(err).Error("failed to get user from context")
		ape.RenderErr(w, problems.Unauthorized("failed to get user from context"))

		return
	}

	err = s.core.Logout(r.Context(), auth.InitiatorData{
		AccountID: initiator.ID,
		SessionID: initiator.SessionID,
	})
	if err != nil {
		s.log.WithError(err).Errorf("failed to logout user")
		switch {
		default:
			ape.RenderErr(w, problems.InternalError())
		}

		return
	}

	w.WriteHeader(http.StatusNoContent)
}
