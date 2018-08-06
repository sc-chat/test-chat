package sha256

import (
	"crypto/sha256"
	"encoding/hex"
)

// NewHash returns sha256 hash of provided string
func NewHash(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}
