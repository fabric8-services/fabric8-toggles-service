package featuretoggles_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	unleash "github.com/Unleash/unleash-client-go"
	unleashapi "github.com/Unleash/unleash-client-go/api"
	authclient "github.com/fabric8-services/fabric8-toggles-service/auth/client"
	"github.com/fabric8-services/fabric8-toggles-service/featuretoggles"
	testfeaturetoggles "github.com/fabric8-services/fabric8-toggles-service/test/featuretoggles"
	"github.com/stretchr/testify/assert"
)

var fooGroupFeature, foobarFeature, fooFeature, barFeature, bazFeature unleashapi.Feature

func init() {

	fooGroupFeature = unleashapi.Feature{
		Name: "foogroup",
	}
	foobarFeature = unleashapi.Feature{
		Name: "foobar",
	}
	fooFeature = unleashapi.Feature{
		Name:    "foogroup.foo",
		Enabled: true,
		Strategies: []unleashapi.Strategy{
			{
				Name: featuretoggles.EnableByLevelStrategyName,
				Parameters: map[string]interface{}{
					featuretoggles.LevelParameter: featuretoggles.BetaLevel,
				},
			},
		},
	}

	barFeature = unleashapi.Feature{
		Name:    "bar",
		Enabled: true,
	}

	bazFeature = unleashapi.Feature{
		Name:    "baz",
		Enabled: false,
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
	Ready                   bool
	Feature                 unleashapi.Feature
	User                    authclient.User
	ExpectedEnablement      bool
	ExpectedEnablementLevel string
}

func TestIsFeatureEnabled(t *testing.T) {
	betaLevel := featuretoggles.BetaLevel
	// given
	allTestData := []FeatureEnablementData{
		{
			Ready:   true,
			Feature: fooFeature,
			User: authclient.User{
				Data: &authclient.UserData{
					Attributes: &authclient.UserDataAttributes{
						FeatureLevel: &betaLevel,
					},
				},
			},
			ExpectedEnablement:      true,
			ExpectedEnablementLevel: featuretoggles.BetaLevel,
		},
		{
			Ready:   true,
			Feature: bazFeature,
			User: authclient.User{
				Data: &authclient.UserData{
					Attributes: &authclient.UserDataAttributes{
						FeatureLevel: &betaLevel,
					},
				},
			},
			ExpectedEnablement:      false,
			ExpectedEnablementLevel: featuretoggles.UnknownLevel,
		},
		{
			Ready:   false,
			Feature: barFeature,
			User: authclient.User{
				Data: &authclient.UserData{
					Attributes: &authclient.UserDataAttributes{
						FeatureLevel: &betaLevel,
					},
				},
			},
			ExpectedEnablement:      false,
			ExpectedEnablementLevel: featuretoggles.UnknownLevel,
		},
	}
	mockClient := testfeaturetoggles.NewUnleashClientMock(t)
	mockClient.IsEnabledFunc = func(feature string, options ...unleash.FeatureOption) (enabled bool) {
		// force client to return `true` when the feature to check is `fooFeature`
		return feature == fooFeature.Name
	}

	for _, testData := range allTestData {
		t.Run(fmt.Sprintf("with feature %s", testData.Feature.Name), func(t *testing.T) {
			// given
			client := featuretoggles.NewClientWithState(mockClient, testData.Ready)
			// when
			enabled, level := client.IsFeatureEnabled(context.Background(), testData.Feature, &testData.User)
			// then
			assert.Equal(t, testData.ExpectedEnablement, enabled)
			assert.Equal(t, testData.ExpectedEnablementLevel, level)
		})
	}

}
