package store

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/ssoready/ssoready/internal/authn"
	"github.com/ssoready/ssoready/internal/store/idformat"
	"github.com/ssoready/ssoready/internal/store/queries"
)

type GetAppSessionRequest struct {
	SessionToken string
	Now          time.Time
}

type GetAppSessionResponse struct {
	AppSessionID       string
	AppUserID          string
	AppOrganizationID  uuid.UUID
	AppUserDisplayName string
	AppUserEmail       string
}

func (s *Store) GetAppSession(ctx context.Context, req *GetAppSessionRequest) (*GetAppSessionResponse, error) {
	token, err := hex.DecodeString(req.SessionToken)
	if err != nil {
		return nil, fmt.Errorf("parse session token: %w", err)
	}

	sha := sha256.Sum256(token)
	appSession, err := s.q.GetAppSessionByTokenSHA256(ctx, queries.GetAppSessionByTokenSHA256Params{
		TokenSha256: sha[:],
		ExpireTime:  req.Now,
	})
	if err != nil {
		return nil, err
	}

	return &GetAppSessionResponse{
		AppSessionID:       idformat.AppSession.Format(appSession.ID),
		AppUserID:          idformat.AppUser.Format(appSession.AppUserID),
		AppOrganizationID:  appSession.AppOrganizationID,
		AppUserDisplayName: appSession.DisplayName,
		AppUserEmail:       appSession.Email,
	}, nil
}

type CreateGoogleSessionRequest struct {
	Email        string
	DisplayName  string
	HostedDomain string
}

type CreateGoogleSessionResponse struct {
	SessionToken string
}

func (s *Store) CreateGoogleSession(ctx context.Context, req *CreateGoogleSessionRequest) (*CreateGoogleSessionResponse, error) {
	_, q, commit, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	appUser, err := s.upsertGoogleAppUser(ctx, q, req)
	if err != nil {
		return nil, err
	}

	// generate token as 32-byte random string
	var tokenBytes [32]byte
	if _, err := rand.Read(tokenBytes[:]); err != nil {
		return nil, err
	}
	tokenHex := hex.EncodeToString(tokenBytes[:])
	tokenSHA := sha256.Sum256(tokenBytes[:])

	revoked := false
	if _, err := q.CreateAppSession(ctx, queries.CreateAppSessionParams{
		ID:          uuid.New(),
		AppUserID:   appUser.ID,
		CreateTime:  time.Now(),
		ExpireTime:  time.Now().Add(time.Hour * 24 * 7),
		TokenSha256: tokenSHA[:],
		Revoked:     &revoked,
	}); err != nil {
		return nil, err
	}

	if err := commit(); err != nil {
		return nil, err
	}

	return &CreateGoogleSessionResponse{SessionToken: tokenHex}, nil
}

func (s *Store) upsertGoogleAppUser(ctx context.Context, q *queries.Queries, req *CreateGoogleSessionRequest) (*queries.AppUser, error) {
	appUser, err := q.GetAppUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			appOrg, err := s.upsertGoogleAppOrg(ctx, q, req)
			if err != nil {
				return nil, err
			}

			appUser, err := q.CreateAppUser(ctx, queries.CreateAppUserParams{
				ID:                uuid.New(),
				AppOrganizationID: appOrg.ID,
				DisplayName:       req.DisplayName,
				Email:             req.Email,
			})
			if err != nil {
				return nil, err
			}

			return &appUser, nil
		}

		return nil, err
	}

	return &appUser, nil
}

func (s *Store) upsertGoogleAppOrg(ctx context.Context, q *queries.Queries, req *CreateGoogleSessionRequest) (*queries.AppOrganization, error) {
	if req.HostedDomain == "" {
		// this is a personal address; give them their own app org
		appOrg, err := q.CreateAppOrganization(ctx, queries.CreateAppOrganizationParams{
			ID: uuid.New(),
		})
		if err != nil {
			return nil, err
		}

		return &appOrg, nil
	}

	appOrg, err := q.GetAppOrganizationByGoogleHostedDomain(ctx, &req.HostedDomain)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			appOrg, err := q.CreateAppOrganization(ctx, queries.CreateAppOrganizationParams{
				ID:                 uuid.New(),
				GoogleHostedDomain: &req.HostedDomain,
			})
			if err != nil {
				return nil, err
			}

			return &appOrg, nil
		}

		return nil, err
	}

	return &appOrg, nil
}

type CreateEmailVerificationChallengeRequest struct {
	Email string
}

type CreateEmailVerificationChallengeResponse struct {
	SecretToken string
}

func (s *Store) CreateEmailVerificationChallenge(ctx context.Context, req *CreateEmailVerificationChallengeRequest) (*CreateEmailVerificationChallengeResponse, error) {
	_, q, commit, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("tx: %w", err)
	}
	defer rollback()

	alreadyExists, err := q.CheckExistsEmailVerificationChallenge(ctx, queries.CheckExistsEmailVerificationChallengeParams{
		Email:      req.Email,
		ExpireTime: time.Now(),
	})
	if err != nil {
		return nil, fmt.Errorf("check exists email verification challenge: %w", err)
	}

	if alreadyExists {
		return nil, fmt.Errorf("outstanding email verification challenge already exists")
	}

	// generate token as 32-byte random string, hex-encoded
	var tokenBytes [32]byte
	if _, err := rand.Read(tokenBytes[:]); err != nil {
		return nil, err
	}
	token := hex.EncodeToString(tokenBytes[:])

	qChallenge, err := q.CreateEmailVerificationChallenge(ctx, queries.CreateEmailVerificationChallengeParams{
		ID:          uuid.New(),
		Email:       req.Email,
		ExpireTime:  time.Now().Add(24 * time.Hour),
		SecretToken: token,
	})
	if err != nil {
		return nil, err
	}

	if err := commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return &CreateEmailVerificationChallengeResponse{SecretToken: qChallenge.SecretToken}, nil
}

type VerifyEmailRequest struct {
	Token string
}

type VerifyEmailResponse struct {
	SessionToken string
}

func (s *Store) VerifyEmail(ctx context.Context, req *VerifyEmailRequest) (*VerifyEmailResponse, error) {
	_, q, commit, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	qChallenge, err := q.GetEmailVerificationChallengeBySecretToken(ctx, queries.GetEmailVerificationChallengeBySecretTokenParams{
		SecretToken: req.Token,
		ExpireTime:  time.Now(),
	})
	if err != nil {
		return nil, err
	}

	appUser, err := s.upsertUserByEmailSoleInOrg(ctx, q, qChallenge.Email)
	if err != nil {
		return nil, err
	}

	// generate token as 32-byte random string
	var tokenBytes [32]byte
	if _, err := rand.Read(tokenBytes[:]); err != nil {
		return nil, err
	}
	tokenHex := hex.EncodeToString(tokenBytes[:])
	tokenSHA := sha256.Sum256(tokenBytes[:])

	revoked := false
	if _, err := q.CreateAppSession(ctx, queries.CreateAppSessionParams{
		ID:          uuid.New(),
		AppUserID:   appUser.ID,
		CreateTime:  time.Now(),
		ExpireTime:  time.Now().Add(time.Hour * 24 * 7),
		TokenSha256: tokenSHA[:],
		Revoked:     &revoked,
	}); err != nil {
		return nil, err
	}

	if err := commit(); err != nil {
		return nil, err
	}

	return &VerifyEmailResponse{SessionToken: tokenHex}, nil
}

func (s *Store) upsertUserByEmailSoleInOrg(ctx context.Context, q *queries.Queries, email string) (*queries.AppUser, error) {
	appUser, err := q.GetAppUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			appOrg, err := q.CreateAppOrganization(ctx, queries.CreateAppOrganizationParams{
				ID: uuid.New(),
			})
			if err != nil {
				return nil, err
			}

			appUser, err := q.CreateAppUser(ctx, queries.CreateAppUserParams{
				ID:                uuid.New(),
				AppOrganizationID: appOrg.ID,
				Email:             email,
			})
			if err != nil {
				return nil, err
			}

			return &appUser, nil
		}

		return nil, err
	}

	return &appUser, nil
}

func (s *Store) RevokeSession(ctx context.Context) error {
	authnData := authn.FullContextData(ctx)
	if authnData.AppSession == nil {
		panic("RevokeSession called on session with no AppSession")
	}

	id, err := idformat.AppSession.Parse(authnData.AppSession.AppSessionID)
	if err != nil {
		return fmt.Errorf("parse app session id: %w", err)
	}

	if _, err := s.q.RevokeAppSessionByID(ctx, id); err != nil {
		return fmt.Errorf("revoke app session by id: %w", err)
	}

	return nil
}
