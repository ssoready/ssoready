package store

import (
	"context"
	"fmt"
	"log/slog"
	"sort"

	"github.com/google/uuid"
	"github.com/ssoready/ssoready/internal/authn"
	ssoreadyv1 "github.com/ssoready/ssoready/internal/gen/ssoready/v1"
	"github.com/ssoready/ssoready/internal/store/idformat"
	"github.com/ssoready/ssoready/internal/store/queries"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Store) AppListOrganizations(ctx context.Context, req *ssoreadyv1.AppListOrganizationsRequest) (*ssoreadyv1.AppListOrganizationsResponse, error) {
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
		AppOrganizationID: authn.AppOrgID(ctx),
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
		orgs = append(orgs, parseOrganization(qOrg, qOrgDomains))
	}

	var nextPageToken string
	if len(orgs) == limit+1 {
		nextPageToken = s.pageEncoder.Marshal(qOrgs[limit].ID)
		orgs = orgs[:limit]
	}

	return &ssoreadyv1.AppListOrganizationsResponse{
		Organizations: orgs,
		NextPageToken: nextPageToken,
	}, nil
}

func (s *Store) AppGetOrganization(ctx context.Context, req *ssoreadyv1.AppGetOrganizationRequest) (*ssoreadyv1.Organization, error) {
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
		AppOrganizationID: authn.AppOrgID(ctx),
		ID:                id,
	})
	if err != nil {
		return nil, err
	}

	qOrgDomains, err := q.ListOrganizationDomains(ctx, []uuid.UUID{qOrg.ID})
	if err != nil {
		return nil, err
	}

	return parseOrganization(qOrg, qOrgDomains), nil
}

func (s *Store) AppCreateOrganization(ctx context.Context, req *ssoreadyv1.AppCreateOrganizationRequest) (*ssoreadyv1.Organization, error) {
	_, q, commit, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	envID, err := idformat.Environment.Parse(req.Organization.EnvironmentId)
	if err != nil {
		return nil, err
	}

	// idor check
	if _, err = q.GetEnvironment(ctx, queries.GetEnvironmentParams{
		AppOrganizationID: authn.AppOrgID(ctx),
		ID:                envID,
	}); err != nil {
		return nil, err
	}

	var displayName *string
	if req.Organization.DisplayName != "" {
		displayName = &req.Organization.DisplayName
	}

	var externalID *string
	if req.Organization.ExternalId != "" {
		externalID = &req.Organization.ExternalId
	}

	qOrg, err := q.CreateOrganization(ctx, queries.CreateOrganizationParams{
		ID:            uuid.New(),
		EnvironmentID: envID,
		DisplayName:   displayName,
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

	return parseOrganization(qOrg, qOrgDomains), nil
}

func (s *Store) AppUpdateOrganization(ctx context.Context, req *ssoreadyv1.AppUpdateOrganizationRequest) (*ssoreadyv1.Organization, error) {
	_, q, commit, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	id, err := idformat.Organization.Parse(req.Organization.Id)
	if err != nil {
		return nil, err
	}

	// authz check
	if _, err = q.GetOrganization(ctx, queries.GetOrganizationParams{
		AppOrganizationID: authn.AppOrgID(ctx),
		ID:                id,
	}); err != nil {
		return nil, err
	}

	if err := q.DeleteOrganizationDomains(ctx, id); err != nil {
		return nil, err
	}

	var displayName *string
	if req.Organization.DisplayName != "" {
		displayName = &req.Organization.DisplayName
	}

	var externalID *string
	if req.Organization.ExternalId != "" {
		externalID = &req.Organization.ExternalId
	}

	qOrg, err := q.UpdateOrganization(ctx, queries.UpdateOrganizationParams{
		ID:          id,
		DisplayName: displayName,
		ExternalID:  externalID,
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

	return parseOrganization(qOrg, qOrgDomains), nil
}

func (s *Store) AppDeleteOrganization(ctx context.Context, req *ssoreadyv1.AppDeleteOrganizationRequest) (*emptypb.Empty, error) {
	_, q, commit, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	id, err := idformat.Organization.Parse(req.OrganizationId)
	if err != nil {
		return nil, err
	}

	// authz check
	if _, err = q.GetOrganization(ctx, queries.GetOrganizationParams{
		AppOrganizationID: authn.AppOrgID(ctx),
		ID:                id,
	}); err != nil {
		return nil, err
	}

	slog.InfoContext(ctx, "delete_organization", "organization_id", req.OrganizationId)

	// delete each saml connection
	samlConnectionIDs, err := q.ListAllSAMLConnectionIDs(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("list all saml connection ids: %w", err)
	}

	for _, samlConnID := range samlConnectionIDs {
		if err := s.deleteSAMLConnection(ctx, q, samlConnID); err != nil {
			return nil, fmt.Errorf("delete saml connection: %w", err)
		}
	}

	// delete each scim directory
	scimDirectoryIDs, err := q.ListAllSCIMDirectoryIDs(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("list all scim directory ids: %w", err)
	}

	for _, scimDirID := range scimDirectoryIDs {
		if err := s.deleteSCIMDirectory(ctx, q, scimDirID); err != nil {
			return nil, fmt.Errorf("delete scim directory: %w", err)
		}
	}

	if err := q.DeleteOrganizationDomains(ctx, id); err != nil {
		return nil, fmt.Errorf("delete organization domains: %w", err)
	}

	// delete admin access tokens
	adminAccessTokenCount, err := q.DeleteOrganizationAdminAccessTokens(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("delete organization admin access tokens: %w", err)
	}

	slog.InfoContext(ctx, "delete_organization", "admin_access_token_count", adminAccessTokenCount)

	if _, err := q.DeleteOrganization(ctx, id); err != nil {
		return nil, fmt.Errorf("delete organization: %w", err)
	}

	if err := commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return &emptypb.Empty{}, nil
}

func parseOrganization(qOrg queries.Organization, qOrgDomains []queries.OrganizationDomain) *ssoreadyv1.Organization {
	var domains []string
	for _, qOrgDomain := range qOrgDomains {
		if qOrgDomain.OrganizationID == qOrg.ID {
			domains = append(domains, qOrgDomain.Domain)
		}
	}
	sort.Strings(domains)

	return &ssoreadyv1.Organization{
		Id:            idformat.Organization.Format(qOrg.ID),
		EnvironmentId: idformat.Environment.Format(qOrg.EnvironmentID),
		DisplayName:   derefOrEmpty(qOrg.DisplayName),
		ExternalId:    derefOrEmpty(qOrg.ExternalID),
		Domains:       domains,
	}
}
