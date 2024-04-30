package store

import (
	"context"
	"fmt"

	"github.com/ssoready/ssoready/internal/appauth"
	ssoreadyv1 "github.com/ssoready/ssoready/internal/gen/ssoready/v1"
	"github.com/ssoready/ssoready/internal/store/idformat"
	"github.com/ssoready/ssoready/internal/store/queries"
)

func (s *Store) Whoami(ctx context.Context, req *ssoreadyv1.WhoamiRequest) (*ssoreadyv1.WhoamiResponse, error) {
	userID, err := idformat.AppUser.Parse(appauth.AppUserID(ctx))
	if err != nil {
		return nil, err
	}

	appUser, err := s.q.GetAppUserByID(ctx, queries.GetAppUserByIDParams{
		AppOrganizationID: appauth.OrgID(ctx),
		ID:                userID,
	})
	if err != nil {
		return nil, fmt.Errorf("get app user by id: %w", err)
	}

	return &ssoreadyv1.WhoamiResponse{
		AppUserId:   idformat.AppUser.Format(appUser.ID),
		DisplayName: appUser.DisplayName,
		Email:       *appUser.Email,
	}, nil
}
