package apiservice

import (
	"context"
	"encoding/pem"
	"io"
	"net/http"

	"connectrpc.com/connect"
	ssoreadyv1 "github.com/ssoready/ssoready/internal/gen/ssoready/v1"
	"github.com/ssoready/ssoready/internal/saml"
)

func (s *Service) GetSAMLRedirectURL(ctx context.Context, req *connect.Request[ssoreadyv1.GetSAMLRedirectURLRequest]) (*connect.Response[ssoreadyv1.GetSAMLRedirectURLResponse], error) {
	res, err := s.Store.GetSAMLRedirectURL(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(res), nil
}

func (s *Service) RedeemSAMLAccessCode(ctx context.Context, req *connect.Request[ssoreadyv1.RedeemSAMLAccessCodeRequest]) (*connect.Response[ssoreadyv1.RedeemSAMLAccessCodeResponse], error) {
	res, err := s.Store.RedeemSAMLAccessCode(ctx, req.Msg)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(res), nil
}

func (s *Service) ParseSAMLMetadata(ctx context.Context, req *connect.Request[ssoreadyv1.ParseSAMLMetadataRequest]) (*connect.Response[ssoreadyv1.ParseSAMLMetadataResponse], error) {
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

	return connect.NewResponse(&ssoreadyv1.ParseSAMLMetadataResponse{
		IdpRedirectUrl: metadataRes.RedirectURL,
		IdpEntityId:    metadataRes.IDPEntityID,
		IdpCertificate: string(pem.EncodeToMemory(&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: metadataRes.IDPCertificate.Raw,
		})),
	}), nil
}
