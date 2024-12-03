package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/cyllective/oauth-labs/oalib"
	"golang.org/x/oauth2"

	"github.com/cyllective/oauth-labs/lab04/client/internal/config"
	"github.com/cyllective/oauth-labs/lab04/client/internal/constants"
)

func Init() error {
	o := NewOAuthClient()
	return o.RegisterWithBackoff(10)
}

type OAuthClient struct {
	client    *http.Client
	meta      *oalib.ClientMetadata
	creds     *oalib.ClientCredentials
	userAgent string
	regURI    string
	regSecret string
	revokeURI string
}

func NewOAuthClient() *OAuthClient {
	cfg := config.Get()
	client := &http.Client{
		Timeout: time.Duration(3) * time.Second,
	}
	if proxyURL := os.Getenv("BURP_PROXY"); proxyURL != "" {
		proxy, _ := url.Parse(proxyURL)
		client.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxy),
		}
	}
	meta := &oalib.ClientMetadata{
		RedirectURIs:            []string{cfg.GetString("client.redirect_uri")},
		TokenEndpointAuthMethod: "client_secret_basic",
		GrantTypes:              []string{"authorization_code", "refresh_token"},
		ResponseTypes:           []string{"code"},
		ClientName:              fmt.Sprintf("client-%s", constants.LabNumber),
		ClientURI:               cfg.GetString("client.uri"),
		LogoURI:                 cfg.GetString("client.logo_uri"),
		Scope:                   strings.Join(cfg.GetStringSlice("client.scopes"), " "),
	}
	creds := &oalib.ClientCredentials{
		ID:     cfg.GetString("client.id"),
		Secret: cfg.GetString("client.secret"),
	}
	return &OAuthClient{
		client:    client,
		meta:      meta,
		creds:     creds,
		regURI:    cfg.GetString("authorization_server.register_uri"),
		regSecret: cfg.GetString("authorization_server.registration_secret"),
		revokeURI: cfg.GetString("authorization_server.revocation_uri"),
	}
}

// Register performs dynamic client registration.
func (o *OAuthClient) Register() error {
	ctx := context.Background()
	rawBody, err := json.Marshal(o.meta)
	if err != nil {
		panic(err)
	}
	req, err := http.NewRequestWithContext(ctx, "POST", o.regURI, bytes.NewReader(rawBody))
	if err != nil {
		panic(err)
	}
	req.SetBasicAuth(o.creds.ID, o.creds.Secret)
	req.Header.Set("User-Agent", o.userAgent)
	req.Header.Set("content-type", "application/json")
	req.Header.Set("x-register-key", o.regSecret)
	res, err := o.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to perform client registration request: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		_, _ = io.Copy(io.Discard, res.Body)
		return fmt.Errorf("failed to register client")
	}

	_, _ = io.Copy(io.Discard, res.Body)
	return nil
}

func (o *OAuthClient) RegisterWithBackoff(tries int) error {
	errs := 0
	for i := 0; i < tries; i++ {
		if err := o.Register(); err == nil {
			log.Printf("[client]: client registration performed!")
			return nil
		}

		errs++
		time.Sleep(time.Duration(errs) * time.Second)
		log.Printf("[client]: dynamic client registration failed... retrying.")
	}

	log.Printf("[client]: dynamic registration failed... aborting.")
	return errors.New("dynamic client registration failed")
}

// Revoke instructs the authorization server to revoke the passed tokens.
func (o *OAuthClient) Revoke(tokens *oauth2.Token) []error {
	return o.revoke(context.Background(), tokens)
}

// RevokeWithContext instructs the authorization server to revoke the passed tokens.
func (o *OAuthClient) RevokeWithContext(ctx context.Context, tokens *oauth2.Token) []error {
	return o.revoke(ctx, tokens)
}

func (o *OAuthClient) revoke(ctx context.Context, tokens *oauth2.Token) []error {
	wg := new(sync.WaitGroup)
	nRequests := 2
	wg.Add(nRequests)
	errChan := make(chan error, nRequests)
	defer close(errChan)

	// Handle the refresh_token revocation.
	go func() {
		defer wg.Done()
		err := o.sendRevocationRequest(ctx, &oalib.RevocationRequest{
			Token:         tokens.RefreshToken,
			TokenTypeHint: "refresh_token",
		})
		errChan <- err
	}()

	// Handle the access_token revocation.
	go func() {
		defer wg.Done()
		err := o.sendRevocationRequest(ctx, &oalib.RevocationRequest{
			Token:         tokens.AccessToken,
			TokenTypeHint: "access_token",
		})
		errChan <- err
	}()

	// Wait for the revocation requests to complete.
	wg.Wait()

	// Collect revocation errors.
	errs := make([]error, 0)
	for i := 0; i < nRequests; i++ {
		errs = append(errs, <-errChan)
	}
	return errs
}

func (o *OAuthClient) sendRevocationRequest(ctx context.Context, request *oalib.RevocationRequest) error {
	rawBody, err := json.Marshal(request)
	if err != nil {
		panic(err)
	}
	req, err := http.NewRequestWithContext(ctx, "POST", o.revokeURI, bytes.NewReader(rawBody))
	if err != nil {
		panic(err)
	}
	req.SetBasicAuth(o.creds.ID, o.creds.Secret)
	req.Header.Set("User-Agent", o.userAgent)
	req.Header.Set("content-type", "application/json")
	res, err := o.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to revoke %s: %w", request.TokenTypeHint, err)
	}
	defer func() {
		_, _ = io.Copy(io.Discard, res.Body)
		res.Body.Close()
	}()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to revoke %s", request.TokenTypeHint)
	}
	return nil
}
