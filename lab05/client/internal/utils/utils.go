package utils

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
)

func RandomBytes(size int) []byte {
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return b
}

func RandomHex(size int) string {
	b := RandomBytes(size)
	return hex.EncodeToString(b)
}

func RandomBase64URL(size int) string {
	b := RandomBytes(size)
	return base64.RawURLEncoding.EncodeToString(b)
}
