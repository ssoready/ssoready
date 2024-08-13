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

func (s *Store) AppListSCIMGroups(ctx context.Context, req *ssoreadyv1.AppListSCIMGroupsRequest) (*ssoreadyv1.AppListSCIMGroupsResponse, error) {
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
	var qSCIMGroups []queries.ScimGroup
	if req.ScimUserId == "" {
		qSCIMGroups, err = q.ListSCIMGroups(ctx, queries.ListSCIMGroupsParams{
			ScimDirectoryID: scimDirID,
			ID:              startID,
			Limit:           int32(limit + 1),
		})
		if err != nil {
			return nil, err
		}
	} else {
		scimUserID, err := idformat.SCIMUser.Parse(req.ScimUserId)
		if err != nil {
			return nil, fmt.Errorf("parse scim user id: %w", err)
		}

		qSCIMGroups, err = q.ListSCIMGroupsBySCIMUserID(ctx, queries.ListSCIMGroupsBySCIMUserIDParams{
			ScimDirectoryID: scimDirID,
			ID:              startID,
			Limit:           int32(limit + 1),
			ScimUserID:      scimUserID,
		})
		if err != nil {
			return nil, err
		}
	}

	var scimGroups []*ssoreadyv1.SCIMGroup
	for _, qSCIMGroup := range qSCIMGroups {
		scimGroups = append(scimGroups, parseSCIMGroup(qSCIMGroup))
	}

	var nextPageToken string
	if len(scimGroups) == limit+1 {
		nextPageToken = s.pageEncoder.Marshal(qSCIMGroups[limit].ID)
		scimGroups = scimGroups[:limit]
	}

	return &ssoreadyv1.AppListSCIMGroupsResponse{
		ScimGroups:    scimGroups,
		NextPageToken: nextPageToken,
	}, nil
}

func (s *Store) AppGetSCIMGroup(ctx context.Context, req *ssoreadyv1.AppGetSCIMGroupRequest) (*ssoreadyv1.SCIMGroup, error) {
	scimGroupID, err := idformat.SCIMGroup.Parse(req.Id)
	if err != nil {
		return nil, err
	}

	qSCIMGroup, err := s.q.AppGetSCIMGroup(ctx, queries.AppGetSCIMGroupParams{
		AppOrganizationID: authn.AppOrgID(ctx),
		ID:                scimGroupID,
	})
	if err != nil {
		return nil, fmt.Errorf("get scim group: %w", err)
	}

	return parseSCIMGroup(qSCIMGroup), nil
}
