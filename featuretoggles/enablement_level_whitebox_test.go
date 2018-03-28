package featuretoggles

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

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
