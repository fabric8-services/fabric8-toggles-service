package featuretoggles_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	unleash "github.com/Unleash/unleash-client-go"
	unleashapi "github.com/Unleash/unleash-client-go/api"
	"github.com/fabric8-services/fabric8-toggles-service/featuretoggles"
	testfeaturetoggles "github.com/fabric8-services/fabric8-toggles-service/test/featuretoggles"
	"github.com/stretchr/testify/assert"
)

var fooGroupFeature, foobarFeature, fooFeature, barFeature unleashapi.Feature

func init() {

	fooGroupFeature = unleashapi.Feature{
		Name: "foogroup",
	}
	foobarFeature = unleashapi.Feature{
		Name: "foobar",
	}
	fooFeature = unleashapi.Feature{
		Name: "foogroup.foo",
	}

	barFeature = unleashapi.Feature{
		Name: "bar",
	}
}

var getFeatureByName func(name string) *unleashapi.Feature

func init() {
	getFeatureByName = func(name string) *unleashapi.Feature {
		switch name {
		case fooFeature.Name:
			return &fooFeature
		case barFeature.Name:
			return &barFeature
		case foobarFeature.Name:
			return &foobarFeature
		case fooGroupFeature.Name:
			return &fooGroupFeature
		}
		return nil
	}
}

func TestGetFeature(t *testing.T) {
	// given
	mockUnleashClient := testfeaturetoggles.NewUnleashClientMock(t)
	mockUnleashClient.GetFeatureFunc = getFeatureByName

	t.Run("client ready", func(t *testing.T) {
		// given
		ft := featuretoggles.NewClientWithState(mockUnleashClient, true)
		// when
		f := ft.GetFeature(context.Background(), "foogroup.foo")
		// then
		require.NotNil(t, f)
		assert.Equal(t, fooFeature, *f)
	})

	t.Run("client not ready", func(t *testing.T) {
		// given
		ft := featuretoggles.NewClientWithState(mockUnleashClient, false)
		// when
		f := ft.GetFeature(context.Background(), "foogroup.foo")
		// then
		assert.Nil(t, f)
	})
}

func TestGetFeaturesByName(t *testing.T) {
	// given
	mockUnleashClient := testfeaturetoggles.NewUnleashClientMock(t)
	mockUnleashClient.GetFeatureFunc = getFeatureByName

	t.Run("client ready", func(t *testing.T) {
		// given
		ft := featuretoggles.NewClientWithState(mockUnleashClient, true)
		t.Run("no matches", func(t *testing.T) {
			// when
			f := ft.GetFeaturesByName(context.Background(), []string{"unknown"})
			// then
			assert.Empty(t, f)
		})
		t.Run("all matches", func(t *testing.T) {
			// when
			f := ft.GetFeaturesByName(context.Background(), []string{"foogroup", "foogroup.foo"})
			// then
			require.Len(t, f, 2)
			assert.ElementsMatch(t, f, []unleashapi.Feature{fooGroupFeature, fooFeature})
		})
	})

	t.Run("client not ready", func(t *testing.T) {
		// given
		ft := featuretoggles.NewClientWithState(mockUnleashClient, false)
		// when
		f := ft.GetFeaturesByName(context.Background(), []string{"foogroup.foo", "bar"})
		// then
		require.Empty(t, f)
	})
}

func TestGetFeaturesByPattern(t *testing.T) {
	// given
	mockUnleashClient := testfeaturetoggles.NewUnleashClientMock(t)
	mockUnleashClient.GetFeatureFunc = getFeatureByName

	mockUnleashClient.GetFeaturesByPatternFunc = func(pattern string) []unleashapi.Feature {
		return []unleashapi.Feature{fooFeature, fooGroupFeature}
	}

	t.Run("client ready", func(t *testing.T) {
		// given
		ft := featuretoggles.NewClientWithState(mockUnleashClient, true)
		// when
		f := ft.GetFeaturesByPattern(context.Background(), "foogroup")
		// then
		require.NotEmpty(t, f)
		assert.Len(t, f, 2)
	})

	t.Run("client not ready", func(t *testing.T) {
		// given
		ft := featuretoggles.NewClientWithState(mockUnleashClient, false)
		// when
		f := ft.GetFeaturesByPattern(context.Background(), "foogroup.foo")
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
		{Ready: true, Feature: fooFeature, UserLevel: featuretoggles.BetaLevel, ExpectedEnablement: true},
		{Ready: false, Feature: barFeature, UserLevel: featuretoggles.BetaLevel, ExpectedEnablement: false},
	}
	mockClient := testfeaturetoggles.NewUnleashClientMock(t)
	mockClient.IsEnabledFunc = func(feature string, options ...unleash.FeatureOption) (enabled bool) {
		// force client to return `true` when the feature to check is `beta`
		return feature == fooFeature.Name
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
