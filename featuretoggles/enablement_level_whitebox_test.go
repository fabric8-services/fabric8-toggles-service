package featuretoggles

import (
	"context"
	"fmt"
	"testing"

	unleashapi "github.com/Unleash/unleash-client-go/api"
	"github.com/stretchr/testify/require"

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
				Name: EnableByLevelStrategyName,
				Parameters: map[string]interface{}{
					LevelParameter: "foo", // invalid value
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
				Name: EnableByLevelStrategyName,
				Parameters: map[string]interface{}{
					LevelParameter: InternalLevel,
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
				Name: EnableByLevelStrategyName,
				Parameters: map[string]interface{}{
					LevelParameter: InternalLevel,
				},
			},
			{
				Name: EnableByLevelStrategyName,
				Parameters: map[string]interface{}{
					LevelParameter: ExperimentalLevel,
				},
			},
			{
				Name: EnableByLevelStrategyName,
				Parameters: map[string]interface{}{
					LevelParameter: BetaLevel,
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
				Name: EnableByLevelStrategyName,
				Parameters: map[string]interface{}{
					LevelParameter: ReleasedLevel,
				},
			},
		},
	}

	devFeature := unleashapi.Feature{
		Name:        "devFeature",
		Description: "Feature in development",
		Enabled:     true,
		Strategies: []unleashapi.Strategy{
			{
				Name: EnableByEmailsStrategyName,
				Parameters: map[string]interface{}{
					EmailsParameter: "adam@foo.com",
				},
			},
		},
	}

	internalUser := true
	externalUser := false
	dataset := map[bool]map[string]string{
		internalUser: {
			disabledFeature.Name:              UnknownLevel,
			noStrategyFeature.Name:            UnknownLevel,
			misconfiguredStrategyFeature.Name: UnknownLevel,
			singleStrategyFeature.Name:        InternalLevel, // user is allowed to access this level of feature
			multiStrategiesFeature.Name:       BetaLevel,
			releasedFeature.Name:              ReleasedLevel,
			devFeature.Name:                   UnknownLevel, // feature cannot be opted-in, user needs to have a matching email address
		},
		externalUser: {
			disabledFeature.Name:              UnknownLevel,
			noStrategyFeature.Name:            UnknownLevel,
			misconfiguredStrategyFeature.Name: UnknownLevel,
			singleStrategyFeature.Name:        UnknownLevel, // user is *not* allowed to access this level of feature
			multiStrategiesFeature.Name:       BetaLevel,
			releasedFeature.Name:              ReleasedLevel,
			devFeature.Name:                   UnknownLevel, // feature cannot be opted-in, user needs to have a matching email address
		},
	}
	features := map[string]unleashapi.Feature{
		disabledFeature.Name:              disabledFeature,
		noStrategyFeature.Name:            noStrategyFeature,
		misconfiguredStrategyFeature.Name: misconfiguredStrategyFeature,
		singleStrategyFeature.Name:        singleStrategyFeature,
		multiStrategiesFeature.Name:       multiStrategiesFeature,
		releasedFeature.Name:              releasedFeature,
		devFeature.Name:                   devFeature,
	}

	for internal, featureData := range dataset {
		t.Run(fmt.Sprintf("internal %t", internal), func(t *testing.T) {
			for featureName, expectedLevel := range featureData {
				f := features[featureName]
				t.Run(f.Description, func(t *testing.T) {
					// when
					result := ComputeEnablementLevel(context.Background(), f, internal)
					// then
					assert.Equal(t, expectedLevel, result)
				})
			}
		})

	}
}

func TestFeatureIsEnabled(t *testing.T) {
	// given
	type FeatureLevelTest struct {
		featureLevel   FeatureLevel
		userLevel      string
		expectedResult bool
	}
	testData := []FeatureLevelTest{
		{internal, InternalLevel, true},
		{internal, ExperimentalLevel, false},
		{internal, BetaLevel, false},
		{internal, ReleasedLevel, false},
		{experimental, InternalLevel, true},
		{experimental, ExperimentalLevel, true},
		{experimental, BetaLevel, false},
		{experimental, ReleasedLevel, false},
		{beta, InternalLevel, true},
		{beta, ExperimentalLevel, true},
		{beta, BetaLevel, true},
		{beta, ReleasedLevel, false},
		{released, InternalLevel, true},
		{released, ExperimentalLevel, true},
		{released, BetaLevel, true},
		{released, ReleasedLevel, true},
		{internal, "", false},     // new user with no level specified is considered as `released` and can not use a non-`released` feature
		{experimental, "", false}, // new user with no level specified is considered as `released` and can not use a non-`released` feature
		{beta, "", false},         // new user with no level specified is considered as `released` and can not use a non-`released` feature
		{released, "", true},      // new user with no level specified is considered as `released` and can use a `released` feature
	}
	for _, test := range testData {
		t.Run(fmt.Sprintf("%s vs %v -> %t", fromFeatureLevel(test.featureLevel), test.userLevel, test.expectedResult), func(t *testing.T) {
			assert.Equal(t, test.expectedResult, test.featureLevel.IsEnabled(test.userLevel))
		})
	}
}
func TestFeatureLevelConversion(t *testing.T) {

	t.Run("convert from feature level type", func(t *testing.T) {
		// given
		dataSet := map[FeatureLevel]string{
			internal:     InternalLevel,
			experimental: ExperimentalLevel,
			beta:         BetaLevel,
			released:     ReleasedLevel,
			unknown:      UnknownLevel,
		}
		// iterate over the data set
		for inputFeatureLevel, expectedValue := range dataSet {
			t.Run(expectedValue, func(t *testing.T) {
				// when
				result := fromFeatureLevel(inputFeatureLevel)
				// then
				require.NotNil(t, result)
				assert.Equal(t, expectedValue, result)
			})
		}
	})

	t.Run("convert to feature level type", func(t *testing.T) {
		// given
		dataSet := map[string]FeatureLevel{
			InternalLevel:     internal,
			ExperimentalLevel: experimental,
			BetaLevel:         beta,
			ReleasedLevel:     released,
			UnknownLevel:      unknown,
		}
		// iterate over the data set
		for inputValue, expectedFeatureLevel := range dataSet {
			t.Run(inputValue, func(t *testing.T) {
				// when
				result := toFeatureLevel(inputValue, unknown)
				// then
				assert.Equal(t, expectedFeatureLevel, result)
			})
		}
	})
}
