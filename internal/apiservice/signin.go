package apiservice

import (
	"context"
	"fmt"
	"log/slog"

	"connectrpc.com/connect"
	"github.com/resend/resend-go/v2"
	ssoreadyv1 "github.com/ssoready/ssoready/internal/gen/ssoready/v1"
	"github.com/ssoready/ssoready/internal/google"
	"github.com/ssoready/ssoready/internal/store"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Service) VerifyEmail(ctx context.Context, req *connect.Request[ssoreadyv1.VerifyEmailRequest]) (*connect.Response[emptypb.Empty], error) {
	challengeRes, err := s.Store.CreateEmailVerificationChallenge(ctx, &store.CreateEmailVerificationChallengeRequest{
		Email: req.Msg.Email,
	})
	if err != nil {
		return nil, err
	}

	if _, err := s.ResendClient.Emails.SendWithContext(ctx, &resend.SendEmailRequest{
		From:    s.EmailChallengeFrom,
		To:      []string{req.Msg.Email},
		Subject: "SSOReady - Verify your email",
		Text:    fmt.Sprintf("Hi,\n\nPlease verify your email address with SSOReady by clicking the link below:\n\n%s?t=%s\n\nThanks,\nSSOReady", s.EmailVerificationEndpoint, challengeRes.SecretToken),
	}); err != nil {
		return nil, err
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (s *Service) SignIn(ctx context.Context, req *connect.Request[ssoreadyv1.SignInRequest]) (*connect.Response[ssoreadyv1.SignInResponse], error) {
	slog.InfoContext(ctx, "sign_in", "google_credential", req.Msg.GoogleCredential, "email_verify_token", req.Msg.EmailVerifyToken)

	if req.Msg.GoogleCredential != "" {
		credRes, err := s.GoogleClient.ParseCredential(ctx, &google.ParseCredentialRequest{
			Credential: req.Msg.GoogleCredential,
		})
		slog.InfoContext(ctx, "parse_credential", "res", credRes, "err", err)

		if err != nil {
			return nil, fmt.Errorf("google: parse credential: %w", err)
		}

		createSessionRes, err := s.Store.CreateGoogleSession(ctx, &store.CreateGoogleSessionRequest{
			Email:        credRes.Email,
			HostedDomain: credRes.HostedDomain,
			DisplayName:  credRes.Name,
		})

		if err != nil {
			return nil, fmt.Errorf("store: %w", err)
		}

		return connect.NewResponse(&ssoreadyv1.SignInResponse{
			SessionToken: createSessionRes.SessionToken,
		}), nil
	}

	verifyRes, err := s.Store.VerifyEmail(ctx, &store.VerifyEmailRequest{
		Token: req.Msg.EmailVerifyToken,
	})
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}

	return connect.NewResponse(&ssoreadyv1.SignInResponse{
		SessionToken: verifyRes.SessionToken,
	}), nil
}
