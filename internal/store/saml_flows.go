package store

import (
	"context"
	"encoding/json"
	"time"

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

	type pageData struct {
		CreateTime time.Time
		ID         uuid.UUID
	}

	var startPageData pageData
	if err := s.pageEncoder.Unmarshal(req.PageToken, &startPageData); err != nil {
		return nil, err
	}

	limit := 10
	var qSAMLFlows []queries.SamlFlow
	if req.PageToken == "" {
		qSAMLFlows, err = q.ListSAMLFlowsFirstPage(ctx, queries.ListSAMLFlowsFirstPageParams{
			SamlConnectionID: samlConnectionID,
			Limit:            int32(limit + 1),
		})
		if err != nil {
			return nil, err
		}
	} else {
		qSAMLFlows, err = q.ListSAMLFlowsNextPage(ctx, queries.ListSAMLFlowsNextPageParams{
			SamlConnectionID: samlConnectionID,
			Limit:            int32(limit + 1),
			CreateTime:       startPageData.CreateTime,
			ID:               startPageData.ID,
		})
		if err != nil {
			return nil, err
		}
	}

	var flows []*ssoreadyv1.SAMLFlow
	for _, qSAMLFlow := range qSAMLFlows {
		flows = append(flows, parseSAMLFlow(qSAMLFlow))
	}

	var nextPageToken string
	if len(flows) == limit+1 {
		nextPageToken = s.pageEncoder.Marshal(pageData{
			CreateTime: qSAMLFlows[limit].CreateTime,
			ID:         qSAMLFlows[limit].ID,
		})
		flows = flows[:limit]
	}

	return &ssoreadyv1.ListSAMLFlowsResponse{
		SamlFlows:     flows,
		NextPageToken: nextPageToken,
	}, nil
}

func (s *Store) GetSAMLFlow(ctx context.Context, req *ssoreadyv1.GetSAMLFlowRequest) (*ssoreadyv1.SAMLFlow, error) {
	id, err := idformat.SAMLFlow.Parse(req.Id)
	if err != nil {
		return nil, err
	}

	_, q, _, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	qSAMLFlow, err := q.GetSAMLFlow(ctx, queries.GetSAMLFlowParams{
		AppOrganizationID: appauth.OrgID(ctx),
		ID:                id,
	})
	if err != nil {
		return nil, err
	}

	return parseSAMLFlow(qSAMLFlow), nil
}

func parseSAMLFlow(qSAMLFlow queries.SamlFlow) *ssoreadyv1.SAMLFlow {
	var attrs map[string]string
	if len(qSAMLFlow.SubjectIdpAttributes) != 0 {
		if err := json.Unmarshal(qSAMLFlow.SubjectIdpAttributes, &attrs); err != nil {
			panic(err)
		}
	}

	var status ssoreadyv1.SAMLFlowStatus
	if qSAMLFlow.Status.Valid {
		switch qSAMLFlow.Status.SamlFlowStatus {
		case queries.SamlFlowStatusInProgress:
			status = ssoreadyv1.SAMLFlowStatus_SAML_FLOW_STATUS_IN_PROGRESS
		case queries.SamlFlowStatusFailed:
			status = ssoreadyv1.SAMLFlowStatus_SAML_FLOW_STATUS_FAILED
		case queries.SamlFlowStatusSucceeded:
			status = ssoreadyv1.SAMLFlowStatus_SAML_FLOW_STATUS_SUCCEEDED
		}
	}

	res := ssoreadyv1.SAMLFlow{
		Id:                   idformat.SAMLFlow.Format(qSAMLFlow.ID),
		SamlConnectionId:     idformat.SAMLConnection.Format(qSAMLFlow.SamlConnectionID),
		Status:               status,
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
		RedeemResponse:       string(qSAMLFlow.RedeemResponse),
	}

	if qSAMLFlow.ErrorBadIssuer != nil {
		res.Error = &ssoreadyv1.SAMLFlow_BadIssuer{BadIssuer: *qSAMLFlow.ErrorBadIssuer}
	}
	if qSAMLFlow.ErrorBadAudience != nil {
		res.Error = &ssoreadyv1.SAMLFlow_BadAudience{BadAudience: *qSAMLFlow.ErrorBadAudience}
	}
	if qSAMLFlow.ErrorBadSubjectID != nil {
		res.Error = &ssoreadyv1.SAMLFlow_BadSubjectId{BadSubjectId: *qSAMLFlow.ErrorBadSubjectID}
	}
	if qSAMLFlow.ErrorEmailOutsideOrganizationDomains != nil {
		res.Error = &ssoreadyv1.SAMLFlow_EmailOutsideOrganizationDomains{EmailOutsideOrganizationDomains: *qSAMLFlow.ErrorEmailOutsideOrganizationDomains}
	}

	return &res
}
