package featuretoggles

import (
	"testing"

	unleashcontext "github.com/Unleash/unleash-client-go/context"
	"github.com/stretchr/testify/assert"
)

func TestFeatureIsEnabledByEmail(t *testing.T) {

	// given
	s := EnableByEmailsStrategy{}
	ctx := &unleashcontext.Context{
		Properties: map[string]string{
			EmailsParameter: "foo@foo.com",
		},
	}
	t.Run("yes", func(t *testing.T) {
		t.Run("single email", func(t *testing.T) {
			// given
			settings := map[string]interface{}{
				EmailsParameter: "foo@foo.com",
			}
			// when
			result := s.IsEnabled(settings, ctx)
			// then
			assert.True(t, result)
		})
		t.Run("multiple emails", func(t *testing.T) {
			// given
			settings := map[string]interface{}{
				EmailsParameter: "baz@foo.com,foo@foo.com,bar@foo.com",
			}
			// when
			result := s.IsEnabled(settings, ctx)
			// then
			assert.True(t, result)
		})
	})

	t.Run("no", func(t *testing.T) {
		// given
		settings := map[string]interface{}{
			EmailsParameter: "baz@foo.com,bar@foo.com",
		}
		// when
		result := s.IsEnabled(settings, ctx)
		// then
		assert.False(t, result)
	})

}
