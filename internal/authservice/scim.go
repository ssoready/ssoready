package authservice

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ssoready/ssoready/internal/scimpatch"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

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

func (s *Service) scimListUsers(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	scimDirectoryID := mux.Vars(r)["scim_directory_id"]

	if err := s.scimVerifyBearerToken(ctx, scimDirectoryID, r.Header.Get("Authorization")); err != nil {
		if errors.Is(err, store.ErrAuthSCIMBadBearerToken) {
			http.Error(w, "invalid bearer token", http.StatusUnauthorized)
			return err
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
				return nil
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
		return nil
	}

	startIndex := 0
	if r.URL.Query().Get("startIndex") != "" {
		i, err := strconv.Atoi(r.URL.Query().Get("startIndex"))
		if err != nil {
			http.Error(w, fmt.Sprintf("parse startIndex: %s", err), http.StatusBadRequest)
			return nil
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
		resources = append(resources, scimUserToResource(scimUser))
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
	return nil
}

func (s *Service) scimGetUser(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	scimDirectoryID := mux.Vars(r)["scim_directory_id"]
	scimUserID := mux.Vars(r)["scim_user_id"]

	scimUser, err := s.Store.AuthGetSCIMUser(ctx, &store.AuthGetSCIMUserRequest{
		SCIMDirectoryID: scimDirectoryID,
		SCIMUserID:      scimUserID,
	})
	if err != nil {
		if errors.Is(err, store.ErrSCIMUserNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return nil
		}

		return err
	}

	resource := scimUserToResource(scimUser)
	resource["schemas"] = []string{"urn:ietf:params:scim:schemas:core:2.0:User"}

	w.Header().Set("Content-Type", "application/scim+json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resource); err != nil {
		return err
	}
	return nil
}

func (s *Service) scimCreateUser(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	scimDirectoryID := mux.Vars(r)["scim_directory_id"]

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
		return &badUsernameError{BadUsername: userName}
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
		return &emailOutsideOrgDomainsError{BadEmail: userName}
	}

	// at this point, all remaining properties are user attributes
	attributes, err := structpb.NewStruct(resource)
	if err != nil {
		panic(fmt.Errorf("convert attributes to structpb: %w", err))
	}

	scimUser, err := s.Store.AuthCreateSCIMUser(ctx, &store.AuthCreateSCIMUserRequest{
		SCIMUser: &ssoreadyv1.SCIMUser{
			ScimDirectoryId: scimDirectoryID,
			Email:           userName,
			Deleted:         false,
			Attributes:      attributes,
		},
	})
	if err != nil {
		panic(fmt.Errorf("store: %w", err))
	}

	w.Header().Set("Content-Type", "application/scim+json")
	w.WriteHeader(http.StatusCreated)

	response := scimUserToResource(scimUser.SCIMUser)
	response["schemas"] = []string{"urn:ietf:params:scim:schemas:core:2.0:User"}

	w.Header().Set("Content-Type", "application/scim+json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		panic(err)
	}
	return nil
}

func (s *Service) scimUpdateUser(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	scimDirectoryID := mux.Vars(r)["scim_directory_id"]
	scimUserID := mux.Vars(r)["scim_user_id"]

	defer r.Body.Close()
	var resource map[string]any
	if err := json.NewDecoder(r.Body).Decode(&resource); err != nil {
		return err
	}

	if resource["userName"] == nil {
		http.Error(w, "userName is required", http.StatusBadRequest)
		return &badUsernameError{BadUsername: ""}
	}
	if _, ok := resource["userName"]; !ok {
		http.Error(w, "userName is required", http.StatusBadRequest)
		return &badUsernameError{BadUsername: ""}
	}

	userName := resource["userName"].(string)
	active := true // may be omitted in request
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
		return &badUsernameError{BadUsername: userName}
	}

	allowedDomains, err := s.Store.AuthGetSCIMDirectoryOrganizationDomains(ctx, scimDirectoryID)
	if err != nil {
		return err
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
		return &emailOutsideOrgDomainsError{BadEmail: userName}
	}

	scimUser, err := s.Store.AuthUpdateSCIMUser(ctx, &store.AuthUpdateSCIMUserRequest{
		SCIMUser: &ssoreadyv1.SCIMUser{
			Id:              scimUserID,
			ScimDirectoryId: scimDirectoryID,
			Email:           userName,
			Deleted:         !active,
			Attributes:      attributes,
		},
	})
	if err != nil {
		return fmt.Errorf("store: %w", err)
	}

	w.Header().Set("Content-Type", "application/scim+json")
	w.WriteHeader(http.StatusOK)

	response := scimUserToResource(scimUser.SCIMUser)
	response["schemas"] = []string{"urn:ietf:params:scim:schemas:core:2.0:User"}

	w.Header().Set("Content-Type", "application/scim+json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		return fmt.Errorf("encode: %w", err)
	}
	return nil
}

func (s *Service) scimPatchUser(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	scimDirectoryID := mux.Vars(r)["scim_directory_id"]
	scimUserID := mux.Vars(r)["scim_user_id"]

	var patch struct {
		Operations []scimpatch.Operation `json:"operations"`
	}

	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		panic(err)
	}

	scimUser, err := s.Store.AuthGetSCIMUserIncludeDeleted(ctx, &store.AuthGetSCIMUserIncludeDeletedRequest{
		SCIMDirectoryID: scimDirectoryID,
		SCIMUserID:      scimUserID,
	})
	if err != nil {
		panic(fmt.Errorf("store: get scim user for patch: %w", err))
	}

	// convert scimUser to its SCIM representation
	scimUserResource := scimUserToResource(scimUser)

	slog.InfoContext(ctx, "patched_user_fetch", "scim_user", scimUser, "scim_user_resource", scimUserResource)

	// apply patches
	if err := scimpatch.Patch(patch.Operations, &scimUserResource); err != nil {
		w.Header().Set("Content-Type", "application/scim+json")
		w.WriteHeader(http.StatusBadRequest)
		errorResponse := map[string]interface{}{
			"schemas":  []string{"urn:ietf:params:scim:api:messages:2.0:Error"},
			"status":   "400",
			"scimType": "invalidPath",
			"detail":   fmt.Sprintf("Unsupported PATCH operation: %s", err.Error()),
		}
		if err := json.NewEncoder(w).Encode(errorResponse); err != nil {
			panic(fmt.Errorf("encode error response: %w", err))
		}
		return nil
	}

	// convert back to our representation
	patchedSCIMUser := scimUserFromResource(scimDirectoryID, scimUserID, scimUserResource)

	// do not allow patches to remove the user email address
	if patchedSCIMUser.Email == "" {
		patchedSCIMUser.Email = scimUser.Email
	}

	slog.InfoContext(ctx, "patched_user_fetch", "patched_scim_user_resource", scimUserResource, "patched_scim_user", patchedSCIMUser)

	// validate email
	emailDomain, err := emailaddr.Parse(patchedSCIMUser.Email)
	if err != nil {
		http.Error(w, "userName is not a valid email address", http.StatusBadRequest)
		return &badUsernameError{BadUsername: patchedSCIMUser.Email}
	}

	allowedDomains, err := s.Store.AuthGetSCIMDirectoryOrganizationDomains(ctx, scimDirectoryID)
	if err != nil {
		return err
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
		return &emailOutsideOrgDomainsError{BadEmail: patchedSCIMUser.Email}
	}

	// write patched scim user back to database
	if _, err := s.Store.AuthUpdateSCIMUser(ctx, &store.AuthUpdateSCIMUserRequest{
		SCIMUser: patchedSCIMUser,
	}); err != nil {
		return fmt.Errorf("store: update patched user: %w", err)
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (s *Service) scimDeleteUser(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	scimDirectoryID := mux.Vars(r)["scim_directory_id"]
	scimUserID := mux.Vars(r)["scim_user_id"]

	if err := s.Store.AuthDeleteSCIMUser(ctx, &store.AuthDeleteSCIMUserRequest{
		SCIMDirectoryID: scimDirectoryID,
		SCIMUserID:      scimUserID,
	}); err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (s *Service) scimListGroups(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	scimDirectoryID := mux.Vars(r)["scim_directory_id"]

	if err := s.scimVerifyBearerToken(ctx, scimDirectoryID, r.Header.Get("Authorization")); err != nil {
		if errors.Is(err, store.ErrAuthSCIMBadBearerToken) {
			http.Error(w, "invalid bearer token", http.StatusUnauthorized)
			return nil
		}
		panic(err)
	}

	startIndex := 0
	if r.URL.Query().Get("startIndex") != "" {
		i, err := strconv.Atoi(r.URL.Query().Get("startIndex"))
		if err != nil {
			http.Error(w, fmt.Sprintf("parse startIndex: %s", err), http.StatusBadRequest)
			return nil
		}

		startIndex = i - 1 // scim is 1-indexed, store is 0-indexed
	}

	var filterDisplayName string
	if r.URL.Query().Has("filter") {
		filterDisplayNamePat := regexp.MustCompile(`displayName eq "(.*)"`)
		match := filterDisplayNamePat.FindStringSubmatch(r.URL.Query().Get("filter"))
		if match == nil {
			panic("unsupported filter param")
		}

		filterDisplayName = match[1]
	}

	scimGroups, err := s.Store.AuthListSCIMGroups(ctx, &store.AuthListSCIMGroupsRequest{
		SCIMDirectoryID: scimDirectoryID,
		StartIndex:      startIndex,

		// Unlike ListUsers, which uses a separate query to list users by email
		// (an operation that's guaranteed to return only one value), ListGroups
		// here instead does a more traditional filter, because we do not
		// enforce group uniqueness by displayName. Multiple values may be
		// returned even if FilterDisplayName is set.
		FilterDisplayName: filterDisplayName,
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
	return nil
}

func (s *Service) scimGetGroup(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	scimDirectoryID := mux.Vars(r)["scim_directory_id"]
	scimGroupID := mux.Vars(r)["scim_group_id"]

	if err := s.scimVerifyBearerToken(ctx, scimDirectoryID, r.Header.Get("Authorization")); err != nil {
		if errors.Is(err, store.ErrAuthSCIMBadBearerToken) {
			http.Error(w, "invalid bearer token", http.StatusUnauthorized)
			return nil
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
			return nil
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
	return nil
}

func (s *Service) scimCreateGroup(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	scimDirectoryID := mux.Vars(r)["scim_directory_id"]

	if err := s.scimVerifyBearerToken(ctx, scimDirectoryID, r.Header.Get("Authorization")); err != nil {
		if errors.Is(err, store.ErrAuthSCIMBadBearerToken) {
			http.Error(w, "invalid bearer token", http.StatusUnauthorized)
			return nil
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
	return nil
}

func (s *Service) scimDeleteGroup(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	scimDirectoryID := mux.Vars(r)["scim_directory_id"]
	scimGroupID := mux.Vars(r)["scim_group_id"]

	if err := s.scimVerifyBearerToken(ctx, scimDirectoryID, r.Header.Get("Authorization")); err != nil {
		if errors.Is(err, store.ErrAuthSCIMBadBearerToken) {
			http.Error(w, "invalid bearer token", http.StatusUnauthorized)
			return nil
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
	return nil
}

func (s *Service) scimUpdateGroup(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	scimDirectoryID := mux.Vars(r)["scim_directory_id"]
	scimGroupID := mux.Vars(r)["scim_group_id"]

	if err := s.scimVerifyBearerToken(ctx, scimDirectoryID, r.Header.Get("Authorization")); err != nil {
		if errors.Is(err, store.ErrAuthSCIMBadBearerToken) {
			http.Error(w, "invalid bearer token", http.StatusUnauthorized)
			return nil
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
	return nil
}

func (s *Service) scimPatchGroup(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	scimDirectoryID := mux.Vars(r)["scim_directory_id"]
	scimGroupID := mux.Vars(r)["scim_group_id"]

	if err := s.scimVerifyBearerToken(ctx, scimDirectoryID, r.Header.Get("Authorization")); err != nil {
		if errors.Is(err, store.ErrAuthSCIMBadBearerToken) {
			http.Error(w, "invalid bearer token", http.StatusUnauthorized)
			return nil
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
		return nil
	}

	// jumpcloud adds members to groups via an `add` on members; entra uses an `Add`
	if len(patch.Operations) == 1 && (patch.Operations[0].Op == "add" || patch.Operations[0].Op == "Add") && patch.Operations[0].Path == "members" {
		value := patch.Operations[0].Value.([]any)
		scimUserID := value[0].(map[string]any)["value"].(string)

		if err := s.Store.AuthAddSCIMGroupMember(ctx, &store.AuthAddSCIMGroupMemberRequest{
			SCIMGroup: &ssoreadyv1.SCIMGroup{
				Id:              scimGroupID,
				ScimDirectoryId: scimDirectoryID,
			},
			SCIMUserID: scimUserID,
		}); err != nil {
			if errors.Is(err, store.ErrBadSCIMUserID) {
				http.Error(w, "bad scim user id", http.StatusUnauthorized)
				return nil
			}

			panic(fmt.Errorf("store: %w", err))
		}

		w.WriteHeader(http.StatusNoContent)
		return nil
	}

	// entra removes members via a `remove` on members with a value
	if len(patch.Operations) == 1 && (patch.Operations[0].Op == "remove" || patch.Operations[0].Op == "Remove") && patch.Operations[0].Path == "members" {
		value := patch.Operations[0].Value.([]any)
		scimUserID := value[0].(map[string]any)["value"].(string)

		if err := s.Store.AuthRemoveSCIMGroupMember(ctx, &store.AuthRemoveSCIMGroupMemberRequest{
			SCIMGroup: &ssoreadyv1.SCIMGroup{
				Id:              scimGroupID,
				ScimDirectoryId: scimDirectoryID,
			},
			SCIMUserID: scimUserID,
		}); err != nil {
			panic(fmt.Errorf("store: %w", err))
		}

		w.WriteHeader(http.StatusNoContent)
		return nil
	}

	panic("unsupported group PATCH operation type")
}

// scimUserToResource converts our representation of a scim user to its SCIM HTTP representation
func scimUserToResource(scimUser *ssoreadyv1.SCIMUser) map[string]any {
	r := scimUser.Attributes.AsMap()
	r["id"] = scimUser.Id
	r["userName"] = scimUser.Email

	// normalize Entra-style "active" property
	if r["active"] == "True" {
		r["active"] = true
	} else if r["active"] == "False" {
		r["active"] = false
	}

	// convert simple manager id reference to complex manager reference for Entra compatibility
	if r["urn:ietf:params:scim:schemas:extension:enterprise:2.0:User"] != nil {
		enterpriseUser := r["urn:ietf:params:scim:schemas:extension:enterprise:2.0:User"].(map[string]any)
		if enterpriseUser["manager"] != nil {
			if managerID, ok := enterpriseUser["manager"].(string); ok {
				enterpriseUser["manager"] = map[string]any{
					"value": managerID,
				}
			}
		}
	}

	return r
}

func scimUserFromResource(scimDirectoryID, scimUserID string, r map[string]any) *ssoreadyv1.SCIMUser {
	// if included, id and schemas are not attributes
	delete(r, "id")
	delete(r, "schemas")

	// normalize Entra-style "active" property
	if r["active"] == "True" {
		r["active"] = true
	} else if r["active"] == "False" {
		r["active"] = false
	}

	attrs, err := structpb.NewStruct(r)
	if err != nil {
		panic(fmt.Errorf("convert attributes to structpb: %w", err))
	}

	// at this point, deliberately throw away non-well-typed values
	email, _ := r["userName"].(string)
	active, _ := r["active"].(bool)

	return &ssoreadyv1.SCIMUser{
		Id:              scimUserID,
		ScimDirectoryId: scimDirectoryID,
		Email:           email,
		Deleted:         !active,
		Attributes:      attrs,
	}
}

type badUsernameError struct {
	BadUsername string
}

func (e *badUsernameError) Error() string {
	return fmt.Sprintf("bad username: %v", e.BadUsername)
}

type emailOutsideOrgDomainsError struct {
	BadEmail string
}

func (e *emailOutsideOrgDomainsError) Error() string {
	return fmt.Sprintf("email outside organization domains: %v", e.BadEmail)
}

// scimMiddleware verifies scim bearer tokens and creates scim request logs in the database.
//
// To detect the error cases of bad usernames and emails outside org domains, scimMiddleware takes an f that returns an
// error. If that error is a badUsernameError or emailOutsideOrgDomainsError, the logged scim request is appropriately
// marked as such.
func (s *Service) scimMiddleware(f func(w http.ResponseWriter, r *http.Request) error) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		scimDirectoryID := mux.Vars(r)["scim_directory_id"]

		// check scim dir exists, and 404 immediately if not; we can't create a scim request for this case anyway
		if err := s.Store.AuthCheckSCIMDirectoryExists(ctx, scimDirectoryID); err != nil {
			if errors.Is(err, store.ErrNoSuchSCIMDirectory) {
				http.Error(w, "scim directory not found", http.StatusNotFound)
				return
			}

			panic(err)
		}

		defer r.Body.Close()
		reqBody, err := io.ReadAll(r.Body)
		if err != nil {
			panic(fmt.Errorf("read body: %w", err))
		}

		var scimRequestMethod ssoreadyv1.SCIMRequestHTTPMethod
		switch r.Method {
		case http.MethodGet:
			scimRequestMethod = ssoreadyv1.SCIMRequestHTTPMethod_SCIM_REQUEST_HTTP_METHOD_GET
		case http.MethodPost:
			scimRequestMethod = ssoreadyv1.SCIMRequestHTTPMethod_SCIM_REQUEST_HTTP_METHOD_POST
		case http.MethodPut:
			scimRequestMethod = ssoreadyv1.SCIMRequestHTTPMethod_SCIM_REQUEST_HTTP_METHOD_PUT
		case http.MethodPatch:
			scimRequestMethod = ssoreadyv1.SCIMRequestHTTPMethod_SCIM_REQUEST_HTTP_METHOD_PATCH
		case http.MethodDelete:
			scimRequestMethod = ssoreadyv1.SCIMRequestHTTPMethod_SCIM_REQUEST_HTTP_METHOD_DELETE
		}

		var scimRequestBody *structpb.Struct
		if len(reqBody) > 0 {
			if err := json.Unmarshal(reqBody, &scimRequestBody); err != nil {
				panic(err)
			}
		}

		scimRequest := &ssoreadyv1.SCIMRequest{
			ScimDirectoryId:   scimDirectoryID,
			Timestamp:         timestamppb.New(time.Now()),
			HttpRequestUrl:    r.URL.String(),
			HttpRequestMethod: scimRequestMethod,
			HttpRequestBody:   scimRequestBody,
		}

		// rewrite the response to be a recorded one, and the request to have the original body
		recorder := httptest.NewRecorder()

		// Make a copy of reqBody to work with later
		bodyCopy := make([]byte, len(reqBody))
		copy(bodyCopy, reqBody)

		// Set the copied reqBody back to r.Body
		r.Body = io.NopCloser(bytes.NewBuffer(bodyCopy))

		// check bearer token before calling f
		bearerToken := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if err := s.Store.AuthSCIMVerifyBearerToken(ctx, scimDirectoryID, bearerToken); err != nil {
			if errors.Is(err, store.ErrAuthSCIMBadBearerToken) {
				// log a failed scim request
				scimRequest.HttpResponseStatus = ssoreadyv1.SCIMRequestHTTPStatus_SCIM_REQUEST_HTTP_STATUS_401
				scimRequest.Error = &ssoreadyv1.SCIMRequest_BadBearerToken{BadBearerToken: &emptypb.Empty{}}
				if _, err := s.Store.AuthCreateSCIMRequest(ctx, scimRequest); err != nil {
					panic(err)
				}

				http.Error(w, "invalid bearer token", http.StatusUnauthorized)
				return
			}

			panic(err)
		}

		// call the underlying f
		if err := f(recorder, r); err != nil {
			var badUsernameError *badUsernameError
			var badEmailError *emailOutsideOrgDomainsError

			if errors.As(err, &badUsernameError) {
				scimRequest.Error = &ssoreadyv1.SCIMRequest_BadUsername{BadUsername: badUsernameError.BadUsername}
			} else if errors.As(err, &badEmailError) {
				scimRequest.Error = &ssoreadyv1.SCIMRequest_EmailOutsideOrganizationDomains{EmailOutsideOrganizationDomains: badEmailError.BadEmail}
			} else {
				panic(err)
			}
		}

		switch recorder.Code {
		case http.StatusOK:
			scimRequest.HttpResponseStatus = ssoreadyv1.SCIMRequestHTTPStatus_SCIM_REQUEST_HTTP_STATUS_200
		case http.StatusCreated:
			scimRequest.HttpResponseStatus = ssoreadyv1.SCIMRequestHTTPStatus_SCIM_REQUEST_HTTP_STATUS_201
		case http.StatusNoContent:
			scimRequest.HttpResponseStatus = ssoreadyv1.SCIMRequestHTTPStatus_SCIM_REQUEST_HTTP_STATUS_204
		case http.StatusBadRequest:
			scimRequest.HttpResponseStatus = ssoreadyv1.SCIMRequestHTTPStatus_SCIM_REQUEST_HTTP_STATUS_400
		case http.StatusNotFound:
			scimRequest.HttpResponseStatus = ssoreadyv1.SCIMRequestHTTPStatus_SCIM_REQUEST_HTTP_STATUS_404
		}

		if err := json.Unmarshal(recorder.Body.Bytes(), &scimRequest.HttpResponseBody); err != nil {
			// can't use errors.Is, because json's errors aren't comparable
			if _, ok := err.(*json.SyntaxError); ok {
				// ignore this error; we only care to record JSON responses, and so just record no response body at all
			} else {
				panic(err)
			}
		}

		if _, err := s.Store.AuthCreateSCIMRequest(ctx, scimRequest); err != nil {
			panic(err)
		}

		// write out recorded response to w
		for k, v := range recorder.Header() {
			w.Header()[k] = v
		}
		w.WriteHeader(recorder.Code)
		if _, err := recorder.Body.WriteTo(w); err != nil {
			panic(fmt.Errorf("write reqBody: %w", err))
		}
	})
}

func (s *Service) scimVerifyBearerToken(ctx context.Context, scimDirectoryID, authorization string) error {
	bearerToken := strings.TrimPrefix(authorization, "Bearer ")
	return s.Store.AuthSCIMVerifyBearerToken(ctx, scimDirectoryID, bearerToken)
}
