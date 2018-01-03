package unleashclient

import (
	unleash "github.com/Unleash/unleash-client-go"
	unleashapi "github.com/Unleash/unleash-client-go/api"
	unleashcontext "github.com/Unleash/unleash-client-go/context"
	unleashstrategy "github.com/Unleash/unleash-client-go/strategy"
	"github.com/fabric8-services/fabric8-wit/log"
)

type MockUnleashClient struct {
	Features   []unleashapi.Feature
	Strategies []unleashstrategy.Strategy
}

// getStrategy looks-up the strategy by its name
func (c *MockUnleashClient) getStrategy(name string) unleashstrategy.Strategy {
	for _, s := range c.Strategies {
		if s.Name() == name {
			return s
		}
	}
	return nil
}

// GetEnabledFeatures mimicks the behaviour of the real client, ie, it uses the strategies to verify the features
func (c *MockUnleashClient) GetEnabledFeatures(ctx *unleashcontext.Context) []string {
	result := make([]string, 0)
	for _, f := range c.Features {
		for _, s := range f.Strategies {
			foundStrategy := c.getStrategy(s.Name)
			if foundStrategy == nil {
				// TODO: warnOnce missingStrategy
				continue
			}
			if foundStrategy.IsEnabled(s.Parameters, ctx) {
				result = append(result, f.Name)
			}
		}
	}
	return result
}

// IsEnabled mimicks the behaviour of the real client
func (c *MockUnleashClient) IsEnabled(name string, options ...unleash.FeatureOption) (enabled bool) {
	defer func() {
		log.Debug(nil, map[string]interface{}{"feature_name": name, "enabled": enabled}, "checked if feature is enabled for user...")
	}()
	for _, f := range c.Features {
		if f.Name == name {
			return f.Enabled
		}
	}
	log.Debug(nil, map[string]interface{}{"feature_name": name}, "unable to find feature by name")
	return false
}

// GetFeature returns a feature by its name
func (c *MockUnleashClient) GetFeature(name string) *unleashapi.Feature {
	for _, f := range c.Features {
		if f.Name == name {
			return &f
		}
	}
	return nil
}

func (m *MockUnleashClient) Close() error {
	return nil
}

func (m *MockUnleashClient) Ready() <-chan bool {
	return nil
}
