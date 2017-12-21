package controller_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"

	unleashapi "github.com/Unleash/unleash-client-go/api"
	unleashstrategy "github.com/Unleash/unleash-client-go/strategy"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/dnaeon/go-vcr/recorder"
	"github.com/fabric8-services/fabric8-toggles-service/app"
	"github.com/fabric8-services/fabric8-toggles-service/app/test"
	"github.com/fabric8-services/fabric8-toggles-service/controller"
	"github.com/fabric8-services/fabric8-toggles-service/featuretoggles"
	unleashtestclient "github.com/fabric8-services/fabric8-toggles-service/test/unleashclient"
	"github.com/goadesign/goa"
	goajwt "github.com/goadesign/goa/middleware/security/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestFeatureControllerConfig struct {
	authServiceURL string
}

func (c *TestFeatureControllerConfig) GetAuthServiceURL() string {
	return c.authServiceURL
}

func NewFeaturesController(r *recorder.Recorder) (*goa.Service, *controller.FeaturesController) {
	svc := goa.New("feature")
	// given
	featureA := unleashapi.Feature{
		Name:        "FeatureA",
		Description: "Feature description",
		Enabled:     true,
		Strategies:  []unleashapi.Strategy{},
	}

	featureB := unleashapi.Feature{
		Name:        "FeatureB",
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

	featureC := unleashapi.Feature{
		Name:        "FeatureC",
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
	unleashClient := &unleashtestclient.MockUnleashClient{
		Features: []unleashapi.Feature{
			featureA,
			featureB,
			featureC,
		},
		Strategies: []unleashstrategy.Strategy{
			&featuretoggles.EnableByGroupIDStrategy{},
		},
	}

	ctrl := controller.NewFeaturesController(svc,
		featuretoggles.NewCustomClient(unleashClient, true),
		&http.Client{
			Transport: r.Transport,
		},
		&TestFeatureControllerConfig{
			authServiceURL: "http://auth",
		},
	)
	return svc, ctrl
}

func TestShowFeature(t *testing.T) {
	// given
	cassetteName := "../test/data/controller/auth_get_user"
	_, err := os.Stat(fmt.Sprintf("%s.yaml", cassetteName))
	require.NoError(t, err)
	r, err := recorder.New(cassetteName)
	require.NoError(t, err)
	defer r.Stop()
	svc, ctrl := NewFeaturesController(r)

	t.Run("fail", func(t *testing.T) {
		t.Run("unauthorized", func(t *testing.T) {
			// when/then
			test.ShowFeaturesUnauthorized(t, createInvalidContext(), svc, ctrl, "FeatureA")
		})
		t.Run("not found", func(t *testing.T) {
			// when/then
			test.ShowFeaturesNotFound(t, createValidContext(), svc, ctrl, "FeatureZ")
		})
	})

	t.Run("ok", func(t *testing.T) {
		t.Run("feature enabled for user", func(t *testing.T) {
			// when
			_, appFeature := test.ShowFeaturesOK(t, createValidContext(), svc, ctrl, "FeatureC")
			// then
			require.NotNil(t, appFeature)
			enablementLevel := "beta"
			expectedFeatureData := &app.Feature{
				ID:   "FeatureC",
				Type: "features",
				Attributes: &app.FeatureAttributes{
					Description:     "Feature description",
					Enabled:         true,
					UserEnabled:     true,
					EnablementLevel: &enablementLevel,
				},
			}
			assert.Equal(t, expectedFeatureData, appFeature.Data)
		})
	})

	// t.Run("OK with jwt token containing groupID for a non-enabled feature", func(t *testing.T) {
	// 	// when
	// 	feature := test.ShowFeatureOK(t, createValidContext(), svc, ctrl, "Planner")
	// 	// then
	// 	require.NotNil(t, feature)
	// })
}

func createValidContext() context.Context {
	claims := jwt.MapClaims{}
	claims["company"] = "Red Hat" // TODO replace by BETA
	token := jwt.NewWithClaims(jwt.SigningMethodRS512, claims)
	return goajwt.WithJWT(context.Background(), token)
}

func createInvalidContext() context.Context {
	return context.Background()
}

// func TestListFeatures(t *testing.T) {
// 	// given
// 	svc, ctrl := NewFeaturesController(nil)

// 	t.Run("Unauhorized - no token", func(t *testing.T) {
// 		// when/then
// 		test.ListFeaturesUnauthorized(t, createInvalidContext(), svc, ctrl)
// 	})
// 	// t.Run("OK with jwt token containing groupID", func(t *testing.T) {
// 	// 	// when
// 	// 	_, featuresList := test.ListFeaturesOK(t, createValidContext(), svc, ctrl)
// 	// 	// then
// 	// 	require.Equal(t, 2, len(featuresList.Data))
// 	// 	assert.Equal(t, *featuresList.Data[0].Attributes.GroupID, "experimental")
// 	// 	assert.Equal(t, *featuresList.Data[1].Attributes.GroupID, "beta")
// 	// })
// 	// t.Run("Not found", func(t *testing.T) {
// 	// 	test.ListFeaturesNotFound(t, createValidContext(), svc, ctrl)
// 	// })
// }
