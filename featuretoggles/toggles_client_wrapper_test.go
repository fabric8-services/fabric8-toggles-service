package featuretoggles_test

import (
	"context"
	"fmt"
	"testing"

	unleash "github.com/Unleash/unleash-client-go"
	unleashapi "github.com/Unleash/unleash-client-go/api"
	unleashstrategy "github.com/Unleash/unleash-client-go/strategy"
	"github.com/fabric8-services/fabric8-toggles-service/featuretoggles"
	unleashtestclient "github.com/fabric8-services/fabric8-toggles-service/test/unleashclient"
	"github.com/stretchr/testify/assert"
)

var betaFeature, releasedFeature unleashapi.Feature

func init() {

	betaFeature = unleashapi.Feature{
		Name: "disabledFeature",
	}

	releasedFeature = unleashapi.Feature{
		Name: "releasedFeature",
	}
}
func NewMockUnleashClient(features ...unleashapi.Feature) *unleashtestclient.MockUnleashClient {
	return &unleashtestclient.MockUnleashClient{
		Features: features,
		Strategies: []unleashstrategy.Strategy{
			&featuretoggles.EnableByLevelStrategy{},
		},
	}
}

func TestGetFeature(t *testing.T) {
}

func TestGetFeatures(t *testing.T) {

}

type FeatureEnablementData struct {
	Ready              bool
	Feature            unleashapi.Feature
	UserLevel          string
	ExpectedEnablement bool
}

func TestIsFeatureEnabled(t *testing.T) {

	allTestData := []FeatureEnablementData{
		{Ready: true, Feature: betaFeature, UserLevel: featuretoggles.BetaLevel, ExpectedEnablement: true},
		{Ready: false, Feature: releasedFeature, UserLevel: featuretoggles.BetaLevel, ExpectedEnablement: false},
	}

	for _, testData := range allTestData {
		t.Run(fmt.Sprintf("with client ready=%t", testData.Ready), func(t *testing.T) {
			// given
			client := featuretoggles.NewClientWithState(&FakeUnleashClient{}, testData.Ready)
			// when
			result := client.IsFeatureEnabled(context.Background(), testData.Feature, testData.UserLevel)
			// then
			assert.Equal(t, testData.ExpectedEnablement, result)
		})
	}

}

type FakeUnleashClient struct {
}

func (c *FakeUnleashClient) Ready() <-chan bool {
	return nil
}

func (c *FakeUnleashClient) GetFeature(name string) *unleashapi.Feature {
	return nil
}

type featureOption struct {
	fallback *bool
	ctx      *context.Context
}

func (c *FakeUnleashClient) IsEnabled(feature string, options ...unleash.FeatureOption) (enabled bool) {
	// force client to return `true` when the feature to check is `beta`
	return feature == betaFeature.Name
}

func (c *FakeUnleashClient) Close() error {
	return nil
}
