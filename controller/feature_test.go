package controller

import (
	unleashstrategy "github.com/Unleash/unleash-client-go/strategy"
	"github.com/fabric8-services/fabric8-toggles-service/app/test"
	"github.com/fabric8-services/fabric8-toggles-service/featuretoggles"
	"github.com/goadesign/goa"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestShowFeatures(t *testing.T) {
	// given
	svc := goa.New("feature")
	ctrl := FeatureController{
		Controller: svc.NewController("FeatureController"),
		client: &featuretoggles.Client{
			UnleashClient: &MockUnleashClient{
				Features:   NewFakeFeatureList(4),
				Strategies: []unleashstrategy.Strategy{&featuretoggles.EnableByGroupIDStrategy{}},
			},
		},
	}

	t.Run("OK with jwt token without groupID claim", func(t *testing.T) {
		// when/then
		test.ShowFeatureUnauthorized(t, createInvalidContext(), svc, &ctrl, "Planner")
	})
	t.Run("OK with jwt token containing groupID for a enabled feature", func(t *testing.T) {
		// when
		feature := test.ShowFeatureOK(t, createValidContext(), svc, &ctrl, "ENABLED")
		// then
		require.NotNil(t, feature)
	})
	t.Run("OK with jwt token containing groupID for a non-enabled feature", func(t *testing.T) {
		// when
		feature := test.ShowFeatureOK(t, createValidContext(), svc, &ctrl, "Planner")
		// then
		require.NotNil(t, feature)
	})
}
