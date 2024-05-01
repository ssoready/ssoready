package store

import (
	"context"
	"crypto/x509"
	"encoding/pem"

	"github.com/google/uuid"
	"github.com/ssoready/ssoready/internal/appauth"
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

	orgID, err := idformat.Organization.Parse(req.OrganizationId)
	if err != nil {
		return nil, err
	}

	// idor check
	if _, err = q.GetOrganization(ctx, queries.GetOrganizationParams{
		AppOrganizationID: appauth.OrgID(ctx),
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

		samlConns = append(samlConns, &ssoreadyv1.SAMLConnection{
			Id:                 idformat.SAMLConnection.Format(qSAMLConn.ID),
			OrganizationId:     idformat.Organization.Format(qSAMLConn.OrganizationID),
			IdpRedirectUrl:     derefOrEmpty(qSAMLConn.IdpRedirectUrl),
			IdpX509Certificate: certPEM,
			IdpEntityId:        derefOrEmpty(qSAMLConn.IdpEntityID),
		})
	}

	var nextPageToken string
	if len(samlConns) == limit+1 {
		nextPageToken = s.pageEncoder.Marshal(samlConns[limit].Id)
		samlConns = samlConns[:limit]
	}

	return &ssoreadyv1.ListSAMLConnectionsResponse{
		SamlConnections: samlConns,
		NextPageToken:   nextPageToken,
	}, nil
}

func (s *Store) GetSAMLConnection(ctx context.Context, req *ssoreadyv1.GetSAMLConnectionRequest) (*ssoreadyv1.SAMLConnection, error) {
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
		AppOrganizationID: appauth.OrgID(ctx),
		ID:                id,
	})
	if err != nil {
		return nil, err
	}

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
		Id:                 idformat.SAMLConnection.Format(qSAMLConn.ID),
		OrganizationId:     idformat.Organization.Format(qSAMLConn.OrganizationID),
		IdpRedirectUrl:     derefOrEmpty(qSAMLConn.IdpRedirectUrl),
		IdpX509Certificate: certPEM,
		IdpEntityId:        derefOrEmpty(qSAMLConn.IdpEntityID),
	}, nil
}

func (s *Store) CreateSAMLConnection(ctx context.Context, req *ssoreadyv1.CreateSAMLConnectionRequest) (*ssoreadyv1.SAMLConnection, error) {
	_, q, _, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	orgID, err := idformat.Organization.Parse(req.SamlConnection.OrganizationId)
	if err != nil {
		return nil, err
	}

	// idor check
	if _, err = q.GetOrganization(ctx, queries.GetOrganizationParams{
		AppOrganizationID: appauth.OrgID(ctx),
		ID:                orgID,
	}); err != nil {
		return nil, err
	}

	qSAMLConn, err := q.CreateSAMLConnection(ctx, queries.CreateSAMLConnectionParams{
		ID:             uuid.New(),
		OrganizationID: orgID,
	})
	if err != nil {
		return nil, err
	}

	return &ssoreadyv1.SAMLConnection{
		Id:             idformat.SAMLConnection.Format(qSAMLConn.ID),
		OrganizationId: idformat.Organization.Format(qSAMLConn.OrganizationID),
	}, nil
}
