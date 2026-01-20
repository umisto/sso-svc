package contracts

import (
	"time"

	"github.com/google/uuid"
)

const AccountsTopicV1 = "accounts.v1"

const AccountCreatedEvent = "account.created"

type AccountCreatedPayload struct {
	Data      AccountCreatedPayloadData `json:"data"`
	Timestamp time.Time                 `json:"timestamp"`
}

type AccountCreatedPayloadData struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email,omitempty"`
	Role  string    `json:"role"`
}

const AccountDeletedEvent = "account.deleted"

type AccountDeletedPayload struct {
	Data      AccountDeletedPayloadData `json:"data"`
	Timestamp time.Time                 `json:"timestamp"`
}

type AccountDeletedPayloadData struct {
	AccountID uuid.UUID `json:"account_id"`
}
