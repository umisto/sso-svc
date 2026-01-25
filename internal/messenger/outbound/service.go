package outbound

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/netbill/evebox/box/outbox"
	"github.com/netbill/logium"
)

type Outbound struct {
	log    *logium.Logger
	outbox outbox.Box
}

func New(log *logium.Logger, pool *pgxpool.Pool) *Outbound {
	return &Outbound{
		log:    log,
		outbox: outbox.New(pool),
	}
}
