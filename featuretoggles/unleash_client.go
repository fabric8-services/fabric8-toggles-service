package featuretoggles

import (
	"context"
	"fmt"
	"github.com/Unleash/unleash-client-go"
	unleashcontext "github.com/Unleash/unleash-client-go/context"
)

type Client interface {
	GetEnabledFeatures(groupId string) []string
}

// ToggleClient the toggle client
type ToggleClient struct {
	unleashClient unleash.Client
}

// ToggleServiceConfiguration the configuration to the Toggle service
type ToggleServiceConfiguration interface {
	// GetToggleServiceAppName() string
	GetTogglesURL() string
}

// NewFeatureToggleClient returns a new client to the toggle feature service
func NewFeatureToggleClient(ctx context.Context, config ToggleServiceConfiguration) (*ToggleClient, error) {
	a := config.GetTogglesURL()
	fmt.Println(a)
	client, err := unleash.NewClient(
		unleash.WithAppName("fabric8-ui"),
		unleash.WithUrl(config.GetTogglesURL()),
		unleash.WithStrategies(&EnableByGroupIDStrategy{}),
	)
	if err != nil {
		return nil, err
	}
	// wait until client did perform a data sync
	<-client.Ready()
	return &ToggleClient{unleashClient: *client}, nil

}

// Close closes the underlying Unleash client
func (c *ToggleClient) Close() error {
	return c.unleashClient.Close()
}

// GetEnabledFeatures returns the names of enabled features for the given user
func (c *ToggleClient) GetEnabledFeatures(groupId string) []string {
	return c.unleashClient.GetEnabledFeatures(&unleashcontext.Context{
		Properties: map[string]string{
			GroupID: groupId,
		},
	})
}

const (
	EnableByGroupID string = "enableByGroupID"
	GroupID         string = "groupID"
)

// EnableByGroupIDStrategy the strategy to roll out a feature if the user belongs to a given group
type EnableByGroupIDStrategy struct {
}

// Name the name of the stragegy. Must match the name on the Unleash server.
func (s *EnableByGroupIDStrategy) Name() string {
	return EnableByGroupID
}

// IsEnabled returns `true` if the given context is compatible with the settings configured on the Unleash server
func (s *EnableByGroupIDStrategy) IsEnabled(settings map[string]interface{}, ctx *unleashcontext.Context) bool {
	fmt.Printf("Checking %+v vs %+v", settings, ctx)
	return settings[GroupID] == ctx.Properties[GroupID]
}
