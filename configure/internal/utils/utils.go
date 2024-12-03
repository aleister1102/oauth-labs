package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/google/uuid"
)

var (
	charset    = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	charsetLen = big.NewInt(62)
)

func RandomClientID() string {
	return uuid.NewString()
}

func RandomPassword(length int) string {
	b := make([]rune, length)
	for i := range b {
		n, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			panic(err)
		}
		b[i] = charset[n.Int64()]
	}
	return string(b)
}

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

func NewRSAPrivateKey() string {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}
	keyPEM := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(key),
		},
	)
	return string(keyPEM)
}

func GetLabRoot() string {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}

	labRoot, err := filepath.Abs(filepath.Join(ex, "..", "..", ".."))
	if err != nil {
		panic(err)
	}
	return labRoot
}

func WriteTemplateConfig(file string, templ *template.Template, data any) error {
	fh, err := os.OpenFile(file, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		panic(err)
	}
	defer fh.Close()
	if err := templ.Execute(fh, data); err != nil {
		panic(err)
	}
	return nil
}

func IndentPEM(pem string, indentation int) string {
	indentNewline := fmt.Sprintf("\n%s", strings.Repeat(" ", indentation))
	return strings.ReplaceAll(pem, "\n", indentNewline)
}
