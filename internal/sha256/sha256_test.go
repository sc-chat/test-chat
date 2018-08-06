package sha256

import "testing"

func TestNewHash(t *testing.T) {
	str := "example"
	expectedHash := "50d858e0985ecc7f60418aaf0cc5ab587f42c2570a884095a9e8ccacd0f6545c"

	hash := NewHash(str)
	if expectedHash != hash {
		t.Errorf("Hash should be %s but got %s\n", expectedHash, hash)
	}
}
