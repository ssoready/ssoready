package authservice

import (
	"regexp"
	"testing"

	ssoreadyv1 "github.com/ssoready/ssoready/internal/gen/ssoready/v1"
	"github.com/stretchr/testify/assert"
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

// Helper function to create structpb.Struct from map
func mustNewStruct(m map[string]any) *structpb.Struct {
	s, err := structpb.NewStruct(m)
	if err != nil {
		panic(err)
	}
	return s
}
