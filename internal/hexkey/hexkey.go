package hexkey

import (
	"encoding/hex"
	"fmt"
)

func New(s string) ([32]byte, error) {
	if len(s) != 64 {
		return [32]byte{}, fmt.Errorf("key must have length 64")
	}

	b, err := hex.DecodeString(s)
	if err != nil {
		return [32]byte{}, fmt.Errorf("decode hex: %w", err)
	}

	if len(b) > 32 {
		return [32]byte{}, fmt.Errorf("key must encode 32 bytes")
	}

	var k [32]byte
	copy(k[:], b)
	return k, nil
}
