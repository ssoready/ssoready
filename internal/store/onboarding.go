package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/ssoready/ssoready/internal/apikeyauth"
	"github.com/ssoready/ssoready/internal/appauth"
	ssoreadyv1 "github.com/ssoready/ssoready/internal/gen/ssoready/v1"
	"github.com/ssoready/ssoready/internal/store/idformat"
	"github.com/ssoready/ssoready/internal/store/queries"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Store) GetOnboardingState(ctx context.Context) (*ssoreadyv1.GetOnboardingStateResponse, error) {
	qOnboardingState, err := s.q.GetOnboardingState(ctx, appauth.OrgID(ctx))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &ssoreadyv1.GetOnboardingStateResponse{}, nil
		}

		return nil, err
	}

	return &ssoreadyv1.GetOnboardingStateResponse{
		DummyidpAppId:              qOnboardingState.DummyidpAppID,
		OnboardingEnvironmentId:    idformat.Environment.Format(qOnboardingState.OnboardingEnvironmentID),
		OnboardingOrganizationId:   idformat.Organization.Format(qOnboardingState.OnboardingOrganizationID),
		OnboardingSamlConnectionId: idformat.SAMLConnection.Format(qOnboardingState.OnboardingSamlConnectionID),
	}, nil
}

func (s *Store) UpdateOnboardingState(ctx context.Context, req *ssoreadyv1.UpdateOnboardingStateRequest) (*emptypb.Empty, error) {
	environmentID, err := idformat.Environment.Parse(req.OnboardingEnvironmentId)
	if err != nil {
		return nil, err
	}

	organizationID, err := idformat.Organization.Parse(req.OnboardingOrganizationId)
	if err != nil {
		return nil, err
	}

	samlConnectionID, err := idformat.SAMLConnection.Parse(req.OnboardingSamlConnectionId)
	if err != nil {
		return nil, err
	}

	if _, err := s.q.UpdateOnboardingState(ctx, queries.UpdateOnboardingStateParams{
		AppOrganizationID:          appauth.OrgID(ctx),
		DummyidpAppID:              req.DummyidpAppId,
		OnboardingEnvironmentID:    environmentID,
		OnboardingOrganizationID:   organizationID,
		OnboardingSamlConnectionID: samlConnectionID,
	}); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *Store) OnboardingGetSAMLRedirectURL(ctx context.Context, req *ssoreadyv1.OnboardingGetSAMLRedirectURLRequest) (*ssoreadyv1.GetSAMLRedirectURLResponse, error) {
	apiKey, err := s.q.GetAPIKeyBySecretValue(ctx, req.ApiKeySecretToken)
	if err != nil {
		return nil, err
	}

	if apiKey.AppOrganizationID != appauth.OrgID(ctx) {
		panic(fmt.Errorf("mismatch between apiKey.AppOrganizationID and appauth.OrgID: %v, %v", apiKey.AppOrganizationID, appauth.OrgID(ctx)))
	}

	ctx = apikeyauth.WithAPIKey(ctx, apiKey.AppOrganizationID, apiKey.EnvironmentID)

	res, err := s.GetSAMLRedirectURL(ctx, &ssoreadyv1.GetSAMLRedirectURLRequest{SamlConnectionId: req.SamlConnectionId})
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *Store) OnboardingRedeemSAMLAccessCode(ctx context.Context, req *ssoreadyv1.OnboardingRedeemSAMLAccessCodeRequest) (*ssoreadyv1.RedeemSAMLAccessCodeResponse, error) {
	apiKey, err := s.q.GetAPIKeyBySecretValue(ctx, req.ApiKeySecretToken)
	if err != nil {
		return nil, err
	}

	if apiKey.AppOrganizationID != appauth.OrgID(ctx) {
		panic(fmt.Errorf("mismatch between apiKey.AppOrganizationID and appauth.OrgID: %v, %v", apiKey.AppOrganizationID, appauth.OrgID(ctx)))
	}

	ctx = apikeyauth.WithAPIKey(ctx, apiKey.AppOrganizationID, apiKey.EnvironmentID)

	res, err := s.RedeemSAMLAccessCode(ctx, &ssoreadyv1.RedeemSAMLAccessCodeRequest{SamlAccessCode: req.SamlAccessCode})
	if err != nil {
		return nil, err
	}

	return res, nil
}
