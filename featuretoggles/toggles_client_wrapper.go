package featuretoggles

import (
	"os"
	"time"

	"github.com/Unleash/unleash-client-go"
	unleashapi "github.com/Unleash/unleash-client-go/api"
	unleashcontext "github.com/Unleash/unleash-client-go/context"
	"github.com/fabric8-services/fabric8-wit/log"
)

type UnleashClient interface {
	Ready() <-chan bool
	GetEnabledFeatures(ctx *unleashcontext.Context) []string
	GetFeature(name string) *unleashapi.Feature
	IsEnabled(feature string, options ...unleash.FeatureOption) (enabled bool)
	Close() error
}

// Client the toggle client
type Client struct {
	ready         bool
	UnleashClient UnleashClient
}

// ToggleServiceConfiguration the configuration to the Toggle service
type ToggleServiceConfiguration interface {
	// GetToggleServiceAppName() string
	GetTogglesURL() string
}

// NewClient returns a new client to the toggle feature service including the default underlying unleash client initialized
func NewClient(serviceName string, config ToggleServiceConfiguration) (*Client, error) {
	l := clientListener{}
	unleashclient, err := unleash.NewClient(
		unleash.WithAppName(serviceName),
		unleash.WithInstanceId(os.Getenv("HOSTNAME")),
		unleash.WithUrl(config.GetTogglesURL()),
		unleash.WithMetricsInterval(1*time.Minute),
		unleash.WithRefreshInterval(10*time.Second),
		unleash.WithListener(l),
	)
	if err != nil {
		return nil, err
	}
	result := NewCustomClient(unleashclient, false)
	l.client = result
	return result, nil
}

// NewCustomClient returns a new client to the toggle feature service with a pre-initialized unleash client
func NewCustomClient(unleashclient UnleashClient, ready bool) *Client {
	result := &Client{
		UnleashClient: unleashclient,
		ready:         ready,
	}
	return result
}

// Close closes the underlying Unleash client
func (c *Client) Close() error {
	return c.UnleashClient.Close()
}

// GetFeature returns the feature given its name
func (c *Client) GetFeature(groupID string) *unleashapi.Feature {
	return c.UnleashClient.GetFeature(groupID)
}

// GetEnabledFeatures returns the names of enabled features for the given user
func (c *Client) GetEnabledFeatures(groupID string) []string {
	return c.UnleashClient.GetEnabledFeatures(&unleashcontext.Context{
		Properties: map[string]string{
			GroupID: groupID,
		},
	})
}

// IsFeatureEnabled returns a boolean to specify whether on feature is enabled for a given groupID
func (c *Client) IsFeatureEnabled(feature unleashapi.Feature, groupID string) bool {
	if !c.ready {
		log.Warn(nil, nil, "unable to check if feature is enabled due to: client is not ready")
		return false
	}
	ctx := unleashcontext.Context{
		Properties: map[string]string{
			GroupID: groupID,
		},
	}
	return c.UnleashClient.IsEnabled(feature.Name, unleash.WithContext(ctx))
}
