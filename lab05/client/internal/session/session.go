package session

import (
	"github.com/gin-contrib/sessions"
	"golang.org/x/oauth2"

	"github.com/cyllective/oauth-labs/lab05/client/internal/config"
)

func IsAuthenticated(s sessions.Session) bool {
	_, atOK := GetString(s, "access_token")
	_, rtOK := GetString(s, "refresh_token")
	return atOK && rtOK
}

func GetString(s sessions.Session, key string) (string, bool) {
	value := s.Get(key)
	if value == nil {
		return "", false
	}
	if _, ok := value.(string); !ok {
		return "", false
	}
	return value.(string), true
}

func Delete(s sessions.Session) {
	opts := config.GetSessionOptions()
	opts.MaxAge = -1
	s.Set("access_token", false)
	s.Set("refresh_token", false)
	s.Options(opts)
	s.Clear()
	_ = s.Save()
}

func SetTokens(s sessions.Session, tokens *oauth2.Token) error {
	s.Set("access_token", tokens.AccessToken)
	s.Set("refresh_token", tokens.RefreshToken)
	return s.Save()
}
