package scimpatch_test

import (
	"testing"

	"github.com/ssoready/ssoready/internal/scimpatch"
	"github.com/stretchr/testify/assert"
)

func TestPatch(t *testing.T) {
	testCases := []struct {
		name string
		in   map[string]any
		ops  []scimpatch.Operation
		out  map[string]any
	}{
		{
			name: "replace entire value",
			in:   map[string]any{"foo": "xxx"},
			ops:  []scimpatch.Operation{{Op: "replace", Path: "", Value: map[string]any{"bar": "yyy"}}},
			out:  map[string]any{"bar": "yyy"},
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
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			err := scimpatch.Patch(tt.ops, &tt.in)
			assert.NoError(t, err)
			assert.Equal(t, tt.out, tt.in)
		})
	}
}
