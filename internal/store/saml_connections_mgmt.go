package store

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/google/uuid"
	"github.com/ssoready/ssoready/internal/authn"
	ssoreadyv1 "github.com/ssoready/ssoready/internal/gen/ssoready/v1"
	"github.com/ssoready/ssoready/internal/store/idformat"
	"github.com/ssoready/ssoready/internal/store/queries"
)

func (s *Store) ListSAMLConnections(ctx context.Context, req *ssoreadyv1.ListSAMLConnectionsRequest) (*ssoreadyv1.ListSAMLConnectionsResponse, error) {
	_, q, _, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	envID, err := idformat.Environment.Parse(authn.FullContextData(ctx).APIKey.EnvID)
	if err != nil {
		return nil, err
	}

	orgID, err := idformat.Organization.Parse(req.OrganizationId)
	if err != nil {
		return nil, err
	}

	// idor check
	if _, err = q.ManagementGetOrganization(ctx, queries.ManagementGetOrganizationParams{
		EnvironmentID: envID,
		ID:            orgID,
	}); err != nil {
		return nil, err
	}

	var startID uuid.UUID
	if err := s.pageEncoder.Unmarshal(req.PageToken, &startID); err != nil {
		return nil, err
	}

	limit := 10
	qSAMLConns, err := q.ListSAMLConnections(ctx, queries.ListSAMLConnectionsParams{
		OrganizationID: orgID,
		ID:             startID,
		Limit:          int32(limit + 1),
	})
	if err != nil {
		return nil, err
	}

	var samlConns []*ssoreadyv1.SAMLConnection
	for _, qSAMLConn := range qSAMLConns {
		samlConns = append(samlConns, parseSAMLConnection(qSAMLConn))
	}

	var nextPageToken string
	if len(samlConns) == limit+1 {
		nextPageToken = s.pageEncoder.Marshal(qSAMLConns[limit].ID)
		samlConns = samlConns[:limit]
	}

	return &ssoreadyv1.ListSAMLConnectionsResponse{
		SamlConnections: samlConns,
		NextPageToken:   nextPageToken,
	}, nil
}

func (s *Store) GetSAMLConnection(ctx context.Context, req *ssoreadyv1.GetSAMLConnectionRequest) (*ssoreadyv1.GetSAMLConnectionResponse, error) {
	envID, err := idformat.Environment.Parse(authn.FullContextData(ctx).APIKey.EnvID)
	if err != nil {
		return nil, err
	}

	id, err := idformat.SAMLConnection.Parse(req.Id)
	if err != nil {
		return nil, err
	}

	_, q, _, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	qSAMLConn, err := q.ManagementGetSAMLConnection(ctx, queries.ManagementGetSAMLConnectionParams{
		EnvironmentID: envID,
		ID:            id,
	})
	if err != nil {
		return nil, err
	}

	return &ssoreadyv1.GetSAMLConnectionResponse{SamlConnection: parseSAMLConnection(qSAMLConn)}, nil
}

func (s *Store) CreateSAMLConnection(ctx context.Context, req *ssoreadyv1.CreateSAMLConnectionRequest) (*ssoreadyv1.CreateSAMLConnectionResponse, error) {
	envID, err := idformat.Environment.Parse(authn.FullContextData(ctx).APIKey.EnvID)
	if err != nil {
		return nil, err
	}

	_, q, commit, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	orgID, err := idformat.Organization.Parse(req.SamlConnection.OrganizationId)
	if err != nil {
		return nil, err
	}

	// idor check
	if _, err := q.ManagementGetOrganization(ctx, queries.ManagementGetOrganizationParams{
		EnvironmentID: envID,
		ID:            orgID,
	}); err != nil {
		return nil, err
	}

	env, err := q.GetEnvironmentByID(ctx, envID)
	if err != nil {
		return nil, err
	}

	authURL := s.defaultAuthURL
	if env.AuthUrl != nil {
		authURL = *env.AuthUrl
	}

	var idpCert []byte
	if req.SamlConnection.IdpCertificate != "" {
		blk, _ := pem.Decode([]byte(req.SamlConnection.IdpCertificate))
		if blk == nil || blk.Type != "CERTIFICATE" {
			return nil, fmt.Errorf("idp certificate must be a PEM-encoded CERTIFICATE block")
		}
		if _, err := x509.ParseCertificate(blk.Bytes); err != nil {
			return nil, fmt.Errorf("parse idp certificate: %w", err)
		}
		idpCert = blk.Bytes
	}

	id := uuid.New()
	entityID := fmt.Sprintf("%s/v1/saml/%s", authURL, idformat.SAMLConnection.Format(id))
	acsURL := fmt.Sprintf("%s/v1/saml/%s/acs", authURL, idformat.SAMLConnection.Format(id))
	qSAMLConn, err := q.CreateSAMLConnection(ctx, queries.CreateSAMLConnectionParams{
		ID:                 id,
		OrganizationID:     orgID,
		IsPrimary:          req.SamlConnection.Primary,
		SpAcsUrl:           acsURL,
		SpEntityID:         entityID,
		IdpEntityID:        &req.SamlConnection.IdpEntityId,
		IdpRedirectUrl:     &req.SamlConnection.IdpRedirectUrl,
		IdpX509Certificate: idpCert,
	})
	if err != nil {
		return nil, err
	}

	if qSAMLConn.IsPrimary {
		if err := q.UpdatePrimarySAMLConnection(ctx, queries.UpdatePrimarySAMLConnectionParams{
			OrganizationID: qSAMLConn.OrganizationID,
			ID:             qSAMLConn.ID,
		}); err != nil {
			return nil, err
		}
	}

	if err := commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return &ssoreadyv1.CreateSAMLConnectionResponse{SamlConnection: parseSAMLConnection(qSAMLConn)}, nil
}

func (s *Store) UpdateSAMLConnection(ctx context.Context, req *ssoreadyv1.UpdateSAMLConnectionRequest) (*ssoreadyv1.UpdateSAMLConnectionResponse, error) {
	envID, err := idformat.Environment.Parse(authn.FullContextData(ctx).APIKey.EnvID)
	if err != nil {
		return nil, err
	}

	_, q, commit, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	id, err := idformat.SAMLConnection.Parse(req.Id)
	if err != nil {
		return nil, fmt.Errorf("parse saml connection id: %w", err)
	}

	// idor check
	if _, err = q.ManagementGetSAMLConnection(ctx, queries.ManagementGetSAMLConnectionParams{
		EnvironmentID: envID,
		ID:            id,
	}); err != nil {
		return nil, fmt.Errorf("get saml connection: %w", err)
	}

	var idpCert []byte
	if req.SamlConnection.IdpCertificate != "" {
		blk, _ := pem.Decode([]byte(req.SamlConnection.IdpCertificate))
		if blk == nil || blk.Type != "CERTIFICATE" {
			return nil, fmt.Errorf("idp certificate must be a PEM-encoded CERTIFICATE block")
		}
		if _, err := x509.ParseCertificate(blk.Bytes); err != nil {
			return nil, fmt.Errorf("parse idp certificate: %w", err)
		}
		idpCert = blk.Bytes
	}

	qSAMLConn, err := q.UpdateSAMLConnection(ctx, queries.UpdateSAMLConnectionParams{
		ID:                 id,
		IsPrimary:          req.SamlConnection.Primary,
		IdpEntityID:        &req.SamlConnection.IdpEntityId,
		IdpRedirectUrl:     &req.SamlConnection.IdpRedirectUrl,
		IdpX509Certificate: idpCert,
	})
	if err != nil {
		return nil, fmt.Errorf("update saml connection: %w", err)
	}

	if qSAMLConn.IsPrimary {
		if err := q.UpdatePrimarySAMLConnection(ctx, queries.UpdatePrimarySAMLConnectionParams{
			OrganizationID: qSAMLConn.OrganizationID,
			ID:             qSAMLConn.ID,
		}); err != nil {
			return nil, err
		}
	}

	if err := commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return &ssoreadyv1.UpdateSAMLConnectionResponse{SamlConnection: parseSAMLConnection(qSAMLConn)}, nil
}
