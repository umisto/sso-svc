package producer

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/core/models"
	"github.com/netbill/auth-svc/internal/messanger/contracts"
	"github.com/netbill/kafkakit/header"
	"github.com/segmentio/kafka-go"
)

func (p Producer) WriteAccountCreated(
	ctx context.Context,
	account models.Account,
	email string,
) error {
	payload, err := json.Marshal(contracts.AccountCreatedPayload{
		Account: account,
		Email:   email,
	})
	if err != nil {
		return err
	}

	eve, err := p.outbox.CreateOutboxEvent(
		ctx,
		kafka.Message{
			Topic: contracts.AccountsTopicV1,
			Key:   []byte(account.ID.String()),
			Value: payload,
			Headers: []kafka.Header{
				{Key: header.EventID, Value: []byte(uuid.New().String())},
				{Key: header.EventType, Value: []byte(contracts.AccountCreatedEvent)},
				{Key: header.EventVersion, Value: []byte("1")},
				{Key: header.Producer, Value: []byte(contracts.AuthSvcGroup)},
				{Key: header.ContentType, Value: []byte("application/json")},
			},
		},
	)
	if err != nil {
		return err
	}

	p.log.Infof("Produced AccountCreated event for account ID '%s'", eve.ID)

	return err
}
