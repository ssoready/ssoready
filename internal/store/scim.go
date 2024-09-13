package store

import (
	"connectrpc.com/connect"
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/ssoready/ssoready/internal/authn"
	ssoreadyv1 "github.com/ssoready/ssoready/internal/gen/ssoready/v1"
	"github.com/ssoready/ssoready/internal/store/idformat"
	"github.com/ssoready/ssoready/internal/store/queries"
)

func (s *Store) ListSCIMUsers(ctx context.Context, req *ssoreadyv1.ListSCIMUsersRequest) (*ssoreadyv1.ListSCIMUsersResponse, error) {
	_, q, commit, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	authnData := authn.FullContextData(ctx)
	if authnData.APIKey == nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("api key authentication is required"))
	}

	envID, err := idformat.Environment.Parse(authnData.APIKey.EnvID)
	if err != nil {
		return nil, err
	}

	var scimDirID uuid.UUID
	if req.ScimDirectoryId != "" {
		scimDirID, err = idformat.SCIMDirectory.Parse(req.ScimDirectoryId)
		if err != nil {
			return nil, err
		}

		// check that scim dir belongs to env by making sure this query finds something
		if _, err := s.q.GetSCIMDirectoryByIDAndEnvironmentID(ctx, queries.GetSCIMDirectoryByIDAndEnvironmentIDParams{
			EnvironmentID: envID,
			ID:            scimDirID,
		}); err != nil {
			return nil, err
		}
	} else if req.OrganizationId != "" {
		orgID, err := idformat.Organization.Parse(req.OrganizationId)
		if err != nil {
			return nil, err
		}

		scimDirID, err = q.GetPrimarySCIMDirectoryIDByOrganizationID(ctx, queries.GetPrimarySCIMDirectoryIDByOrganizationIDParams{
			EnvironmentID: envID,
			ID:            orgID,
		})
		if err != nil {
			return nil, err
		}
	} else if req.OrganizationExternalId != "" {
		scimDirID, err = q.GetPrimarySCIMDirectoryIDByOrganizationExternalID(ctx, queries.GetPrimarySCIMDirectoryIDByOrganizationExternalIDParams{
			EnvironmentID: envID,
			ExternalID:    &req.OrganizationExternalId,
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("bad organization_external_id: organization not found, or organization does not have a primary SCIM directory"))
			}
			return nil, err
		}
	} else {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("one of scim_directory_id, organization_id, or organization_external_id must be provided"))
	}

	var startID uuid.UUID
	if err := s.pageEncoder.Unmarshal(req.PageToken, &startID); err != nil {
		return nil, err
	}

	limit := 10
	var qSCIMUsers []queries.ScimUser
	if req.ScimGroupId != "" {
		// list by group id
		scimGroupID, err := idformat.SCIMGroup.Parse(req.ScimGroupId)
		if err != nil {
			return nil, fmt.Errorf("parse scim group id: %w", err)
		}

		qSCIMUsers, err = s.q.ListSCIMUsersInSCIMGroup(ctx, queries.ListSCIMUsersInSCIMGroupParams{
			ScimDirectoryID: scimDirID,
			ID:              startID,
			Limit:           int32(limit + 1),
			ScimGroupID:     scimGroupID,
		})
		if err != nil {
			return nil, err
		}
	} else {
		// plain list by scim dir id
		qSCIMUsers, err = s.q.ListSCIMUsers(ctx, queries.ListSCIMUsersParams{
			ScimDirectoryID: scimDirID,
			ID:              startID,
			Limit:           int32(limit + 1),
		})
		if err != nil {
			return nil, err
		}
	}

	if err := commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	var scimUsers []*ssoreadyv1.SCIMUser
	for _, qSCIMUser := range qSCIMUsers {
		scimUsers = append(scimUsers, parseSCIMUser(qSCIMUser))
	}

	var nextPageToken string
	if len(scimUsers) == limit+1 {
		nextPageToken = s.pageEncoder.Marshal(qSCIMUsers[limit].ID)
		scimUsers = scimUsers[:limit]
	}

	return &ssoreadyv1.ListSCIMUsersResponse{
		ScimUsers:     scimUsers,
		NextPageToken: nextPageToken,
	}, nil
}

func (s *Store) GetSCIMUser(ctx context.Context, req *ssoreadyv1.GetSCIMUserRequest) (*ssoreadyv1.GetSCIMUserResponse, error) {
	id, err := idformat.SCIMUser.Parse(req.Id)
	if err != nil {
		return nil, err
	}

	authnData := authn.FullContextData(ctx)
	if authnData.APIKey == nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("api key authentication is required"))
	}

	envID, err := idformat.Environment.Parse(authnData.APIKey.EnvID)
	if err != nil {
		return nil, err
	}

	_, q, _, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	qSCIMUser, err := q.GetSCIMUser(ctx, queries.GetSCIMUserParams{
		EnvironmentID: envID,
		ID:            id,
	})
	if err != nil {
		return nil, err
	}

	return &ssoreadyv1.GetSCIMUserResponse{ScimUser: parseSCIMUser(qSCIMUser)}, nil
}

func (s *Store) ListSCIMGroups(ctx context.Context, req *ssoreadyv1.ListSCIMGroupsRequest) (*ssoreadyv1.ListSCIMGroupsResponse, error) {
	_, q, commit, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	authnData := authn.FullContextData(ctx)
	if authnData.APIKey == nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("api key authentication is required"))
	}

	envID, err := idformat.Environment.Parse(authnData.APIKey.EnvID)
	if err != nil {
		return nil, err
	}

	var scimDirID uuid.UUID
	if req.ScimDirectoryId != "" {
		scimDirID, err = idformat.SCIMDirectory.Parse(req.ScimDirectoryId)
		if err != nil {
			return nil, err
		}

		// check that scim dir belongs to env by making sure this query finds something
		if _, err := s.q.GetSCIMDirectoryByIDAndEnvironmentID(ctx, queries.GetSCIMDirectoryByIDAndEnvironmentIDParams{
			EnvironmentID: envID,
			ID:            scimDirID,
		}); err != nil {
			return nil, err
		}
	} else if req.OrganizationId != "" {
		orgID, err := idformat.Organization.Parse(req.OrganizationId)
		if err != nil {
			return nil, err
		}

		scimDirID, err = q.GetPrimarySCIMDirectoryIDByOrganizationID(ctx, queries.GetPrimarySCIMDirectoryIDByOrganizationIDParams{
			EnvironmentID: envID,
			ID:            orgID,
		})
		if err != nil {
			return nil, err
		}
	} else if req.OrganizationExternalId != "" {
		scimDirID, err = q.GetPrimarySCIMDirectoryIDByOrganizationExternalID(ctx, queries.GetPrimarySCIMDirectoryIDByOrganizationExternalIDParams{
			EnvironmentID: envID,
			ExternalID:    &req.OrganizationExternalId,
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("bad organization_external_id: organization not found, or organization does not have a primary SCIM directory"))
			}
			return nil, err
		}
	} else {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("one of scim_directory_id, organization_id, or organization_external_id must be provided"))
	}

	var startID uuid.UUID
	if err := s.pageEncoder.Unmarshal(req.PageToken, &startID); err != nil {
		return nil, err
	}

	limit := 10
	qSCIMGroups, err := s.q.ListSCIMGroups(ctx, queries.ListSCIMGroupsParams{
		ScimDirectoryID: scimDirID,
		ID:              startID,
		Limit:           int32(limit + 1),
	})
	if err != nil {
		return nil, err
	}

	if err := commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	var scimGroups []*ssoreadyv1.SCIMGroup
	for _, qSCIMgroup := range qSCIMGroups {
		scimGroups = append(scimGroups, parseSCIMGroup(qSCIMgroup))
	}

	var nextPageToken string
	if len(scimGroups) == limit+1 {
		nextPageToken = s.pageEncoder.Marshal(qSCIMGroups[limit].ID)
		scimGroups = scimGroups[:limit]
	}

	return &ssoreadyv1.ListSCIMGroupsResponse{
		ScimGroups:    scimGroups,
		NextPageToken: nextPageToken,
	}, nil
}

func (s *Store) GetSCIMGroup(ctx context.Context, req *ssoreadyv1.GetSCIMGroupRequest) (*ssoreadyv1.GetSCIMGroupResponse, error) {
	id, err := idformat.SCIMGroup.Parse(req.Id)
	if err != nil {
		return nil, err
	}

	authnData := authn.FullContextData(ctx)
	if authnData.APIKey == nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("api key authentication is required"))
	}

	envID, err := idformat.Environment.Parse(authnData.APIKey.EnvID)
	if err != nil {
		return nil, err
	}

	_, q, _, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	qSCIMGroup, err := q.GetSCIMGroup(ctx, queries.GetSCIMGroupParams{
		EnvironmentID: envID,
		ID:            id,
	})
	if err != nil {
		return nil, err
	}

	return &ssoreadyv1.GetSCIMGroupResponse{ScimGroup: parseSCIMGroup(qSCIMGroup)}, nil
}
