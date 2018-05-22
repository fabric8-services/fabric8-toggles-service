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

type FeatureEnablementData struct {
	Ready                   bool
	Feature                 unleashapi.Feature
	User                    *authclient.User
	ExpectedUserEnablement  bool
	ExpectedEnablementLevel string
}

func TestGetFeature(t *testing.T) {
	// given
	betaLevel := featuretoggles.BetaLevel
	allTestData := []FeatureEnablementData{
		{
			Ready: true,
			Feature: unleashapi.Feature{
				Name:    "matching strategy",
				Enabled: true,
				Strategies: []unleashapi.Strategy{
					{
						Name: featuretoggles.EnableByLevelStrategyName,
						Parameters: map[string]interface{}{
							featuretoggles.LevelParameter: featuretoggles.BetaLevel,
						},
					},
				},
			},
			User: &authclient.User{
				Data: &authclient.UserData{
					Attributes: &authclient.UserDataAttributes{
						FeatureLevel: &betaLevel,
					},
				},
			},
			ExpectedUserEnablement:  true,
			ExpectedEnablementLevel: featuretoggles.BetaLevel,
		},
		{
			Ready: true,
			Feature: unleashapi.Feature{
				Name:    "disabled",
				Enabled: false,
			},
			User: &authclient.User{
				Data: &authclient.UserData{
					Attributes: &authclient.UserDataAttributes{
						FeatureLevel: &betaLevel,
					},
				},
			},
			ExpectedUserEnablement:  false,
			ExpectedEnablementLevel: featuretoggles.UnknownLevel,
		},
		{
			Ready: false,
			Feature: unleashapi.Feature{
				Name:    "no strategy",
				Enabled: true,
			},
			User: &authclient.User{
				Data: &authclient.UserData{
					Attributes: &authclient.UserDataAttributes{
						FeatureLevel: &betaLevel,
					},
				},
			},
			ExpectedUserEnablement:  false,
			ExpectedEnablementLevel: featuretoggles.UnknownLevel,
		},
	}
	mockUnleashClient := testfeaturetoggles.NewUnleashClientMock(t)
	mockUnleashClient.GetFeatureFunc = func(name string) *unleashapi.Feature {
		for _, testData := range allTestData {
			if testData.Feature.Name == name {
				return &testData.Feature
			}
		}
		return nil
	}
	mockUnleashClient.IsEnabledFunc = func(feature string, options ...unleash.FeatureOption) (enabled bool) {
		// force client to return `true` when the feature to check is `fooFeature`
		return feature == "matching strategy"
	}

	for _, testData := range allTestData {
		t.Run(fmt.Sprintf("%s", testData.Feature.Name), func(t *testing.T) {
			// given
			ft := featuretoggles.NewClientWithState(mockUnleashClient, true)
			// when
			f := ft.GetFeature(context.Background(), testData.Feature.Name, testData.User)
			// then
			require.NotNil(t, f)
			assert.Equal(t, featuretoggles.UserFeature{
				Name:            testData.Feature.Name,
				Description:     testData.Feature.Description,
				Enabled:         testData.Feature.Enabled,
				EnablementLevel: testData.ExpectedEnablementLevel,
				UserEnabled:     testData.ExpectedUserEnablement,
			}, f)
		})
	}
}

func TestGetFeaturesByName(t *testing.T) {
	// given
	matchingFeature := unleashapi.Feature{
		Name:    "matching strategy",
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
	disabledFeature := unleashapi.Feature{
		Name:    "disabled",
		Enabled: false,
	}
	noStrategyFeature := unleashapi.Feature{
		Name:    "no strategy",
		Enabled: true,
	}
	allFeatures := []unleashapi.Feature{
		matchingFeature, disabledFeature, noStrategyFeature,
	}
	mockUnleashClient := testfeaturetoggles.NewUnleashClientMock(t)
	mockUnleashClient.GetFeatureFunc = func(name string) *unleashapi.Feature {
		for _, f := range allFeatures {
			if f.Name == name {
				return &f
			}
		}
		return nil
	}
	mockUnleashClient.IsEnabledFunc = func(feature string, options ...unleash.FeatureOption) (enabled bool) {
		// force client to return `true` when the feature to check is `fooFeature`
		return feature == "matching strategy"
	}
	betaLevel := featuretoggles.BetaLevel
	user := &authclient.User{
		Data: &authclient.UserData{
			Attributes: &authclient.UserDataAttributes{
				FeatureLevel: &betaLevel,
			},
		},
	}

	t.Run("client ready", func(t *testing.T) {
		// given
		ft := featuretoggles.NewClientWithState(mockUnleashClient, true)
		t.Run("no matches", func(t *testing.T) {
			// when
			f := ft.GetFeaturesByName(context.Background(), []string{"unknown"}, user)
			// then
			assert.Empty(t, f)
		})
		t.Run("all matches", func(t *testing.T) {
			// when
			f := ft.GetFeaturesByName(context.Background(), []string{"matching strategy", "disabled"}, user)
			// then
			require.Len(t, f, 2)
			assert.ElementsMatch(t, f, []featuretoggles.UserFeature{
				{
					Name:            matchingFeature.Name,
					Description:     matchingFeature.Description,
					Enabled:         matchingFeature.Enabled,
					EnablementLevel: featuretoggles.BetaLevel,
					UserEnabled:     true,
				},
				{
					Name:            disabledFeature.Name,
					Description:     disabledFeature.Description,
					Enabled:         disabledFeature.Enabled,
					EnablementLevel: featuretoggles.UnknownLevel,
					UserEnabled:     false,
				},
			})
		})
	})

	t.Run("client not ready", func(t *testing.T) {
		// given
		ft := featuretoggles.NewClientWithState(mockUnleashClient, false)
		// when
		f := ft.GetFeaturesByName(context.Background(), []string{"matching strategy", "disabled"}, user)
		// then
		require.Empty(t, f)
	})
}

func TestGetFeaturesByPattern(t *testing.T) {
	// given
	fooGroupFeature := unleashapi.Feature{
		Name:    "foogroup",
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
	fooFeature := unleashapi.Feature{
		Name:    "foogroup.foo",
		Enabled: false,
	}
	mockUnleashClient := testfeaturetoggles.NewUnleashClientMock(t)
	mockUnleashClient.GetFeaturesByPatternFunc = func(pattern string) []unleashapi.Feature {
		return []unleashapi.Feature{fooGroupFeature, fooFeature}
	}
	mockUnleashClient.IsEnabledFunc = func(feature string, options ...unleash.FeatureOption) (enabled bool) {
		// force client to return `true` when the feature to check is `fooFeature`
		return feature == fooGroupFeature.Name
	}
	betaLevel := featuretoggles.BetaLevel
	user := &authclient.User{
		Data: &authclient.UserData{
			Attributes: &authclient.UserDataAttributes{
				FeatureLevel: &betaLevel,
			},
		},
	}

	t.Run("client ready", func(t *testing.T) {
		// given
		ft := featuretoggles.NewClientWithState(mockUnleashClient, true)
		// when
		f := ft.GetFeaturesByPattern(context.Background(), "fooGroup", user)
		// then
		require.Len(t, f, 2)
		assert.ElementsMatch(t, f, []featuretoggles.UserFeature{
			{
				Name:            fooGroupFeature.Name,
				Description:     fooGroupFeature.Description,
				Enabled:         fooGroupFeature.Enabled,
				EnablementLevel: featuretoggles.BetaLevel,
				UserEnabled:     true,
			},
			{
				Name:            fooFeature.Name,
				Description:     fooFeature.Description,
				Enabled:         fooFeature.Enabled,
				EnablementLevel: featuretoggles.UnknownLevel,
				UserEnabled:     false,
			},
		})
	})

	t.Run("client not ready", func(t *testing.T) {
		// given
		ft := featuretoggles.NewClientWithState(mockUnleashClient, false)
		// when
		f := ft.GetFeaturesByPattern(context.Background(), "fooGroup", user)
		// then
		require.Empty(t, f)
	})
}

func TestGetFeaturesByStrategy(t *testing.T) {
	// given
	feat1 := unleashapi.Feature{
		Name:    "feat1",
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
	feat2 := unleashapi.Feature{
		Name:    "feat2",
		Enabled: false,
		Strategies: []unleashapi.Strategy{
			{
				Name: featuretoggles.EnableByLevelStrategyName,
				Parameters: map[string]interface{}{
					featuretoggles.LevelParameter: featuretoggles.BetaLevel,
				},
			},
		},
	}
	mockUnleashClient := testfeaturetoggles.NewUnleashClientMock(t)
	mockUnleashClient.GetFeaturesByStrategyFunc = func(name string) []unleashapi.Feature {
		return []unleashapi.Feature{feat1, feat2}
	}
	mockUnleashClient.IsEnabledFunc = func(feature string, options ...unleash.FeatureOption) (enabled bool) {
		// force client to return `true` when the feature to check is `fooFeature`
		return feature == feat1.Name
	}
	betaLevel := featuretoggles.BetaLevel
	user := &authclient.User{
		Data: &authclient.UserData{
			Attributes: &authclient.UserDataAttributes{
				FeatureLevel: &betaLevel,
			},
		},
	}

	t.Run("2 matches", func(t *testing.T) {
		// given
		ft := featuretoggles.NewClientWithState(mockUnleashClient, true)
		// when
		f := ft.GetFeaturesByStrategy(context.Background(), "enableByLevel", user)
		// then
		require.Len(t, f, 2)
		assert.ElementsMatch(t, f, []featuretoggles.UserFeature{
			{
				Name:            feat1.Name,
				Description:     feat1.Description,
				Enabled:         feat1.Enabled,
				EnablementLevel: featuretoggles.BetaLevel,
				UserEnabled:     true,
			},
			{
				Name:            feat2.Name,
				Description:     feat2.Description,
				Enabled:         feat2.Enabled,
				EnablementLevel: featuretoggles.UnknownLevel,
				UserEnabled:     false,
			},
		})
	})

	t.Run("client not ready", func(t *testing.T) {
		// given
		ft := featuretoggles.NewClientWithState(mockUnleashClient, false)
		// when
		f := ft.GetFeaturesByStrategy(context.Background(), "enableByLevel", user)
		// then
		require.Empty(t, f)
	})
}
