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

func (s *Store) ListSAMLFlowSteps(ctx context.Context, req *ssoreadyv1.ListSAMLFlowStepsRequest) (*ssoreadyv1.ListSAMLFlowStepsResponse, error) {
	_, q, _, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	samlFlowId, err := idformat.SAMLFlow.Parse(req.SamlFlowId)
	if err != nil {
		return nil, err
	}

	// idor check
	if _, err = q.GetSAMLFlow(ctx, queries.GetSAMLFlowParams{
		AppOrganizationID: appauth.OrgID(ctx),
		ID:                samlFlowId,
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
	qEntries, err := q.ListSAMLFlowSteps(ctx, queries.ListSAMLFlowStepsParams{
		SamlFlowID: samlFlowId,
		Timestamp:  start.Timestamp,
		ID:         start.ID,
		Limit:      int32(limit + 1),
	})
	if err != nil {
		return nil, err
	}

	var entries []*ssoreadyv1.SAMLFlowStep
	for _, qEntry := range qEntries {
		entries = append(entries, parseSAMLFlowStep(qEntry))
	}

	var nextPageToken string
	if len(entries) == limit+1 {
		nextPageToken = s.pageEncoder.Marshal(pageData{Timestamp: entries[limit].Timestamp.AsTime(), ID: uuid.MustParse(entries[limit].Id)})
		entries = entries[:limit]
	}

	return &ssoreadyv1.ListSAMLFlowStepsResponse{
		SamlFlowSteps: entries,
		NextPageToken: nextPageToken,
	}, nil
}

func parseSAMLFlowStep(qStep queries.SamlFlowStep) *ssoreadyv1.SAMLFlowStep {
	e := &ssoreadyv1.SAMLFlowStep{
		Id:         idformat.SAMLFlowStep.Format(qStep.ID),
		SamlFlowId: idformat.SAMLFlowStep.Format(qStep.SamlFlowID),
		Timestamp:  timestamppb.New(qStep.Timestamp),
	}

	switch qStep.Type {
	case queries.SamlFlowStepTypeGetRedirect:
		e.Details = &ssoreadyv1.SAMLFlowStep_GetRedirect{GetRedirect: *qStep.GetRedirectUrl}
	case queries.SamlFlowStepTypeSamlInitiate:
		e.Details = &ssoreadyv1.SAMLFlowStep_SamlInitiate{SamlInitiate: *qStep.SamlInitiateUrl}
	case queries.SamlFlowStepTypeSamlReceiveAssertion:
		e.Details = &ssoreadyv1.SAMLFlowStep_SamlReceiveAssertion{SamlReceiveAssertion: *qStep.SamlReceiveAssertionPayload}
	case queries.SamlFlowStepTypeRedeem:
		e.Details = &ssoreadyv1.SAMLFlowStep_Redeem{}
	}

	return e
}
