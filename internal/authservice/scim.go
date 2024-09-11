package authservice

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/ssoready/ssoready/internal/emailaddr"
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

	slog.InfoContext(ctx, "scim_list_users", "scim_directory_id", scimDirectoryID, "filter", r.URL.Query().Get("filter"))

	if r.URL.Query().Has("filter") {
		filterEmailPat := regexp.MustCompile(`userName eq "(.*)"`)
		match := filterEmailPat.FindStringSubmatch(r.URL.Query().Get("filter"))
		if match == nil {
			panic("unsupported filter param")
		}

		// scimvalidator.microsoft.com sends url-encoded values; harmless to "normal" emails to url-parse them
		email, err := url.QueryUnescape(match[1])
		if err != nil {
			panic(err)
		}

		scimUser, err := s.Store.AuthGetSCIMUserByEmail(ctx, &store.AuthGetSCIMUserByEmailRequest{
			SCIMDirectoryID: scimDirectoryID,
			Email:           email,
		})
		if err != nil {
			if errors.Is(err, store.ErrSCIMUserNotFound) {
				w.Header().Set("Content-Type", "application/scim+json")
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

		w.Header().Set("Content-Type", "application/scim+json")
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
		return
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

	resources := []any{} // intentionally initialized to avoid returning `null` instead of `[]`
	for _, scimUser := range scimUsers.SCIMUsers {
		resource := scimUser.Attributes.AsMap()
		resource["id"] = scimUser.Id
		resource["userName"] = scimUser.Email

		resources = append(resources, resource)
	}

	w.Header().Set("Content-Type", "application/scim+json")
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
		if errors.Is(err, store.ErrSCIMUserNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		panic(err)
	}

	resource := scimUser.Attributes.AsMap()
	resource["id"] = scimUser.Id
	resource["userName"] = scimUser.Email
	resource["schemas"] = []string{"urn:ietf:params:scim:schemas:core:2.0:User"}

	w.Header().Set("Content-Type", "application/scim+json")
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

	emailDomain, err := emailaddr.Parse(userName)
	if err != nil {
		http.Error(w, "userName is not a valid email address", http.StatusBadRequest)
		return
	}

	allowedDomains, err := s.Store.AuthGetSCIMDirectoryOrganizationDomains(ctx, scimDirectoryID)
	if err != nil {
		panic(err)
	}

	var domainOk bool
	for _, domain := range allowedDomains {
		if emailDomain == domain {
			domainOk = true
		}
	}

	if !domainOk {
		msg, err := json.Marshal(map[string]any{
			"status": http.StatusBadRequest,
			"detail": fmt.Sprintf("userName is not from the list of allowed domains: %s", strings.Join(allowedDomains, ", ")),
		})
		if err != nil {
			panic(err)
		}

		http.Error(w, string(msg), http.StatusBadRequest)
		return
	}

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

	w.Header().Set("Content-Type", "application/scim+json")
	w.WriteHeader(http.StatusCreated)

	response := scimUser.SCIMUser.Attributes.AsMap()
	response["schemas"] = []string{"urn:ietf:params:scim:schemas:core:2.0:User"}
	response["id"] = scimUser.SCIMUser.Id

	w.Header().Set("Content-Type", "application/scim+json")
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
	active := true                            // may be omitted in request
	if _, ok := resource["active"]; ok {
		active = resource["active"].(bool)
	}

	delete(resource, "schemas")

	// at this point, all remaining properties are user attributes
	attributes, err := structpb.NewStruct(resource)
	if err != nil {
		panic(fmt.Errorf("convert attributes to structpb: %w", err))
	}

	emailDomain, err := emailaddr.Parse(userName)
	if err != nil {
		http.Error(w, "userName is not a valid email address", http.StatusBadRequest)
		return
	}

	allowedDomains, err := s.Store.AuthGetSCIMDirectoryOrganizationDomains(ctx, scimDirectoryID)
	if err != nil {
		panic(err)
	}

	var domainOk bool
	for _, domain := range allowedDomains {
		if emailDomain == domain {
			domainOk = true
		}
	}

	if !domainOk {
		msg, err := json.Marshal(map[string]any{
			"status": http.StatusBadRequest,
			"detail": fmt.Sprintf("userName is not from the list of allowed domains: %s", strings.Join(allowedDomains, ", ")),
		})
		if err != nil {
			panic(err)
		}

		http.Error(w, string(msg), http.StatusBadRequest)
		return
	}

	scimUser, err := s.Store.AuthUpdateSCIMUser(ctx, &store.AuthUpdateSCIMUserRequest{
		SCIMUser: &ssoreadyv1.SCIMUser{
			Id:              scimUserID,
			ScimDirectoryId: scimDirectoryID,
			Email:           userName, // todo validate it's an email
			Deleted:         !active,
			Attributes:      attributes,
		},
	})
	if err != nil {
		panic(fmt.Errorf("store: %w", err))
	}

	w.Header().Set("Content-Type", "application/scim+json")
	w.WriteHeader(http.StatusOK)

	response := scimUser.SCIMUser.Attributes.AsMap()
	response["schemas"] = []string{"urn:ietf:params:scim:schemas:core:2.0:User"}
	response["id"] = scimUser.SCIMUser.Id

	w.Header().Set("Content-Type", "application/scim+json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		panic(err)
	}
}

func (s *Service) scimPatchUser(w http.ResponseWriter, r *http.Request) {
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

	var patch struct {
		Operations []struct {
			Op    string `json:"op"`
			Path  string `json:"path"`
			Value any    `json:"value"`
		} `json:"operations"`
	}

	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		panic(err)
	}

	// this is how entra deactivates users
	if len(patch.Operations) == 1 && patch.Operations[0].Op == "Replace" && patch.Operations[0].Path == "active" && patch.Operations[0].Value == "False" {
		if err := s.Store.AuthDeleteSCIMUser(ctx, &store.AuthDeleteSCIMUserRequest{
			SCIMDirectoryID: scimDirectoryID,
			SCIMUserID:      scimUserID,
		}); err != nil {
			panic(err)
		}

		w.WriteHeader(http.StatusNoContent)
		return
	}

	panic("unsupported group PATCH operation type")
}

func (s *Service) scimDeleteUser(w http.ResponseWriter, r *http.Request) {
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

	if err := s.Store.AuthDeleteSCIMUser(ctx, &store.AuthDeleteSCIMUserRequest{
		SCIMDirectoryID: scimDirectoryID,
		SCIMUserID:      scimUserID,
	}); err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Service) scimListGroups(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	scimDirectoryID := mux.Vars(r)["scim_directory_id"]

	if err := s.scimVerifyBearerToken(ctx, scimDirectoryID, r.Header.Get("Authorization")); err != nil {
		if errors.Is(err, store.ErrAuthSCIMBadBearerToken) {
			http.Error(w, "invalid bearer token", http.StatusUnauthorized)
			return
		}
		panic(err)
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

	scimGroups, err := s.Store.AuthListSCIMGroups(ctx, &store.AuthListSCIMGroupsRequest{
		SCIMDirectoryID: scimDirectoryID,
		StartIndex:      startIndex,
	})
	if err != nil {
		panic(fmt.Errorf("store: %w", err))
	}

	resources := []any{} // intentionally initialized to avoid returning `null` instead of `[]`
	for _, scimGroup := range scimGroups.SCIMGroups {
		resource := scimGroup.Attributes.AsMap()
		resource["id"] = scimGroup.Id
		resource["displayName"] = scimGroup.DisplayName

		resources = append(resources, resource)
	}

	w.Header().Set("Content-Type", "application/scim+json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(scimListResponse{
		TotalResults: scimGroups.TotalResults,
		ItemsPerPage: len(resources),
		StartIndex:   startIndex,
		Schemas:      []string{"urn:ietf:params:scim:schemas:core:2.0:Group"},
		Resources:    resources,
	}); err != nil {
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
		if errors.Is(err, store.ErrSCIMGroupNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		panic(err)
	}

	resource := scimGroup.Attributes.AsMap()
	resource["id"] = scimGroup.Id
	resource["displayName"] = scimGroup.DisplayName
	resource["schemas"] = []string{"urn:ietf:params:scim:schemas:core:2.0:Group"}

	w.Header().Set("Content-Type", "application/scim+json")
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

	var memberSCIMUserIDs []string
	if members, ok := resource["members"]; ok {
		members := members.([]any)
		for _, member := range members {
			member := member.(map[string]any)
			userID := member["value"].(string)
			memberSCIMUserIDs = append(memberSCIMUserIDs, userID)
		}
	}

	displayName := resource["displayName"].(string)
	delete(resource, "schemas")

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

	w.Header().Set("Content-Type", "application/scim+json")
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

	w.Header().Set("Content-Type", "application/scim+json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		panic(err)
	}
}

func (s *Service) scimDeleteGroup(w http.ResponseWriter, r *http.Request) {
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

	if err := s.Store.AuthDeleteSCIMGroup(ctx, &store.AuthDeleteSCIMGroupRequest{
		SCIMDirectoryID: scimDirectoryID,
		SCIMGroupID:     scimGroupID,
	}); err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Service) scimUpdateGroup(w http.ResponseWriter, r *http.Request) {
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

	defer r.Body.Close()
	var resource map[string]any
	if err := json.NewDecoder(r.Body).Decode(&resource); err != nil {
		panic(err)
	}

	var memberSCIMUserIDs []string
	if resource["members"] != nil {
		members := resource["members"].([]any)
		for _, member := range members {
			member := member.(map[string]any)
			userID := member["value"].(string)
			memberSCIMUserIDs = append(memberSCIMUserIDs, userID)
		}
	}

	displayName := resource["displayName"].(string)
	delete(resource, "schemas")

	// at this point, all remaining properties are user attributes
	attributes, err := structpb.NewStruct(resource)
	if err != nil {
		panic(fmt.Errorf("convert attributes to structpb: %w", err))
	}

	scimGroup, err := s.Store.AuthUpdateSCIMGroup(ctx, &store.AuthUpdateSCIMGroupRequest{
		SCIMGroup: &ssoreadyv1.SCIMGroup{
			Id:              scimGroupID,
			ScimDirectoryId: scimDirectoryID,
			DisplayName:     displayName,
			Attributes:      attributes,
		},
		MemberSCIMUserIDs: memberSCIMUserIDs,
	})
	if err != nil {
		panic(fmt.Errorf("store: %w", err))
	}

	w.Header().Set("Content-Type", "application/scim+json")
	w.WriteHeader(http.StatusOK)

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

func (s *Service) scimPatchGroup(w http.ResponseWriter, r *http.Request) {
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

	var patch struct {
		Operations []struct {
			Op    string `json:"op"`
			Path  string `json:"path"`
			Value any    `json:"value"`
		} `json:"operations"`
	}

	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		panic(err)
	}

	// jumpcloud changes group display names via a top-level replace
	if len(patch.Operations) == 1 && patch.Operations[0].Op == "replace" && patch.Operations[0].Path == "" {
		value := patch.Operations[0].Value.(map[string]any)
		displayName := value["displayName"].(string)

		if err := s.Store.AuthUpdateSCIMGroupDisplayName(ctx, &ssoreadyv1.SCIMGroup{
			Id:              scimGroupID,
			ScimDirectoryId: scimDirectoryID,
			DisplayName:     displayName,
		}); err != nil {
			panic(fmt.Errorf("store: %w", err))
		}

		w.WriteHeader(http.StatusNoContent)
		return
	}

	// jumpcloud adds members to groups via an `add` on members
	if len(patch.Operations) == 1 && patch.Operations[0].Op == "add" && patch.Operations[0].Path == "members" {
		value := patch.Operations[0].Value.([]any)
		scimUserID := value[0].(map[string]any)["value"].(string)

		if err := s.Store.AuthAddSCIMGroupMember(ctx, &store.AuthAddSCIMGroupMemberRequest{
			SCIMGroup: &ssoreadyv1.SCIMGroup{
				Id:              scimGroupID,
				ScimDirectoryId: scimDirectoryID,
			},
			SCIMUserID: scimUserID,
		}); err != nil {
			panic(fmt.Errorf("store: %w", err))
		}

		w.WriteHeader(http.StatusNoContent)
		return
	}

	panic("unsupported group PATCH operation type")
}

func (s *Service) scimVerifyBearerToken(ctx context.Context, scimDirectoryID, authorization string) error {
	bearerToken := strings.TrimPrefix(authorization, "Bearer ")
	return s.Store.AuthSCIMVerifyBearerToken(ctx, scimDirectoryID, bearerToken)
}
