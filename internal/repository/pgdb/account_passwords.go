package pgdb

import (
	"context"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/netbill/pgxtx"
)

const accountPasswordsTable = "account_passwords"

const accountPasswordsColumns = "account_id, hash, created_at, updated_at"

type AccountPassword struct {
	AccountID pgtype.UUID        `db:"account_id"`
	Hash      pgtype.Text        `db:"hash"`
	UpdatedAt pgtype.Timestamptz `db:"updated_at"`
	CreatedAt pgtype.Timestamptz `db:"created_at"`
}

func (a *AccountPassword) scan(row sq.RowScanner) error {
	err := row.Scan(
		&a.AccountID,
		&a.Hash,
		&a.CreatedAt,
		&a.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("scanning account password: %w", err)
	}
	return nil
}

type AccountPasswordsQ struct {
	db       pgxtx.DBTX
	selector sq.SelectBuilder
	inserter sq.InsertBuilder
	updater  sq.UpdateBuilder
	deleter  sq.DeleteBuilder
	counter  sq.SelectBuilder
}

func NewAccountPasswordsQ(db pgxtx.DBTX) AccountPasswordsQ {
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	return AccountPasswordsQ{
		db:       db,
		selector: builder.Select(accountPasswordsTable + ".*").From(accountPasswordsTable),
		inserter: builder.Insert(accountPasswordsTable),
		updater:  builder.Update(accountPasswordsTable),
		deleter:  builder.Delete(accountPasswordsTable),
		counter:  builder.Select("COUNT(*) AS count").From(accountPasswordsTable),
	}
}

func (q AccountPasswordsQ) Insert(ctx context.Context, input AccountPassword) (AccountPassword, error) {
	query, args, err := q.inserter.SetMap(map[string]interface{}{
		"account_id": input.AccountID,
		"hash":       input.Hash,
		"updated_at": input.UpdatedAt,
		"created_at": input.CreatedAt,
	}).Suffix("RETURNING " + accountPasswordsColumns).ToSql()
	if err != nil {
		return AccountPassword{}, fmt.Errorf("building insert query for %s: %w", accountPasswordsTable, err)
	}

	var inserted AccountPassword
	if err = inserted.scan(q.db.QueryRow(ctx, query, args...)); err != nil {
		return AccountPassword{}, err
	}

	return inserted, nil
}

func (q AccountPasswordsQ) UpdateMany(ctx context.Context) (int64, error) {
	q.updater = q.updater.Set("updated_at", pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true})

	query, args, err := q.updater.ToSql()
	if err != nil {
		return 0, fmt.Errorf("building update query for %s: %w", accountPasswordsTable, err)
	}

	tag, err := q.db.Exec(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("executing update query for %s: %w", accountPasswordsTable, err)
	}

	return tag.RowsAffected(), nil
}

func (q AccountPasswordsQ) UpdateOne(ctx context.Context) (AccountPassword, error) {
	q.updater = q.updater.Set("updated_at", pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true})

	query, args, err := q.updater.
		Suffix("RETURNING " + accountPasswordsColumns).
		ToSql()
	if err != nil {
		return AccountPassword{}, fmt.Errorf("building update query for %s: %w", accountPasswordsTable, err)
	}

	var updated AccountPassword
	if err = updated.scan(q.db.QueryRow(ctx, query, args...)); err != nil {
		return AccountPassword{}, err
	}

	return updated, nil
}

func (q AccountPasswordsQ) UpdateHash(hash string) AccountPasswordsQ {
	q.updater = q.updater.Set("hash", pgtype.Text{String: hash, Valid: true})
	return q
}

func (q AccountPasswordsQ) Get(ctx context.Context) (AccountPassword, error) {
	query, args, err := q.selector.Limit(1).ToSql()
	if err != nil {
		return AccountPassword{}, fmt.Errorf("building get query for %s: %w", accountPasswordsTable, err)
	}

	var p AccountPassword
	err = p.scan(q.db.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return AccountPassword{}, nil
		}
		return AccountPassword{}, err
	}

	return p, nil
}

func (q AccountPasswordsQ) Select(ctx context.Context) ([]AccountPassword, error) {
	query, args, err := q.selector.ToSql()
	if err != nil {
		return nil, fmt.Errorf("building select query for %s: %w", accountPasswordsTable, err)
	}

	rows, err := q.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []AccountPassword
	for rows.Next() {
		var p AccountPassword
		err = p.scan(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning account_password: %w", err)
		}
		out = append(out, p)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

func (q AccountPasswordsQ) Exists(ctx context.Context) (bool, error) {
	query, args, err := q.selector.
		Columns("1").
		Limit(1).
		ToSql()
	if err != nil {
		return false, err
	}

	var one int
	err = q.db.QueryRow(ctx, query, args...).Scan(&one)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (q AccountPasswordsQ) Delete(ctx context.Context) error {
	query, args, err := q.deleter.ToSql()
	if err != nil {
		return fmt.Errorf("building delete query for %s: %w", accountPasswordsTable, err)
	}

	_, err = q.db.Exec(ctx, query, args...)
	return err
}

func (q AccountPasswordsQ) FilterAccountID(accountID uuid.UUID) AccountPasswordsQ {
	id := pgtype.UUID{Bytes: [16]byte(accountID), Valid: true}

	q.selector = q.selector.Where(sq.Eq{"account_id": id})
	q.counter = q.counter.Where(sq.Eq{"account_id": id})
	q.deleter = q.deleter.Where(sq.Eq{"account_id": id})
	q.updater = q.updater.Where(sq.Eq{"account_id": id})
	return q
}

func (q AccountPasswordsQ) Count(ctx context.Context) (uint, error) {
	query, args, err := q.counter.ToSql()
	if err != nil {
		return 0, fmt.Errorf("building count query for %s: %w", accountPasswordsTable, err)
	}

	var count int64
	err = q.db.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, err
	}
	if count < 0 {
		return 0, fmt.Errorf("invalid count for %s: %d", accountPasswordsTable, count)
	}

	return uint(count), nil
}

func (q AccountPasswordsQ) Page(limit, offset uint) AccountPasswordsQ {
	q.selector = q.selector.Limit(uint64(limit)).Offset(uint64(offset))
	return q
}
