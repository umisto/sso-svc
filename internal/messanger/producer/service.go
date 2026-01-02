package producer

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/netbill/kafkakit/box"
	"github.com/netbill/logium"
	"github.com/segmentio/kafka-go"
)

type Producer struct {
	log    logium.Logger
	addr   []string
	outbox outbox
}

type outbox interface {
	CreateOutboxEvent(
		ctx context.Context,
		message kafka.Message,
	) (box.OutboxEvent, error)

	GetOutboxEventByID(ctx context.Context, id uuid.UUID) (box.OutboxEvent, error)
	GetPendingOutboxEvents(ctx context.Context, limit int32) ([]box.OutboxEvent, error)
	MarkOutboxEventsSent(ctx context.Context, ids []uuid.UUID) ([]box.OutboxEvent, error)
	MarkOutboxEventsAsFailed(ctx context.Context, ids []uuid.UUID) ([]box.OutboxEvent, error)
	MarkOutboxEventsAsPending(ctx context.Context, ids []uuid.UUID, delay time.Duration) ([]box.OutboxEvent, error)
}

func New(log logium.Logger, addr []string, outbox outbox) *Producer {
	return &Producer{
		log:    log,
		addr:   addr,
		outbox: outbox,
	}
}

const eventOutboxRetryDelay = 1 * time.Minute

func (p Producer) Run(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	publisher := kafka.Writer{
		Addr:         kafka.TCP(p.addr...),
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireAll,
		Compression:  kafka.Snappy,
		BatchTimeout: 50 * time.Millisecond,
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			events, err := p.outbox.GetPendingOutboxEvents(ctx, 100)
			if err != nil {
				p.log.Errorf("outbox.GetPendingOutboxEvents: %v", err)
				continue
			}

			var sentIDs []uuid.UUID
			var NonSentIDs []uuid.UUID

			for _, event := range events {
				err = publisher.WriteMessages(ctx, event.ToMessage())
				if err != nil {
					NonSentIDs = append(NonSentIDs, event.ID)
					p.log.Debugf("outbox: publish event ID %p: %v", event.ID, err)
					continue
				}
				sentIDs = append(sentIDs, event.ID)
			}

			if len(sentIDs) > 0 {
				_, err = p.outbox.MarkOutboxEventsSent(ctx, sentIDs)
				if err != nil {
					p.log.Debugf("outbox: mark events as sent: %v", err)
				}
			}

			if len(NonSentIDs) > 0 {
				_, err = p.outbox.MarkOutboxEventsAsPending(ctx, NonSentIDs, eventOutboxRetryDelay)
				if err != nil {
					p.log.Debugf("outbox: delay events: %v", err)
				}
			}
		}
	}
}
