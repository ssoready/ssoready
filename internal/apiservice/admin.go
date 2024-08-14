package apiservice

import (
	"context"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"

	"connectrpc.com/connect"
	ssoreadyv1 "github.com/ssoready/ssoready/internal/gen/ssoready/v1"
	"github.com/ssoready/ssoready/internal/saml"
)

func (s *Service) CreateAdminSetupURL(ctx context.Context, req *connect.Request[ssoreadyv1.CreateAdminSetupURLRequest]) (*connect.Response[ssoreadyv1.CreateAdminSetupURLResponse], error) {
	res, err := s.Store.CreateAdminSetupURL(ctx, req.Msg)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}

	return connect.NewResponse(res), nil
}

func (s *Service) AdminRedeemOneTimeToken(ctx context.Context, req *connect.Request[ssoreadyv1.AdminRedeemOneTimeTokenRequest]) (*connect.Response[ssoreadyv1.AdminRedeemOneTimeTokenResponse], error) {
	res, err := s.Store.AdminRedeemOneTimeToken(ctx, req.Msg)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}

	return connect.NewResponse(res), nil
}

func (s *Service) AdminListSAMLConnections(ctx context.Context, req *connect.Request[ssoreadyv1.AdminListSAMLConnectionsRequest]) (*connect.Response[ssoreadyv1.AdminListSAMLConnectionsResponse], error) {
	res, err := s.Store.AdminListSAMLConnections(ctx, req.Msg)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}
	return connect.NewResponse(res), nil
}

func (s *Service) AdminGetSAMLConnection(ctx context.Context, req *connect.Request[ssoreadyv1.AdminGetSAMLConnectionRequest]) (*connect.Response[ssoreadyv1.AdminGetSAMLConnectionResponse], error) {
	res, err := s.Store.AdminGetSAMLConnection(ctx, req.Msg)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}
	return connect.NewResponse(res), nil
}

func (s *Service) AdminCreateSAMLConnection(ctx context.Context, req *connect.Request[ssoreadyv1.AdminCreateSAMLConnectionRequest]) (*connect.Response[ssoreadyv1.AdminCreateSAMLConnectionResponse], error) {
	res, err := s.Store.AdminCreateSAMLConnection(ctx, req.Msg)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}
	return connect.NewResponse(res), nil
}

func (s *Service) AdminUpdateSAMLConnection(ctx context.Context, req *connect.Request[ssoreadyv1.AdminUpdateSAMLConnectionRequest]) (*connect.Response[ssoreadyv1.AdminUpdateSAMLConnectionResponse], error) {
	res, err := s.Store.AdminUpdateSAMLConnection(ctx, req.Msg)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}
	return connect.NewResponse(res), nil
}

func (s *Service) AdminParseSAMLMetadata(ctx context.Context, req *connect.Request[ssoreadyv1.AdminParseSAMLMetadataRequest]) (*connect.Response[ssoreadyv1.AdminParseSAMLMetadataResponse], error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, req.Msg.Url, nil)
	if err != nil {
		return nil, err
	}

	httpRes, err := s.SAMLMetadataHTTPClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer httpRes.Body.Close()

	body, err := io.ReadAll(httpRes.Body)
	if err != nil {
		return nil, err
	}

	metadataRes, err := saml.ParseMetadata(body)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&ssoreadyv1.AdminParseSAMLMetadataResponse{
		IdpRedirectUrl: metadataRes.RedirectURL,
		IdpEntityId:    metadataRes.IDPEntityID,
		IdpCertificate: string(pem.EncodeToMemory(&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: metadataRes.IDPCertificate.Raw,
		})),
	}), nil
}

func (s *Service) AdminListSCIMDirectories(ctx context.Context, req *connect.Request[ssoreadyv1.AdminListSCIMDirectoriesRequest]) (*connect.Response[ssoreadyv1.AdminListSCIMDirectoriesResponse], error) {
	res, err := s.Store.AdminListSCIMDirectories(ctx, req.Msg)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}
	return connect.NewResponse(res), nil
}
func (s *Service) AdminGetSCIMDirectory(ctx context.Context, req *connect.Request[ssoreadyv1.AdminGetSCIMDirectoryRequest]) (*connect.Response[ssoreadyv1.AdminGetSCIMDirectoryResponse], error) {
	res, err := s.Store.AdminGetSCIMDirectory(ctx, req.Msg)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}
	return connect.NewResponse(res), nil
}
func (s *Service) AdminCreateSCIMDirectory(ctx context.Context, req *connect.Request[ssoreadyv1.AdminCreateSCIMDirectoryRequest]) (*connect.Response[ssoreadyv1.AdminCreateSCIMDirectoryResponse], error) {
	res, err := s.Store.AdminCreateSCIMDirectory(ctx, req.Msg)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}
	return connect.NewResponse(res), nil
}
func (s *Service) AdminUpdateSCIMDirectory(ctx context.Context, req *connect.Request[ssoreadyv1.AdminUpdateSCIMDirectoryRequest]) (*connect.Response[ssoreadyv1.AdminUpdateSCIMDirectoryResponse], error) {
	res, err := s.Store.AdminUpdateSCIMDirectory(ctx, req.Msg)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}
	return connect.NewResponse(res), nil
}
func (s *Service) AdminRotateSCIMDirectoryBearerToken(ctx context.Context, req *connect.Request[ssoreadyv1.AdminRotateSCIMDirectoryBearerTokenRequest]) (*connect.Response[ssoreadyv1.AdminRotateSCIMDirectoryBearerTokenResponse], error) {
	res, err := s.Store.AdminRotateSCIMDirectoryBearerToken(ctx, req.Msg)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}
	return connect.NewResponse(res), nil
}
