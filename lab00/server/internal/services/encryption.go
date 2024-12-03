package services

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
)

type CryptoService struct{}

func NewCryptoService() *CryptoService {
	return &CryptoService{}
}

func (c CryptoService) Encrypt(data []byte, key []byte) []byte {
	gcm := c.newGCM(key)
	nonce := c.newNonce(gcm)
	return gcm.Seal(nonce, nonce, data, nil)
}

func (c CryptoService) Decrypt(data []byte, key []byte) ([]byte, error) {
	gcm := c.newGCM(key)
	nonce := data[:gcm.NonceSize()]
	ciphertext := data[gcm.NonceSize():]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

func (c CryptoService) newGCM(key []byte) cipher.AEAD {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err)
	}
	return gcm
}

func (c CryptoService) newNonce(gcm cipher.AEAD) []byte {
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err)
	}
	return nonce
}
