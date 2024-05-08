package store

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/ssoready/ssoready/internal/store/idformat"
	"github.com/ssoready/ssoready/internal/store/queries"
)

type GetAppSessionRequest struct {
	SessionToken string
	Now          time.Time
}

type GetAppSessionResponse struct {
	AppUserID         string
	AppOrganizationID uuid.UUID
}

func (s *Store) GetAppSession(ctx context.Context, req *GetAppSessionRequest) (*GetAppSessionResponse, error) {
	appSession, err := s.q.GetAppSessionByToken(ctx, queries.GetAppSessionByTokenParams{
		Token:      req.SessionToken,
		ExpireTime: req.Now,
	})
	if err != nil {
		return nil, err
	}

	return &GetAppSessionResponse{
		AppUserID:         idformat.AppUser.Format(appSession.AppUserID),
		AppOrganizationID: appSession.AppOrganizationID,
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

	// generate token as 32-byte random string, hex-encoded
	var tokenBytes [32]byte
	if _, err := rand.Read(tokenBytes[:]); err != nil {
		return nil, err
	}
	token := hex.EncodeToString(tokenBytes[:])

	appSession, err := q.CreateAppSession(ctx, queries.CreateAppSessionParams{
		ID:         uuid.New(),
		AppUserID:  appUser.ID,
		CreateTime: time.Now(),
		ExpireTime: time.Now().Add(time.Hour * 24 * 7),
		Token:      token,
	})
	if err != nil {
		return nil, err
	}

	if err := commit(); err != nil {
		return nil, err
	}

	return &CreateGoogleSessionResponse{SessionToken: appSession.Token}, nil
}

func (s *Store) upsertGoogleAppUser(ctx context.Context, q *queries.Queries, req *CreateGoogleSessionRequest) (*queries.AppUser, error) {
	appUser, err := q.GetAppUserByEmail(ctx, &req.Email)
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
				Email:             &req.Email,
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
	// generate token as 32-byte random string, hex-encoded
	var tokenBytes [32]byte
	if _, err := rand.Read(tokenBytes[:]); err != nil {
		return nil, err
	}
	token := hex.EncodeToString(tokenBytes[:])

	qChallenge, err := s.q.CreateEmailVerificationChallenge(ctx, queries.CreateEmailVerificationChallengeParams{
		ID:          uuid.New(),
		Email:       req.Email,
		ExpireTime:  time.Now().Add(24 * time.Hour),
		SecretToken: token,
	})
	if err != nil {
		return nil, err
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

	// generate token as 32-byte random string, hex-encoded
	var tokenBytes [32]byte
	if _, err := rand.Read(tokenBytes[:]); err != nil {
		return nil, err
	}
	token := hex.EncodeToString(tokenBytes[:])

	appSession, err := q.CreateAppSession(ctx, queries.CreateAppSessionParams{
		ID:         uuid.New(),
		AppUserID:  appUser.ID,
		CreateTime: time.Now(),
		ExpireTime: time.Now().Add(time.Hour * 24 * 7),
		Token:      token,
	})
	if err != nil {
		return nil, err
	}

	if err := commit(); err != nil {
		return nil, err
	}

	return &VerifyEmailResponse{SessionToken: appSession.Token}, nil
}

func (s *Store) upsertUserByEmailSoleInOrg(ctx context.Context, q *queries.Queries, email string) (*queries.AppUser, error) {
	appUser, err := q.GetAppUserByEmail(ctx, &email)
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
				Email:             &email,
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
