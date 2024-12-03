package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"golang.org/x/oauth2"

	"github.com/cyllective/oauth-labs/lab03/client/internal/config"
	"github.com/cyllective/oauth-labs/lab03/client/internal/constants"
	"github.com/cyllective/oauth-labs/lab03/client/internal/dto"
	ierrors "github.com/cyllective/oauth-labs/lab03/client/internal/errors"
)

type APIClient struct {
	ctx       context.Context
	config    *oauth2.Config
	client    *http.Client
	baseURL   string
	userAgent string
}

func NewAPIClient(ctx context.Context, tokens *oauth2.Token) *APIClient {
	cfg := config.Get()
	client := &http.Client{Timeout: time.Duration(3) * time.Second}
	if proxyURL := os.Getenv("BURP_PROXY"); proxyURL != "" {
		proxy, _ := url.Parse(proxyURL)
		client.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxy),
		}
	}
	oacfg := config.GetOAuthConfig()
	oactx := context.WithValue(ctx, oauth2.HTTPClient, client)
	client = oauth2.NewClient(oactx, oacfg.TokenSource(oactx, tokens))
	return &APIClient{
		ctx:       oactx,
		config:    oacfg,
		client:    client,
		baseURL:   cfg.GetString("resource_server.base_url"),
		userAgent: fmt.Sprintf("oauth-labs/client-%s", constants.LabNumber),
	}
}

func (o *APIClient) GetProfile(id string) (*dto.Profile, error) {
	url := fmt.Sprintf("%s/api/users/%s", o.baseURL, id)
	var profile dto.Profile
	if err := o.getJSON(o.ctx, url, &profile); err != nil {
		return nil, err
	}
	return &profile, nil
}

func (o *APIClient) getJSON(ctx context.Context, url string, model any) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("User-Agent", o.userAgent)
	res, err := o.client.Do(req)
	if err != nil {
		return ierrors.APIError{Err: err, StatusCode: -1}
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return ierrors.APIError{Err: errors.New("api request failed"), StatusCode: res.StatusCode}
	}
	rawBody, err := io.ReadAll(res.Body)
	if err != nil {
		return ierrors.APIError{Err: errors.New("failed to read response body"), StatusCode: res.StatusCode}
	}
	if err := json.Unmarshal(rawBody, &model); err != nil {
		return ierrors.APIError{Err: errors.New("failed to unmarshal body"), StatusCode: res.StatusCode}
	}
	return nil
}
