package session

import (
	"github.com/gin-contrib/sessions"

	"github.com/cyllective/oauth-labs/lab04/server/internal/config"
)

func IsAuthenticated(s sessions.Session) bool {
	uid, ok := GetString(s, "user_id")
	return uid != "" && ok
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
	s.Set("user_id", false)
	s.Set("access_token", false)
	s.Options(opts)
	s.Clear()
	_ = s.Save()
}
