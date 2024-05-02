package statesign

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/nacl/auth"
)

type Signer struct {
	Key [32]byte
}

type Data struct {
	SAMLLoginEventID string
	State            string
}

func (signer *Signer) Encode(d Data) string {
	payload := fmt.Sprintf("%s.%s", d.SAMLLoginEventID, base64.RawURLEncoding.EncodeToString([]byte(d.State)))
	digest := auth.Sum([]byte(payload), &signer.Key)
	return fmt.Sprintf("%s.%s", base64.RawURLEncoding.EncodeToString([]byte(payload)), base64.RawURLEncoding.EncodeToString(digest[:]))
}

func (signer *Signer) Decode(s string) (*Data, error) {
	payloadBase64, digestBase64, ok := strings.Cut(s, ".")
	if !ok {
		return nil, fmt.Errorf("invalid signature: missing '.'")
	}

	payload, err := base64.RawURLEncoding.DecodeString(payloadBase64)
	if err != nil {
		return nil, fmt.Errorf("invalid signature: parse payload: %w", err)
	}

	digest, err := base64.RawURLEncoding.DecodeString(digestBase64)
	if err != nil {
		return nil, fmt.Errorf("invalid signature: parse digest: %w", err)
	}

	if !auth.Verify(digest, payload, &signer.Key) {
		return nil, fmt.Errorf("invalid signature: digest mismatch")
	}

	samlLoginEventID, stateBase64, ok := bytes.Cut(payload, []byte("."))
	if !ok {
		return nil, fmt.Errorf("invalid signature: missing '.' in payload")
	}

	state, err := base64.RawURLEncoding.DecodeString(string(stateBase64))
	if err != nil {
		return nil, fmt.Errorf("invalid signature: parse state: %w", err)
	}

	return &Data{SAMLLoginEventID: string(samlLoginEventID), State: string(state)}, nil
}
