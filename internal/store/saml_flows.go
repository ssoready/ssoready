package store

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/ssoready/ssoready/internal/appauth"
	ssoreadyv1 "github.com/ssoready/ssoready/internal/gen/ssoready/v1"
	"github.com/ssoready/ssoready/internal/store/idformat"
	"github.com/ssoready/ssoready/internal/store/queries"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Store) ListSAMLFlows(ctx context.Context, req *ssoreadyv1.ListSAMLFlowsRequest) (*ssoreadyv1.ListSAMLFlowsResponse, error) {
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
	qSAMLFlows, err := q.ListSAMLFlows(ctx, queries.ListSAMLFlowsParams{
		SamlConnectionID: samlConnectionID,
		ID:               startID,
		Limit:            int32(limit + 1),
	})
	if err != nil {
		return nil, err
	}

	var flows []*ssoreadyv1.SAMLFlow
	for _, qSAMLFlow := range qSAMLFlows {
		flows = append(flows, parseSAMLFlow(qSAMLFlow))
	}

	var nextPageToken string
	if len(flows) == limit+1 {
		nextPageToken = s.pageEncoder.Marshal(flows[limit].Id)
		flows = flows[:limit]
	}

	return &ssoreadyv1.ListSAMLFlowsResponse{
		SamlFlows:     flows,
		NextPageToken: nextPageToken,
	}, nil
}

func parseSAMLFlow(qSAMLFlow queries.SamlFlow) *ssoreadyv1.SAMLFlow {
	var attrs map[string]string
	if len(qSAMLFlow.SubjectIdpAttributes) != 0 {
		if err := json.Unmarshal(qSAMLFlow.SubjectIdpAttributes, &attrs); err != nil {
			panic(err)
		}
	}

	return &ssoreadyv1.SAMLFlow{
		Id:                   idformat.SAMLFlow.Format(qSAMLFlow.ID),
		SamlConnectionId:     idformat.SAMLConnection.Format(qSAMLFlow.SamlConnectionID),
		State:                qSAMLFlow.State,
		SubjectIdpId:         derefOrEmpty(qSAMLFlow.SubjectIdpID),
		SubjectIdpAttributes: attrs,
		CreateTime:           timestamppb.New(qSAMLFlow.CreateTime),
		UpdateTime:           timestamppb.New(qSAMLFlow.UpdateTime),
		AuthRedirectUrl:      derefOrEmpty(qSAMLFlow.AuthRedirectUrl),
		GetRedirectTime:      ptrTimeToTimestamp(qSAMLFlow.GetRedirectTime),
		InitiateRequest:      derefOrEmpty(qSAMLFlow.InitiateRequest),
		InitiateTime:         ptrTimeToTimestamp(qSAMLFlow.InitiateTime),
		Assertion:            derefOrEmpty(qSAMLFlow.Assertion),
		AppRedirectUrl:       derefOrEmpty(qSAMLFlow.AppRedirectUrl),
		ReceiveAssertionTime: ptrTimeToTimestamp(qSAMLFlow.ReceiveAssertionTime),
		RedeemTime:           ptrTimeToTimestamp(qSAMLFlow.RedeemTime),
	}
}
