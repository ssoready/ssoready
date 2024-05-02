package store

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/ssoready/ssoready/internal/appauth"
	ssoreadyv1 "github.com/ssoready/ssoready/internal/gen/ssoready/v1"
	"github.com/ssoready/ssoready/internal/store/idformat"
	"github.com/ssoready/ssoready/internal/store/queries"
)

func (s *Store) ListSAMLLoginEvents(ctx context.Context, req *ssoreadyv1.ListSAMLLoginEventsRequest) (*ssoreadyv1.ListSAMLLoginEventsResponse, error) {
	_, q, _, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	samlConnectionID, err := idformat.SAMLConnection.Parse(req.SamlConnectionId)
	if err != nil {
		return nil, err
	}

	// idor check
	if _, err = q.GetSAMLConnection(ctx, queries.GetSAMLConnectionParams{
		AppOrganizationID: appauth.OrgID(ctx),
		ID:                samlConnectionID,
	}); err != nil {
		return nil, err
	}

	var startID uuid.UUID
	if err := s.pageEncoder.Unmarshal(req.PageToken, &startID); err != nil {
		return nil, err
	}

	limit := 10
	qSAMLLoginEvents, err := q.ListSAMLLoginEvents(ctx, queries.ListSAMLLoginEventsParams{
		SamlConnectionID: samlConnectionID,
		ID:               startID,
		Limit:            int32(limit + 1),
	})
	if err != nil {
		return nil, err
	}

	var samlLoginEvents []*ssoreadyv1.SAMLLoginEvent
	for _, qSAMLLoginEvent := range qSAMLLoginEvents {
		samlLoginEvents = append(samlLoginEvents, parseSAMLLoginEvent(qSAMLLoginEvent))
	}

	var nextPageToken string
	if len(samlLoginEvents) == limit+1 {
		nextPageToken = s.pageEncoder.Marshal(samlLoginEvents[limit].Id)
		samlLoginEvents = samlLoginEvents[:limit]
	}

	return &ssoreadyv1.ListSAMLLoginEventsResponse{
		SamlLoginEvents: samlLoginEvents,
		NextPageToken:   nextPageToken,
	}, nil
}

func parseSAMLLoginEvent(qSAMLLoginEvent queries.SamlLoginEvent) *ssoreadyv1.SAMLLoginEvent {
	var attrs map[string]string
	if len(qSAMLLoginEvent.SubjectIdpAttributes) != 0 {
		if err := json.Unmarshal(qSAMLLoginEvent.SubjectIdpAttributes, &attrs); err != nil {
			panic(err)
		}
	}

	return &ssoreadyv1.SAMLLoginEvent{
		Id:                   idformat.SAMLLoginEvent.Format(qSAMLLoginEvent.ID),
		SamlConnectionId:     idformat.SAMLConnection.Format(qSAMLLoginEvent.SamlConnectionID),
		State:                qSAMLLoginEvent.State,
		SubjectIdpId:         derefOrEmpty(qSAMLLoginEvent.SubjectIdpID),
		SubjectIdpAttributes: attrs,
	}
}
