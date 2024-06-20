package store

import (
	"context"
	"fmt"

	"github.com/ssoready/ssoready/internal/authn"
	ssoreadyv1 "github.com/ssoready/ssoready/internal/gen/ssoready/v1"
	"github.com/ssoready/ssoready/internal/store/idformat"
	"github.com/ssoready/ssoready/internal/store/queries"
)

func (s *Store) Whoami(ctx context.Context, req *ssoreadyv1.WhoamiRequest) (*ssoreadyv1.WhoamiResponse, error) {
	authnData := authn.FullContextData(ctx)
	if authnData.AppSession == nil {
		return nil, fmt.Errorf("whoami must only be called when authenticated over an app session")
	}

	userID, err := idformat.AppUser.Parse(authnData.AppSession.AppUserID)
	if err != nil {
		return nil, err
	}

	appUser, err := s.q.GetAppUserByID(ctx, queries.GetAppUserByIDParams{
		AppOrganizationID: authn.AppOrgID(ctx),
		ID:                userID,
	})
	if err != nil {
		return nil, fmt.Errorf("get app user by id: %w", err)
	}

	return &ssoreadyv1.WhoamiResponse{
		AppUserId:   idformat.AppUser.Format(appUser.ID),
		DisplayName: appUser.DisplayName,
		Email:       appUser.Email,
	}, nil
}
