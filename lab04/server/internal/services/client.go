package services

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/url"

	"github.com/cyllective/oauth-labs/oalib"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"github.com/cyllective/oauth-labs/lab04/server/internal/config"
	"github.com/cyllective/oauth-labs/lab04/server/internal/dto"
	"github.com/cyllective/oauth-labs/lab04/server/internal/entities"
	"github.com/cyllective/oauth-labs/lab04/server/internal/repositories"
)

type ClientService struct {
	clientRepository *repositories.ClientRepository
}

func NewClientService(clientRepository *repositories.ClientRepository) *ClientService {
	return &ClientService{clientRepository}
}

var ErrClientNotFound = errors.New("client not found")

func (cs *ClientService) Get(ctx context.Context, id string) (*dto.Client, error) {
	if id == "" {
		log.Println("[ClientService.Get] id was empty, aborting with client not found")
		return nil, ErrClientNotFound
	}

	dbClient, err := cs.clientRepository.Get(ctx, id)
	if err != nil {
		log.Printf("[ClientService.Get] failed to fetch client: %s\n", err.Error())
		return nil, err
	}

	var client dto.Client
	err = json.Unmarshal([]byte(dbClient.Config), &client)
	if err != nil {
		log.Printf("[ClientService.Get] failed to unmarshal client config: %s\n", err.Error())
		return nil, err
	}
	return &client, nil
}

func (cs *ClientService) GetMany(ctx context.Context, ids ...string) ([]*dto.Client, error) {
	clients := make([]*dto.Client, 0)
	if len(ids) == 0 {
		return clients, nil
	}

	dbClients, err := cs.clientRepository.GetMany(ctx, ids...)
	if err != nil {
		return nil, err
	}
	for _, c := range dbClients {
		var client dto.Client
		err := json.Unmarshal([]byte(c.Config), &client)
		if err != nil {
			return nil, err
		}
		clients = append(clients, &client)
	}

	return clients, nil
}

func (cs *ClientService) GetAll(ctx context.Context) ([]*dto.Client, error) {
	clients := make([]*dto.Client, 0)
	dbClients, err := cs.clientRepository.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	for _, c := range dbClients {
		var client dto.Client
		err := json.Unmarshal([]byte(c.Config), &client)
		if err != nil {
			return nil, err
		}
		clients = append(clients, &client)
	}

	return clients, nil
}

func (cs *ClientService) ParseCredentials(c *gin.Context) (*oalib.ClientCredentials, *oalib.VerboseError) {
	baClientID, baClientSecret, hasBasicAuth := c.Request.BasicAuth()
	hasPostAuth := c.PostForm("client_secret") != ""
	if !hasBasicAuth && !hasPostAuth {
		return nil, &oalib.VerboseError{
			Err:         "invalid_client",
			Description: "client included no supported authentication parameters.",
		}
	}
	if hasBasicAuth && hasPostAuth {
		return nil, &oalib.VerboseError{
			Err:         "invalid_client",
			Description: "client included multiple forms of authentication.",
		}
	}

	var clientID string
	var clientSecret string
	if hasBasicAuth {
		clientID = baClientID
		clientSecret = baClientSecret

		if baClientID == "" || baClientSecret == "" {
			return nil, &oalib.VerboseError{
				Err:         "invalid_client",
				Description: "client misses client_id or client_secret in basic-auth request.",
			}
		}
		// Ensure that if a client_id was included in the post body that it
		// matches the client_id in the basic auth credentials.
		if c.PostForm("client_id") != "" && baClientID != c.PostForm("client_id") {
			return nil, &oalib.VerboseError{
				Err:         "invalid_client",
				Description: "client included mismatched client_id in body and basic-auth.",
			}
		}
	} else { // hasPostAuth
		clientID = c.PostForm("client_id")
		clientSecret = c.PostForm("client_secret")

		if clientID == "" || clientSecret == "" {
			return nil, &oalib.VerboseError{
				Err:         "invalid_client",
				Description: "client misses client_id or client_secret in post request.",
			}
		}
	}

	return &oalib.ClientCredentials{ID: clientID, Secret: clientSecret}, nil
}

func (cs *ClientService) Authenticate(ctx context.Context, credentials *oalib.ClientCredentials) (*dto.Client, *oalib.VerboseError) {
	client, err := cs.Get(ctx, credentials.ID)
	if err != nil {
		return nil, &oalib.VerboseError{
			Err:         "invalid_client",
			Description: err.Error(),
		}
	}
	if ok := cs.CheckSecret(client, credentials.Secret); !ok {
		return nil, &oalib.VerboseError{
			Err:         "invalid_client",
			Description: "invalid client_secret",
		}
	}
	return client, nil
}

func (cs *ClientService) GetFromURL(ctx context.Context, u string) (*dto.Client, error) {
	uri, err := url.Parse(u)
	if err != nil {
		return nil, ErrClientNotFound
	}
	cid := uri.Query().Get("client_id")
	if cid == "" {
		return nil, ErrClientNotFound
	}
	client, err := cs.Get(ctx, cid)
	if err != nil {
		return nil, ErrClientNotFound
	}
	return client, nil
}

func (cs *ClientService) Register(ctx context.Context, request *dto.ClientRegister) error {
	if request.ID == "" || request.Secret == "" {
		return errors.New("client_id and client_secret are required")
	}

	// Ensure the client secret is hashed.
	cfg := config.Get()
	hash, err := bcrypt.GenerateFromPassword([]byte(request.Secret), cfg.GetInt("server.bcrypt_cost"))
	if err != nil {
		panic(err)
	}
	request.Secret = string(hash)
	data, err := json.Marshal(request)
	if err != nil {
		return err
	}

	err = cs.clientRepository.Register(ctx, &entities.Client{
		ID:     request.ID,
		Config: string(data),
	})
	return err
}

func (cs *ClientService) CheckSecret(client *dto.Client, secret string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(client.Secret), []byte(secret))
	return err == nil
}
