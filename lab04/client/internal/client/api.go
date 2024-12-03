package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"golang.org/x/oauth2"

	"github.com/cyllective/oauth-labs/lab04/client/internal/config"
	"github.com/cyllective/oauth-labs/lab04/client/internal/constants"
	"github.com/cyllective/oauth-labs/lab04/client/internal/dto"
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

func (o *APIClient) GetProfile() (*dto.Profile, error) {
	url := fmt.Sprintf("%s/api/users/me", o.baseURL)
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
		return fmt.Errorf("api request to %#v failed: %w", url, err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("api request to %#v failed: invalid status code", url)
	}
	rawBody, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("api request to %#v failed: reading response body: %w", url, err)
	}
	if err := json.Unmarshal(rawBody, &model); err != nil {
		return fmt.Errorf("api request to %#v failed: unmarshal failed: %w", url, err)
	}
	return nil
}
