package featuretoggles

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func TestFeatureLevelConversion(t *testing.T) {

	t.Run("convert from feature level type", func(t *testing.T) {
		// given
		internalLevel := InternalLevel
		experimentalLevel := ExperimentalLevel
		betaLevel := BetaLevel
		releasedLevel := ReleasedLevel
		dataSet := map[FeatureLevel]*string{
			internal:     &internalLevel,
			experimental: &experimentalLevel,
			beta:         &betaLevel,
			released:     &releasedLevel,
			unknown:      nil,
		}
		// iterate over the data set
		for inputFeatureLevel, expectedValue := range dataSet {
			if expectedValue != nil {
				t.Run(*expectedValue, func(t *testing.T) {
					// when
					result := fromFeatureLevel(inputFeatureLevel)
					// then
					require.NotNil(t, result)
					assert.Equal(t, *expectedValue, *result)
				})
			} else {
				t.Run("unknown value", func(t *testing.T) {
					// when
					result := fromFeatureLevel(inputFeatureLevel)
					// then
					require.Nil(t, result)
				})
			}
		}
	})

	t.Run("convert to feature level type", func(t *testing.T) {
		t.Run("with internal user", func(t *testing.T) {
			// given
			internalUserDataSet := map[string]FeatureLevel{
				InternalLevel:     internal,
				ExperimentalLevel: experimental,
				BetaLevel:         beta,
				ReleasedLevel:     released,
				UnknownLevel:      unknown,
			}
			// iterate over the data set
			for inputValue, expectedFeatureLevel := range internalUserDataSet {
				t.Run(inputValue, func(t *testing.T) {
					// when
					result := ToFeatureLevel(inputValue, true)
					// then
					assert.Equal(t, expectedFeatureLevel, result)
				})
			}
		})
		t.Run("with external user", func(t *testing.T) {
			// given
			externalUserDataSet := map[string]FeatureLevel{
				InternalLevel:     unknown,
				ExperimentalLevel: experimental,
				BetaLevel:         beta,
				ReleasedLevel:     released,
				UnknownLevel:      unknown,
			}
			// iterate over the data set
			for inputValue, expectedFeatureLevel := range externalUserDataSet {
				t.Run(inputValue, func(t *testing.T) {
					// when
					result := ToFeatureLevel(inputValue, false)
					// then
					assert.Equal(t, expectedFeatureLevel, result)
				})
			}
		})
	})
}
