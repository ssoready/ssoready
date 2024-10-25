package store

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/ssoready/ssoready/internal/authn"
	ssoreadyv1 "github.com/ssoready/ssoready/internal/gen/ssoready/v1"
	"github.com/ssoready/ssoready/internal/store/idformat"
	"github.com/ssoready/ssoready/internal/store/queries"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Store) AppListSCIMDirectories(ctx context.Context, req *ssoreadyv1.AppListSCIMDirectoriesRequest) (*ssoreadyv1.AppListSCIMDirectoriesResponse, error) {
	_, q, _, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	orgID, err := idformat.Organization.Parse(req.OrganizationId)
	if err != nil {
		return nil, err
	}

	// idor check
	if _, err = q.GetOrganization(ctx, queries.GetOrganizationParams{
		AppOrganizationID: authn.AppOrgID(ctx),
		ID:                orgID,
	}); err != nil {
		return nil, err
	}

	var startID uuid.UUID
	if err := s.pageEncoder.Unmarshal(req.PageToken, &startID); err != nil {
		return nil, err
	}

	limit := 10
	qSCIMDirectories, err := q.ListSCIMDirectories(ctx, queries.ListSCIMDirectoriesParams{
		OrganizationID: orgID,
		ID:             startID,
		Limit:          int32(limit + 1),
	})
	if err != nil {
		return nil, err
	}

	var scimDirectories []*ssoreadyv1.SCIMDirectory
	for _, qSCIMDirectory := range qSCIMDirectories {
		scimDirectories = append(scimDirectories, parseSCIMDirectory(qSCIMDirectory))
	}

	var nextPageToken string
	if len(scimDirectories) == limit+1 {
		nextPageToken = s.pageEncoder.Marshal(qSCIMDirectories[limit].ID)
		scimDirectories = scimDirectories[:limit]
	}

	return &ssoreadyv1.AppListSCIMDirectoriesResponse{
		ScimDirectories: scimDirectories,
		NextPageToken:   nextPageToken,
	}, nil
}

func (s *Store) AppGetSCIMDirectory(ctx context.Context, req *ssoreadyv1.AppGetSCIMDirectoryRequest) (*ssoreadyv1.SCIMDirectory, error) {
	_, q, _, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	scimDirID, err := idformat.SCIMDirectory.Parse(req.Id)
	if err != nil {
		return nil, fmt.Errorf("parse scim directory id: %w", err)
	}

	qSCIMDir, err := q.GetSCIMDirectory(ctx, queries.GetSCIMDirectoryParams{
		AppOrganizationID: authn.AppOrgID(ctx),
		ID:                scimDirID,
	})
	if err != nil {
		return nil, fmt.Errorf("get scim directory: %w", err)
	}

	return parseSCIMDirectory(qSCIMDir), nil
}

func (s *Store) AppCreateSCIMDirectory(ctx context.Context, req *ssoreadyv1.AppCreateSCIMDirectoryRequest) (*ssoreadyv1.SCIMDirectory, error) {
	_, q, commit, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	orgID, err := idformat.Organization.Parse(req.ScimDirectory.OrganizationId)
	if err != nil {
		return nil, fmt.Errorf("parse organization id: %w", err)
	}

	// idor check
	org, err := q.GetOrganization(ctx, queries.GetOrganizationParams{
		AppOrganizationID: authn.AppOrgID(ctx),
		ID:                orgID,
	})
	if err != nil {
		return nil, err
	}

	env, err := q.GetEnvironment(ctx, queries.GetEnvironmentParams{
		AppOrganizationID: authn.AppOrgID(ctx),
		ID:                org.EnvironmentID,
	})
	if err != nil {
		return nil, err
	}

	authURL := s.defaultAuthURL
	if env.AuthUrl != nil {
		authURL = *env.AuthUrl
	}

	id := uuid.New()
	scimBaseURL := fmt.Sprintf("%s/v1/scim/%s", authURL, idformat.SCIMDirectory.Format(id))
	qSCIMDirectory, err := q.CreateSCIMDirectory(ctx, queries.CreateSCIMDirectoryParams{
		ID:             id,
		OrganizationID: orgID,
		IsPrimary:      req.ScimDirectory.Primary,
		ScimBaseUrl:    scimBaseURL,
	})
	if err != nil {
		return nil, fmt.Errorf("create scim directory: %w", err)
	}

	if qSCIMDirectory.IsPrimary {
		if err := q.UpdatePrimarySCIMDirectory(ctx, queries.UpdatePrimarySCIMDirectoryParams{
			ID:             qSCIMDirectory.ID,
			OrganizationID: qSCIMDirectory.OrganizationID,
		}); err != nil {
			return nil, fmt.Errorf("update primary scim directory: %w", err)
		}
	}

	if err := commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return parseSCIMDirectory(qSCIMDirectory), nil
}

func (s *Store) AppUpdateSCIMDirectory(ctx context.Context, req *ssoreadyv1.AppUpdateSCIMDirectoryRequest) (*ssoreadyv1.SCIMDirectory, error) {
	_, q, commit, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	scimDirID, err := idformat.SCIMDirectory.Parse(req.ScimDirectory.Id)
	if err != nil {
		return nil, fmt.Errorf("parse scim directory id: %w", err)
	}

	if _, err := q.GetSCIMDirectory(ctx, queries.GetSCIMDirectoryParams{
		AppOrganizationID: authn.AppOrgID(ctx),
		ID:                scimDirID,
	}); err != nil {
		return nil, fmt.Errorf("get scim directory: %w", err)
	}

	qSCIMDir, err := q.UpdateSCIMDirectory(ctx, queries.UpdateSCIMDirectoryParams{
		ID:        scimDirID,
		IsPrimary: req.ScimDirectory.Primary,
	})
	if err != nil {
		return nil, fmt.Errorf("update scim directory: %w", err)
	}

	if qSCIMDir.IsPrimary {
		if err := q.UpdatePrimarySCIMDirectory(ctx, queries.UpdatePrimarySCIMDirectoryParams{
			ID:             qSCIMDir.ID,
			OrganizationID: qSCIMDir.OrganizationID,
		}); err != nil {
			return nil, fmt.Errorf("update primary scim directory: %w", err)
		}
	}

	if err := commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return parseSCIMDirectory(qSCIMDir), nil
}

func (s *Store) AppRotateSCIMDirectoryBearerToken(ctx context.Context, req *ssoreadyv1.AppRotateSCIMDirectoryBearerTokenRequest) (*ssoreadyv1.AppRotateSCIMDirectoryBearerTokenResponse, error) {
	_, q, commit, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	scimDirID, err := idformat.SCIMDirectory.Parse(req.ScimDirectoryId)
	if err != nil {
		return nil, fmt.Errorf("parse scim directory id: %w", err)
	}

	if _, err := q.GetSCIMDirectory(ctx, queries.GetSCIMDirectoryParams{
		AppOrganizationID: authn.AppOrgID(ctx),
		ID:                scimDirID,
	}); err != nil {
		return nil, fmt.Errorf("get scim directory: %w", err)
	}

	bearerToken := uuid.New()
	bearerTokenSHA := sha256.Sum256(bearerToken[:])

	if _, err := q.UpdateSCIMDirectoryBearerToken(ctx, queries.UpdateSCIMDirectoryBearerTokenParams{
		BearerTokenSha256: bearerTokenSHA[:],
		ID:                scimDirID,
	}); err != nil {
		return nil, fmt.Errorf("update scim directory access token: %w", err)
	}

	if err := commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return &ssoreadyv1.AppRotateSCIMDirectoryBearerTokenResponse{
		BearerToken: idformat.SCIMBearerToken.Format(bearerToken),
	}, nil
}

func (s *Store) AppDeleteSCIMDirectory(ctx context.Context, req *ssoreadyv1.AppDeleteSCIMDirectoryRequest) (*emptypb.Empty, error) {
	_, q, commit, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollback()

	slog.InfoContext(ctx, "delete_scim_directory", "scim_directory_id", req.ScimDirectoryId)

	scimDirID, err := idformat.SCIMDirectory.Parse(req.ScimDirectoryId)
	if err != nil {
		return nil, fmt.Errorf("parse scim directory id: %w", err)
	}

	if _, err := q.GetSCIMDirectory(ctx, queries.GetSCIMDirectoryParams{
		AppOrganizationID: authn.AppOrgID(ctx),
		ID:                scimDirID,
	}); err != nil {
		return nil, fmt.Errorf("get scim directory: %w", err)
	}

	scimUserGroupMembershipsCount, err := q.DeleteSCIMUserGroupMembershipsBySCIMDirectory(ctx, scimDirID)
	if err != nil {
		return nil, fmt.Errorf("delete user group memberships: %w", err)
	}

	slog.InfoContext(ctx, "delete_scim_directory", "scim_user_group_memberships_count", scimUserGroupMembershipsCount)

	scimGroupsCount, err := q.DeleteSCIMGroupsBySCIMDirectory(ctx, scimDirID)
	if err != nil {
		return nil, fmt.Errorf("delete groups: %w", err)
	}

	slog.InfoContext(ctx, "delete_scim_directory", "scim_groups_count", scimGroupsCount)

	scimUsersCount, err := q.DeleteSCIMUsersBySCIMDirectory(ctx, scimDirID)
	if err != nil {
		return nil, fmt.Errorf("delete users: %w", err)
	}

	slog.InfoContext(ctx, "delete_scim_directory", "scim_user_count", scimUsersCount)

	scimRequestsCount, err := q.DeleteSCIMRequestsBySCIMDirectory(ctx, scimDirID)
	if err != nil {
		return nil, fmt.Errorf("delete scim requests: %w", err)
	}

	slog.InfoContext(ctx, "delete_scim_directory", "scim_request_count", scimRequestsCount)

	scimDirectoriesCount, err := q.DeleteSCIMDirectory(ctx, scimDirID)
	if err != nil {
		return nil, fmt.Errorf("delete scim directory: %w", err)
	}

	slog.InfoContext(ctx, "delete_scim_directory", "scim_directories_count", scimDirectoriesCount)

	if err := commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return &emptypb.Empty{}, nil
}

func parseSCIMDirectory(qSCIMDirectory queries.ScimDirectory) *ssoreadyv1.SCIMDirectory {
	return &ssoreadyv1.SCIMDirectory{
		Id:                   idformat.SCIMDirectory.Format(qSCIMDirectory.ID),
		OrganizationId:       idformat.Organization.Format(qSCIMDirectory.OrganizationID),
		Primary:              qSCIMDirectory.IsPrimary,
		ScimBaseUrl:          qSCIMDirectory.ScimBaseUrl,
		HasClientBearerToken: len(qSCIMDirectory.BearerTokenSha256) > 0,
	}
}
