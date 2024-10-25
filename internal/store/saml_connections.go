package store

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/ssoready/ssoready/internal/authn"
	ssoreadyv1 "github.com/ssoready/ssoready/internal/gen/ssoready/v1"
	"github.com/ssoready/ssoready/internal/store/idformat"
	"github.com/ssoready/ssoready/internal/store/queries"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Store) AppListSAMLConnections(ctx context.Context, req *ssoreadyv1.AppListSAMLConnectionsRequest) (*ssoreadyv1.AppListSAMLConnectionsResponse, error) {
	_, q, _, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	orgID, err := idformat.Organization.Parse(req.OrganizationId)
	if err != nil {
		return nil, err
	}

	// idor check
	if _, err = q.GetOrganization(ctx, queries.GetOrganizationParams{
		AppOrganizationID: authn.AppOrgID(ctx),
		ID:                orgID,
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

	return &ssoreadyv1.AppListSAMLConnectionsResponse{
		SamlConnections: samlConns,
		NextPageToken:   nextPageToken,
	}, nil
}

func (s *Store) AppGetSAMLConnection(ctx context.Context, req *ssoreadyv1.AppGetSAMLConnectionRequest) (*ssoreadyv1.SAMLConnection, error) {
	id, err := idformat.SAMLConnection.Parse(req.Id)
	if err != nil {
		return nil, err
	}

	_, q, _, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	qSAMLConn, err := q.GetSAMLConnection(ctx, queries.GetSAMLConnectionParams{
		AppOrganizationID: authn.AppOrgID(ctx),
		ID:                id,
	})
	if err != nil {
		return nil, err
	}

	return parseSAMLConnection(qSAMLConn), nil
}

func (s *Store) AppCreateSAMLConnection(ctx context.Context, req *ssoreadyv1.AppCreateSAMLConnectionRequest) (*ssoreadyv1.SAMLConnection, error) {
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
	org, err := q.GetOrganization(ctx, queries.GetOrganizationParams{
		AppOrganizationID: authn.AppOrgID(ctx),
		ID:                orgID,
	})
	if err != nil {
		return nil, err
	}

	env, err := q.GetEnvironment(ctx, queries.GetEnvironmentParams{
		AppOrganizationID: authn.AppOrgID(ctx),
		ID:                org.EnvironmentID,
	})
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

	return parseSAMLConnection(qSAMLConn), nil
}

func (s *Store) AppUpdateSAMLConnection(ctx context.Context, req *ssoreadyv1.AppUpdateSAMLConnectionRequest) (*ssoreadyv1.SAMLConnection, error) {
	_, q, commit, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	id, err := idformat.SAMLConnection.Parse(req.SamlConnection.Id)
	if err != nil {
		return nil, fmt.Errorf("parse saml connection id: %w", err)
	}

	// idor check
	if _, err = q.GetSAMLConnection(ctx, queries.GetSAMLConnectionParams{
		AppOrganizationID: authn.AppOrgID(ctx),
		ID:                id,
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

	return parseSAMLConnection(qSAMLConn), nil
}

func (s *Store) AppDeleteSAMLConnection(ctx context.Context, req *ssoreadyv1.AppDeleteSAMLConnectionRequest) (*emptypb.Empty, error) {
	_, q, commit, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	slog.InfoContext(ctx, "delete_saml_connection", "saml_connection_id", req.SamlConnectionId)

	samlConnID, err := idformat.SAMLConnection.Parse(req.SamlConnectionId)
	if err != nil {
		return nil, fmt.Errorf("parse saml connection id: %w", err)
	}

	// idor check
	if _, err = q.GetSAMLConnection(ctx, queries.GetSAMLConnectionParams{
		AppOrganizationID: authn.AppOrgID(ctx),
		ID:                samlConnID,
	}); err != nil {
		return nil, fmt.Errorf("get saml connection: %w", err)
	}

	flowsCount, err := q.DeleteSAMLFlowsBySAMLConnectionID(ctx, samlConnID)
	if err != nil {
		return nil, fmt.Errorf("delete saml flows: %w", err)
	}

	slog.InfoContext(ctx, "delete_saml_connection", "flows_count", flowsCount)

	samlConnsCount, err := q.DeleteSAMLConnection(ctx, samlConnID)
	if err != nil {
		return nil, fmt.Errorf("delete saml connection: %w", err)
	}

	slog.InfoContext(ctx, "delete_saml_connection", "saml_conns_count", samlConnsCount)

	if err := commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return &emptypb.Empty{}, nil
}

func parseSAMLConnection(qSAMLConn queries.SamlConnection) *ssoreadyv1.SAMLConnection {
	var certPEM string
	if len(qSAMLConn.IdpX509Certificate) != 0 {
		cert, err := x509.ParseCertificate(qSAMLConn.IdpX509Certificate)
		if err != nil {
			panic(err)
		}

		certPEM = string(pem.EncodeToMemory(&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: cert.Raw,
		}))
	}

	return &ssoreadyv1.SAMLConnection{
		Id:             idformat.SAMLConnection.Format(qSAMLConn.ID),
		OrganizationId: idformat.Organization.Format(qSAMLConn.OrganizationID),
		Primary:        qSAMLConn.IsPrimary,
		IdpRedirectUrl: derefOrEmpty(qSAMLConn.IdpRedirectUrl),
		IdpCertificate: certPEM,
		IdpEntityId:    derefOrEmpty(qSAMLConn.IdpEntityID),
		SpEntityId:     qSAMLConn.SpEntityID,
		SpAcsUrl:       qSAMLConn.SpAcsUrl,
	}
}
