package authservice

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/ssoready/ssoready/internal/store"
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

	var resources []map[string]any
	for _, scimUser := range scimUsers.SCIMUsers {
		resources = append(resources, map[string]any{
			"id": scimUser.Email,
			""
		})
	}
}
