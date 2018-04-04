package featuretoggles_test

import (
	"context"
	"fmt"
	"testing"

	unleashapi "github.com/Unleash/unleash-client-go/api"
	"github.com/fabric8-services/fabric8-toggles-service/featuretoggles"
	"github.com/stretchr/testify/assert"
)

func TestComputeEnablementLevel(t *testing.T) {

	// given the following features
	disabledFeature := unleashapi.Feature{
		Name:        "disabledFeature",
		Description: "Disabled feature",
		Enabled:     false,
		Strategies:  []unleashapi.Strategy{},
	}

	noStrategyFeature := unleashapi.Feature{
		Name:        "noStrategyFeature",
		Description: "Feature with no strategy",
		Enabled:     true,
		Strategies:  []unleashapi.Strategy{},
	}

	misconfiguredStrategyFeature := unleashapi.Feature{
		Name:        "singleStrategyFeature",
		Description: "Feature with single strategy",
		Enabled:     true,
		Strategies: []unleashapi.Strategy{
			{
				Name: featuretoggles.EnableByLevelStrategyName,
				Parameters: map[string]interface{}{
					featuretoggles.LevelParameter: "foo", // invalid value
				},
			},
		},
	}

	singleStrategyFeature := unleashapi.Feature{
		Name:        "singleStrategyFeature",
		Description: "Feature with single strategy",
		Enabled:     true,
		Strategies: []unleashapi.Strategy{
			{
				Name: featuretoggles.EnableByLevelStrategyName,
				Parameters: map[string]interface{}{
					featuretoggles.LevelParameter: featuretoggles.InternalLevel,
				},
			},
		},
	}

	multiStrategiesFeature := unleashapi.Feature{
		Name:        "multiStrategiesFeature",
		Description: "Feature with multiple strategies",
		Enabled:     true,
		Strategies: []unleashapi.Strategy{
			{
				Name: featuretoggles.EnableByLevelStrategyName,
				Parameters: map[string]interface{}{
					featuretoggles.LevelParameter: featuretoggles.InternalLevel,
				},
			},
			{
				Name: featuretoggles.EnableByLevelStrategyName,
				Parameters: map[string]interface{}{
					featuretoggles.LevelParameter: featuretoggles.ExperimentalLevel,
				},
			},
			{
				Name: featuretoggles.EnableByLevelStrategyName,
				Parameters: map[string]interface{}{
					featuretoggles.LevelParameter: featuretoggles.BetaLevel,
				},
			},
		},
	}

	releasedFeature := unleashapi.Feature{
		Name:        "releasedFeature",
		Description: "Feature released",
		Enabled:     true,
		Strategies: []unleashapi.Strategy{
			{
				Name: featuretoggles.EnableByLevelStrategyName,
				Parameters: map[string]interface{}{
					featuretoggles.LevelParameter: featuretoggles.ReleasedLevel,
				},
			},
		},
	}

	internalUser := true
	externalUser := false
	dataset := map[bool]map[string]string{
		internalUser: {
			disabledFeature.Name:              featuretoggles.UnknownLevel,
			noStrategyFeature.Name:            featuretoggles.UnknownLevel,
			misconfiguredStrategyFeature.Name: featuretoggles.UnknownLevel,
			singleStrategyFeature.Name:        featuretoggles.InternalLevel, // user is allowed to access this level of feature
			multiStrategiesFeature.Name:       featuretoggles.BetaLevel,
			releasedFeature.Name:              featuretoggles.ReleasedLevel,
		},
		externalUser: {
			disabledFeature.Name:              featuretoggles.UnknownLevel,
			noStrategyFeature.Name:            featuretoggles.UnknownLevel,
			misconfiguredStrategyFeature.Name: featuretoggles.UnknownLevel,
			singleStrategyFeature.Name:        featuretoggles.UnknownLevel, // user is *not* allowed to access this level of feature
			multiStrategiesFeature.Name:       featuretoggles.BetaLevel,
			releasedFeature.Name:              featuretoggles.ReleasedLevel,
		},
	}
	features := map[string]unleashapi.Feature{
		disabledFeature.Name:              disabledFeature,
		noStrategyFeature.Name:            noStrategyFeature,
		misconfiguredStrategyFeature.Name: misconfiguredStrategyFeature,
		singleStrategyFeature.Name:        singleStrategyFeature,
		multiStrategiesFeature.Name:       multiStrategiesFeature,
		releasedFeature.Name:              releasedFeature,
	}

	for internal, featureData := range dataset {
		t.Run(fmt.Sprintf("internal %t", internal), func(t *testing.T) {
			for featureName, expectedLevel := range featureData {
				f := features[featureName]
				t.Run(f.Description, func(t *testing.T) {
					// when
					result := featuretoggles.ComputeEnablementLevel(context.Background(), f, internal)
					// then
					assert.Equal(t, expectedLevel, result)
				})
			}
		})

	}
}
