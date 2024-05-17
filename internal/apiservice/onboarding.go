package apiservice

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	ssoreadyv1 "github.com/ssoready/ssoready/internal/gen/ssoready/v1"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Service) GetOnboardingState(ctx context.Context, req *connect.Request[ssoreadyv1.GetOnboardingStateRequest]) (*connect.Response[ssoreadyv1.GetOnboardingStateResponse], error) {
	res, err := s.Store.GetOnboardingState(ctx)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}

	return connect.NewResponse(res), nil
}

func (s *Service) UpdateOnboardingState(ctx context.Context, req *connect.Request[ssoreadyv1.UpdateOnboardingStateRequest]) (*connect.Response[emptypb.Empty], error) {
	res, err := s.Store.UpdateOnboardingState(ctx, req.Msg)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}

	return connect.NewResponse(res), nil
}

func (s *Service) OnboardingGetSAMLRedirectURL(ctx context.Context, req *connect.Request[ssoreadyv1.OnboardingGetSAMLRedirectURLRequest]) (*connect.Response[ssoreadyv1.GetSAMLRedirectURLResponse], error) {
	res, err := s.Store.OnboardingGetSAMLRedirectURL(ctx, req.Msg)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}

	return connect.NewResponse(res), nil
}

func (s *Service) OnboardingRedeemSAMLAccessCode(ctx context.Context, req *connect.Request[ssoreadyv1.OnboardingRedeemSAMLAccessCodeRequest]) (*connect.Response[ssoreadyv1.RedeemSAMLAccessCodeResponse], error) {
	res, err := s.Store.OnboardingRedeemSAMLAccessCode(ctx, req.Msg)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}

	return connect.NewResponse(res), nil
}
