package store

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	ssoreadyv1 "github.com/ssoready/ssoready/internal/gen/ssoready/v1"
	"github.com/ssoready/ssoready/internal/store/idformat"
	"github.com/ssoready/ssoready/internal/store/queries"
	"google.golang.org/protobuf/types/known/structpb"
)

var ErrAuthSCIMBadBearerToken = errors.New("store: bad scim directory bearer token")

func (s *Store) AuthSCIMVerifyBearerToken(ctx context.Context, scimDirectoryID, bearerToken string) error {
	scimDirID, err := idformat.SCIMDirectory.Parse(scimDirectoryID)
	if err != nil {
		return fmt.Errorf("parse scim directory id: %w", err)
	}

	bearerTokenID, err := idformat.SCIMBearerToken.Parse(bearerToken)
	if err != nil {
		return fmt.Errorf("parse bearer token: %w", err)
	}

	fmt.Println("bearer token id", uuid.UUID(bearerTokenID))

	bearerTokenSHA := sha256.Sum256(bearerTokenID[:])

	fmt.Println("check sha", uuid.UUID(scimDirID), hex.EncodeToString(bearerTokenSHA[:]))

	if _, err := s.q.AuthGetSCIMDirectoryByIDAndBearerToken(ctx, queries.AuthGetSCIMDirectoryByIDAndBearerTokenParams{
		ID:                scimDirID,
		BearerTokenSha256: bearerTokenSHA[:],
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrAuthSCIMBadBearerToken
		}

		return fmt.Errorf("get scim directory: %w", err)
	}

	return nil
}

type AuthListSCIMUsersRequest struct {
	SCIMDirectoryID string
	StartIndex      int
}

type AuthListSCIMUsersResponse struct {
	TotalResults int
	SCIMUsers    []*ssoreadyv1.SCIMUser
}

func (s *Store) AuthListSCIMUsers(ctx context.Context, req *AuthListSCIMUsersRequest) (*AuthListSCIMUsersResponse, error) {
	scimDirID, err := idformat.SCIMDirectory.Parse(req.SCIMDirectoryID)
	if err != nil {
		return nil, fmt.Errorf("parse scim directory id: %w", err)
	}

	_, q, _, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("tx: %w", err)
	}
	defer rollback()

	count, err := q.AuthCountSCIMUsers(ctx, scimDirID)
	if err != nil {
		return nil, fmt.Errorf("count scim users: %w", err)
	}

	qSCIMUsers, err := q.AuthListSCIMUsers(ctx, queries.AuthListSCIMUsersParams{
		ScimDirectoryID: scimDirID,
		Offset:          int32(req.StartIndex),
		Limit:           10,
	})
	if err != nil {
		return nil, fmt.Errorf("list scim users: %w", err)
	}

	var scimUsers []*ssoreadyv1.SCIMUser
	for _, qSCIMUser := range qSCIMUsers {
		scimUsers = append(scimUsers, parseSCIMUser(qSCIMUser))
	}

	return &AuthListSCIMUsersResponse{
		TotalResults: int(count),
		SCIMUsers:    scimUsers,
	}, nil
}

type AuthGetSCIMUserByEmailRequest struct {
	SCIMDirectoryID string
	Email           string
}

var ErrSCIMUserNotFound = errors.New("store: scim user not found")

func (s *Store) AuthGetSCIMUserByEmail(ctx context.Context, req *AuthGetSCIMUserByEmailRequest) (*ssoreadyv1.SCIMUser, error) {
	scimDirID, err := idformat.SCIMDirectory.Parse(req.SCIMDirectoryID)
	if err != nil {
		return nil, fmt.Errorf("parse scim directory id: %w", err)
	}

	qSCIMUser, err := s.q.AuthGetSCIMUserByEmail(ctx, queries.AuthGetSCIMUserByEmailParams{
		ScimDirectoryID: scimDirID,
		Email:           req.Email,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSCIMUserNotFound
		}

		return nil, fmt.Errorf("get scim user by email: %w", err)
	}

	return parseSCIMUser(qSCIMUser), nil
}

type AuthGetSCIMUserRequest struct {
	SCIMDirectoryID string
	SCIMUserID      string
}

func (s *Store) AuthGetSCIMUser(ctx context.Context, req *AuthGetSCIMUserRequest) (*ssoreadyv1.SCIMUser, error) {
	scimDirID, err := idformat.SCIMDirectory.Parse(req.SCIMDirectoryID)
	if err != nil {
		return nil, fmt.Errorf("parse scim directory id: %w", err)
	}

	scimUserID, err := idformat.SCIMUser.Parse(req.SCIMUserID)
	if err != nil {
		return nil, fmt.Errorf("parse scim user id: %w", err)
	}

	qSCIMUser, err := s.q.AuthGetSCIMUser(ctx, queries.AuthGetSCIMUserParams{
		ScimDirectoryID: scimDirID,
		ID:              scimUserID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSCIMUserNotFound
		}

		return nil, fmt.Errorf("get scim user: %w", err)
	}

	return parseSCIMUser(qSCIMUser), nil
}

type AuthCreateSCIMUserRequest struct {
	SCIMUser *ssoreadyv1.SCIMUser
}

type AuthCreateSCIMUserResponse struct {
	SCIMUser *ssoreadyv1.SCIMUser
}

func (s *Store) AuthCreateSCIMUser(ctx context.Context, req *AuthCreateSCIMUserRequest) (*AuthCreateSCIMUserResponse, error) {
	_, q, commit, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("tx: %w", err)
	}
	defer rollback()

	scimDirID, err := idformat.SCIMDirectory.Parse(req.SCIMUser.ScimDirectoryId)
	if err != nil {
		return nil, fmt.Errorf("parse scim directory id: %w", err)
	}

	attrs, err := json.Marshal(req.SCIMUser.Attributes.AsMap())
	if err != nil {
		panic(fmt.Errorf("marshal scim user attributes: %w", err))
	}

	qSCIMUser, err := q.AuthCreateSCIMUser(ctx, queries.AuthCreateSCIMUserParams{
		ID:              uuid.New(),
		ScimDirectoryID: scimDirID,
		Email:           req.SCIMUser.Email,
		Deleted:         false,
		Attributes:      attrs,
	})
	if err != nil {
		return nil, fmt.Errorf("create scim user: %w", err)
	}

	if err := commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return &AuthCreateSCIMUserResponse{
		SCIMUser: parseSCIMUser(qSCIMUser),
	}, nil
}

type AuthUpdateSCIMUserRequest struct {
	SCIMUser *ssoreadyv1.SCIMUser
}

type AuthUpdateSCIMUserResponse struct {
	SCIMUser *ssoreadyv1.SCIMUser
}

func (s *Store) AuthUpdateSCIMUser(ctx context.Context, req *AuthUpdateSCIMUserRequest) (*AuthUpdateSCIMUserResponse, error) {
	_, q, commit, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("tx: %w", err)
	}
	defer rollback()

	scimDirID, err := idformat.SCIMDirectory.Parse(req.SCIMUser.ScimDirectoryId)
	if err != nil {
		return nil, fmt.Errorf("parse scim directory id: %w", err)
	}

	scimUserID, err := idformat.SCIMUser.Parse(req.SCIMUser.Id)
	if err != nil {
		return nil, fmt.Errorf("parse scim user id: %w", err)
	}

	attrs, err := json.Marshal(req.SCIMUser.Attributes.AsMap())
	if err != nil {
		panic(fmt.Errorf("marshal scim user attributes: %w", err))
	}

	qSCIMUser, err := q.AuthUpdateSCIMUser(ctx, queries.AuthUpdateSCIMUserParams{
		ID:              scimUserID,
		ScimDirectoryID: scimDirID,
		Email:           req.SCIMUser.Email,
		Deleted:         req.SCIMUser.Deleted,
		Attributes:      attrs,
	})
	if err != nil {
		return nil, fmt.Errorf("create scim user: %w", err)
	}

	if err := commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return &AuthUpdateSCIMUserResponse{
		SCIMUser: parseSCIMUser(qSCIMUser),
	}, nil
}

type AuthGetSCIMGroupRequest struct {
	SCIMDirectoryID string
	SCIMGroupID     string
}

var ErrSCIMGroupNotFound = errors.New("store: scim group not found")

func (s *Store) AuthGetSCIMGroup(ctx context.Context, req *AuthGetSCIMGroupRequest) (*ssoreadyv1.SCIMGroup, error) {
	scimDirID, err := idformat.SCIMDirectory.Parse(req.SCIMDirectoryID)
	if err != nil {
		return nil, fmt.Errorf("parse scim directory id: %w", err)
	}

	scimGroupID, err := idformat.SCIMGroup.Parse(req.SCIMGroupID)
	if err != nil {
		return nil, fmt.Errorf("parse scim group id: %w", err)
	}

	qSCIMGroup, err := s.q.AuthGetSCIMGroup(ctx, queries.AuthGetSCIMGroupParams{
		ScimDirectoryID: scimDirID,
		ID:              scimGroupID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSCIMGroupNotFound
		}
		return nil, fmt.Errorf("get scim group: %w", err)
	}

	return parseSCIMGroup(qSCIMGroup), nil
}

type AuthCreateSCIMGroupRequest struct {
	SCIMGroup         *ssoreadyv1.SCIMGroup
	MemberSCIMUserIDs []string
}

type AuthCreateSCIMGroupResponse struct {
	SCIMGroup         *ssoreadyv1.SCIMGroup
	MemberSCIMUserIDs []string
}

func (s *Store) AuthCreateSCIMGroup(ctx context.Context, req *AuthCreateSCIMGroupRequest) (*AuthCreateSCIMGroupResponse, error) {
	_, q, commit, rollback, err := s.tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("tx: %w", err)
	}
	defer rollback()

	scimDirID, err := idformat.SCIMDirectory.Parse(req.SCIMGroup.ScimDirectoryId)
	if err != nil {
		return nil, fmt.Errorf("parse scim directory id: %w", err)
	}

	// check every member user belongs to same directory as group does
	var scimUserIDs []uuid.UUID
	for _, scimUserID := range req.MemberSCIMUserIDs {
		scimUserID, err := idformat.SCIMUser.Parse(scimUserID)
		if err != nil {
			return nil, fmt.Errorf("parse scim user id: %w", err)
		}

		if _, err := q.AuthGetSCIMUser(ctx, queries.AuthGetSCIMUserParams{
			ScimDirectoryID: scimDirID,
			ID:              scimUserID,
		}); err != nil {
			return nil, fmt.Errorf("get scim user: %w", err)
		}

		scimUserIDs = append(scimUserIDs, scimUserID)
	}

	attrs, err := json.Marshal(req.SCIMGroup.Attributes.AsMap())
	if err != nil {
		panic(fmt.Errorf("marshal scim group attributes: %w", err))
	}

	qSCIMGroup, err := q.AuthCreateSCIMGroup(ctx, queries.AuthCreateSCIMGroupParams{
		ID:              uuid.New(),
		ScimDirectoryID: scimDirID,
		DisplayName:     req.SCIMGroup.DisplayName,
		Attributes:      attrs,
		Deleted:         false,
	})
	if err != nil {
		return nil, fmt.Errorf("create scim group: %w", err)
	}

	for _, scimUserID := range scimUserIDs {
		if _, err := q.AuthCreateSCIMUserGroupMembership(ctx, queries.AuthCreateSCIMUserGroupMembershipParams{
			ID:              uuid.New(),
			ScimDirectoryID: scimDirID,
			ScimUserID:      scimUserID,
			ScimGroupID:     qSCIMGroup.ID,
		}); err != nil {
			return nil, fmt.Errorf("create scim user group membership: %w", err)
		}
	}

	if err := commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return &AuthCreateSCIMGroupResponse{
		SCIMGroup:         parseSCIMGroup(qSCIMGroup),
		MemberSCIMUserIDs: req.MemberSCIMUserIDs,
	}, nil
}

type AuthDeleteSCIMGroupRequest struct {
	SCIMDirectoryID string
	SCIMGroupID     string
}

func (s *Store) AuthDeleteSCIMGroup(ctx context.Context, req *AuthDeleteSCIMGroupRequest) error {
	_, q, commit, rollback, err := s.tx(ctx)
	if err != nil {
		return fmt.Errorf("tx: %w", err)
	}
	defer rollback()

	scimDirID, err := idformat.SCIMDirectory.Parse(req.SCIMDirectoryID)
	if err != nil {
		return fmt.Errorf("parse scim directory id: %w", err)
	}

	scimGroupID, err := idformat.SCIMGroup.Parse(req.SCIMGroupID)
	if err != nil {
		return fmt.Errorf("parse scim group id: %w", err)
	}

	// check that the group belongs to the scim dir
	if _, err := q.AuthGetSCIMGroup(ctx, queries.AuthGetSCIMGroupParams{
		ScimDirectoryID: scimDirID,
		ID:              scimGroupID,
	}); err != nil {
		return fmt.Errorf("get scim group: %w", err)
	}

	if _, err := q.AuthMarkSCIMGroupDeleted(ctx, scimGroupID); err != nil {
		return fmt.Errorf("mark scim group deleted: %w", err)
	}

	if err := commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	return nil
}

func parseSCIMUser(qSCIMUser queries.ScimUser) *ssoreadyv1.SCIMUser {
	var attrs map[string]any
	if err := json.Unmarshal(qSCIMUser.Attributes, &attrs); err != nil {
		panic(fmt.Errorf("unmarshal scim user attributes: %w", err))
	}

	attrsStruct, err := structpb.NewStruct(attrs)
	if err != nil {
		panic(fmt.Errorf("build struct from scim user attributes: %w", err))
	}

	return &ssoreadyv1.SCIMUser{
		Id:              idformat.SCIMUser.Format(qSCIMUser.ID),
		ScimDirectoryId: idformat.SCIMDirectory.Format(qSCIMUser.ScimDirectoryID),
		Email:           qSCIMUser.Email,
		Deleted:         qSCIMUser.Deleted,
		Attributes:      attrsStruct,
	}
}

func parseSCIMGroup(qSCIMGroup queries.ScimGroup) *ssoreadyv1.SCIMGroup {
	var attrs map[string]any
	if err := json.Unmarshal(qSCIMGroup.Attributes, &attrs); err != nil {
		panic(fmt.Errorf("unmarshal scim group attributes: %w", err))
	}

	attrsStruct, err := structpb.NewStruct(attrs)
	if err != nil {
		panic(fmt.Errorf("build struct from scim group attributes: %w", err))
	}

	return &ssoreadyv1.SCIMGroup{
		Id:              idformat.SCIMGroup.Format(qSCIMGroup.ID),
		ScimDirectoryId: idformat.SCIMDirectory.Format(qSCIMGroup.ScimDirectoryID),
		DisplayName:     qSCIMGroup.DisplayName,
		Attributes:      attrsStruct,
		Deleted:         qSCIMGroup.Deleted,
	}
}
