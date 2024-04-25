package stack_test

import (
	"testing"

	"github.com/ssoready/ssoready/internal/saml/uxml/stack"
	"github.com/stretchr/testify/assert"
)

func TestStack(t *testing.T) {
	var s stack.Stack

	assert.Equal(t, 0, s.Len())
	assertGet(t, &s, "unknown", "", false)
	assert.Equal(t, map[string]string{}, s.GetAll())

	s.Push(map[string]string{
		"foo": "http://example.com/foo",
		"bar": "http://example.com/bar",
		"baz": "http://example.com/baz",
	})

	assert.Equal(t, 1, s.Len())
	assertGet(t, &s, "foo", "http://example.com/foo", true)
	assertGet(t, &s, "bar", "http://example.com/bar", true)
	assertGet(t, &s, "unknown", "", false)
	assert.Equal(t, map[string]string{
		"foo": "http://example.com/foo",
		"bar": "http://example.com/bar",
		"baz": "http://example.com/baz",
	}, s.GetAll())

	s.Push(map[string]string{
		"foo": "http://example.com/foo/new",
		"bar": "http://example.com/bar",
	})

	assert.Equal(t, 2, s.Len())
	assertGet(t, &s, "foo", "http://example.com/foo/new", true)
	assertGet(t, &s, "bar", "http://example.com/bar", true)
	assertGet(t, &s, "unknown", "", false)
	assert.Equal(t, map[string]string{
		"foo": "http://example.com/foo/new",
		"bar": "http://example.com/bar",
		"baz": "http://example.com/baz",
	}, s.GetAll())

	s.Pop()

	assert.Equal(t, 1, s.Len())
	assertGet(t, &s, "foo", "http://example.com/foo", true)
	assertGet(t, &s, "bar", "http://example.com/bar", true)
	assertGet(t, &s, "unknown", "", false)
	assert.Equal(t, map[string]string{
		"foo": "http://example.com/foo",
		"bar": "http://example.com/bar",
		"baz": "http://example.com/baz",
	}, s.GetAll())

	s.Pop()

	assert.Equal(t, 0, s.Len())
	assert.Equal(t, map[string]string{}, s.GetAll())
}

func assertGet(t *testing.T, s *stack.Stack, k, v string, ok bool) {
	actualV, actualOk := s.Get(k)
	assert.Equal(t, v, actualV)
	assert.Equal(t, ok, actualOk)
}
