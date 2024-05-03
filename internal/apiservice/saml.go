package apiservice

import (
	"context"

	"connectrpc.com/connect"
	ssoreadyv1 "github.com/ssoready/ssoready/internal/gen/ssoready/v1"
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
