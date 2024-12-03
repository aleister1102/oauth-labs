package utils

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"math/big"
	"strings"

	"github.com/google/uuid"
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

func RandomUUIDBase64URL() string {
	uid := uuid.New()
	return base64.RawURLEncoding.EncodeToString(uid[:])
}

func GetUUID(id string) (string, error) {
	rawUUID, err := hex.DecodeString(strings.ReplaceAll(id, "-", ""))
	if err != nil {
		return "", err
	}
	uid, err := uuid.FromBytes(rawUUID)
	if err != nil {
		return "", err
	}
	return uid.String(), nil
}

func RandomPassword(length int) string {
	charset := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~")
	b := make([]rune, length)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			panic(err)
		}
		b[i] = charset[n.Int64()]
	}
	return string(b)
}
