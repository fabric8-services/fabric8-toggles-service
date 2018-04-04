package featuretoggles_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	unleash "github.com/Unleash/unleash-client-go"
	unleashapi "github.com/Unleash/unleash-client-go/api"
	unleashcontext "github.com/Unleash/unleash-client-go/context"
	"github.com/fabric8-services/fabric8-toggles-service/featuretoggles"
	testfeaturetoggles "github.com/fabric8-services/fabric8-toggles-service/test/featuretoggles"
	"github.com/stretchr/testify/assert"
)

var betaFeature, releasedFeature unleashapi.Feature

func init() {

	betaFeature = unleashapi.Feature{
		Name: "foo.disabledFeature",
	}

	releasedFeature = unleashapi.Feature{
		Name: "bar.releasedFeature",
	}
}

func TestGetFeature(t *testing.T) {
	// given
	mockUnleashClient := testfeaturetoggles.NewUnleashClientMock(t)
	mockUnleashClient.GetFeatureFunc = func(name string) *unleashapi.Feature {
		return &betaFeature
	}

	t.Run("client ready", func(t *testing.T) {
		// given
		ft := featuretoggles.NewClientWithState(mockUnleashClient, true)
		// when
		f := ft.GetFeature(context.Background(), "foo.disabledFeature")
		// then
		require.NotNil(t, f)
		assert.Equal(t, betaFeature, *f)
	})

	t.Run("client not ready", func(t *testing.T) {
		// given
		ft := featuretoggles.NewClientWithState(mockUnleashClient, false)
		// when
		f := ft.GetFeature(context.Background(), "foo.disabledFeature")
		// then
		assert.Nil(t, f)
	})
}

func TestGetFeaturesByName(t *testing.T) {
	// given
	mockUnleashClient := testfeaturetoggles.NewUnleashClientMock(t)
	mockUnleashClient.GetFeatureFunc = func(name string) *unleashapi.Feature {
		return &betaFeature
	}

	t.Run("client ready", func(t *testing.T) {
		// given
		ft := featuretoggles.NewClientWithState(mockUnleashClient, true)
		t.Run("no matches", func(t *testing.T) {
			// when
			f := ft.GetFeaturesByName(context.Background(), []string{"foo.disabledFeature", "bar.releasedFeature"})
			// then
			require.NotEmpty(t, f)
			assert.Len(t, f, 2)
		})
		t.Run("all matches", func(t *testing.T) {
			// when
			f := ft.GetFeaturesByName(context.Background(), []string{"foo.disabledFeature", "bar.releasedFeature"})
			// then
			require.NotEmpty(t, f)
			assert.Len(t, f, 2)
		})
	})

	t.Run("client not ready", func(t *testing.T) {
		// given
		ft := featuretoggles.NewClientWithState(mockUnleashClient, false)
		// when
		f := ft.GetFeaturesByName(context.Background(), []string{"foo.disabledFeature", "bar.releasedFeature"})
		// then
		require.Empty(t, f)
	})
}

func TestGetFeaturesByPattern(t *testing.T) {
	// given
	mockUnleashClient := testfeaturetoggles.NewUnleashClientMock(t)
	mockUnleashClient.GetFeatureFunc = func(name string) *unleashapi.Feature {
		return &betaFeature
	}

	mockUnleashClient.GetEnabledFeaturesFunc = func(ctx *unleashcontext.Context) []string {
		return []string{betaFeature.Name, releasedFeature.Name}
	}

	t.Run("client ready", func(t *testing.T) {
		// given
		ft := featuretoggles.NewClientWithState(mockUnleashClient, true)
		// when
		f := ft.GetFeaturesByPattern(context.Background(), "foo")
		// then
		require.NotEmpty(t, f)
		assert.Len(t, f, 1)
	})

	t.Run("client not ready", func(t *testing.T) {
		// given
		ft := featuretoggles.NewClientWithState(mockUnleashClient, false)
		// when
		f := ft.GetFeaturesByPattern(context.Background(), "foo")
		// then
		require.Empty(t, f)
	})
}

type FeatureEnablementData struct {
	Ready              bool
	Feature            unleashapi.Feature
	UserLevel          string
	ExpectedEnablement bool
}

func TestIsFeatureEnabled(t *testing.T) {

	// given
	allTestData := []FeatureEnablementData{
		{Ready: true, Feature: betaFeature, UserLevel: featuretoggles.BetaLevel, ExpectedEnablement: true},
		{Ready: false, Feature: releasedFeature, UserLevel: featuretoggles.BetaLevel, ExpectedEnablement: false},
	}
	mockClient := testfeaturetoggles.NewUnleashClientMock(t)
	mockClient.IsEnabledFunc = func(feature string, options ...unleash.FeatureOption) (enabled bool) {
		// force client to return `true` when the feature to check is `beta`
		return feature == betaFeature.Name
	}

	for _, testData := range allTestData {
		t.Run(fmt.Sprintf("with client ready=%t", testData.Ready), func(t *testing.T) {
			// given
			client := featuretoggles.NewClientWithState(mockClient, testData.Ready)
			// when
			result := client.IsFeatureEnabled(context.Background(), testData.Feature, testData.UserLevel)
			// then
			assert.Equal(t, testData.ExpectedEnablement, result)
		})
	}

}
