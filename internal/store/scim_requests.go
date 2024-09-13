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

func (s *Store) AppListSCIMRequests(ctx context.Context, req *ssoreadyv1.AppListSCIMRequestsRequest) (*ssoreadyv1.AppListSCIMRequestsResponse, error) {
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

	if startID == uuid.Nil {
		// scim requests are sorted by their ID descending; initial page is max value
		startID = uuid.Max
	}

	limit := 10
	qSCIMRequests, err := q.AppListSCIMRequests(ctx, queries.AppListSCIMRequestsParams{
		ScimDirectoryID: scimDirID,
		ID:              startID,
		Limit:           int32(limit + 1),
	})
	if err != nil {
		return nil, err
	}

	var scimRequests []*ssoreadyv1.SCIMRequest
	for _, qSCIMRequest := range qSCIMRequests {
		scimRequests = append(scimRequests, parseSCIMRequest(qSCIMRequest))
	}

	var nextPageToken string
	if len(scimRequests) == limit+1 {
		nextPageToken = s.pageEncoder.Marshal(qSCIMRequests[limit].ID)
		scimRequests = scimRequests[:limit]
	}

	return &ssoreadyv1.AppListSCIMRequestsResponse{
		ScimRequests:  scimRequests,
		NextPageToken: nextPageToken,
	}, nil
}

func (s *Store) AppGetSCIMRequest(ctx context.Context, req *ssoreadyv1.AppGetSCIMRequestRequest) (*ssoreadyv1.AppGetSCIMRequestResponse, error) {
	scimRequestID, err := idformat.SCIMRequest.Parse(req.Id)
	if err != nil {
		return nil, err
	}

	qSCIMRequest, err := s.q.AppGetSCIMRequest(ctx, queries.AppGetSCIMRequestParams{
		AppOrganizationID: authn.AppOrgID(ctx),
		ID:                scimRequestID,
	})
	if err != nil {
		return nil, fmt.Errorf("get scim request: %w", err)
	}

	return &ssoreadyv1.AppGetSCIMRequestResponse{ScimRequest: parseSCIMRequest(qSCIMRequest)}, nil
}
