package scimpatch_test

import (
	"testing"

	"github.com/ssoready/ssoready/internal/scimpatch"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPatch(t *testing.T) {
	testCases := []struct {
		name string
		in   map[string]any
		ops  []scimpatch.Operation
		out  map[string]any
		err  string
	}{
		{
			name: "replace entire value",
			in:   map[string]any{"foo": "xxx"},
			ops:  []scimpatch.Operation{{Op: "replace", Path: "", Value: map[string]any{"bar": "yyy"}}},
			out:  map[string]any{"bar": "yyy"},
		},
		{
			name: "replace entire value with non-object",
			in:   map[string]any{"foo": "xxx"},
			ops:  []scimpatch.Operation{{Op: "replace", Path: "", Value: "notanobject"}},
			err:  "top-level 'replace' operation must have an object value",
		},
		{
			name: "replace top-level prop",
			in:   map[string]any{"foo": "xxx"},
			ops:  []scimpatch.Operation{{Op: "replace", Path: "foo", Value: "yyy"}},
			out:  map[string]any{"foo": "yyy"},
		},
		{
			name: "replace nested prop",
			in:   map[string]any{"foo": map[string]any{"bar": "xxx"}},
			ops:  []scimpatch.Operation{{Op: "replace", Path: "foo.bar", Value: "yyy"}},
			out:  map[string]any{"foo": map[string]any{"bar": "yyy"}},
		},
		{
			name: "replace map prop",
			in:   map[string]any{"foo": map[string]any{"bar": "xxx"}},
			ops:  []scimpatch.Operation{{Op: "replace", Path: "foo", Value: map[string]any{"bar": "yyy"}}},
			out:  map[string]any{"foo": map[string]any{"bar": "yyy"}},
		},
		{
			name: "replace scalar with map",
			in:   map[string]any{"foo": map[string]any{"bar": "xxx"}},
			ops:  []scimpatch.Operation{{Op: "replace", Path: "foo.bar", Value: map[string]any{"baz": "yyy"}}},
			out:  map[string]any{"foo": map[string]any{"bar": map[string]any{"baz": "yyy"}}},
		},
		{
			name: "add to slice",
			in:   map[string]any{"foo": []any{"xxx"}},
			ops:  []scimpatch.Operation{{Op: "add", Path: "foo", Value: []any{"yyy"}}},
			out:  map[string]any{"foo": []any{"xxx", "yyy"}},
		},
		{
			name: "add multiple to slice", // this is inferred from spec; unclear if used in the wild
			in:   map[string]any{"foo": []any{"xxx"}},
			ops:  []scimpatch.Operation{{Op: "add", Path: "foo", Value: []any{"yyy", "zzz"}}},
			out:  map[string]any{"foo": []any{"xxx", "yyy", "zzz"}},
		},
		{
			name: "add to empty property",
			in:   map[string]any{},
			ops:  []scimpatch.Operation{{Op: "add", Path: "foo", Value: "yyy"}},
			out:  map[string]any{"foo": "yyy"},
		},
		{
			name: "add to sub-object",
			in:   map[string]any{"foo": map[string]any{"bar": "xxx"}},
			ops:  []scimpatch.Operation{{Op: "add", Path: "foo", Value: map[string]any{"baz": "yyy"}}},
			out:  map[string]any{"foo": map[string]any{"bar": "xxx", "baz": "yyy"}},
		},
		{
			name: "add to top-level-object",
			in:   map[string]any{"foo": map[string]any{"bar": "xxx"}},
			ops:  []scimpatch.Operation{{Op: "add", Path: "", Value: map[string]any{"baz": "yyy"}}},
			err:  "unsupported 'add' operation on top-level object",
		},
		{
			name: "add creates intermediate object even when sibling exists",
			in:   map[string]any{"foo": map[string]any{"bar": map[string]any{"baz": "xxx"}}},
			ops:  []scimpatch.Operation{{Op: "add", Path: "foo.notbar.baz", Value: map[string]any{"baz": "yyy"}}},
			out: map[string]any{"foo": map[string]any{
				"bar":    map[string]any{"baz": "xxx"},
				"notbar": map[string]any{"baz": map[string]any{"baz": "yyy"}},
			}},
		},

		{
			name: "uppercase Replace op",
			in:   map[string]any{"foo": "xxx"},
			ops:  []scimpatch.Operation{{Op: "Replace", Path: "", Value: map[string]any{"bar": "yyy"}}},
			out:  map[string]any{"bar": "yyy"},
		},
		{
			name: "uppercase Add op",
			in:   map[string]any{"foo": []any{"xxx"}},
			ops:  []scimpatch.Operation{{Op: "Add", Path: "foo", Value: []any{"yyy"}}},
			out:  map[string]any{"foo": []any{"xxx", "yyy"}},
		},

		{
			name: "invalid op",
			in:   map[string]any{"foo": []any{"xxx"}},
			ops:  []scimpatch.Operation{{Op: "BadOp", Path: "foo", Value: []any{"yyy"}}},
			err:  `unsupported SCIM PATCH operation: "BadOp"`,
		},

		{
			name: "special-case for entra patches on enterprise user",
			in: map[string]any{
				"urn:ietf:params:scim:schemas:extension:enterprise:2.0:User": map[string]any{
					"foo": "xxx",
				},
			},
			ops: []scimpatch.Operation{
				{
					Op:    "Add",
					Path:  "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User:bar",
					Value: "yyy",
				},
			},
			out: map[string]any{
				"urn:ietf:params:scim:schemas:extension:enterprise:2.0:User": map[string]any{
					"foo": "xxx",
					"bar": "yyy",
				},
			},
		},
		{
			// inferred behavior; not seen in wild -- case where there's no sub-":" in the path
			name: "special-case for entra patches on enterprise user",
			in:   map[string]any{},
			ops: []scimpatch.Operation{
				{
					Op:    "Add",
					Path:  "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User",
					Value: map[string]any{"foo": "xxx"},
				},
			},
			out: map[string]any{
				"urn:ietf:params:scim:schemas:extension:enterprise:2.0:User": map[string]any{
					"foo": "xxx",
				},
			},
		},
		{
			name: "specific err for non-object Add to enterprise user",
			in: map[string]any{
				"urn:ietf:params:scim:schemas:extension:enterprise:2.0:User": map[string]any{
					"foo": "xxx",
				},
			},
			ops: []scimpatch.Operation{
				{
					Op:    "Add",
					Path:  "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User",
					Value: "notanobject",
				},
			},
			err: "enterprise user schema value must be an object",
		},

		{
			name: "filter expression on non-array",
			in: map[string]any{
				"items": "notanarray",
			},
			ops: []scimpatch.Operation{{Op: "Replace", Path: "items[type eq \"bar\"]", Value: "zzz"}},
			err: `invalid path: not an array: items[type eq "bar"]`,
		},
		{
			name: "filter expression on array containing non-object",
			in: map[string]any{
				"items": []any{
					map[string]any{
						"type": "foo",
						"str":  "xxx",
					},
					"notanobject",
				},
			},
			ops: []scimpatch.Operation{{Op: "Replace", Path: "items[type eq \"bar\"]", Value: "zzz"}},
			err: `invalid path: applied filter on array containing non-object: items[type eq "bar"]`,
		},
		{
			name: "replace with filter expression in path",
			in: map[string]any{
				"items": []any{
					map[string]any{
						"type": "foo",
						"str":  "xxx",
					},
					map[string]any{
						"type": "bar",
						"str":  "yyy",
					},
				},
			},
			ops: []scimpatch.Operation{{Op: "Replace", Path: "items[type eq \"bar\"].str", Value: "zzz"}},
			out: map[string]any{
				"items": []any{
					map[string]any{
						"type": "foo",
						"str":  "xxx",
					},
					map[string]any{
						"type": "bar",
						"str":  "zzz",
					},
				},
			},
		},
		{
			name: "replace entire object with filter expression",
			in: map[string]any{
				"items": []any{
					map[string]any{
						"type": "foo",
						"str":  "xxx",
					},
					map[string]any{
						"type": "bar",
						"str":  "yyy",
					},
				},
			},
			ops: []scimpatch.Operation{{Op: "Replace", Path: "items[type eq \"bar\"]", Value: map[string]any{
				"type": "baz",
				"str":  "zzz",
			}}},
			out: map[string]any{
				"items": []any{
					map[string]any{
						"type": "foo",
						"str":  "xxx",
					},
					map[string]any{
						"type": "baz",
						"str":  "zzz",
					},
				},
			},
		},
		{
			name: "replace entire object with filter expression",
			in: map[string]any{
				"items": []any{
					map[string]any{
						"type": "foo",
						"str":  "xxx",
					},
					map[string]any{
						"type": "bar",
						"str":  "yyy",
					},
				},
			},
			ops: []scimpatch.Operation{{Op: "Replace", Path: "items[type eq \"bar\"]", Value: map[string]any{
				"type": "baz",
				"str":  "zzz",
			}}},
			out: map[string]any{
				"items": []any{
					map[string]any{
						"type": "foo",
						"str":  "xxx",
					},
					map[string]any{
						"type": "baz",
						"str":  "zzz",
					},
				},
			},
		},
		{
			name: "replace with not-equal filter expression",
			in: map[string]any{
				"items": []any{
					map[string]any{
						"type": "foo",
						"str":  "xxx",
					},
					map[string]any{
						"type": "bar",
						"str":  "yyy",
					},
					map[string]any{
						"type": "baz",
						"str":  "zzz",
					},
				},
			},
			ops: []scimpatch.Operation{{Op: "Replace", Path: "items[type ne \"foo\"].str", Value: "aaa"}},
			out: map[string]any{
				"items": []any{
					map[string]any{
						"type": "foo",
						"str":  "xxx",
					},
					map[string]any{
						"type": "bar",
						"str":  "aaa",
					},
					map[string]any{
						"type": "baz",
						"str":  "aaa",
					},
				},
			},
		},
		{
			name: "replace entire object with not-equal filter expression",
			in: map[string]any{
				"items": []any{
					map[string]any{
						"type": "foo",
						"str":  "xxx",
					},
					map[string]any{
						"type": "bar",
						"str":  "yyy",
					},
					map[string]any{
						"type": "baz",
						"str":  "zzz",
					},
				},
			},
			ops: []scimpatch.Operation{{Op: "Replace", Path: "items[type ne \"foo\"]", Value: map[string]any{
				"type": "aaa",
				"str":  "bbb",
			}}},
			out: map[string]any{
				"items": []any{
					map[string]any{
						"type": "foo",
						"str":  "xxx",
					},
					map[string]any{
						"type": "aaa",
						"str":  "bbb",
					},
					map[string]any{
						"type": "aaa",
						"str":  "bbb",
					},
				},
			},
		},
		{
			name: "replace with contains filter expression",
			in: map[string]any{
				"items": []any{
					map[string]any{
						"type": "foo",
						"str":  "xxx_abc",
					},
					map[string]any{
						"type": "bar",
						"str":  "yyy_abc",
					},
					map[string]any{
						"type": "baz",
						"str":  "zzz",
					},
				},
			},
			ops: []scimpatch.Operation{{Op: "Replace", Path: "items[str co \"abc\"].type", Value: "aaa"}},
			out: map[string]any{
				"items": []any{
					map[string]any{
						"type": "aaa",
						"str":  "xxx_abc",
					},
					map[string]any{
						"type": "aaa",
						"str":  "yyy_abc",
					},
					map[string]any{
						"type": "baz",
						"str":  "zzz",
					},
				},
			},
		},
		{
			name: "contain operator with non-string",
			in: map[string]any{
				"items": []any{
					map[string]any{
						"type": "foo",
						"val":  true,
					},
				},
			},
			ops: []scimpatch.Operation{{Op: "Replace", Path: "items[val co \"whatever\"].type", Value: "xxx"}},
			err: "'co' operator can only be used with string values",
		},
		{
			name: "replace with starts-with filter expression",
			in: map[string]any{
				"items": []any{
					map[string]any{
						"type": "foo",
						"str":  "xxx_abc",
					},
					map[string]any{
						"type": "bar",
						"str":  "xxx_def",
					},
					map[string]any{
						"type": "baz",
						"str":  "yyy_abc",
					},
				},
			},
			ops: []scimpatch.Operation{{Op: "Replace", Path: "items[str sw \"xxx\"].type", Value: "aaa"}},
			out: map[string]any{
				"items": []any{
					map[string]any{
						"type": "aaa",
						"str":  "xxx_abc",
					},
					map[string]any{
						"type": "aaa",
						"str":  "xxx_def",
					},
					map[string]any{
						"type": "baz",
						"str":  "yyy_abc",
					},
				},
			},
		},
		{
			name: "starts-with operator with non-string",
			in: map[string]any{
				"items": []any{
					map[string]any{
						"type": "foo",
						"val":  true,
					},
				},
			},
			ops: []scimpatch.Operation{{Op: "Replace", Path: "items[val sw \"whatever\"].type", Value: "xxx"}},
			err: "'sw' operator can only be used with string values",
		},
		{
			name: "replace with ends-with filter expression",
			in: map[string]any{
				"items": []any{
					map[string]any{
						"type": "foo",
						"str":  "xxx_abc",
					},
					map[string]any{
						"type": "bar",
						"str":  "yyy_def",
					},
					map[string]any{
						"type": "baz",
						"str":  "zzz_abc",
					},
				},
			},
			ops: []scimpatch.Operation{{Op: "Replace", Path: "items[str ew \"abc\"].type", Value: "aaa"}},
			out: map[string]any{
				"items": []any{
					map[string]any{
						"type": "aaa",
						"str":  "xxx_abc",
					},
					map[string]any{
						"type": "bar",
						"str":  "yyy_def",
					},
					map[string]any{
						"type": "aaa",
						"str":  "zzz_abc",
					},
				},
			},
		},
		{
			name: "ends-with operator with non-string",
			in: map[string]any{
				"items": []any{
					map[string]any{
						"type": "foo",
						"val":  true,
					},
				},
			},
			ops: []scimpatch.Operation{{Op: "Replace", Path: "items[val ew \"whatever\"].type", Value: "xxx"}},
			err: "'ew' operator can only be used with string values",
		},
		{
			name: "replace with present filter expression",
			in: map[string]any{
				"items": []any{
					map[string]any{
						"type": "foo",
						"str":  "xxx",
					},
					map[string]any{
						"type": "bar",
					},
					map[string]any{
						"type": "baz",
						"str":  "",
					},
					map[string]any{
						"type": "qux",
						"str":  nil,
					},
					map[string]any{
						"type": "qux",
						"str":  "zzz",
					},
				},
			},
			ops: []scimpatch.Operation{{Op: "Replace", Path: "items[str pr].type", Value: "aaa"}},
			out: map[string]any{
				"items": []any{
					map[string]any{
						"type": "aaa",
						"str":  "xxx",
					},
					map[string]any{
						"type": "bar",
					},
					map[string]any{
						"type": "baz",
						"str":  "",
					},
					map[string]any{
						"type": "qux",
						"str":  nil,
					},
					map[string]any{
						"type": "aaa",
						"str":  "zzz",
					},
				},
			},
		},
		{
			name: "replace with greater than filter expression on strings",
			in: map[string]any{
				"items": []any{
					map[string]any{
						"type": "foo",
						"str":  "aaa",
					},
					map[string]any{
						"type": "bar",
						"str":  "mmm",
					},
					map[string]any{
						"type": "baz",
						"str":  "zzz",
					},
				},
			},
			ops: []scimpatch.Operation{{Op: "Replace", Path: "items[str gt \"mmm\"].type", Value: "xxx"}},
			out: map[string]any{
				"items": []any{
					map[string]any{
						"type": "foo",
						"str":  "aaa",
					},
					map[string]any{
						"type": "bar",
						"str":  "mmm",
					},
					map[string]any{
						"type": "xxx",
						"str":  "zzz",
					},
				},
			},
		},
		{
			name: "replace with greater than or equal filter expression on strings",
			in: map[string]any{
				"items": []any{
					map[string]any{
						"type": "foo",
						"str":  "aaa",
					},
					map[string]any{
						"type": "bar",
						"str":  "mmm",
					},
					map[string]any{
						"type": "baz",
						"str":  "zzz",
					},
				},
			},
			ops: []scimpatch.Operation{{Op: "Replace", Path: "items[str ge \"mmm\"].type", Value: "xxx"}},
			out: map[string]any{
				"items": []any{
					map[string]any{
						"type": "foo",
						"str":  "aaa",
					},
					map[string]any{
						"type": "xxx",
						"str":  "mmm",
					},
					map[string]any{
						"type": "xxx",
						"str":  "zzz",
					},
				},
			},
		},
		{
			name: "replace with less than filter expression on strings",
			in: map[string]any{
				"items": []any{
					map[string]any{
						"type": "foo",
						"str":  "aaa",
					},
					map[string]any{
						"type": "bar",
						"str":  "mmm",
					},
					map[string]any{
						"type": "baz",
						"str":  "zzz",
					},
				},
			},
			ops: []scimpatch.Operation{{Op: "Replace", Path: "items[str lt \"mmm\"].type", Value: "xxx"}},
			out: map[string]any{
				"items": []any{
					map[string]any{
						"type": "xxx",
						"str":  "aaa",
					},
					map[string]any{
						"type": "bar",
						"str":  "mmm",
					},
					map[string]any{
						"type": "baz",
						"str":  "zzz",
					},
				},
			},
		},
		{
			name: "replace with less than or equal filter expression on strings",
			in: map[string]any{
				"items": []any{
					map[string]any{
						"type": "foo",
						"str":  "aaa",
					},
					map[string]any{
						"type": "bar",
						"str":  "mmm",
					},
					map[string]any{
						"type": "baz",
						"str":  "zzz",
					},
				},
			},
			ops: []scimpatch.Operation{{Op: "Replace", Path: "items[str le \"mmm\"].type", Value: "xxx"}},
			out: map[string]any{
				"items": []any{
					map[string]any{
						"type": "xxx",
						"str":  "aaa",
					},
					map[string]any{
						"type": "xxx",
						"str":  "mmm",
					},
					map[string]any{
						"type": "baz",
						"str":  "zzz",
					},
				},
			},
		},
		{
			name: "replace with comparison operators on numbers",
			in: map[string]any{
				"items": []any{
					map[string]any{
						"type": "foo",
						"num":  10.0,
					},
					map[string]any{
						"type": "bar",
						"num":  20.0,
					},
					map[string]any{
						"type": "baz",
						"num":  30.0,
					},
				},
			},
			ops: []scimpatch.Operation{
				{Op: "Replace", Path: "items[num gt \"20\"].type", Value: "xxx"},
				{Op: "Replace", Path: "items[num lt \"20\"].type", Value: "yyy"},
			},
			out: map[string]any{
				"items": []any{
					map[string]any{
						"type": "yyy",
						"num":  10.0,
					},
					map[string]any{
						"type": "bar",
						"num":  20.0,
					},
					map[string]any{
						"type": "xxx",
						"num":  30.0,
					},
				},
			},
		},
		{
			name: "invalid number in comparison filter",
			in: map[string]any{
				"items": []any{
					map[string]any{
						"type": "foo",
						"num":  10.0,
					},
					map[string]any{
						"type": "bar",
						"num":  20.0,
					},
					map[string]any{
						"type": "baz",
						"num":  30.0,
					},
				},
			},
			ops: []scimpatch.Operation{
				{Op: "Replace", Path: "items[num gt \"badnumber\"].type", Value: "xxx"},
			},
			err: `invalid number in comparison: "badnumber"`,
		},
		{
			name: "comparison operators with invalid types should fail",
			in: map[string]any{
				"items": []any{
					map[string]any{
						"type": "foo",
						"val":  true,
					},
				},
			},
			ops: []scimpatch.Operation{{Op: "Replace", Path: "items[val gt \"true\"].type", Value: "xxx"}},
			err: "comparison operators can only be used with string or numeric values",
		},
		{
			name: "replace with comparison operators on dates",
			in: map[string]any{
				"items": []any{
					map[string]any{
						"type": "foo",
						"date": "2024-01-01T00:00:00Z",
					},
					map[string]any{
						"type": "bar",
						"date": "2024-02-01T00:00:00Z",
					},
					map[string]any{
						"type": "baz",
						"date": "2024-03-01T00:00:00Z",
					},
				},
			},
			ops: []scimpatch.Operation{
				{Op: "Replace", Path: "items[date gt \"2024-02-01T00:00:00Z\"].type", Value: "xxx"},
				{Op: "Replace", Path: "items[date lt \"2024-02-01T00:00:00Z\"].type", Value: "yyy"},
			},
			out: map[string]any{
				"items": []any{
					map[string]any{
						"type": "yyy",
						"date": "2024-01-01T00:00:00Z",
					},
					map[string]any{
						"type": "bar",
						"date": "2024-02-01T00:00:00Z",
					},
					map[string]any{
						"type": "xxx",
						"date": "2024-03-01T00:00:00Z",
					},
				},
			},
		},
		{
			name: "unsupported filter operator",
			in: map[string]any{
				"items": []any{
					map[string]any{
						"type": "foo",
						"val":  "whatever",
					},
				},
			},
			ops: []scimpatch.Operation{{Op: "Replace", Path: "items[val unknown \"alsowhatever\"].type", Value: "xxx"}},
			err: `unsupported filter operator: "unknown"`,
		},
		{
			name: "replace with filter in enterprise user path",
			in: map[string]any{
				"urn:ietf:params:scim:schemas:extension:enterprise:2.0:User": map[string]any{
					"items": []any{
						map[string]any{
							"type": "foo",
							"str":  "xxx",
						},
						map[string]any{
							"type": "bar",
							"str":  "yyy",
						},
					},
				},
			},
			ops: []scimpatch.Operation{{Op: "Replace", Path: "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User:items[type eq \"bar\"].str", Value: "zzz"}},
			out: map[string]any{
				"urn:ietf:params:scim:schemas:extension:enterprise:2.0:User": map[string]any{
					"items": []any{
						map[string]any{
							"type": "foo",
							"str":  "xxx",
						},
						map[string]any{
							"type": "bar",
							"str":  "zzz",
						},
					},
				},
			},
		},
		{
			name: "create enterprise user schema with department using Add when schema doesn't exist",
			in:   map[string]any{},
			ops: []scimpatch.Operation{
				{
					Op:    "Add",
					Path:  "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User:department",
					Value: "Engineering",
				},
			},
			out: map[string]any{
				"urn:ietf:params:scim:schemas:extension:enterprise:2.0:User": map[string]any{
					"department": "Engineering",
				},
			},
		},
		{
			name: "should fail when trying to replace enterprise user schema without a field",
			in:   map[string]any{},
			ops: []scimpatch.Operation{
				{
					Op:    "Replace",
					Path:  "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User",
					Value: map[string]any{},
				},
			},
			err: "invalid path: \"urn:ietf:params:scim:schemas:extension:enterprise:2.0:User\"",
		},

		{
			name: "add filter with eq creates a new field",
			in: map[string]any{
				"foo": "bar",
			},
			ops: []scimpatch.Operation{
				{
					Op:    "Add",
					Path:  `emails[type eq "work"].value`,
					Value: "aaa",
				},
				{
					Op:    "Add",
					Path:  `addresses[type eq "work"].region`,
					Value: "bbb",
				},
				{
					Op:    "Add",
					Path:  `addresses[type eq "work"].country`,
					Value: "ccc",
				},
				{
					Op:    "Add",
					Path:  `phoneNumbers[type eq "work"].value`,
					Value: "ddd",
				},
				{
					Op:    "Add",
					Path:  `phoneNumbers[type eq "mobile"].value`,
					Value: "eee",
				},
			},
			out: map[string]any{
				"foo": "bar",
				"emails": []any{
					map[string]any{
						"type":  "work",
						"value": "aaa",
					},
				},
				"addresses": []any{
					map[string]any{
						"type":    "work",
						"region":  "bbb",
						"country": "ccc",
					},
				},
				"phoneNumbers": []any{
					map[string]any{
						"type":  "work",
						"value": "ddd",
					},
					map[string]any{
						"type":  "mobile",
						"value": "eee",
					},
				},
			},
		},
		{
			name: "add to nested property creates parent objects automatically",
			in:   map[string]any{},
			ops: []scimpatch.Operation{
				{
					Op:    "Add",
					Path:  "name.givenName",
					Value: "John",
				},
				{
					Op:    "Add",
					Path:  "name.familyName",
					Value: "Doe",
				},
				{
					Op:    "Add",
					Path:  "name.formatted",
					Value: "John Doe",
				},
			},
			out: map[string]any{
				"name": map[string]any{
					"givenName":  "John",
					"familyName": "Doe",
					"formatted":  "John Doe",
				},
			},
		},
		{
			name: "add to deeply nested property creates all parent objects",
			in:   map[string]any{},
			ops: []scimpatch.Operation{
				{
					Op:    "Add",
					Path:  "level1.level2.level3.value",
					Value: "deep",
				},
			},
			out: map[string]any{
				"level1": map[string]any{
					"level2": map[string]any{
						"level3": map[string]any{
							"value": "deep",
						},
					},
				},
			},
		},
		{
			name: "replace on nested property still requires parent to exist",
			in:   map[string]any{},
			ops: []scimpatch.Operation{
				{
					Op:    "Replace",
					Path:  "name.givenName",
					Value: "John",
				},
			},
			err: `invalid path: "name.givenName"`,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			err := scimpatch.Patch(tt.ops, &tt.in)
			if tt.err != "" {
				require.Error(t, err)
				assert.Equal(t, tt.err, err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.out, tt.in)
			}
		})
	}
}
