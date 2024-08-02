package store

import (
	"context"
	"encoding/json"
	"fmt"

	ssoreadyv1 "github.com/ssoready/ssoready/internal/gen/ssoready/v1"
	"github.com/ssoready/ssoready/internal/store/idformat"
	"github.com/ssoready/ssoready/internal/store/queries"
	"google.golang.org/protobuf/types/known/structpb"
)

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

	qSCIMUsers, err := s.q.AuthListSCIMUsers(ctx, queries.AuthListSCIMUsersParams{
		ScimDirectoryID: scimDirID,
		Offset:          int32(req.StartIndex),
		Limit:           10,
	})
	if err != nil {
		return nil, fmt.Errorf("list scim users: %w", err)
	}

	var scimUsers []*ssoreadyv1.SCIMUser
	for _, qSCIMUser := range qSCIMUsers {
		var attrs map[string]any
		if err := json.Unmarshal(qSCIMUser.Attributes, &attrs); err != nil {
			panic(fmt.Errorf("unmarshal scim user attributes: %w", err))
		}

		attrsStruct, err := structpb.NewStruct(attrs)
		if err != nil {
			panic(fmt.Errorf("build struct from scim user attributes: %w", err))
		}

		scimUsers = append(scimUsers, &ssoreadyv1.SCIMUser{
			Id:              idformat.SCIMUser.Format(qSCIMUser.ID),
			ScimDirectoryId: idformat.SCIMDirectory.Format(qSCIMUser.ScimDirectoryID),
			Email:           qSCIMUser.Email,
			Deleted:         qSCIMUser.Deleted,
			Attributes:      attrsStruct,
		})
	}

	return &AuthListSCIMUsersResponse{
		TotalResults: int(count),
		SCIMUsers:    scimUsers,
	}, nil
}
