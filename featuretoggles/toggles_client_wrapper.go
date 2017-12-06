package featuretoggles

import (
	"fmt"
	"time"

	"github.com/Unleash/unleash-client-go"
	unleashcontext "github.com/Unleash/unleash-client-go/context"
)

type UnleashClient interface {
	Ready() <-chan bool
	GetEnabledFeatures(ctx *unleashcontext.Context) []string
	IsEnabled(feature string, options ...unleash.FeatureOption) (enabled bool)
	Close() error
}

// Client the toggle client
type Client struct {
	UnleashClient UnleashClient
}

// ToggleServiceConfiguration the configuration to the Toggle service
type ToggleServiceConfiguration interface {
	// GetToggleServiceAppName() string
	GetTogglesURL() string
}

// NewClient returns a new client to the toggle feature service
func NewClient(config ToggleServiceConfiguration) (*Client, error) {
	a := config.GetTogglesURL()
	fmt.Println(a)
	client, err := unleash.NewClient(
		unleash.WithAppName("fabric8-ui"),
		unleash.WithUrl(config.GetTogglesURL()),
		unleash.WithStrategies(&EnableByGroupIDStrategy{}),
		unleash.WithRefreshInterval(1*time.Second),
		unleash.WithMetricsInterval(5*time.Second),
		unleash.WithListener(&MetricsListener{}),
	)
	if err != nil {
		return nil, err
	}
	return &Client{UnleashClient: client}, nil

}

// Close closes the underlying Unleash client
func (c *Client) Close() error {
	return c.UnleashClient.Close()
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
func (c *Client) IsFeatureEnabled(featureID string, groupID string) bool {
	if !ready {
		return false
	}
	ctx := unleashcontext.Context{
		Properties: map[string]string{
			GroupID: groupID,
		},
	}

	return c.UnleashClient.IsEnabled(featureID, unleash.WithContext(ctx))
}
