package controller

import (
	"context"
	"fmt"
	"testing"

	"github.com/Unleash/unleash-client-go"
	unleashcontext "github.com/Unleash/unleash-client-go/context"
	unleashstrategy "github.com/Unleash/unleash-client-go/strategy"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/fabric8-services/fabric8-toggles-service/app/test"
	"github.com/fabric8-services/fabric8-toggles-service/featuretoggles"
	"github.com/goadesign/goa"
	goajwt "github.com/goadesign/goa/middleware/security/jwt"
	"github.com/magiconair/properties/assert"
	"github.com/stretchr/testify/require"
)

type MockUnleashClient struct {
	Features   []MockFeature
	Strategies []unleashstrategy.Strategy
}

// getStrategy looks-up the strategy by its name
func (m *MockUnleashClient) getStrategy(name string) unleashstrategy.Strategy {
	for _, s := range m.Strategies {
		if s.Name() == name {
			return s
		}
	}
	return nil
}

// GetEnabledFeatures mimicks the behaviour of the real client, ie, it uses the strategies to verify the features
func (m *MockUnleashClient) GetEnabledFeatures(ctx *unleashcontext.Context) []string {
	result := make([]string, 0)
	for _, f := range m.Features {
		for _, s := range f.Strategies {
			foundStrategy := m.getStrategy(s.Name())
			if foundStrategy == nil {
				// TODO: warnOnce missingStrategy
				continue
			}
			if foundStrategy.IsEnabled(f.Parameters, ctx) {
				result = append(result, f.Name)
			}
		}
	}
	return result
}

// IsFeatureEnabled mimicks the behaviour of the real client, always returns true
func (c *MockUnleashClient) IsEnabled(feature string, options ...unleash.FeatureOption) (enabled bool) {
	if feature == "ENABLED" {
		return true
	}
	return false
}

func (m *MockUnleashClient) Close() error {
	return nil
}

func (m *MockUnleashClient) Ready() <-chan bool {
	return nil
}

type MockFeature struct {
	Name       string
	Parameters map[string]interface{}
	Strategies []unleashstrategy.Strategy
}

func NewFakeFeatureList(length int) []MockFeature {
	res := make([]MockFeature, length)
	for i := 0; i < length; i++ {
		f := MockFeature{}
		f.Name = fmt.Sprintf("Feature %d", i)
		if i%2 == 0 {
			f.Parameters = map[string]interface{}{"groupID": "BETA"}
		} else {
			f.Parameters = map[string]interface{}{"groupID": "Red Hat"}
		}
		f.Strategies = []unleashstrategy.Strategy{&featuretoggles.EnableByGroupIDStrategy{}}
		res = append(res, f)
	}
	return res
}

func TestListFeatures(t *testing.T) {
	// given
	svc := goa.New("feature")
	ctrl := FeaturesController{
		Controller: svc.NewController("FeaturesController"),
		client: &featuretoggles.Client{
			UnleashClient: &MockUnleashClient{
				Features:   NewFakeFeatureList(4),
				Strategies: []unleashstrategy.Strategy{&featuretoggles.EnableByGroupIDStrategy{}},
			},
		},
	}

	t.Run("OK with jwt token without groupID claim", func(t *testing.T) {
		// when/then
		test.ListFeaturesUnauthorized(t, createInvalidContext(), svc, &ctrl)
	})
	t.Run("OK with jwt token containing groupID", func(t *testing.T) {
		// when
		_, featuresList := test.ListFeaturesOK(t, createValidContext(), svc, &ctrl)
		// then
		require.Equal(t, 2, len(featuresList.Data))
		assert.Equal(t, *featuresList.Data[0].Attributes.GroupID, "Red Hat")
		assert.Equal(t, *featuresList.Data[1].Attributes.GroupID, "Red Hat")
	})
	//t.Run("Unauhorized - no token", func(t *testing.T) {
	//	test.ListFeaturesUnauthorized(t, context.Background(), svc, ctrl)
	//})
	//t.Run("Not found", func(t *testing.T) {
	//	test.ListFeaturesNotFound(t, context.Background(), svc, ctrl)
	//})
}

func createValidContext() context.Context {
	claims := jwt.MapClaims{}
	claims["company"] = "Red Hat" // TODO replace by BETA
	token := jwt.NewWithClaims(jwt.SigningMethodRS512, claims)
	return goajwt.WithJWT(context.Background(), token)
}

func createInvalidContext() context.Context {
	claims := jwt.MapClaims{}
	token := jwt.NewWithClaims(jwt.SigningMethodRS512, claims)
	return goajwt.WithJWT(context.Background(), token)
}
