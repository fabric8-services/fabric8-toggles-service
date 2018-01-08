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
	UnleashClient  UnleashClient
	clientListener *UnleashClientListener
}

// ToggleServiceConfiguration the configuration to the Toggle service
type ToggleServiceConfiguration interface {
	// GetToggleServiceAppName() string
	GetTogglesURL() string
}

// NewClient returns a new client to the toggle feature service including the default underlying unleash client initialized
func NewClient(serviceName string, config ToggleServiceConfiguration) (*Client, error) {
	l := UnleashClientListener{ready: false}
	unleashclient, err := unleash.NewClient(
		unleash.WithAppName(serviceName),
		unleash.WithInstanceId(os.Getenv("HOSTNAME")),
		unleash.WithUrl(config.GetTogglesURL()),
		unleash.WithStrategies(&EnableByLevelStrategy{}),
		unleash.WithMetricsInterval(1*time.Minute),
		unleash.WithRefreshInterval(10*time.Second),
		unleash.WithListener(&l),
	)
	if err != nil {
		return nil, err
	}
	result := NewCustomClient(unleashclient, &l)
	return result, nil
}

// NewClientWithState returns a new client to the toggle feature service with a pre-initialized unleash client listener
func NewClientWithState(unleashclient UnleashClient, ready bool) *Client {
	return NewCustomClient(unleashclient, &UnleashClientListener{ready: ready})
}

// NewCustomClient returns a new client to the toggle feature service with a pre-initialized unleash client
func NewCustomClient(unleashclient UnleashClient, l *UnleashClientListener) *Client {
	result := &Client{
		UnleashClient:  unleashclient,
		clientListener: l,
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
func (c *Client) GetEnabledFeatures(level string) []string {
	return c.UnleashClient.GetEnabledFeatures(&unleashcontext.Context{
		Properties: map[string]string{
			Level: level,
		},
	})
}

// IsFeatureEnabled returns a boolean to specify whether on feature is enabled for a given groupID
func (c *Client) IsFeatureEnabled(feature unleashapi.Feature, level *string) bool {
	if level == nil {
		log.Warn(nil, nil, "skipping check for toggle feature due to: user level is nil")
		return false
	}
	if !c.clientListener.ready {
		log.Warn(nil, nil, "unable to check if feature is enabled due to: client is not ready")
		return false
	}
	ctx := unleashcontext.Context{
		Properties: map[string]string{
			Level: *level,
		},
	}
	return c.UnleashClient.IsEnabled(feature.Name, unleash.WithContext(ctx))
}
