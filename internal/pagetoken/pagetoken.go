package pagetoken

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"io"

	"connectrpc.com/connect"
	"golang.org/x/crypto/nacl/secretbox"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
)

type Encoder struct {
	Secret [32]byte
}

func (e *Encoder) Marshal(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}

	var nonce [24]byte
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		panic(err)
	}

	d := secretbox.Seal(nonce[:], b, &nonce, &e.Secret)
	return base64.URLEncoding.EncodeToString(d)
}

func (e *Encoder) Unmarshal(s string, v any) error {
	if s == "" {
		return nil
	}

	d, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		return validateErr()
	}

	var nonce [24]byte
	copy(nonce[:], d[:24])

	b, ok := secretbox.Open(nil, d[24:], &nonce, &e.Secret)
	if !ok {
		return validateErr()
	}

	if err := json.Unmarshal(b, v); err != nil {
		return validateErr()
	}

	return nil
}

func validateErr() error {
	err := connect.NewError(connect.CodeInvalidArgument, nil)
	detail, _ := connect.NewErrorDetail(&errdetails.BadRequest_FieldViolation{
		Field:       "page_token",
		Description: "invalid pagination token",
	})
	err.AddDetail(detail)
	return err
}
