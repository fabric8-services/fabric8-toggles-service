package featuretoggles_test

import (
	"context"
	"testing"

	unleashapi "github.com/Unleash/unleash-client-go/api"
	"github.com/fabric8-services/fabric8-toggles-service/featuretoggles"
	"github.com/stretchr/testify/assert"
)

func TestFeatureEnablementLevel(t *testing.T) {

	// given
	featureA := &unleashapi.Feature{
		Name:        "Feature",
		Description: "Feature description",
		Enabled:     true,
		Strategies:  []unleashapi.Strategy{},
	}

	featureB := &unleashapi.Feature{
		Name:        "Feature",
		Description: "Feature description",
		Enabled:     true,
		Strategies: []unleashapi.Strategy{
			unleashapi.Strategy{
				Name: featuretoggles.EnableByGroupID,
				Parameters: map[string]interface{}{
					"groupID": "internal",
				},
			},
		},
	}

	featureC := &unleashapi.Feature{
		Name:        "Feature",
		Description: "Feature description",
		Enabled:     true,
		Strategies: []unleashapi.Strategy{
			unleashapi.Strategy{
				Name: featuretoggles.EnableByGroupID,
				Parameters: map[string]interface{}{
					"groupID": "internal",
				},
			},
			unleashapi.Strategy{
				Name: featuretoggles.EnableByGroupID,
				Parameters: map[string]interface{}{
					"groupID": "experimental",
				},
			},
			unleashapi.Strategy{
				Name: featuretoggles.EnableByGroupID,
				Parameters: map[string]interface{}{
					"groupID": "beta",
				},
			},
		},
	}

	t.Run("internal user", func(t *testing.T) {
		// given
		internalUser := true

		t.Run("feature with no strategy", func(t *testing.T) {
			// when
			level := featuretoggles.ComputeEnablementLevel(context.Background(), featureA, internalUser)
			// then
			assert.Nil(t, level)
		})

		t.Run("feature with single strategy", func(t *testing.T) {
			// when
			level := featuretoggles.ComputeEnablementLevel(context.Background(), featureB, internalUser)
			// then
			assert.Equal(t, "internal", *level)
		})

		t.Run("feature with multiple strategies", func(t *testing.T) {
			// when
			level := featuretoggles.ComputeEnablementLevel(context.Background(), featureC, internalUser)
			// then
			assert.Equal(t, "beta", *level)
		})
	})

	t.Run("external user", func(t *testing.T) {
		// given
		internalUser := false

		t.Run("feature with no strategy", func(t *testing.T) {
			// when
			level := featuretoggles.ComputeEnablementLevel(context.Background(), featureA, internalUser)
			// then
			assert.Nil(t, level)
		})

		t.Run("feature with single strategy", func(t *testing.T) {
			// when
			level := featuretoggles.ComputeEnablementLevel(context.Background(), featureB, internalUser)
			// then
			assert.Nil(t, level)
		})

		t.Run("feature with multiple strategies", func(t *testing.T) {
			// given

			// when
			level := featuretoggles.ComputeEnablementLevel(context.Background(), featureC, internalUser)
			// then
			assert.Equal(t, "beta", *level)
		})
	})

}
