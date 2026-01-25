package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/netbill/auth-svc/internal/core/models"
	"github.com/netbill/auth-svc/internal/repository/pgdb"
)

func (r Repository) CreateOrgMember(ctx context.Context, member models.Member) error {
	_, err := r.orgMembersQ(ctx).Insert(ctx, pgdb.OrganizationMemberInsertInput{
		ID:              member.ID,
		AccountID:       member.AccountID,
		OrganizationID:  member.OrganizationID,
		SourceCreatedAt: member.CreatedAt,
	})
	if err != nil {
		return fmt.Errorf("failed to insert organization member, cause: %w", err)
	}

	return err
}

func (r Repository) DeleteOrgMember(ctx context.Context, memberID uuid.UUID) error {
	err := r.orgMembersQ(ctx).FilterByID(memberID).Delete(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete organization member with id %s, cause: %w", memberID, err)
	}

	return nil
}

func (r Repository) ExistOrgMemberByAccount(ctx context.Context, accountID uuid.UUID) (bool, error) {
	exist, err := r.orgMembersQ(ctx).FilterByID(accountID).Exists(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to check existence of organization member with account id %s, cause: %w", accountID, err)
	}

	return exist, nil
}
