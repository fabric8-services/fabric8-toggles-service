package auth

import (
	"context"
	"net/http"
	"net/url"

	"github.com/fabric8-services/fabric8-auth/goasupport"
	"github.com/fabric8-services/fabric8-auth/log"
	"github.com/fabric8-services/fabric8-toggles-service/auth/client"
	goaclient "github.com/goadesign/goa/client"
	goajwt "github.com/goadesign/goa/middleware/security/jwt"
)

// ClientConfig the config to use when initializing a new Client
type ClientConfig struct {
	ctx        context.Context
	httpClient *http.Client
}

// ClientConfigOption an option to customize the config to use when initializing a client
type ClientConfigOption func(*ClientConfig)

// WithHTTPClient uses a custom http client
func WithHTTPClient(httpClient *http.Client) ClientConfigOption {
	return func(c *ClientConfig) {
		log.Warn(nil,
			map[string]interface{}{"http_transport": httpClient.Transport},
			"configuring custom HTTP client for the auth client")
		c.httpClient = httpClient
	}
}

// NewClient initializes a new client to the `auth` service
func NewClient(ctx context.Context, authURL string, options ...ClientConfigOption) (*client.Client, error) {
	u, err := url.Parse(authURL)
	if err != nil {
		return nil, err
	}
	clientConfig := ClientConfig{
		httpClient: http.DefaultClient,
	}
	for _, configure := range options {
		configure(&clientConfig)
	}
	c := client.New(goaclient.HTTPClientDoer(clientConfig.httpClient))
	c.Host = u.Host
	c.Scheme = u.Scheme
	// allow requests with no JWT in the context
	if goajwt.ContextJWT(ctx) != nil {
		c.SetJWTSigner(goasupport.NewForwardSigner(ctx))
	} else {
		log.Info(ctx, nil, "no token in context")
	}
	return c, nil
}
