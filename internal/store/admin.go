package store

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"net/url"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/ssoready/ssoready/internal/authn"
	ssoreadyv1 "github.com/ssoready/ssoready/internal/gen/ssoready/v1"
	"github.com/ssoready/ssoready/internal/store/idformat"
	"github.com/ssoready/ssoready/internal/store/queries"
)

func (s *Store) AppCreateAdminSetupURL(ctx context.Context, req *ssoreadyv1.AppCreateAdminSetupURLRequest) (*ssoreadyv1.AppCreateAdminSetupURLResponse, error) {
	_, q, commit, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	orgID, err := idformat.Organization.Parse(req.OrganizationId)
	if err != nil {
		return nil, fmt.Errorf("parse organization id: %w", err)
	}

	// idor check
	org, err := q.GetOrganization(ctx, queries.GetOrganizationParams{
		AppOrganizationID: authn.AppOrgID(ctx),
		ID:                orgID,
	})
	if err != nil {
		return nil, fmt.Errorf("get organization: %w", err)
	}

	oneTimeToken := uuid.New()
	oneTimeTokenSHA := sha256.Sum256(oneTimeToken[:])

	if _, err := q.CreateAdminAccessToken(ctx, queries.CreateAdminAccessTokenParams{
		ID:                 uuid.New(),
		OrganizationID:     org.ID,
		OneTimeTokenSha256: oneTimeTokenSHA[:],
		CreateTime:         time.Now(),
		ExpireTime:         time.Now().Add(time.Hour * 24),
		CanManageSaml:      &req.CanManageSaml,
		CanManageScim:      &req.CanManageScim,
	}); err != nil {
		return nil, fmt.Errorf("create admin access token: %w", err)
	}

	if err := commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	loginURL, err := url.Parse(s.defaultAdminSetupURL)
	if err != nil {
		panic(fmt.Errorf("parse default admin login url: %w", err))
	}

	query := url.Values{}
	query.Set("one-time-token", idformat.AdminOneTimeToken.Format(oneTimeToken))

	loginURL.RawQuery = query.Encode()

	return &ssoreadyv1.AppCreateAdminSetupURLResponse{
		Url: loginURL.String(),
	}, nil
}

func (s *Store) AdminRedeemOneTimeToken(ctx context.Context, req *ssoreadyv1.AdminRedeemOneTimeTokenRequest) (*ssoreadyv1.AdminRedeemOneTimeTokenResponse, error) {
	_, q, commit, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	oneTimeToken, err := idformat.AdminOneTimeToken.Parse(req.OneTimeToken)
	if err != nil {
		return nil, fmt.Errorf("")
	}

	oneTimeTokenSHA := sha256.Sum256(oneTimeToken[:])
	adminAccessToken, err := q.AdminGetAdminAccessTokenByOneTimeToken(ctx, oneTimeTokenSHA[:])
	if err != nil {
		return nil, fmt.Errorf("get admin access token: %w", err)
	}

	// generate token as 32-byte random string
	var tokenBytes [32]byte
	if _, err := rand.Read(tokenBytes[:]); err != nil {
		return nil, err
	}
	accessTokenHex := hex.EncodeToString(tokenBytes[:])
	accessTokenSHA := sha256.Sum256(tokenBytes[:])

	if _, err := q.AdminConvertAdminAccessTokenToSession(ctx, queries.AdminConvertAdminAccessTokenToSessionParams{
		AccessTokenSha256: accessTokenSHA[:],
		ID:                adminAccessToken.ID,
	}); err != nil {
		return nil, fmt.Errorf("convert admin access token to session: %w", err)
	}

	if err := commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return &ssoreadyv1.AdminRedeemOneTimeTokenResponse{
		AdminSessionToken: accessTokenHex,
	}, nil
}

type AdminGetAdminSessionResponse struct {
	OrganizationID uuid.UUID
	CanManageSAML  bool
	CanManageSCIM  bool
}

func (s *Store) AdminGetAdminSession(ctx context.Context, sessionToken string) (*AdminGetAdminSessionResponse, error) {
	token, err := hex.DecodeString(sessionToken)
	if err != nil {
		return nil, fmt.Errorf("parse session token: %w", err)
	}

	sha := sha256.Sum256(token)
	qAdminAccessToken, err := s.q.AdminGetAdminAccessTokenByAccessToken(ctx, queries.AdminGetAdminAccessTokenByAccessTokenParams{
		AccessTokenSha256: sha[:],
		ExpireTime:        time.Now(),
	})
	if err != nil {
		return nil, fmt.Errorf("get organization by admin access token: %w", err)
	}

	return &AdminGetAdminSessionResponse{
		OrganizationID: qAdminAccessToken.OrganizationID,
		CanManageSAML:  derefOrEmpty(qAdminAccessToken.CanManageSaml),
		CanManageSCIM:  derefOrEmpty(qAdminAccessToken.CanManageScim),
	}, nil
}

func (s *Store) AdminWhoami(ctx context.Context, req *ssoreadyv1.AdminWhoamiRequest) (*ssoreadyv1.AdminWhoamiResponse, error) {
	tokenAuthnData := authn.FullContextData(ctx).AdminAccessToken
	return &ssoreadyv1.AdminWhoamiResponse{
		CanManageSaml: tokenAuthnData.CanManageSAML,
		CanManageScim: tokenAuthnData.CanManageSCIM,
	}, nil
}

func (s *Store) AdminListSAMLConnections(ctx context.Context, req *ssoreadyv1.AdminListSAMLConnectionsRequest) (*ssoreadyv1.AdminListSAMLConnectionsResponse, error) {
	_, q, _, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	if !authn.FullContextData(ctx).AdminAccessToken.CanManageSAML {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("not authorized to manage saml"))
	}

	orgID := authn.FullContextData(ctx).AdminAccessToken.OrganizationID

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

	return &ssoreadyv1.AdminListSAMLConnectionsResponse{
		SamlConnections: samlConns,
		NextPageToken:   nextPageToken,
	}, nil
}

func (s *Store) AdminGetSAMLConnection(ctx context.Context, req *ssoreadyv1.AdminGetSAMLConnectionRequest) (*ssoreadyv1.AdminGetSAMLConnectionResponse, error) {
	id, err := idformat.SAMLConnection.Parse(req.Id)
	if err != nil {
		return nil, err
	}

	if !authn.FullContextData(ctx).AdminAccessToken.CanManageSAML {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("not authorized to manage saml"))
	}

	orgID := authn.FullContextData(ctx).AdminAccessToken.OrganizationID

	_, q, _, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	qSAMLConn, err := q.AdminGetSAMLConnection(ctx, queries.AdminGetSAMLConnectionParams{
		OrganizationID: orgID,
		ID:             id,
	})
	if err != nil {
		return nil, err
	}

	return &ssoreadyv1.AdminGetSAMLConnectionResponse{SamlConnection: parseSAMLConnection(qSAMLConn)}, nil
}

func (s *Store) AdminCreateSAMLConnection(ctx context.Context, req *ssoreadyv1.AdminCreateSAMLConnectionRequest) (*ssoreadyv1.AdminCreateSAMLConnectionResponse, error) {
	_, q, commit, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	if !authn.FullContextData(ctx).AdminAccessToken.CanManageSAML {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("not authorized to manage saml"))
	}

	orgID := authn.FullContextData(ctx).AdminAccessToken.OrganizationID

	org, err := q.GetOrganizationByID(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("get org by id: %w", err)
	}

	env, err := q.GetEnvironmentByID(ctx, org.EnvironmentID)
	if err != nil {
		return nil, fmt.Errorf("get env by id: %w", err)
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

	return &ssoreadyv1.AdminCreateSAMLConnectionResponse{SamlConnection: parseSAMLConnection(qSAMLConn)}, nil
}

func (s *Store) AdminUpdateSAMLConnection(ctx context.Context, req *ssoreadyv1.AdminUpdateSAMLConnectionRequest) (*ssoreadyv1.AdminUpdateSAMLConnectionResponse, error) {
	_, q, commit, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	if !authn.FullContextData(ctx).AdminAccessToken.CanManageSAML {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("not authorized to manage saml"))
	}

	orgID := authn.FullContextData(ctx).AdminAccessToken.OrganizationID

	id, err := idformat.SAMLConnection.Parse(req.SamlConnection.Id)
	if err != nil {
		return nil, fmt.Errorf("parse saml connection id: %w", err)
	}

	// authz check; check saml connection has the same org ID as the admin access token
	samlConn, err := q.GetSAMLConnectionByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get saml conn by id: %w", err)
	}

	if samlConn.OrganizationID != orgID {
		return nil, fmt.Errorf("saml conn organization id != admin access token org id")
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

	return &ssoreadyv1.AdminUpdateSAMLConnectionResponse{SamlConnection: parseSAMLConnection(qSAMLConn)}, nil
}

func (s *Store) AdminListSCIMDirectories(ctx context.Context, req *ssoreadyv1.AdminListSCIMDirectoriesRequest) (*ssoreadyv1.AdminListSCIMDirectoriesResponse, error) {
	_, q, _, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	if !authn.FullContextData(ctx).AdminAccessToken.CanManageSAML {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("not authorized to manage saml"))
	}

	orgID := authn.FullContextData(ctx).AdminAccessToken.OrganizationID

	org, err := q.GetOrganizationByID(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("get org by id: %w", err)
	}

	var startID uuid.UUID
	if err := s.pageEncoder.Unmarshal(req.PageToken, &startID); err != nil {
		return nil, err
	}

	limit := 10
	qSCIMDirectories, err := q.ListSCIMDirectories(ctx, queries.ListSCIMDirectoriesParams{
		OrganizationID: org.ID,
		ID:             startID,
		Limit:          int32(limit + 1),
	})
	if err != nil {
		return nil, err
	}

	var scimDirectories []*ssoreadyv1.SCIMDirectory
	for _, qSCIMDirectory := range qSCIMDirectories {
		scimDirectories = append(scimDirectories, parseSCIMDirectory(qSCIMDirectory))
	}

	var nextPageToken string
	if len(scimDirectories) == limit+1 {
		nextPageToken = s.pageEncoder.Marshal(qSCIMDirectories[limit].ID)
		scimDirectories = scimDirectories[:limit]
	}

	return &ssoreadyv1.AdminListSCIMDirectoriesResponse{
		ScimDirectories: scimDirectories,
		NextPageToken:   nextPageToken,
	}, nil
}

func (s *Store) AdminGetSCIMDirectory(ctx context.Context, req *ssoreadyv1.AdminGetSCIMDirectoryRequest) (*ssoreadyv1.AdminGetSCIMDirectoryResponse, error) {
	id, err := idformat.SCIMDirectory.Parse(req.Id)
	if err != nil {
		return nil, err
	}

	if !authn.FullContextData(ctx).AdminAccessToken.CanManageSCIM {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("not authorized to manage scim"))
	}

	orgID := authn.FullContextData(ctx).AdminAccessToken.OrganizationID

	_, q, _, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	qSCIMDir, err := q.AdminGetSCIMDirectory(ctx, queries.AdminGetSCIMDirectoryParams{
		OrganizationID: orgID,
		ID:             id,
	})
	if err != nil {
		return nil, err
	}

	return &ssoreadyv1.AdminGetSCIMDirectoryResponse{ScimDirectory: parseSCIMDirectory(qSCIMDir)}, nil
}

func (s *Store) AdminCreateSCIMDirectory(ctx context.Context, req *ssoreadyv1.AdminCreateSCIMDirectoryRequest) (*ssoreadyv1.AdminCreateSCIMDirectoryResponse, error) {
	_, q, commit, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	if !authn.FullContextData(ctx).AdminAccessToken.CanManageSCIM {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("not authorized to manage scim"))
	}

	orgID := authn.FullContextData(ctx).AdminAccessToken.OrganizationID

	org, err := q.GetOrganizationByID(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("get org by id: %w", err)
	}

	env, err := q.GetEnvironmentByID(ctx, org.EnvironmentID)
	if err != nil {
		return nil, fmt.Errorf("get env by id: %w", err)
	}

	authURL := s.defaultAuthURL
	if env.AuthUrl != nil {
		authURL = *env.AuthUrl
	}

	id := uuid.New()
	scimBaseURL := fmt.Sprintf("%s/v1/scim/%s", authURL, idformat.SCIMDirectory.Format(id))
	qSCIMDirectory, err := q.CreateSCIMDirectory(ctx, queries.CreateSCIMDirectoryParams{
		ID:             id,
		OrganizationID: orgID,
		IsPrimary:      req.ScimDirectory.Primary,
		ScimBaseUrl:    scimBaseURL,
	})
	if err != nil {
		return nil, fmt.Errorf("create scim directory: %w", err)
	}

	if qSCIMDirectory.IsPrimary {
		if err := q.UpdatePrimarySCIMDirectory(ctx, queries.UpdatePrimarySCIMDirectoryParams{
			ID:             qSCIMDirectory.ID,
			OrganizationID: qSCIMDirectory.OrganizationID,
		}); err != nil {
			return nil, fmt.Errorf("update primary scim directory: %w", err)
		}
	}

	if err := commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return &ssoreadyv1.AdminCreateSCIMDirectoryResponse{ScimDirectory: parseSCIMDirectory(qSCIMDirectory)}, nil
}

func (s *Store) AdminUpdateSCIMDirectory(ctx context.Context, req *ssoreadyv1.AdminUpdateSCIMDirectoryRequest) (*ssoreadyv1.AdminUpdateSCIMDirectoryResponse, error) {
	_, q, commit, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	if !authn.FullContextData(ctx).AdminAccessToken.CanManageSCIM {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("not authorized to manage scim"))
	}

	orgID := authn.FullContextData(ctx).AdminAccessToken.OrganizationID

	scimDirID, err := idformat.SCIMDirectory.Parse(req.ScimDirectory.Id)
	if err != nil {
		return nil, fmt.Errorf("parse scim directory id: %w", err)
	}

	// authz check; check scim directory has the same org ID as the admin access token
	scimDir, err := q.GetSCIMDirectoryByID(ctx, scimDirID)
	if err != nil {
		return nil, fmt.Errorf("get saml conn by id: %w", err)
	}

	if scimDir.OrganizationID != orgID {
		return nil, fmt.Errorf("scim dir organization id != admin access token org id")
	}

	qSCIMDir, err := q.UpdateSCIMDirectory(ctx, queries.UpdateSCIMDirectoryParams{
		ID:        scimDirID,
		IsPrimary: req.ScimDirectory.Primary,
	})
	if err != nil {
		return nil, fmt.Errorf("update scim directory: %w", err)
	}

	if qSCIMDir.IsPrimary {
		if err := q.UpdatePrimarySCIMDirectory(ctx, queries.UpdatePrimarySCIMDirectoryParams{
			ID:             qSCIMDir.ID,
			OrganizationID: qSCIMDir.OrganizationID,
		}); err != nil {
			return nil, fmt.Errorf("update primary scim directory: %w", err)
		}
	}

	if err := commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return &ssoreadyv1.AdminUpdateSCIMDirectoryResponse{ScimDirectory: parseSCIMDirectory(qSCIMDir)}, nil
}

func (s *Store) AdminRotateSCIMDirectoryBearerToken(ctx context.Context, req *ssoreadyv1.AdminRotateSCIMDirectoryBearerTokenRequest) (*ssoreadyv1.AdminRotateSCIMDirectoryBearerTokenResponse, error) {
	_, q, commit, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	if !authn.FullContextData(ctx).AdminAccessToken.CanManageSCIM {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("not authorized to manage scim"))
	}

	orgID := authn.FullContextData(ctx).AdminAccessToken.OrganizationID

	scimDirID, err := idformat.SCIMDirectory.Parse(req.ScimDirectoryId)
	if err != nil {
		return nil, fmt.Errorf("parse scim directory id: %w", err)
	}

	// authz check; check scim directory has the same org ID as the admin access token
	scimDir, err := q.GetSCIMDirectoryByID(ctx, scimDirID)
	if err != nil {
		return nil, fmt.Errorf("get saml conn by id: %w", err)
	}

	if scimDir.OrganizationID != orgID {
		return nil, fmt.Errorf("scim dir organization id != admin access token org id")
	}

	bearerToken := uuid.New()
	bearerTokenSHA := sha256.Sum256(bearerToken[:])

	if _, err := q.UpdateSCIMDirectoryBearerToken(ctx, queries.UpdateSCIMDirectoryBearerTokenParams{
		BearerTokenSha256: bearerTokenSHA[:],
		ID:                scimDirID,
	}); err != nil {
		return nil, fmt.Errorf("update scim directory access token: %w", err)
	}

	if err := commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return &ssoreadyv1.AdminRotateSCIMDirectoryBearerTokenResponse{
		BearerToken: idformat.SCIMBearerToken.Format(bearerToken),
	}, nil
}
