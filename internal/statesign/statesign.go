package statesign

import (
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/nacl/auth"
)

type Signer struct {
	Key [32]byte
}

type Data struct {
	SAMLFlowID string
}

func (signer *Signer) Encode(d Data) string {
	digest := auth.Sum([]byte(d.SAMLFlowID), &signer.Key)
	return fmt.Sprintf("%s.%s", base64.RawURLEncoding.EncodeToString([]byte(d.SAMLFlowID)), base64.RawURLEncoding.EncodeToString(digest[:]))
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

	return &Data{SAMLFlowID: string(payload)}, nil
}
