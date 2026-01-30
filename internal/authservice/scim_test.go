package authservice

import (
	"regexp"
	"testing"

	"github.com/ssoready/ssoready/internal/emailaddr"
	ssoreadyv1 "github.com/ssoready/ssoready/internal/gen/ssoready/v1"
	"github.com/ssoready/ssoready/internal/scimpatch"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestSCIMUserToResource_ManagerReference(t *testing.T) {
	tests := []struct {
		name     string
		input    *ssoreadyv1.SCIMUser
		expected map[string]any
	}{
		{
			name: "simple manager ID is converted to complex reference",
			input: &ssoreadyv1.SCIMUser{
				Id:    "user123",
				Email: "test@example.com",
				Attributes: mustNewStruct(map[string]any{
					"urn:ietf:params:scim:schemas:extension:enterprise:2.0:User": map[string]any{
						"manager": "manager123",
					},
				}),
			},
			expected: map[string]any{
				"id":       "user123",
				"userName": "test@example.com",
				"urn:ietf:params:scim:schemas:extension:enterprise:2.0:User": map[string]any{
					"manager": map[string]any{
						"value": "manager123",
					},
				},
			},
		},
		{
			name: "already complex manager reference is preserved",
			input: &ssoreadyv1.SCIMUser{
				Id:    "user123",
				Email: "test@example.com",
				Attributes: mustNewStruct(map[string]any{
					"urn:ietf:params:scim:schemas:extension:enterprise:2.0:User": map[string]any{
						"manager": map[string]any{
							"value": "manager123",
						},
					},
				}),
			},
			expected: map[string]any{
				"id":       "user123",
				"userName": "test@example.com",
				"urn:ietf:params:scim:schemas:extension:enterprise:2.0:User": map[string]any{
					"manager": map[string]any{
						"value": "manager123",
					},
				},
			},
		},
		{
			name: "no manager reference remains unchanged",
			input: &ssoreadyv1.SCIMUser{
				Id:    "user123",
				Email: "test@example.com",
				Attributes: mustNewStruct(map[string]any{
					"urn:ietf:params:scim:schemas:extension:enterprise:2.0:User": map[string]any{},
				}),
			},
			expected: map[string]any{
				"id":       "user123",
				"userName": "test@example.com",
				"urn:ietf:params:scim:schemas:extension:enterprise:2.0:User": map[string]any{},
			},
		},
		{
			name: "no enterprise extension remains unchanged",
			input: &ssoreadyv1.SCIMUser{
				Id:         "user123",
				Email:      "test@example.com",
				Attributes: mustNewStruct(map[string]any{}),
			},
			expected: map[string]any{
				"id":       "user123",
				"userName": "test@example.com",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scimUserToResource(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSCIMFilterRegex(t *testing.T) {
	tests := []struct {
		name        string
		filter      string
		expected    string
		shouldMatch bool
	}{
		{
			name:        "userName filter",
			filter:      `userName eq "john@example.com"`,
			expected:    "john@example.com",
			shouldMatch: true,
		},
		{
			name:        "email.value filter",
			filter:      `email.value eq "jane@example.com"`,
			expected:    "jane@example.com",
			shouldMatch: true,
		},
		{
			name:        "email.value filter with special characters",
			filter:      `email.value eq "user+tag@example.com"`,
			expected:    "user+tag@example.com",
			shouldMatch: true,
		},
		{
			name:        "unsupported filter - different attribute",
			filter:      `displayName eq "John Doe"`,
			expected:    "",
			shouldMatch: false,
		},
		{
			name:        "unsupported filter - different operator",
			filter:      `userName ne "john@example.com"`,
			expected:    "",
			shouldMatch: false,
		},
		{
			name:        "unsupported filter - malformed",
			filter:      `userName eq john@example.com`,
			expected:    "",
			shouldMatch: false,
		},
	}

	filterEmailPat := regexp.MustCompile(`(userName|email\.value) eq "(.*)"`)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match := filterEmailPat.FindStringSubmatch(tt.filter)

			if tt.shouldMatch {
				assert.NotNil(t, match, "Expected filter to match but it didn't")
				assert.Len(t, match, 3, "Expected 3 capture groups (full match, attribute, value)")
				assert.Equal(t, tt.expected, match[2], "Expected email value to match")
			} else {
				assert.Nil(t, match, "Expected filter to not match but it did")
			}
		})
	}
}

func TestSCIMPatchUser_NestedProperties(t *testing.T) {
	tests := []struct {
		name           string
		input          *ssoreadyv1.SCIMUser
		patchOps       []scimpatch.Operation
		expectedResult map[string]any
	}{
		{
			name: "add nested name properties when parent doesn't exist",
			input: &ssoreadyv1.SCIMUser{
				Id:    "user123",
				Email: "test@example.com",
				Attributes: mustNewStruct(map[string]any{
					"displayName": "Test User",
				}),
			},
			patchOps: []scimpatch.Operation{
				{Op: "Replace", Path: "displayName", Value: "John Doe (Changed)"},
				{Op: "Add", Path: "name.givenName", Value: "John"},
				{Op: "Add", Path: "name.familyName", Value: "Doe"},
				{Op: "Add", Path: "name.formatted", Value: "John Doe"},
			},
			expectedResult: map[string]any{
				"id":          "user123",
				"userName":    "test@example.com",
				"displayName": "John Doe (Changed)",
				"name": map[string]any{
					"givenName":  "John",
					"familyName": "Doe",
					"formatted":  "John Doe",
				},
			},
		},
		{
			name: "add nested properties to existing parent object",
			input: &ssoreadyv1.SCIMUser{
				Id:    "user456",
				Email: "user@example.com",
				Attributes: mustNewStruct(map[string]any{
					"name": map[string]any{
						"givenName": "John",
					},
				}),
			},
			patchOps: []scimpatch.Operation{
				{Op: "Add", Path: "name.familyName", Value: "Doe"},
				{Op: "Add", Path: "name.formatted", Value: "John Doe"},
			},
			expectedResult: map[string]any{
				"id":       "user456",
				"userName": "user@example.com",
				"name": map[string]any{
					"givenName":  "John",
					"familyName": "Doe",
					"formatted":  "John Doe",
				},
			},
		},
		{
			name: "deeply nested property creation with filter expressions",
			input: &ssoreadyv1.SCIMUser{
				Id:         "user789",
				Email:      "deep@example.com",
				Attributes: mustNewStruct(map[string]any{}),
			},
			patchOps: []scimpatch.Operation{
				{Op: "Add", Path: "addresses[type eq \"work\"].streetAddress", Value: "123 Main St"},
				{Op: "Add", Path: "addresses[type eq \"work\"].locality", Value: "San Francisco"},
			},
			expectedResult: map[string]any{
				"id":       "user789",
				"userName": "deep@example.com",
				"addresses": []any{
					map[string]any{
						"type":          "work",
						"streetAddress": "123 Main St",
						"locality":      "San Francisco",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert SCIMUser to resource
			resource := scimUserToResource(tt.input)

			// Apply patch operations using scimpatch
			err := scimpatch.Patch(tt.patchOps, &resource)
			require.NoError(t, err, "patch should succeed")

			// Verify the result matches expected
			assert.Equal(t, tt.expectedResult, resource)
		})
	}
}

// Helper function to create structpb.Struct from map
func mustNewStruct(m map[string]any) *structpb.Struct {
	s, err := structpb.NewStruct(m)
	if err != nil {
		panic(err)
	}
	return s
}

func TestExtractEmailFromResource(t *testing.T) {
	tests := []struct {
		name        string
		resource    map[string]any
		expected    string
		expectError bool
	}{
		{
			name: "extract primary email",
			resource: map[string]any{
				"userName": "m.buschner",
				"emails": []any{
					map[string]any{
						"primary": false,
						"value":   "other@example.com",
					},
					map[string]any{
						"primary": true,
						"value":   "m.buschner@first-colo.net",
					},
				},
			},
			expected:    "m.buschner@first-colo.net",
			expectError: false,
		},
		{
			name: "extract first email when no primary",
			resource: map[string]any{
				"userName": "m.buschner",
				"emails": []any{
					map[string]any{
						"value": "m.buschner@first-colo.net",
					},
					map[string]any{
						"value": "other@example.com",
					},
				},
			},
			expected:    "m.buschner@first-colo.net",
			expectError: false,
		},
		{
			name: "error when emails array missing",
			resource: map[string]any{
				"userName": "m.buschner",
			},
			expected:    "",
			expectError: true,
		},
		{
			name: "error when emails array empty",
			resource: map[string]any{
				"userName": "m.buschner",
				"emails":   []any{},
			},
			expected:    "",
			expectError: true,
		},
		{
			name: "extract single email",
			resource: map[string]any{
				"userName": "john",
				"emails": []any{
					map[string]any{
						"primary": true,
						"value":   "john@example.com",
					},
				},
			},
			expected:    "john@example.com",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := extractEmailFromResource(tt.resource)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestSCIMUserToResource_NonEmailUsername(t *testing.T) {
	tests := []struct {
		name     string
		input    *ssoreadyv1.SCIMUser
		expected map[string]any
	}{
		{
			name: "userName different from email",
			input: &ssoreadyv1.SCIMUser{
				Id:    "scim_user_123",
				Email: "m.buschner@first-colo.net",
				Attributes: mustNewStruct(map[string]any{
					"userName": "m.buschner",
					"active":   true,
					"emails": []any{
						map[string]any{
							"primary": true,
							"value":   "m.buschner@first-colo.net",
						},
					},
					"name": map[string]any{
						"familyName": "Buschner",
						"givenName":  "Martin",
					},
				}),
			},
			expected: map[string]any{
				"id":       "scim_user_123",
				"userName": "m.buschner", // userName preserved from attributes
				"active":   true,
				"emails": []any{
					map[string]any{
						"primary": true,
						"value":   "m.buschner@first-colo.net",
					},
				},
				"name": map[string]any{
					"familyName": "Buschner",
					"givenName":  "Martin",
				},
			},
		},
		{
			name: "userName not in attributes - fallback to email",
			input: &ssoreadyv1.SCIMUser{
				Id:    "scim_user_456",
				Email: "john@example.com",
				Attributes: mustNewStruct(map[string]any{
					"active": true,
				}),
			},
			expected: map[string]any{
				"id":       "scim_user_456",
				"userName": "john@example.com", // fallback to email
				"active":   true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scimUserToResource(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSCIMUserFromResource(t *testing.T) {
	tests := []struct {
		name           string
		resource       map[string]any
		expectedEmail  string
		expectedActive bool
	}{
		{
			name: "extract email from emails array with non-email username",
			resource: map[string]any{
				"userName": "m.buschner",
				"active":   true,
				"emails": []any{
					map[string]any{
						"primary": true,
						"value":   "m.buschner@first-colo.net",
					},
				},
			},
			expectedEmail:  "m.buschner@first-colo.net",
			expectedActive: true,
		},
		{
			name: "extract email from emails array with email-format username",
			resource: map[string]any{
				"userName": "john@example.com",
				"active":   true,
				"emails": []any{
					map[string]any{
						"primary": true,
						"value":   "john@different.com",
					},
				},
			},
			expectedEmail:  "john@different.com", // email from emails array takes precedence
			expectedActive: true,
		},
		{
			name: "fallback to userName when emails array missing - backward compatibility",
			resource: map[string]any{
				"userName": "john@example.com",
				"active":   false,
			},
			expectedEmail:  "john@example.com", // fallback to userName
			expectedActive: false,
		},
		{
			name: "extract first email when no primary marked",
			resource: map[string]any{
				"userName": "testuser",
				"active":   true,
				"emails": []any{
					map[string]any{
						"value": "first@example.com",
					},
					map[string]any{
						"value": "second@example.com",
					},
				},
			},
			expectedEmail:  "first@example.com",
			expectedActive: true,
		},
		{
			name: "handle string 'True' for active - Entra compatibility",
			resource: map[string]any{
				"userName": "user@example.com",
				"active":   "True",
				"emails": []any{
					map[string]any{
						"primary": true,
						"value":   "user@example.com",
					},
				},
			},
			expectedEmail:  "user@example.com",
			expectedActive: true,
		},
		{
			name: "handle string 'False' for active - Entra compatibility",
			resource: map[string]any{
				"userName": "user@example.com",
				"active":   "False",
				"emails": []any{
					map[string]any{
						"primary": true,
						"value":   "user@example.com",
					},
				},
			},
			expectedEmail:  "user@example.com",
			expectedActive: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scimUserFromResource("scim_dir_123", "scim_user_456", tt.resource)

			assert.Equal(t, "scim_user_456", result.Id)
			assert.Equal(t, "scim_dir_123", result.ScimDirectoryId)
			assert.Equal(t, tt.expectedEmail, result.Email)
			assert.Equal(t, !tt.expectedActive, result.Deleted) // Deleted is inverse of active
		})
	}
}

func TestEmailDomainExtraction(t *testing.T) {
	tests := []struct {
		name           string
		email          string
		expectedDomain string
		expectError    bool
	}{
		{
			name:           "lowercase domain",
			email:          "user@example.com",
			expectedDomain: "example.com",
			expectError:    false,
		},
		{
			name:           "mixed case domain - extracted as lowercase",
			email:          "alex@OmnifactGmbH.onmicrosoft.com",
			expectedDomain: "omnifactgmbh.onmicrosoft.com",
			expectError:    false,
		},
		{
			name:           "uppercase domain - extracted as lowercase",
			email:          "user@EXAMPLE.COM",
			expectedDomain: "example.com",
			expectError:    false,
		},
		{
			name:           "subdomain with mixed case",
			email:          "user@Mail.Example.COM",
			expectedDomain: "mail.example.com",
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			domain, err := emailaddr.Parse(tt.email)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedDomain, domain, "Domain should be extracted as lowercase")
			}
		})
	}
}

func TestEmailFormatDetection(t *testing.T) {
	tests := []struct {
		name        string
		value       string
		isEmailFmt  bool
		description string
	}{
		{
			name:        "valid email format",
			value:       "john@example.com",
			isEmailFmt:  true,
			description: "standard email should be detected as email format",
		},
		{
			name:        "valid email with subdomain",
			value:       "user@mail.example.com",
			isEmailFmt:  true,
			description: "email with subdomain should be detected as email format",
		},
		{
			name:        "valid email with plus",
			value:       "user+tag@example.com",
			isEmailFmt:  true,
			description: "email with plus sign should be detected as email format",
		},
		{
			name:        "non-email username - simple",
			value:       "m.buschner",
			isEmailFmt:  false,
			description: "username without @ should NOT be detected as email format",
		},
		{
			name:        "non-email username - with numbers",
			value:       "user123",
			isEmailFmt:  false,
			description: "alphanumeric username should NOT be detected as email format",
		},
		{
			name:        "non-email username - with underscore",
			value:       "john_doe",
			isEmailFmt:  false,
			description: "username with underscore should NOT be detected as email format",
		},
		{
			name:        "UUID-like username",
			value:       "081ed936-c63c-4660-865c-e9708c163555",
			isEmailFmt:  false,
			description: "UUID should NOT be detected as email format",
		},
		{
			name:        "empty string",
			value:       "",
			isEmailFmt:  false,
			description: "empty string should NOT be detected as email format",
		},
		{
			name:        "malformed email - missing domain",
			value:       "user@",
			isEmailFmt:  false,
			description: "malformed email should NOT be detected as email format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use the same emailaddr.Parse logic used in the filter code
			_, err := emailaddr.Parse(tt.value)
			isEmail := (err == nil)

			assert.Equal(t, tt.isEmailFmt, isEmail, tt.description)
		})
	}
}

func TestBackwardCompatibility(t *testing.T) {
	t.Run("old data where userName equals email in attributes", func(t *testing.T) {
		// Simulates old data where userName was always an email
		scimUser := &ssoreadyv1.SCIMUser{
			Id:    "scim_user_legacy",
			Email: "old.user@example.com",
			Attributes: mustNewStruct(map[string]any{
				"userName": "old.user@example.com", // In old data, userName == email
				"active":   true,
				"displayName": "Old User",
			}),
		}

		resource := scimUserToResource(scimUser)

		// Should return the userName from attributes (which happens to be an email)
		assert.Equal(t, "old.user@example.com", resource["userName"])
		assert.Equal(t, "scim_user_legacy", resource["id"])
		assert.Equal(t, true, resource["active"])
	})

	t.Run("old data where userName field is missing - uses email", func(t *testing.T) {
		// Simulates very old data where userName might not be in attributes
		scimUser := &ssoreadyv1.SCIMUser{
			Id:    "scim_user_very_old",
			Email: "very.old@example.com",
			Attributes: mustNewStruct(map[string]any{
				"active": true,
			}),
		}

		resource := scimUserToResource(scimUser)

		// Should fallback to email when userName not in attributes
		assert.Equal(t, "very.old@example.com", resource["userName"])
		assert.Equal(t, "scim_user_very_old", resource["id"])
	})

	t.Run("scimUserFromResource with old format - userName is email", func(t *testing.T) {
		resource := map[string]any{
			"userName": "legacy@example.com",
			"active":   true,
			"displayName": "Legacy User",
		}

		result := scimUserFromResource("scim_dir_123", "scim_user_789", resource)

		// Should fallback to userName when emails array is missing
		assert.Equal(t, "legacy@example.com", result.Email)
		assert.Equal(t, false, result.Deleted)
		
		// Verify userName is preserved in attributes
		attrs := result.Attributes.AsMap()
		assert.Equal(t, "legacy@example.com", attrs["userName"])
	})
}

func TestNonEmailUsernameScenarios(t *testing.T) {
	t.Run("create user with non-email username and emails array", func(t *testing.T) {
		resource := map[string]any{
			"userName": "m.buschner",
			"active":   true,
			"emails": []any{
				map[string]any{
					"primary": true,
					"value":   "m.buschner@first-colo.net",
				},
			},
			"name": map[string]any{
				"familyName": "Buschner",
				"givenName":  "Martin",
			},
			"externalId": "081ed936-c63c-4660-865c-e9708c163555",
		}

		// Extract email for storage
		email, err := extractEmailFromResource(resource)
		require.NoError(t, err)
		assert.Equal(t, "m.buschner@first-colo.net", email)

		// Simulate storage and retrieval
		scimUser := &ssoreadyv1.SCIMUser{
			Id:              "scim_user_new",
			ScimDirectoryId: "scim_dir_123",
			Email:           email,
			Deleted:         false,
			Attributes:      mustNewStruct(resource),
		}

		// Convert back to response format
		response := scimUserToResource(scimUser)

		// Verify response has correct userName (not email)
		assert.Equal(t, "m.buschner", response["userName"])
		assert.Equal(t, "scim_user_new", response["id"])
		
		// Verify emails array is preserved
		emails, ok := response["emails"].([]any)
		require.True(t, ok)
		assert.Len(t, emails, 1)
		
		firstEmail := emails[0].(map[string]any)
		assert.Equal(t, "m.buschner@first-colo.net", firstEmail["value"])
		assert.Equal(t, true, firstEmail["primary"])
	})

	t.Run("username can be anything - not just valid identifiers", func(t *testing.T) {
		testCases := []struct {
			userName string
			email    string
		}{
			{"user@123", "user@example.com"},      // @ but not valid email
			{"first.last", "first.last@corp.com"}, // dots
			{"user-name", "user@example.com"},     // hyphens
			{"user_name", "user@example.com"},     // underscores
			{"123456", "numeric@example.com"},     // pure numbers
			{"用户名", "user@example.com"},          // unicode
		}

		for _, tc := range testCases {
			t.Run(tc.userName, func(t *testing.T) {
				scimUser := &ssoreadyv1.SCIMUser{
					Id:    "test_user",
					Email: tc.email,
					Attributes: mustNewStruct(map[string]any{
						"userName": tc.userName,
						"emails": []any{
							map[string]any{"primary": true, "value": tc.email},
						},
					}),
				}

				response := scimUserToResource(scimUser)
				assert.Equal(t, tc.userName, response["userName"], "userName should be preserved exactly as stored")
			})
		}
	})
}
