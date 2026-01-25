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

const OrganizationMemberTable = "organization_members"

const OrganizationMemberColumns = "id, account_id, organization_id, source_created_at, replica_created_at"
const OrganizationMemberColumnsM = "m.id, m.account_id, m.organization_id, m.source_created_at, m.replica_created_at"

type OrganizationMember struct {
	ID             pgtype.UUID `json:"id"`
	AccountID      pgtype.UUID `json:"account_id"`
	OrganizationID pgtype.UUID `json:"organization_id"`

	SourceCreatedAt  pgtype.Timestamptz `json:"source_created_at"`
	ReplicaCreatedAt pgtype.Timestamptz `json:"replica_created_at"`
}

func (m *OrganizationMember) scan(row sq.RowScanner) error {
	err := row.Scan(
		&m.ID,
		&m.AccountID,
		&m.OrganizationID,
		&m.SourceCreatedAt,
		&m.ReplicaCreatedAt,
	)
	if err != nil {
		return fmt.Errorf("scanning organization member: %w", err)
	}
	return nil
}

type OrganizationMembersQ struct {
	db       pgxtx.DBTX
	selector sq.SelectBuilder
	inserter sq.InsertBuilder
	deleter  sq.DeleteBuilder
	counter  sq.SelectBuilder
}

func NewOrganizationMembersQ(db pgxtx.DBTX) OrganizationMembersQ {
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	return OrganizationMembersQ{
		db:       db,
		selector: builder.Select(OrganizationMemberColumnsM).From(OrganizationMemberTable + " m"),
		inserter: builder.Insert(OrganizationMemberTable),
		deleter:  builder.Delete(OrganizationMemberTable + " m"),
		counter:  builder.Select("COUNT(*)").From(OrganizationMemberTable + " m"),
	}
}

type OrganizationMemberInsertInput struct {
	ID             uuid.UUID
	AccountID      uuid.UUID
	OrganizationID uuid.UUID

	SourceCreatedAt time.Time
}

func (q OrganizationMembersQ) Insert(ctx context.Context, data OrganizationMemberInsertInput) (OrganizationMember, error) {
	query, args, err := q.inserter.SetMap(map[string]interface{}{
		"id":              pgtype.UUID{Bytes: [16]byte(data.ID), Valid: true},
		"account_id":      pgtype.UUID{Bytes: [16]byte(data.AccountID), Valid: true},
		"organization_id": pgtype.UUID{Bytes: [16]byte(data.OrganizationID), Valid: true},
		"source_created_at": pgtype.Timestamptz{
			Time:  data.SourceCreatedAt.UTC(),
			Valid: true,
		},
		"replica_created_at": pgtype.Timestamptz{
			Time:  time.Now().UTC(),
			Valid: true,
		},
	}).Suffix("RETURNING " + OrganizationMemberColumns).ToSql()
	if err != nil {
		return OrganizationMember{}, fmt.Errorf("building insert query for %s: %w", OrganizationMemberTable, err)
	}

	var inserted OrganizationMember
	if err = inserted.scan(q.db.QueryRow(ctx, query, args...)); err != nil {
		return OrganizationMember{}, err
	}
	return inserted, nil
}

func (q OrganizationMembersQ) Get(ctx context.Context) (OrganizationMember, error) {
	query, args, err := q.selector.Limit(1).ToSql()
	if err != nil {
		return OrganizationMember{}, fmt.Errorf("building select query for %s: %w", OrganizationMemberTable, err)
	}

	var m OrganizationMember
	if err = m.scan(q.db.QueryRow(ctx, query, args...)); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return OrganizationMember{}, nil
		}
		return OrganizationMember{}, err
	}
	return m, nil
}

func (q OrganizationMembersQ) Select(ctx context.Context) ([]OrganizationMember, error) {
	query, args, err := q.selector.ToSql()
	if err != nil {
		return nil, fmt.Errorf("building select query for %s: %w", OrganizationMemberTable, err)
	}

	rows, err := q.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("executing select query for %s: %w", OrganizationMemberTable, err)
	}
	defer rows.Close()

	var out []OrganizationMember
	for rows.Next() {
		var m OrganizationMember
		if err = m.scan(rows); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

func (q OrganizationMembersQ) Exists(ctx context.Context) (bool, error) {
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

func (q OrganizationMembersQ) Delete(ctx context.Context) error {
	query, args, err := q.deleter.ToSql()
	if err != nil {
		return fmt.Errorf("building delete query for %s: %w", OrganizationMemberTable, err)
	}

	_, err = q.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("executing delete query for %s: %w", OrganizationMemberTable, err)
	}

	return nil
}

func (q OrganizationMembersQ) FilterByID(id uuid.UUID) OrganizationMembersQ {
	pid := pgtype.UUID{Bytes: [16]byte(id), Valid: true}

	q.selector = q.selector.Where(sq.Eq{"m.id": pid})
	q.counter = q.counter.Where(sq.Eq{"m.id": pid})
	q.deleter = q.deleter.Where(sq.Eq{"m.id": pid})
	return q
}

func (q OrganizationMembersQ) FilterByAccountID(accountID uuid.UUID) OrganizationMembersQ {
	pid := pgtype.UUID{Bytes: [16]byte(accountID), Valid: true}

	q.selector = q.selector.Where(sq.Eq{"m.account_id": pid})
	q.counter = q.counter.Where(sq.Eq{"m.account_id": pid})
	q.deleter = q.deleter.Where(sq.Eq{"m.account_id": pid})
	return q
}

func (q OrganizationMembersQ) FilterByOrganizationID(organizationID uuid.UUID) OrganizationMembersQ {
	pid := pgtype.UUID{Bytes: [16]byte(organizationID), Valid: true}

	q.selector = q.selector.Where(sq.Eq{"m.organization_id": pid})
	q.counter = q.counter.Where(sq.Eq{"m.organization_id": pid})
	q.deleter = q.deleter.Where(sq.Eq{"m.organization_id": pid})
	return q
}

func (q OrganizationMembersQ) Page(limit, offset uint) OrganizationMembersQ {
	q.selector = q.selector.Limit(uint64(limit)).Offset(uint64(offset))
	return q
}

func (q OrganizationMembersQ) Count(ctx context.Context) (uint, error) {
	query, args, err := q.counter.ToSql()
	if err != nil {
		return 0, fmt.Errorf("building count query for %s: %w", OrganizationMemberTable, err)
	}

	var count int64
	if err = q.db.QueryRow(ctx, query, args...).Scan(&count); err != nil {
		return 0, fmt.Errorf("scanning count for %s: %w", OrganizationMemberTable, err)
	}
	if count < 0 {
		return 0, fmt.Errorf("invalid count for %s: %d", OrganizationMemberTable, count)
	}

	return uint(count), nil
}
