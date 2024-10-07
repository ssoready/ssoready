package apiservice

import (
	"context"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"connectrpc.com/connect"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
	ssoreadyv1 "github.com/ssoready/ssoready/internal/gen/ssoready/v1"
	"github.com/ssoready/ssoready/internal/saml"
)

func (s *Service) AppGetAdminSettings(ctx context.Context, req *connect.Request[ssoreadyv1.AppGetAdminSettingsRequest]) (*connect.Response[ssoreadyv1.AppGetAdminSettingsResponse], error) {
	res, err := s.Store.AppGetAdminSettings(ctx, req.Msg)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}

	var adminLogoURL string
	if res.AdminLogoConfigured {
		adminLogoURL, err = s.adminLogoURL(ctx, req.Msg.EnvironmentId)
		if err != nil {
			return nil, fmt.Errorf("admin logo url: %w", err)
		}
	}

	res.AppGetAdminSettingsResponse.AdminLogoUrl = adminLogoURL
	return connect.NewResponse(res.AppGetAdminSettingsResponse), nil
}

// note well: adminLogoURL does no authz checks
func (s *Service) adminLogoURL(ctx context.Context, environmentID string) (string, error) {
	_, err := s.S3Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: &s.AdminLogosS3BucketName,
		Key:    aws.String(fmt.Sprintf("v1/%s", environmentID)),
	})
	if err != nil {
		var apiError smithy.APIError
		if errors.As(err, &apiError) && apiError.ErrorCode() == "NotFound" {
			return "", nil
		}

		return "", fmt.Errorf("s3: %w", err)
	}

	presignRequest, err := s.S3PresignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: &s.AdminLogosS3BucketName,
		Key:    aws.String(fmt.Sprintf("v1/%s", environmentID)),
	}, s3.WithPresignExpires(time.Hour))
	if err != nil {
		return "", fmt.Errorf("s3: %w", err)
	}

	return presignRequest.URL, nil
}

func (s *Service) AppUpdateAdminSettings(ctx context.Context, req *connect.Request[ssoreadyv1.AppUpdateAdminSettingsRequest]) (*connect.Response[ssoreadyv1.AppUpdateAdminSettingsResponse], error) {
	res, err := s.Store.AppUpdateAdminSettings(ctx, req.Msg)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}

	return connect.NewResponse(res), nil
}

func (s *Service) AppUpdateAdminSettingsLogo(ctx context.Context, req *connect.Request[ssoreadyv1.AppUpdateAdminSettingsLogoRequest]) (*connect.Response[ssoreadyv1.AppUpdateAdminSettingsLogoResponse], error) {
	if _, err := s.Store.GetEnvironment(ctx, &ssoreadyv1.GetEnvironmentRequest{
		Id: req.Msg.EnvironmentId,
	}); err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}

	presignRequest, err := s.S3PresignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: &s.AdminLogosS3BucketName,
		Key:    aws.String(fmt.Sprintf("v1/%s", req.Msg.EnvironmentId)),
	}, s3.WithPresignExpires(24*time.Hour))
	if err != nil {
		return nil, fmt.Errorf("s3: %w", err)
	}

	return connect.NewResponse(&ssoreadyv1.AppUpdateAdminSettingsLogoResponse{
		UploadUrl: presignRequest.URL,
	}), nil
}

func (s *Service) AppCreateAdminSetupURL(ctx context.Context, req *connect.Request[ssoreadyv1.AppCreateAdminSetupURLRequest]) (*connect.Response[ssoreadyv1.AppCreateAdminSetupURLResponse], error) {
	res, err := s.Store.AppCreateAdminSetupURL(ctx, req.Msg)
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

func (s *Service) AdminWhoami(ctx context.Context, req *connect.Request[ssoreadyv1.AdminWhoamiRequest]) (*connect.Response[ssoreadyv1.AdminWhoamiResponse], error) {
	whoamiRes, err := s.Store.AdminWhoami(ctx, req.Msg)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}

	var adminLogoURL string
	if whoamiRes.AdminLogoConfigured {
		adminLogoURL, err = s.adminLogoURL(ctx, whoamiRes.EnvironmentID)
		if err != nil {
			return nil, fmt.Errorf("admin logo url: %w", err)
		}
	}

	return connect.NewResponse(&ssoreadyv1.AdminWhoamiResponse{
		CanManageSaml:        whoamiRes.CanManageSAML,
		CanManageScim:        whoamiRes.CanManageSCIM,
		AdminApplicationName: whoamiRes.AdminApplicationName,
		AdminReturnUrl:       whoamiRes.AdminReturnURL,
		AdminLogoUrl:         adminLogoURL,
	}), nil
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
	var metadata []byte
	if req.Msg.Url != "" {
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

		metadata = body
	} else if req.Msg.Xml != "" {
		metadata = []byte(req.Msg.Xml)
	}

	metadataRes, err := saml.ParseMetadata(metadata)
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

func (s *Service) AdminListSAMLFlows(ctx context.Context, req *connect.Request[ssoreadyv1.AdminListSAMLFlowsRequest]) (*connect.Response[ssoreadyv1.AdminListSAMLFlowsResponse], error) {
	res, err := s.Store.AdminListSAMLFlows(ctx, req.Msg)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}
	return connect.NewResponse(res), nil
}

func (s *Service) AdminGetSAMLFlow(ctx context.Context, req *connect.Request[ssoreadyv1.AdminGetSAMLFlowRequest]) (*connect.Response[ssoreadyv1.AdminGetSAMLFlowResponse], error) {
	res, err := s.Store.AdminGetSAMLFlow(ctx, req.Msg)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}
	return connect.NewResponse(res), nil
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
