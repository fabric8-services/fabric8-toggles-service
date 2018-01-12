package controller_test

import (
	"context"
	"crypto/rsa"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	unleashapi "github.com/Unleash/unleash-client-go/api"
	unleashstrategy "github.com/Unleash/unleash-client-go/strategy"
	jwt "github.com/dgrijalva/jwt-go"
	jwtrequest "github.com/dgrijalva/jwt-go/request"
	"github.com/dnaeon/go-vcr/cassette"
	"github.com/dnaeon/go-vcr/recorder"
	"github.com/fabric8-services/fabric8-auth/log"
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

type TestfeatureBontrollerConfig struct {
	authServiceURL string
}

func (c *TestfeatureBontrollerConfig) GetAuthServiceURL() string {
	return c.authServiceURL
}

func NewFeaturesController(r *recorder.Recorder) (*goa.Service, *controller.FeaturesController) {
	svc := goa.New("feature")
	// given
	featureA := unleashapi.Feature{
		Name:        "FeatureA",
		Description: "Feature description",
		Enabled:     false,
		Strategies:  []unleashapi.Strategy{},
	}

	featureB := unleashapi.Feature{
		Name:        "FeatureB",
		Description: "Feature description",
		Enabled:     true,
		Strategies: []unleashapi.Strategy{
			{
				Name: featuretoggles.EnableByLevel,
				Parameters: map[string]interface{}{
					"level": featuretoggles.InternalLevel,
				},
			},
			{
				Name: featuretoggles.EnableByLevel,
				Parameters: map[string]interface{}{
					"level": featuretoggles.ExperimentalLevel,
				},
			},
		},
	}
	unleashClient := &unleashtestclient.MockUnleashClient{
		Features: []unleashapi.Feature{
			featureA,
			featureB,
			featureB,
		},
		Strategies: []unleashstrategy.Strategy{
			&featuretoggles.EnableByLevelStrategy{},
		},
	}

	ctrl := controller.NewFeaturesController(svc,
		featuretoggles.NewClientWithState(unleashClient, true),
		&http.Client{
			Transport: r.Transport,
		},
		&TestfeatureBontrollerConfig{
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
	_, err = PublicKey()
	require.NoError(t, err)

	// custom cassette matcher that will compare the HTTP requests' token subject with the `sub` header of the recorded data (the yaml file)
	r.SetMatcher(JWTMatcher())
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
			test.ShowFeaturesNotFound(t, createValidContext(t, "user_foo"), svc, ctrl, "FeatureZ")
		})
	})

	t.Run("disabled for user", func(t *testing.T) {

		t.Run("did not opt-in", func(t *testing.T) {
			// when
			_, appFeature := test.ShowFeaturesOK(t, createValidContext(t, "user_baz"), svc, ctrl, "FeatureB")
			// then
			require.NotNil(t, appFeature)
			enablementLevel := featuretoggles.ExperimentalLevel
			expectedFeatureData := &app.Feature{
				ID:   "FeatureB",
				Type: "features",
				Attributes: &app.FeatureAttributes{
					Description:     "Feature description",
					Enabled:         true,
					UserEnabled:     false,
					EnablementLevel: &enablementLevel,
				},
			}
			assert.Equal(t, expectedFeatureData, appFeature.Data)
		})

		t.Run("disabled for all", func(t *testing.T) {
			// when
			_, appFeature := test.ShowFeaturesOK(t, createValidContext(t, "user_foo"), svc, ctrl, "FeatureA")
			// then
			require.NotNil(t, appFeature)
			expectedFeatureData := &app.Feature{
				ID:   "FeatureA",
				Type: "features",
				Attributes: &app.FeatureAttributes{
					Description:     "Feature description",
					Enabled:         false,
					UserEnabled:     false,
					EnablementLevel: nil,
				},
			}
			assert.Equal(t, expectedFeatureData, appFeature.Data)
		})
	})

	t.Run("enabled for user", func(t *testing.T) {
		// when
		_, appFeature := test.ShowFeaturesOK(t, createValidContext(t, "user_bar"), svc, ctrl, "FeatureB")
		// then
		require.NotNil(t, appFeature)
		enablementLevel := featuretoggles.ExperimentalLevel
		expectedFeatureData := &app.Feature{
			ID:   "FeatureB",
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
}

func TestListFeatures(t *testing.T) {
	// given
	cassetteName := "../test/data/controller/auth_get_user"
	_, err := os.Stat(fmt.Sprintf("%s.yaml", cassetteName))
	require.NoError(t, err)
	r, err := recorder.New(cassetteName)
	require.NoError(t, err)
	_, err = PublicKey()
	require.NoError(t, err)

	// custom cassette matcher that will compare the HTTP requests' token subject with the `sub` header of the recorded data (the yaml file)
	r.SetMatcher(JWTMatcher())
	require.NoError(t, err)
	defer r.Stop()
	svc, ctrl := NewFeaturesController(r)

	t.Run("fail", func(t *testing.T) {
		t.Run("unauthorized", func(t *testing.T) {
			// when/then
			test.ListFeaturesUnauthorized(t, createInvalidContext(), svc, ctrl, []string{"FeatureA", "FeatureB"})
		})
	})

	t.Run("ok", func(t *testing.T) {
		t.Run("2 matches", func(t *testing.T) {
			// when
			_, featuresList := test.ListFeaturesOK(t, createValidContext(t, "user_foo"), svc, ctrl, []string{"FeatureA", "FeatureB"})
			// then
			experimentalLevel := featuretoggles.ExperimentalLevel
			expectedData := []*app.Feature{
				&app.Feature{
					ID:   "FeatureA",
					Type: "features",
					Attributes: &app.FeatureAttributes{
						Description:     "Feature description",
						Enabled:         false,
						UserEnabled:     false,
						EnablementLevel: nil,
					},
				},
				&app.Feature{
					ID:   "FeatureB",
					Type: "features",
					Attributes: &app.FeatureAttributes{
						Description:     "Feature description",
						Enabled:         true,
						UserEnabled:     true,
						EnablementLevel: &experimentalLevel,
					},
				},
			}
			assert.Equal(t, expectedData, featuresList.Data)
		})

		t.Run("no feature found", func(t *testing.T) {
			// when
			_, featuresList := test.ListFeaturesOK(t, createValidContext(t, "user_foo"), svc, ctrl, []string{"FeatureX", "FeatureY", "FeatureZ"})
			// then
			expectedData := []*app.Feature{}
			assert.Equal(t, expectedData, featuresList.Data)
		})
	})

}

func JWTMatcher() cassette.Matcher {
	return func(httpRequest *http.Request, cassetteRequest cassette.Request) bool {
		// look-up the JWT's "sub" claim and compare with the request
		token, err := jwtrequest.ParseFromRequest(httpRequest, jwtrequest.AuthorizationHeaderExtractor, func(*jwt.Token) (interface{}, error) {
			return PublicKey()
		})
		if err != nil {
			log.Panic(nil, map[string]interface{}{"error": err.Error()}, "failed to parse token from request")
		}
		claims := token.Claims.(jwt.MapClaims)
		if sub, found := cassetteRequest.Headers["sub"]; found {
			return sub[0] == claims["sub"]
		}
		return false
	}
}

func createValidContext(t *testing.T, userID string) context.Context {
	claims := jwt.MapClaims{}
	claims["sub"] = userID
	token := jwt.NewWithClaims(jwt.SigningMethodRS512, claims)
	// use the test private key to sign the token
	key, err := PrivateKey()
	require.NoError(t, err)
	signed, err := token.SignedString(key)
	require.NoError(t, err)
	token.Raw = signed
	return goajwt.WithJWT(context.Background(), token)
}

func createInvalidContext() context.Context {
	return context.Background()
}

func PrivateKey() (*rsa.PrivateKey, error) {
	rsaPrivateKey, err := ioutil.ReadFile("../test/private_key.pem")
	if err != nil {
		return nil, err
	}
	return jwt.ParseRSAPrivateKeyFromPEM(rsaPrivateKey)
}

func PublicKey() (*rsa.PublicKey, error) {
	rsaPublicKey, err := ioutil.ReadFile("../test/public_key.pem")
	if err != nil {
		return nil, err
	}
	return jwt.ParseRSAPublicKeyFromPEM(rsaPublicKey)
}
