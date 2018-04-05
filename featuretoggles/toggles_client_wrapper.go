package featuretoggles

import (
	"context"
	"fmt"
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
	GetFeaturesByPattern(pattern string) []unleashapi.Feature
	Close() error
}

// Client the toggle client interface
type Client interface {
	GetFeature(ctx context.Context, name string) *unleashapi.Feature
	GetFeaturesByName(ctx context.Context, names []string) []unleashapi.Feature
	GetFeaturesByPattern(ctx context.Context, pattern string) []unleashapi.Feature
	IsFeatureEnabled(ctx context.Context, feature unleashapi.Feature, userLevel string) bool
	Close() error
}

// ClientImpl the toggle client default impl
type ClientImpl struct {
	UnleashClient  UnleashClient
	clientListener *UnleashClientListener
}

// verify that `ClientImpl`` is a valid impl of the `Client`` interface
var _ Client = &ClientImpl{}

// ToggleServiceConfiguration the configuration to the Toggle service
type ToggleServiceConfiguration interface {
	// GetToggleServiceAppName() string
	GetTogglesURL() string
}

// NewDefaultClient returns a new client to the toggle feature service including the default underlying unleash client initialized
func NewDefaultClient(serviceName string, config ToggleServiceConfiguration) (Client, error) {
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
	return &ClientImpl{
		UnleashClient:  unleashclient,
		clientListener: &l,
	}, nil
}

// NewClientWithState returns a new client to the toggle feature service with a pre-initialized unleash client listener
func NewClientWithState(unleashclient UnleashClient, ready bool) Client {
	return &ClientImpl{
		UnleashClient:  unleashclient,
		clientListener: &UnleashClientListener{ready: ready},
	}
}

// Close closes the underlying Unleash client
func (c *ClientImpl) Close() error {
	return c.UnleashClient.Close()
}

// GetFeature returns the feature given its name
func (c *ClientImpl) GetFeature(ctx context.Context, name string) *unleashapi.Feature {
	if !c.clientListener.ready {
		log.Error(ctx, map[string]interface{}{"error": "client is not ready"}, "unable to list features by name")
		return nil
	}
	return c.UnleashClient.GetFeature(name)
}

// GetFeaturesByName returns the features from their names
func (c *ClientImpl) GetFeaturesByName(ctx context.Context, names []string) []unleashapi.Feature {
	result := make([]unleashapi.Feature, 0)
	if !c.clientListener.ready {
		log.Error(ctx, map[string]interface{}{"error": "client is not ready"}, "unable to list features by name")
		return result
	}
	for _, name := range names {
		f := c.UnleashClient.GetFeature(name)
		if f != nil {
			result = append(result, *f)
		}
	}
	return result
}

// GetFeaturesByPattern returns the features whose ID matches the given pattern
func (c *ClientImpl) GetFeaturesByPattern(ctx context.Context, pattern string) []unleashapi.Feature {
	result := make([]unleashapi.Feature, 0)
	if !c.clientListener.ready {
		log.Error(ctx, map[string]interface{}{"error": "client is not ready"}, "unable to list features by pattern")
		return result
	}
	return c.UnleashClient.GetFeaturesByPattern(fmt.Sprintf("^%[1]s$|^%[1]s\\.(.*)", pattern))
}

// IsFeatureEnabled returns a boolean to specify whether on feature is enabled for a given user level
func (c *ClientImpl) IsFeatureEnabled(ctx context.Context, feature unleashapi.Feature, userLevel string) bool {
	if !c.clientListener.ready {
		log.Warn(ctx, nil, "unable to check if feature is enabled due to: client is not ready")
		return false
	}
	log.Debug(ctx, map[string]interface{}{"user_level": userLevel}, "checking if feature is enabled for user...")
	return c.UnleashClient.IsEnabled(
		feature.Name,
		unleash.WithContext(unleashcontext.Context{
			Properties: map[string]string{
				LevelParameter: userLevel,
			},
		}),
	)
}
