package store

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/ssoready/ssoready/internal/appauth"
	ssoreadyv1 "github.com/ssoready/ssoready/internal/gen/ssoready/v1"
	"github.com/ssoready/ssoready/internal/store/idformat"
	"github.com/ssoready/ssoready/internal/store/queries"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Store) ListSAMLLoginEventTimelineEntries(ctx context.Context, req *ssoreadyv1.ListSAMLLoginEventTimelineEntriesRequest) (*ssoreadyv1.ListSAMLLoginEventTimelineEntriesResponse, error) {
	_, q, _, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	samlLoginEventID, err := idformat.SAMLLoginEvent.Parse(req.SamlLoginEventId)
	if err != nil {
		return nil, err
	}

	// idor check
	if _, err = q.GetSAMLLoginEvent(ctx, queries.GetSAMLLoginEventParams{
		AppOrganizationID: appauth.OrgID(ctx),
		ID:                samlLoginEventID,
	}); err != nil {
		return nil, err
	}

	type pageData struct {
		Timestamp time.Time
		ID        uuid.UUID
	}

	var start pageData
	if err := s.pageEncoder.Unmarshal(req.PageToken, &start); err != nil {
		return nil, err
	}

	limit := 10
	qEntries, err := q.ListSAMLLoginEventTimelineEntries(ctx, queries.ListSAMLLoginEventTimelineEntriesParams{
		SamlLoginEventID: samlLoginEventID,
		Timestamp:        start.Timestamp,
		ID:               start.ID,
		Limit:            int32(limit + 1),
	})
	if err != nil {
		return nil, err
	}

	var entries []*ssoreadyv1.SAMLLoginEventTimelineEntry
	for _, qEntry := range qEntries {
		entries = append(entries, parseSAMLLoginEventTimelineEntry(qEntry))
	}

	var nextPageToken string
	if len(entries) == limit+1 {
		nextPageToken = s.pageEncoder.Marshal(pageData{Timestamp: entries[limit].Timestamp.AsTime(), ID: uuid.MustParse(entries[limit].Id)})
		entries = entries[:limit]
	}

	return &ssoreadyv1.ListSAMLLoginEventTimelineEntriesResponse{
		SamlLoginEventTimelineEntries: entries,
		NextPageToken:                 nextPageToken,
	}, nil
}

func parseSAMLLoginEventTimelineEntry(qEntry queries.SamlLoginEventTimelineEntry) *ssoreadyv1.SAMLLoginEventTimelineEntry {
	e := &ssoreadyv1.SAMLLoginEventTimelineEntry{
		Id:               idformat.SAMLLoginEventTimelineEntry.Format(qEntry.ID),
		SamlLoginEventId: idformat.SAMLLoginEvent.Format(qEntry.SamlLoginEventID),
		Timestamp:        timestamppb.New(qEntry.Timestamp),
	}

	switch qEntry.Type {
	case queries.SamlLoginEventTimelineEntryTypeGetRedirect:
		e.Details = &ssoreadyv1.SAMLLoginEventTimelineEntry_GetRedirect{GetRedirect: *qEntry.GetRedirectUrl}
	case queries.SamlLoginEventTimelineEntryTypeSamlInitiate:
		e.Details = &ssoreadyv1.SAMLLoginEventTimelineEntry_SamlInitiate{SamlInitiate: *qEntry.SamlInitiateUrl}
	case queries.SamlLoginEventTimelineEntryTypeSamlReceiveAssertion:
		e.Details = &ssoreadyv1.SAMLLoginEventTimelineEntry_SamlReceiveAssertion{SamlReceiveAssertion: *qEntry.SamlReceiveAssertionPayload}
	case queries.SamlLoginEventTimelineEntryTypeRedeem:
		e.Details = &ssoreadyv1.SAMLLoginEventTimelineEntry_Redeem{}
	}

	return e
}
