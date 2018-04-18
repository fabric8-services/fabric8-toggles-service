package featuretoggles

import (
	"testing"

	unleashcontext "github.com/Unleash/unleash-client-go/context"
	"github.com/stretchr/testify/assert"
)

func TestFeatureIsEnabledByLevel(t *testing.T) {

	// given
	s := EnableByLevelStrategy{}
	settings := map[string]interface{}{
		LevelParameter: BetaLevel,
	}

	t.Run("yes", func(t *testing.T) {
		t.Run("same level", func(t *testing.T) {
			// given
			ctx := &unleashcontext.Context{
				Properties: map[string]string{
					LevelParameter: "beta",
				},
			}
			// when
			result := s.IsEnabled(settings, ctx)
			// then
			assert.True(t, result)
		})

		t.Run("lower level", func(t *testing.T) {
			// given
			ctx := &unleashcontext.Context{
				Properties: map[string]string{
					LevelParameter: "internal",
				},
			}
			// when
			result := s.IsEnabled(settings, ctx)
			// then
			assert.True(t, result)
		})
	})

	t.Run("no", func(t *testing.T) {
		t.Run("upper level", func(t *testing.T) {
			// given
			ctx := &unleashcontext.Context{
				Properties: map[string]string{
					EmailsParameter: "released",
				},
			}
			// when
			result := s.IsEnabled(settings, ctx)
			// then
			assert.False(t, result)
		})
	})

}
