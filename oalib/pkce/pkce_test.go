package pkce

import (
	"crypto/sha256"
	"encoding/base64"
	"testing"
)

func TestPKCEPlain(t *testing.T) {
	verifier := "foobar"
	challenge := "foobar"
	ok, err := Verify("plain", challenge, verifier)
	if err != nil || !ok {
		t.FailNow()
	}
}

func TestPKCES256(t *testing.T) {
	verifier := "foobar"
	sha := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(sha[:])
	ok, err := Verify("S256", challenge, verifier)
	if err != nil || !ok {
		t.FailNow()
	}
}
