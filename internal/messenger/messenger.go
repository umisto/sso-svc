package messenger

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/netbill/logium"
)

type Messenger struct {
	addr []string
	pool *pgxpool.Pool
	log  *logium.Logger
}

func New(
	log *logium.Logger,
	pool *pgxpool.Pool,
	addr ...string,
) Messenger {
	return Messenger{
		addr: addr,
		pool: pool,
		log:  log,
	}
}
