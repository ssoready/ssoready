package authservice

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	ssoreadyv1 "github.com/ssoready/ssoready/internal/gen/ssoready/v1"
	"github.com/ssoready/ssoready/internal/store"
	"google.golang.org/protobuf/types/known/structpb"
)

type scimListResponse struct {
	TotalResults int      `json:"totalResults"`
	ItemsPerPage int      `json:"itemsPerPage"`
	StartIndex   int      `json:"startIndex"`
	Schemas      []string `json:"schemas"`
	Resources    []any    `json:"resources"`
}

func (s *Service) scimListUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	scimDirectoryID := mux.Vars(r)["scim_directory_id"]

	if err := s.scimVerifyBearerToken(ctx, scimDirectoryID, r.Header.Get("Authorization")); err != nil {
		if errors.Is(err, store.ErrAuthSCIMBadBearerToken) {
			http.Error(w, "invalid bearer token", http.StatusUnauthorized)
			return
		}
		panic(err)
	}

	if r.URL.Query().Has("filter") {
		filterEmailPat := regexp.MustCompile(`userName eq "(.*)"`)
		match := filterEmailPat.FindStringSubmatch(r.URL.Query().Get("filter"))
		if match == nil {
			panic("unsupported filter param")
		}

		email := match[1]
		scimUser, err := s.Store.AuthGetSCIMUserByEmail(ctx, &store.AuthGetSCIMUserByEmailRequest{
			SCIMDirectoryID: scimDirectoryID,
			Email:           email,
		})
		if err != nil {
			if errors.Is(err, store.ErrSCIMUserNotFound) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				if err := json.NewEncoder(w).Encode(scimListResponse{
					TotalResults: 0,
					ItemsPerPage: 1,
					StartIndex:   1,
					Schemas:      []string{"urn:ietf:params:scim:schemas:core:2.0:User"},
					Resources:    []any{},
				}); err != nil {
					panic(err)
				}
				return
			}

			panic(err)
		}

		resource := scimUser.Attributes.AsMap()
		resource["id"] = scimUser.Id
		resource["userName"] = scimUser.Email

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(scimListResponse{
			TotalResults: 1,
			ItemsPerPage: 1,
			StartIndex:   1,
			Schemas:      []string{"urn:ietf:params:scim:schemas:core:2.0:User"},
			Resources:    []any{resource},
		}); err != nil {
			panic(err)
		}
	}

	startIndex := 0
	if r.URL.Query().Get("startIndex") != "" {
		i, err := strconv.Atoi(r.URL.Query().Get("startIndex"))
		if err != nil {
			http.Error(w, fmt.Sprintf("parse startIndex: %s", err), http.StatusBadRequest)
			return
		}

		startIndex = i - 1 // scim is 1-indexed, store is 0-indexed
	}

	scimUsers, err := s.Store.AuthListSCIMUsers(ctx, &store.AuthListSCIMUsersRequest{
		SCIMDirectoryID: scimDirectoryID,
		StartIndex:      startIndex,
	})
	if err != nil {
		panic(fmt.Errorf("store: %w", err))
	}

	var resources []any
	for _, scimUser := range scimUsers.SCIMUsers {
		resource := scimUser.Attributes.AsMap()
		resource["id"] = scimUser.Id
		resource["userName"] = scimUser.Email

		resources = append(resources, resource)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(scimListResponse{
		TotalResults: scimUsers.TotalResults,
		ItemsPerPage: len(resources),
		StartIndex:   startIndex,
		Schemas:      []string{"urn:ietf:params:scim:schemas:core:2.0:User"},
		Resources:    resources,
	}); err != nil {
		panic(err)
	}
}

func (s *Service) scimGetUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	scimDirectoryID := mux.Vars(r)["scim_directory_id"]
	scimUserID := mux.Vars(r)["scim_user_id"]

	if err := s.scimVerifyBearerToken(ctx, scimDirectoryID, r.Header.Get("Authorization")); err != nil {
		if errors.Is(err, store.ErrAuthSCIMBadBearerToken) {
			http.Error(w, "invalid bearer token", http.StatusUnauthorized)
			return
		}
		panic(err)
	}

	scimUser, err := s.Store.AuthGetSCIMUser(ctx, &store.AuthGetSCIMUserRequest{
		SCIMDirectoryID: scimDirectoryID,
		SCIMUserID:      scimUserID,
	})
	if err != nil {
		panic(err)
	}

	resource := scimUser.Attributes.AsMap()
	resource["id"] = scimUser.Id
	resource["userName"] = scimUser.Email
	resource["schemas"] = []string{"urn:ietf:params:scim:schemas:core:2.0:User"}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resource); err != nil {
		panic(err)
	}
}

func (s *Service) scimCreateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	scimDirectoryID := mux.Vars(r)["scim_directory_id"]

	if err := s.scimVerifyBearerToken(ctx, scimDirectoryID, r.Header.Get("Authorization")); err != nil {
		if errors.Is(err, store.ErrAuthSCIMBadBearerToken) {
			http.Error(w, "invalid bearer token", http.StatusUnauthorized)
			return
		}
		panic(err)
	}

	defer r.Body.Close()
	var resource map[string]any
	if err := json.NewDecoder(r.Body).Decode(&resource); err != nil {
		panic(err)
	}

	userName := resource["userName"].(string) // todo this may panic
	delete(resource, "schemas")

	// at this point, all remaining properties are user attributes
	attributes, err := structpb.NewStruct(resource)
	if err != nil {
		panic(fmt.Errorf("convert attributes to structpb: %w", err))
	}

	scimUser, err := s.Store.AuthCreateSCIMUser(ctx, &store.AuthCreateSCIMUserRequest{
		SCIMUser: &ssoreadyv1.SCIMUser{
			ScimDirectoryId: scimDirectoryID,
			Email:           userName, // todo validate it's an email
			Deleted:         false,
			Attributes:      attributes,
		},
	})
	if err != nil {
		panic(fmt.Errorf("store: %w", err))
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	response := scimUser.SCIMUser.Attributes.AsMap()
	response["schemas"] = []string{"urn:ietf:params:scim:schemas:core:2.0:User"}
	response["id"] = scimUser.SCIMUser.Id

	if err := json.NewEncoder(w).Encode(response); err != nil {
		panic(err)
	}
}

func (s *Service) scimUpdateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	scimDirectoryID := mux.Vars(r)["scim_directory_id"]
	scimUserID := mux.Vars(r)["scim_user_id"]

	if err := s.scimVerifyBearerToken(ctx, scimDirectoryID, r.Header.Get("Authorization")); err != nil {
		if errors.Is(err, store.ErrAuthSCIMBadBearerToken) {
			http.Error(w, "invalid bearer token", http.StatusUnauthorized)
			return
		}
		panic(err)
	}

	defer r.Body.Close()
	var resource map[string]any
	if err := json.NewDecoder(r.Body).Decode(&resource); err != nil {
		panic(err)
	}

	userName := resource["userName"].(string) // todo this may panic
	delete(resource, "schemas")

	// at this point, all remaining properties are user attributes
	attributes, err := structpb.NewStruct(resource)
	if err != nil {
		panic(fmt.Errorf("convert attributes to structpb: %w", err))
	}

	scimUser, err := s.Store.AuthUpdateSCIMUser(ctx, &store.AuthUpdateSCIMUserRequest{
		SCIMUser: &ssoreadyv1.SCIMUser{
			Id:              scimUserID,
			ScimDirectoryId: scimDirectoryID,
			Email:           userName, // todo validate it's an email
			Deleted:         false,
			Attributes:      attributes,
		},
	})
	if err != nil {
		panic(fmt.Errorf("store: %w", err))
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := scimUser.SCIMUser.Attributes.AsMap()
	response["schemas"] = []string{"urn:ietf:params:scim:schemas:core:2.0:User"}
	response["id"] = scimUser.SCIMUser.Id

	if err := json.NewEncoder(w).Encode(response); err != nil {
		panic(err)
	}
}

func (s *Service) scimGetGroup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	scimDirectoryID := mux.Vars(r)["scim_directory_id"]
	scimGroupID := mux.Vars(r)["scim_group_id"]

	if err := s.scimVerifyBearerToken(ctx, scimDirectoryID, r.Header.Get("Authorization")); err != nil {
		if errors.Is(err, store.ErrAuthSCIMBadBearerToken) {
			http.Error(w, "invalid bearer token", http.StatusUnauthorized)
			return
		}
		panic(err)
	}

	scimGroup, err := s.Store.AuthGetSCIMGroup(ctx, &store.AuthGetSCIMGroupRequest{
		SCIMDirectoryID: scimDirectoryID,
		SCIMGroupID:     scimGroupID,
	})
	if err != nil {
		panic(err)
	}

	resource := scimGroup.Attributes.AsMap()
	resource["id"] = scimGroup.Id
	resource["displayName"] = scimGroup.DisplayName
	resource["schemas"] = []string{"urn:ietf:params:scim:schemas:core:2.0:Group"}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resource); err != nil {
		panic(err)
	}
}

func (s *Service) scimCreateGroup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	scimDirectoryID := mux.Vars(r)["scim_directory_id"]

	if err := s.scimVerifyBearerToken(ctx, scimDirectoryID, r.Header.Get("Authorization")); err != nil {
		if errors.Is(err, store.ErrAuthSCIMBadBearerToken) {
			http.Error(w, "invalid bearer token", http.StatusUnauthorized)
			return
		}
		panic(err)
	}

	defer r.Body.Close()
	var resource map[string]any
	if err := json.NewDecoder(r.Body).Decode(&resource); err != nil {
		panic(err)
	}

	members := resource["members"].([]any)
	var memberSCIMUserIDs []string
	for _, member := range members {
		member := member.(map[string]any)
		userID := member["value"].(string)
		memberSCIMUserIDs = append(memberSCIMUserIDs, userID)
	}

	displayName := resource["displayName"].(string)
	delete(resource, "schemas")
	delete(resource, "members")

	// at this point, all remaining properties are user attributes
	attributes, err := structpb.NewStruct(resource)
	if err != nil {
		panic(fmt.Errorf("convert attributes to structpb: %w", err))
	}

	scimGroup, err := s.Store.AuthCreateSCIMGroup(ctx, &store.AuthCreateSCIMGroupRequest{
		SCIMGroup: &ssoreadyv1.SCIMGroup{
			ScimDirectoryId: scimDirectoryID,
			DisplayName:     displayName,
			Attributes:      attributes,
		},
		MemberSCIMUserIDs: memberSCIMUserIDs,
	})
	if err != nil {
		panic(fmt.Errorf("store: %w", err))
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	response := scimGroup.SCIMGroup.Attributes.AsMap()
	response["schemas"] = []string{"urn:ietf:params:scim:schemas:core:2.0:Group"}
	response["id"] = scimGroup.SCIMGroup.Id
	var responseMembers []map[string]any
	for _, userID := range scimGroup.MemberSCIMUserIDs {
		responseMembers = append(responseMembers, map[string]any{
			"value": userID,
		})
	}
	response["members"] = responseMembers

	if err := json.NewEncoder(w).Encode(response); err != nil {
		panic(err)
	}
}

func (s *Service) scimUpdateGroup(w http.ResponseWriter, r *http.Request) {
	// todo
	s.scimGetGroup(w, r)
}

func (s *Service) scimVerifyBearerToken(ctx context.Context, scimDirectoryID, authorization string) error {
	bearerToken := strings.TrimPrefix(authorization, "Bearer ")
	return s.Store.AuthSCIMVerifyBearerToken(ctx, scimDirectoryID, bearerToken)
}