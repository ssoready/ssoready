package store

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/ssoready/ssoready/internal/authn"
	ssoreadyv1 "github.com/ssoready/ssoready/internal/gen/ssoready/v1"
	"github.com/ssoready/ssoready/internal/store/idformat"
	"github.com/ssoready/ssoready/internal/store/queries"
)

func (s *Store) AppListSCIMUsers(ctx context.Context, req *ssoreadyv1.AppListSCIMUsersRequest) (*ssoreadyv1.AppListSCIMUsersResponse, error) {
	_, q, _, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	scimDirID, err := idformat.SCIMDirectory.Parse(req.ScimDirectoryId)
	if err != nil {
		return nil, err
	}

	// idor check
	if _, err = q.GetSCIMDirectory(ctx, queries.GetSCIMDirectoryParams{
		AppOrganizationID: authn.AppOrgID(ctx),
		ID:                scimDirID,
	}); err != nil {
		return nil, err
	}

	var startID uuid.UUID
	if err := s.pageEncoder.Unmarshal(req.PageToken, &startID); err != nil {
		return nil, err
	}

	limit := 10
	var qSCIMUsers []queries.ScimUser
	if req.ScimGroupId == "" {
		qSCIMUsers, err = q.ListSCIMUsers(ctx, queries.ListSCIMUsersParams{
			ScimDirectoryID: scimDirID,
			ID:              startID,
			Limit:           int32(limit + 1),
		})
		if err != nil {
			return nil, err
		}
	} else {
		scimGroupID, err := idformat.SCIMGroup.Parse(req.ScimGroupId)
		if err != nil {
			return nil, fmt.Errorf("parse scim group id: %", err)
		}

		qSCIMUsers, err = q.ListSCIMUsersInSCIMGroup(ctx, queries.ListSCIMUsersInSCIMGroupParams{
			ScimDirectoryID: scimDirID,
			ID:              startID,
			Limit:           int32(limit + 1),
			ScimGroupID:     scimGroupID,
		})
		if err != nil {
			return nil, err
		}
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

	return &ssoreadyv1.AppListSCIMUsersResponse{
		ScimUsers:     scimUsers,
		NextPageToken: nextPageToken,
	}, nil
}

func (s *Store) AppGetSCIMUser(ctx context.Context, req *ssoreadyv1.AppGetSCIMUserRequest) (*ssoreadyv1.SCIMUser, error) {
	scimUserID, err := idformat.SCIMUser.Parse(req.Id)
	if err != nil {
		return nil, err
	}

	qSCIMUser, err := s.q.AppGetSCIMUser(ctx, queries.AppGetSCIMUserParams{
		AppOrganizationID: authn.AppOrgID(ctx),
		ID:                scimUserID,
	})
	if err != nil {
		return nil, fmt.Errorf("get scim user: %w", err)
	}

	return parseSCIMUser(qSCIMUser), nil
}
