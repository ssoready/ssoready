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

	//tx, err := s.DB.BeginTx(ctx, nil)
	//if err != nil {
	//	return nil, fmt.Errorf("begin: %w", err)
	//}
	//defer tx.Rollback()
	//
	//q := queries.New(tx)

	//users, err := q.GetAppUserByGoogleEmail(ctx, &req.Email)
	//if err != nil {
	//	return nil, fmt.Errorf("db: %w", err)
	//}
	//
	//var userID uuid.UUID
	//
	//if len(users) == 0 {
	//	// no user with this email, create one
	//	org, err := q.CreateAppOrganization(ctx, uuid.New())
	//	if err != nil {
	//		return nil, fmt.Errorf("db: %w", err)
	//	}
	//
	//	// create a dev env
	//	if _, err := q.CreateEnvironment(ctx, queries.CreateEnvironmentParams{
	//		ID:                uuid.New(),
	//		AppOrganizationID: org.ID,
	//		RedirectUrl:       "http://localhost:8080",
	//	}); err != nil {
	//		return nil, err
	//	}
	//
	//	user, err := q.CreateAppUser(ctx, queries.CreateAppUserParams{
	//		ID:                uuid.New(),
	//		AppOrganizationID: org.ID,
	//		Email:       &req.Email,
	//	})
	//
	//	if err != nil {
	//		return nil, fmt.Errorf("db: %w", err)
	//	}
	//
	//	userID = user.ID
	//} else {
	//	userID = users[0].ID
	//}
	//
	//sess, err := q.CreateAppSession(ctx, queries.CreateAppSessionParams{
	//	ID:         uuid.New(),
	//	AppUserID:  userID,
	//	ExpireTime: time.Now().Add(7 * 24 * time.Hour),
	//})
	//
	//if err != nil {
	//	return nil, fmt.Errorf("db: %w", err)
	//}
	//
	//if err := tx.Commit(); err != nil {
	//	return nil, fmt.Errorf("db: %w", err)
	//}
	//
	//return &sess.ID, nil
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
