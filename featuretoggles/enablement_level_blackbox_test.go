package featuretoggles_test

import (
	"context"
	"fmt"
	"testing"

	unleashapi "github.com/Unleash/unleash-client-go/api"
	"github.com/fabric8-services/fabric8-toggles-service/featuretoggles"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComputeEnablementLevel(t *testing.T) {

	// given the following features
	disabledFeature := &unleashapi.Feature{
		Name:        "disabledFeature",
		Description: "Disabled feature",
		Enabled:     false,
		Strategies:  []unleashapi.Strategy{},
	}

	noStrategyFeature := &unleashapi.Feature{
		Name:        "noStrategyFeature",
		Description: "Feature with no strategy",
		Enabled:     true,
		Strategies:  []unleashapi.Strategy{},
	}

	singleStrategyFeature := &unleashapi.Feature{
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

	multiStrategiesFeature := &unleashapi.Feature{
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

	releasedFeature := &unleashapi.Feature{
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
	internalLevel := featuretoggles.InternalLevel
	betaLevel := featuretoggles.BetaLevel
	releasedLevel := featuretoggles.ReleasedLevel
	dataset := map[bool]map[*unleashapi.Feature]*string{
		internalUser: {
			disabledFeature:        nil,
			noStrategyFeature:      nil,
			singleStrategyFeature:  &internalLevel, // user is allowed to access this level of feature
			multiStrategiesFeature: &betaLevel,
			releasedFeature:        &releasedLevel,
		},
		externalUser: {
			disabledFeature:        nil,
			noStrategyFeature:      nil,
			singleStrategyFeature:  nil, // user is *not* allowed to access this level of feature
			multiStrategiesFeature: &betaLevel,
			releasedFeature:        &releasedLevel,
		},
	}

	for user, featureData := range dataset {
		t.Run(fmt.Sprintf("internal user %t", user), func(t *testing.T) {
			for inputFeature, expectedLevel := range featureData {
				t.Run(inputFeature.Description, func(t *testing.T) {
					// when
					result := featuretoggles.ComputeEnablementLevel(context.Background(), inputFeature, user)
					// then
					if expectedLevel == nil {
						assert.Nil(t, result)
					} else {
						require.NotNil(t, result)
						assert.Equal(t, *expectedLevel, *result)
					}
				})
			}
		})

	}
}
