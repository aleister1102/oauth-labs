package redirecturi

import (
	"net/url"
	"strings"
)

type RedirectURI struct {
	uri *url.URL
}

func New(u string) (*RedirectURI, error) {
	uri, err := url.Parse(u)
	if err != nil {
		return nil, err
	}
	return &RedirectURI{uri: uri}, nil
}

func NewWithState(u string, state string) (*RedirectURI, error) {
	r, err := New(u)
	if err != nil {
		return nil, err
	}
	r.SetState(state)
	return r, nil
}

func NewWithError(u string, error string) (*RedirectURI, error) {
	r, err := New(u)
	if err != nil {
		return nil, err
	}
	r.SetError(error)
	return r, nil
}

func (r *RedirectURI) addQuery(key string, value string) {
	q := r.uri.Query()
	q.Del(key)
	q.Add(key, value)
	r.uri.RawQuery = q.Encode()
}

func (r *RedirectURI) GetQuery(key string) string {
	q := r.uri.Query()
	return q.Get(key)
}

func (r *RedirectURI) SetState(state string) {
	if strings.TrimSpace(state) == "" {
		return
	}
	r.addQuery("state", state)
}

func (r *RedirectURI) SetError(err string) {
	if strings.TrimSpace(err) == "" {
		return
	}
	r.addQuery("error", err)
}

func (r *RedirectURI) SetCode(code string) {
	if strings.TrimSpace(code) == "" {
		return
	}
	r.addQuery("code", code)
}

func (r *RedirectURI) String() string {
	return r.uri.String()
}

func (r *RedirectURI) URL() url.URL {
	return *r.uri
}
