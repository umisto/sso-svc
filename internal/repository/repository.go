package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/netbill/auth-svc/internal/repository/pgdb"
	"github.com/netbill/pgxtx"
)

type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) Repository {
	return Repository{pool: pool}
}

func (r Repository) accountsQ(ctx context.Context) pgdb.AccountsQ {
	return pgdb.NewAccountsQ(pgxtx.Exec(r.pool, ctx))
}

func (r Repository) sessionsQ(ctx context.Context) pgdb.SessionsQ {
	return pgdb.NewSessionsQ(pgxtx.Exec(r.pool, ctx))
}

func (r Repository) passwordsQ(ctx context.Context) pgdb.AccountPasswordsQ {
	return pgdb.NewAccountPasswordsQ(pgxtx.Exec(r.pool, ctx))
}

func (r Repository) emailsQ(ctx context.Context) pgdb.AccountEmailsQ {
	return pgdb.NewAccountEmailsQ(pgxtx.Exec(r.pool, ctx))
}

func (r Repository) orgMembersQ(ctx context.Context) pgdb.OrganizationMembersQ {
	return pgdb.NewOrganizationMembersQ(pgxtx.Exec(r.pool, ctx))
}

func (r Repository) Transaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return pgxtx.Transaction(r.pool, ctx, fn)
}
