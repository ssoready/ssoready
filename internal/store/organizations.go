package store

import (
	"context"
	"sort"

	"github.com/google/uuid"
	"github.com/ssoready/ssoready/internal/appauth"
	ssoreadyv1 "github.com/ssoready/ssoready/internal/gen/ssoready/v1"
	"github.com/ssoready/ssoready/internal/store/idformat"
	"github.com/ssoready/ssoready/internal/store/queries"
)

func (s *Store) ListOrganizations(ctx context.Context, req *ssoreadyv1.ListOrganizationsRequest) (*ssoreadyv1.ListOrganizationsResponse, error) {
	_, q, _, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	envID, err := idformat.Environment.Parse(req.EnvironmentId)
	if err != nil {
		return nil, err
	}

	// idor check
	if _, err = q.GetEnvironment(ctx, queries.GetEnvironmentParams{
		AppOrganizationID: appauth.OrgID(ctx),
		ID:                envID,
	}); err != nil {
		return nil, err
	}

	var startID uuid.UUID
	if err := s.pageEncoder.Unmarshal(req.PageToken, &startID); err != nil {
		return nil, err
	}

	limit := 10
	qOrgs, err := q.ListOrganizations(ctx, queries.ListOrganizationsParams{
		EnvironmentID: envID,
		ID:            startID,
		Limit:         int32(limit + 1),
	})
	if err != nil {
		return nil, err
	}

	var qOrgIDs []uuid.UUID
	for _, qOrg := range qOrgs {
		qOrgIDs = append(qOrgIDs, qOrg.ID)
	}

	qOrgDomains, err := q.ListOrganizationDomains(ctx, qOrgIDs)
	if err != nil {
		return nil, err
	}

	var orgs []*ssoreadyv1.Organization
	for _, qOrg := range qOrgs {
		var domains []string
		for _, qOrgDomain := range qOrgDomains {
			if qOrgDomain.OrganizationID == qOrg.ID {
				domains = append(domains, qOrgDomain.Domain)
			}
		}
		sort.Strings(domains)

		orgs = append(orgs, &ssoreadyv1.Organization{
			Id:            idformat.Organization.Format(qOrg.ID),
			EnvironmentId: idformat.Environment.Format(qOrg.EnvironmentID),
			ExternalId:    derefOrEmpty(qOrg.ExternalID),
			Domains:       domains,
		})
	}

	var nextPageToken string
	if len(orgs) == limit+1 {
		nextPageToken = s.pageEncoder.Marshal(orgs[limit].Id)
		orgs = orgs[:limit]
	}

	return &ssoreadyv1.ListOrganizationsResponse{
		Organizations: orgs,
		NextPageToken: nextPageToken,
	}, nil
}

func (s *Store) GetOrganization(ctx context.Context, req *ssoreadyv1.GetOrganizationRequest) (*ssoreadyv1.Organization, error) {
	id, err := idformat.Organization.Parse(req.Id)
	if err != nil {
		return nil, err
	}

	_, q, _, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	qOrg, err := q.GetOrganization(ctx, queries.GetOrganizationParams{
		AppOrganizationID: appauth.OrgID(ctx),
		ID:                id,
	})
	if err != nil {
		return nil, err
	}

	qOrgDomains, err := q.ListOrganizationDomains(ctx, []uuid.UUID{qOrg.ID})
	if err != nil {
		return nil, err
	}

	var domains []string
	for _, qOrgDomain := range qOrgDomains {
		domains = append(domains, qOrgDomain.Domain)
	}
	sort.Strings(domains)

	return &ssoreadyv1.Organization{
		Id:            idformat.Organization.Format(qOrg.ID),
		EnvironmentId: idformat.Environment.Format(qOrg.EnvironmentID),
		ExternalId:    derefOrEmpty(qOrg.ExternalID),
		Domains:       domains,
	}, nil
}
