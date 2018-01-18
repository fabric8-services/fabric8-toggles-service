package featuretoggles

import (
	"context"
	"os"
	"time"

	"github.com/Unleash/unleash-client-go"
	unleashapi "github.com/Unleash/unleash-client-go/api"
	unleashcontext "github.com/Unleash/unleash-client-go/context"
	"github.com/fabric8-services/fabric8-auth/log"
)

// UnleashClient the interface to the unleash client
type UnleashClient interface {
	Ready() <-chan bool
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
func (c *Client) GetFeature(name string) *unleashapi.Feature {
	return c.UnleashClient.GetFeature(name)
}

// GetFeatures returns the features fron their names
func (c *Client) GetFeatures(names []string) []*unleashapi.Feature {
	result := make([]*unleashapi.Feature, 0)
	for _, name := range names {
		f := c.UnleashClient.GetFeature(name)
		if f != nil {
			result = append(result, f)
		}
	}
	return result
}

// IsFeatureEnabled returns a boolean to specify whether on feature is enabled for a given user level
func (c *Client) IsFeatureEnabled(ctx context.Context, feature unleashapi.Feature, userLevel *string) bool {
	if userLevel == nil {
		// accept if feature has (at least) one strategy with the `released` level
		for _, s := range feature.Strategies {
			if s.Parameters[LevelParameter] == ReleasedLevel {
				log.Debug(ctx, nil, "considering feature as enabled as it is released")
				return true
			}
		}
		log.Warn(ctx, nil, "skipping check for toggle feature due to: user level is nil")
		return false
	}
	if !c.clientListener.ready {
		log.Warn(ctx, nil, "unable to check if feature is enabled due to: client is not ready")
		return false
	}
	return c.UnleashClient.IsEnabled(
		feature.Name,
		unleash.WithContext(unleashcontext.Context{
			Properties: map[string]string{
				LevelParameter: *userLevel,
			},
		}),
	)
}
