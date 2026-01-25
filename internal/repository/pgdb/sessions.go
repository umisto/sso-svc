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

const sessionsTable = "sessions"

type Session struct {
	ID        pgtype.UUID        `db:"id"`
	AccountID pgtype.UUID        `db:"account_id"`
	HashToken pgtype.Text        `db:"hash_token"`
	LastUsed  pgtype.Timestamptz `db:"last_used"`
	CreatedAt pgtype.Timestamptz `db:"created_at"`
}

func (s *Session) scan(row sq.RowScanner) error {
	err := row.Scan(
		&s.ID,
		&s.AccountID,
		&s.HashToken,
		&s.LastUsed,
		&s.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("scanning session: %w", err)
	}
	return nil
}

type SessionsQ struct {
	db       pgxtx.DBTX
	selector sq.SelectBuilder
	inserter sq.InsertBuilder
	updater  sq.UpdateBuilder
	deleter  sq.DeleteBuilder
	counter  sq.SelectBuilder
}

func NewSessionsQ(db pgxtx.DBTX) SessionsQ {
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	return SessionsQ{
		db:       db,
		selector: builder.Select(sessionsTable + ".*").From(sessionsTable),
		inserter: builder.Insert(sessionsTable),
		updater:  builder.Update(sessionsTable),
		deleter:  builder.Delete(sessionsTable),
		counter:  builder.Select("COUNT(*) AS count").From(sessionsTable),
	}
}

type InsertSessionParams struct {
	ID        uuid.UUID
	AccountID uuid.UUID
	HashToken string
}

func (q SessionsQ) Insert(ctx context.Context, input InsertSessionParams) (Session, error) {
	query, args, err := q.inserter.SetMap(map[string]interface{}{
		"id":         pgtype.UUID{Bytes: [16]byte(input.ID), Valid: true},
		"account_id": pgtype.UUID{Bytes: [16]byte(input.AccountID), Valid: true},
		"hash_token": pgtype.Text{String: input.HashToken, Valid: true},
	}).Suffix("RETURNING id, account_id, hash_token, last_used, created_at").ToSql()
	if err != nil {
		return Session{}, fmt.Errorf("building insert query for %s: %w", sessionsTable, err)
	}

	var sess Session
	err = sess.scan(q.db.QueryRow(ctx, query, args...))
	if err != nil {
		return Session{}, err
	}

	return sess, nil
}

func (q SessionsQ) Update(ctx context.Context) ([]Session, error) {
	q.updater = q.updater.
		Set("last_used", pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true}).
		Suffix("RETURNING " + sessionsTable + ".*")

	query, args, err := q.updater.ToSql()
	if err != nil {
		return nil, fmt.Errorf("building update query for %s: %w", sessionsTable, err)
	}

	rows, err := q.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Session
	for rows.Next() {
		var s Session
		err = s.scan(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning updated session: %w", err)
		}
		out = append(out, s)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

func (q SessionsQ) UpdateToken(token string) SessionsQ {
	q.updater = q.updater.Set("hash_token", pgtype.Text{String: token, Valid: true})
	return q
}

func (q SessionsQ) UpdateLastUsed(lastUsed time.Time) SessionsQ {
	q.updater = q.updater.Set("last_used", pgtype.Timestamptz{Time: lastUsed.UTC(), Valid: true})
	return q
}

func (q SessionsQ) Get(ctx context.Context) (Session, error) {
	query, args, err := q.selector.Limit(1).ToSql()
	if err != nil {
		return Session{}, fmt.Errorf("building get query for sessions: %w", err)
	}

	var sess Session
	err = sess.scan(q.db.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Session{}, nil
		}
		return Session{}, err
	}

	return sess, nil
}

func (q SessionsQ) Select(ctx context.Context) ([]Session, error) {
	query, args, err := q.selector.ToSql()
	if err != nil {
		return nil, fmt.Errorf("building select query for sessions: %w", err)
	}

	rows, err := q.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []Session
	for rows.Next() {
		var sess Session
		err = sess.scan(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning session row: %w", err)
		}
		sessions = append(sessions, sess)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return sessions, nil
}

func (q SessionsQ) Exists(ctx context.Context) (bool, error) {
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

func (q SessionsQ) Delete(ctx context.Context) error {
	query, args, err := q.deleter.ToSql()
	if err != nil {
		return fmt.Errorf("building delete query for sessions: %w", err)
	}

	_, err = q.db.Exec(ctx, query, args...)
	return err
}

func (q SessionsQ) FilterID(ID uuid.UUID) SessionsQ {
	pid := pgtype.UUID{Bytes: [16]byte(ID), Valid: true}

	q.selector = q.selector.Where(sq.Eq{"id": pid})
	q.deleter = q.deleter.Where(sq.Eq{"id": pid})
	q.updater = q.updater.Where(sq.Eq{"id": pid})
	q.counter = q.counter.Where(sq.Eq{"id": pid})

	return q
}

func (q SessionsQ) FilterAccountID(accountID uuid.UUID) SessionsQ {
	pid := pgtype.UUID{Bytes: [16]byte(accountID), Valid: true}

	q.selector = q.selector.Where(sq.Eq{"account_id": pid})
	q.deleter = q.deleter.Where(sq.Eq{"account_id": pid})
	q.updater = q.updater.Where(sq.Eq{"account_id": pid})
	q.counter = q.counter.Where(sq.Eq{"account_id": pid})

	return q
}

func (q SessionsQ) OrderCreatedAt(ascending bool) SessionsQ {
	if ascending {
		q.selector = q.selector.OrderBy("created_at ASC")
	} else {
		q.selector = q.selector.OrderBy("created_at DESC")
	}
	return q
}

func (q SessionsQ) Count(ctx context.Context) (uint, error) {
	query, args, err := q.counter.ToSql()
	if err != nil {
		return 0, fmt.Errorf("building count query for sessions: %w", err)
	}

	var count int64
	err = q.db.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, err
	}
	if count < 0 {
		return 0, fmt.Errorf("invalid count for sessions: %d", count)
	}

	return uint(count), nil
}

func (q SessionsQ) Page(limit, offset uint) SessionsQ {
	q.selector = q.selector.Limit(uint64(limit)).Offset(uint64(offset))
	return q
}
