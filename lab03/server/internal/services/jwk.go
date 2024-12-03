package services

import (
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"

	"github.com/cyllective/oauth-labs/lab03/server/internal/config"
)

type JWKService struct {
	publicKeys jwk.Set
	keys       jwk.Set
	keyid      string
}

func NewJWKService() *JWKService {
	keyid := "01924316-6800-4804-8332-a0730422e016"
	keys := loadKeys(keyid)
	publicKeys, err := jwk.PublicSetOf(keys)
	if err != nil {
		panic(err)
	}
	return &JWKService{publicKeys, keys, keyid}
}

func (j JWKService) PublicKeys() jwk.Set {
	return j.publicKeys
}

func (j JWKService) PrivateKey() jwk.Key {
	key, ok := j.keys.LookupKeyID(j.keyid)
	if !ok {
		panic("primary private key not found")
	}
	return key
}

func (j JWKService) PublicKey() jwk.Key {
	pk, err := j.PrivateKey().PublicKey()
	if err != nil {
		panic("primary public key not found")
	}
	return pk
}

func (j JWKService) Keys() jwk.Set {
	return j.keys
}

func loadKeys(keyid string) jwk.Set {
	keys := jwk.NewSet()
	private, err := jwk.ParseKey(config.GetJWTPrivateKey(), jwk.WithPEM(true))
	if err != nil {
		panic(err)
	}
	_ = private.Set(jwk.KeyIDKey, keyid)
	_ = private.Set(jwk.KeyUsageKey, jwk.ForSignature)
	_ = private.Set(jwk.KeyTypeKey, jwa.RSA)
	_ = private.Set(jwk.AlgorithmKey, jwa.RS256)
	public, err := private.PublicKey()
	if err != nil {
		panic(err)
	}
	_ = public.Set(jwk.KeyIDKey, keyid)
	_ = public.Set(jwk.KeyUsageKey, jwk.ForEncryption)
	_ = public.Set(jwk.KeyTypeKey, jwa.RSA)
	_ = public.Set(jwk.AlgorithmKey, jwa.RS256)
	_ = keys.AddKey(private)
	return keys
}
