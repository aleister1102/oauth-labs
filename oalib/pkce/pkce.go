package pkce

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
)

func Verify(method string, challenge string, verifier string) (bool, error) {
	switch method {
	case "plain":
		return verifier == challenge, nil
	case "S256":
		sha := sha256.Sum256([]byte(verifier))
		calcChallenge := base64.RawURLEncoding.EncodeToString(sha[:])
		return challenge == calcChallenge, nil
	}

	return false, errors.New("unsupported code_challenge_method")
}
