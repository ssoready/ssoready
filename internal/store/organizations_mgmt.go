package store

import (
	"context"

	"github.com/google/uuid"
	"github.com/ssoready/ssoready/internal/authn"
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

	envID, err := idformat.Environment.Parse(authn.FullContextData(ctx).APIKey.EnvID)
	if err != nil {
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
		orgs = append(orgs, parseOrganization(qOrg, qOrgDomains))
	}

	var nextPageToken string
	if len(orgs) == limit+1 {
		nextPageToken = s.pageEncoder.Marshal(qOrgs[limit].ID)
		orgs = orgs[:limit]
	}

	return &ssoreadyv1.ListOrganizationsResponse{
		Organizations: orgs,
		NextPageToken: nextPageToken,
	}, nil
}

func (s *Store) GetOrganization(ctx context.Context, req *ssoreadyv1.GetOrganizationRequest) (*ssoreadyv1.GetOrganizationResponse, error) {
	id, err := idformat.Organization.Parse(req.Id)
	if err != nil {
		return nil, err
	}

	envID, err := idformat.Environment.Parse(authn.FullContextData(ctx).APIKey.EnvID)
	if err != nil {
		return nil, err
	}

	_, q, _, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	qOrg, err := q.ManagementGetOrganization(ctx, queries.ManagementGetOrganizationParams{
		EnvironmentID: envID,
		ID:            id,
	})
	if err != nil {
		return nil, err
	}

	qOrgDomains, err := q.ListOrganizationDomains(ctx, []uuid.UUID{qOrg.ID})
	if err != nil {
		return nil, err
	}

	return &ssoreadyv1.GetOrganizationResponse{Organization: parseOrganization(qOrg, qOrgDomains)}, nil
}

func (s *Store) CreateOrganization(ctx context.Context, req *ssoreadyv1.CreateOrganizationRequest) (*ssoreadyv1.CreateOrganizationResponse, error) {
	_, q, commit, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	envID, err := idformat.Environment.Parse(authn.FullContextData(ctx).APIKey.EnvID)
	if err != nil {
		return nil, err
	}

	var externalID *string
	if req.Organization.ExternalId != "" {
		externalID = &req.Organization.ExternalId
	}

	qOrg, err := q.CreateOrganization(ctx, queries.CreateOrganizationParams{
		ID:            uuid.New(),
		EnvironmentID: envID,
		ExternalID:    externalID,
	})
	if err != nil {
		return nil, err
	}

	var qOrgDomains []queries.OrganizationDomain
	for _, d := range req.Organization.Domains {
		qOrgDomain, err := q.CreateOrganizationDomain(ctx, queries.CreateOrganizationDomainParams{
			ID:             uuid.New(),
			OrganizationID: qOrg.ID,
			Domain:         d,
		})
		if err != nil {
			return nil, err
		}
		qOrgDomains = append(qOrgDomains, qOrgDomain)
	}

	if err := commit(); err != nil {
		return nil, err
	}

	return &ssoreadyv1.CreateOrganizationResponse{Organization: parseOrganization(qOrg, qOrgDomains)}, nil
}

func (s *Store) UpdateOrganization(ctx context.Context, req *ssoreadyv1.UpdateOrganizationRequest) (*ssoreadyv1.UpdateOrganizationResponse, error) {
	_, q, commit, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	id, err := idformat.Organization.Parse(req.Id)
	if err != nil {
		return nil, err
	}

	envID, err := idformat.Environment.Parse(authn.FullContextData(ctx).APIKey.EnvID)
	if err != nil {
		return nil, err
	}

	// authz check
	if _, err = q.ManagementGetOrganization(ctx, queries.ManagementGetOrganizationParams{
		EnvironmentID: envID,
		ID:            id,
	}); err != nil {
		return nil, err
	}

	if err := q.DeleteOrganizationDomains(ctx, id); err != nil {
		return nil, err
	}

	var externalID *string
	if req.Organization.ExternalId != "" {
		externalID = &req.Organization.ExternalId
	}

	qOrg, err := q.UpdateOrganization(ctx, queries.UpdateOrganizationParams{
		ID:         id,
		ExternalID: externalID,
	})
	if err != nil {
		return nil, err
	}

	var qOrgDomains []queries.OrganizationDomain
	for _, d := range req.Organization.Domains {
		qOrgDomain, err := q.CreateOrganizationDomain(ctx, queries.CreateOrganizationDomainParams{
			ID:             uuid.New(),
			OrganizationID: qOrg.ID,
			Domain:         d,
		})
		if err != nil {
			return nil, err
		}
		qOrgDomains = append(qOrgDomains, qOrgDomain)
	}

	if err := commit(); err != nil {
		return nil, err
	}

	return &ssoreadyv1.UpdateOrganizationResponse{Organization: parseOrganization(qOrg, qOrgDomains)}, nil
}
