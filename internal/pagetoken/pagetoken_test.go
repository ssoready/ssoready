package pagetoken_test

import (
	"reflect"
	"testing"

	"github.com/ssoready/ssoready/internal/pagetoken"
)

func TestEncoder(t *testing.T) {
	type data struct {
		Foo string
		Bar string
	}

	in := data{
		Foo: "foo",
		Bar: "bar",
	}

	e := pagetoken.Encoder{Secret: [32]byte{}}
	encoded := e.Marshal(in)

	var out data
	err := e.Unmarshal(encoded, &out)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	if !reflect.DeepEqual(in, out) {
		t.Fatalf("round-trip failure")
	}
}
