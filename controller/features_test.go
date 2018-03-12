package controller_test

import (
	"context"
	"net/http"
	"reflect"
	"testing"
	"time"

	unleashapi "github.com/Unleash/unleash-client-go/api"
	jwt "github.com/dgrijalva/jwt-go"
	jwtrequest "github.com/dgrijalva/jwt-go/request"
	"github.com/dnaeon/go-vcr/cassette"
	"github.com/fabric8-services/fabric8-auth/log"
	authtoken "github.com/fabric8-services/fabric8-auth/token"
	"github.com/fabric8-services/fabric8-toggles-service/app"
	"github.com/fabric8-services/fabric8-toggles-service/app/test"
	"github.com/fabric8-services/fabric8-toggles-service/auth"
	"github.com/fabric8-services/fabric8-toggles-service/controller"
	"github.com/fabric8-services/fabric8-toggles-service/featuretoggles"
	testsupport "github.com/fabric8-services/fabric8-toggles-service/test"
	"github.com/fabric8-services/fabric8-toggles-service/test/recorder"
	"github.com/fabric8-services/fabric8-toggles-service/token"
	"github.com/goadesign/goa"
	goajwt "github.com/goadesign/goa/middleware/security/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var disabledFeature, singleStrategyFeature, multiStrategiesFeature, releasedFeature unleashapi.Feature

func init() {
	// features
	disabledFeature = unleashapi.Feature{
		Name:        "disabledFeature",
		Description: "Disabled feature",
		Enabled:     false,
		Strategies:  []unleashapi.Strategy{},
	}

	singleStrategyFeature = unleashapi.Feature{
		Name:        "singleStrategyFeature",
		Description: "Feature with single strategy",
		Enabled:     true,
		Strategies: []unleashapi.Strategy{
			{
				Name: featuretoggles.EnableByLevelStrategyName,
				Parameters: map[string]interface{}{
					featuretoggles.LevelParameter: featuretoggles.InternalLevel,
				},
			},
		},
	}

	multiStrategiesFeature = unleashapi.Feature{
		Name:        "multiStrategiesFeature",
		Description: "Feature with multiple strategies",
		Enabled:     true,
		Strategies: []unleashapi.Strategy{
			{
				Name: featuretoggles.EnableByLevelStrategyName,
				Parameters: map[string]interface{}{
					featuretoggles.LevelParameter: featuretoggles.InternalLevel,
				},
			},
			{
				Name: featuretoggles.EnableByLevelStrategyName,
				Parameters: map[string]interface{}{
					featuretoggles.LevelParameter: featuretoggles.ExperimentalLevel,
				},
			},
			{
				Name: featuretoggles.EnableByLevelStrategyName,
				Parameters: map[string]interface{}{
					featuretoggles.LevelParameter: featuretoggles.BetaLevel,
				},
			},
		},
	}

	releasedFeature = unleashapi.Feature{
		Name:        "releasedFeature",
		Description: "Feature released",
		Enabled:     true,
		Strategies: []unleashapi.Strategy{
			{
				Name: featuretoggles.EnableByLevelStrategyName,
				Parameters: map[string]interface{}{
					featuretoggles.LevelParameter: featuretoggles.ReleasedLevel,
				},
			},
		},
	}
}

// MockTogglesClient a mock of the toggles client
type MockTogglesClient struct {
}

func (m *MockTogglesClient) IsFeatureEnabled(ctx context.Context, feature unleashapi.Feature, userLevel string) bool {
	if reflect.DeepEqual(feature, disabledFeature) {
		return false
	}
	if reflect.DeepEqual(feature, multiStrategiesFeature) && userLevel == featuretoggles.ExperimentalLevel {
		return true
	}
	if reflect.DeepEqual(feature, multiStrategiesFeature) && userLevel == featuretoggles.BetaLevel {
		return true
	}
	if reflect.DeepEqual(feature, releasedFeature) {
		return true
	}

	return false
}

func (m *MockTogglesClient) GetFeatures(ctx context.Context, names []string) []*unleashapi.Feature {
	if reflect.DeepEqual(names, []string{disabledFeature.Name, multiStrategiesFeature.Name}) {
		return []*unleashapi.Feature{&disabledFeature, &multiStrategiesFeature}
	} else if reflect.DeepEqual(names, []string{releasedFeature.Name, disabledFeature.Name, multiStrategiesFeature.Name}) {
		return []*unleashapi.Feature{&releasedFeature, &disabledFeature, &multiStrategiesFeature}
	}
	return nil
}

func (m *MockTogglesClient) GetFeature(name string) *unleashapi.Feature {
	switch name {
	case disabledFeature.Name:
		return &disabledFeature
	case singleStrategyFeature.Name:
		return &singleStrategyFeature
	case multiStrategiesFeature.Name:
		return &multiStrategiesFeature
	case releasedFeature.Name:
		return &releasedFeature
	default:
		return nil
	}
}

func (m *MockTogglesClient) Close() error {
	return nil
}

type TestFeatureControllerConfig struct {
	authServiceURL string
}

func (c *TestFeatureControllerConfig) GetAuthServiceURL() string {
	return c.authServiceURL
}

func (c *TestFeatureControllerConfig) GetTogglesURL() string {
	return ""
}

func NewFeaturesController(tokenParser authtoken.Parser, httpClient *http.Client, togglesClient featuretoggles.Client) (*goa.Service, *controller.FeaturesController) {
	svc := goa.New("feature")
	ctrl := controller.NewFeaturesController(svc,
		tokenParser,
		&TestFeatureControllerConfig{
			authServiceURL: "http://auth",
		},
		controller.WithHTTPClient(httpClient),
		controller.WithTogglesClient(&MockTogglesClient{}),
	)
	return svc, ctrl
}

func TestShowFeatures(t *testing.T) {
	// given
	r1, err := recorder.New("../test/data/controller/auth_get_user", recorder.WithMatcher(JWTMatcher()))
	require.NoError(t, err)
	defer r1.Stop()
	r2, err := recorder.New("../test/data/token/auth_get_keys")
	require.NoError(t, err)
	c, err := auth.NewClient(
		context.Background(),
		"http://authservice",
		auth.WithHTTPClient(
			&http.Client{
				Transport: r2.Transport,
			}),
	)
	require.NoError(t, err)
	p, err := token.NewParser(c)
	require.NoError(t, err)
	svc, ctrl := NewFeaturesController(p, &http.Client{Transport: r1.Transport}, &MockTogglesClient{})

	t.Run("disabled for user", func(t *testing.T) {

		t.Run("disabled for all", func(t *testing.T) {
			// when
			ctx, err := createValidContext("../test/private_key.pem", "user_beta_level", time.Now().Add(1*time.Hour))
			require.NoError(t, err)
			_, appFeature := test.ShowFeaturesOK(t, ctx, svc, ctrl, disabledFeature.Name)
			// then
			require.NotNil(t, appFeature)
			expectedFeatureData := &app.Feature{
				ID:   disabledFeature.Name,
				Type: "features",
				Attributes: &app.FeatureAttributes{
					Description:     disabledFeature.Description,
					Enabled:         false,
					UserEnabled:     false,
					EnablementLevel: nil, // feature is disabled, hence no opt-in level would work anyways
				},
			}
			assert.Equal(t, expectedFeatureData, appFeature.Data)
		})

		t.Run("user with no level", func(t *testing.T) {
			// when
			ctx, err := createValidContext("../test/private_key.pem", "user_no_level", time.Now().Add(1*time.Hour))
			require.NoError(t, err)
			_, appFeature := test.ShowFeaturesOK(t, ctx, svc, ctrl, singleStrategyFeature.Name)
			// then
			require.NotNil(t, appFeature)
			expectedFeatureData := &app.Feature{
				ID:   singleStrategyFeature.Name,
				Type: "features",
				Attributes: &app.FeatureAttributes{
					Description:     singleStrategyFeature.Description,
					Enabled:         true,
					UserEnabled:     false,
					EnablementLevel: nil, // because the feature level is internal but the user is external
				},
			}
			assert.Equal(t, expectedFeatureData, appFeature.Data)
		})

	})

	t.Run("enabled for user", func(t *testing.T) {

		t.Run("unreleased feature", func(t *testing.T) {

			t.Run("user with enough level", func(t *testing.T) {

				t.Run("experimental user", func(t *testing.T) {
					// given
					ctx, err := createValidContext("../test/private_key.pem", "user_experimental_level", time.Now().Add(1*time.Hour))
					require.NoError(t, err)
					// when
					_, appFeature := test.ShowFeaturesOK(t, ctx, svc, ctrl, multiStrategiesFeature.Name)
					// then
					require.NotNil(t, appFeature)
					enablementLevel := featuretoggles.BetaLevel
					expectedFeatureData := &app.Feature{
						ID:   multiStrategiesFeature.Name,
						Type: "features",
						Attributes: &app.FeatureAttributes{
							Description:     multiStrategiesFeature.Description,
							Enabled:         true,
							UserEnabled:     true,
							EnablementLevel: &enablementLevel,
						},
					}
					assert.Equal(t, expectedFeatureData, appFeature.Data)
				})
			})
		})

		t.Run("released feature", func(t *testing.T) {

			t.Run("user with no level", func(t *testing.T) {
				// given
				ctx, err := createValidContext("../test/private_key.pem", "user_no_level", time.Now().Add(1*time.Hour))
				require.NoError(t, err)
				// when
				_, appFeature := test.ShowFeaturesOK(t, ctx, svc, ctrl, releasedFeature.Name)
				// then
				require.NotNil(t, appFeature)
				enablementLevel := featuretoggles.ReleasedLevel
				expectedFeatureData := &app.Feature{
					ID:   releasedFeature.Name,
					Type: "features",
					Attributes: &app.FeatureAttributes{
						Description:     releasedFeature.Description,
						Enabled:         true,
						UserEnabled:     true,
						EnablementLevel: &enablementLevel,
					},
				}
				assert.Equal(t, expectedFeatureData, appFeature.Data)
			})

			t.Run("user with enough level", func(t *testing.T) {

				t.Run("experimental level", func(t *testing.T) {
					// given
					ctx, err := createValidContext("../test/private_key.pem", "user_experimental_level", time.Now().Add(1*time.Hour))
					require.NoError(t, err)
					// when
					_, appFeature := test.ShowFeaturesOK(t, ctx, svc, ctrl, releasedFeature.Name)
					// then
					require.NotNil(t, appFeature)
					enablementLevel := featuretoggles.ReleasedLevel
					expectedFeatureData := &app.Feature{
						ID:   releasedFeature.Name,
						Type: "features",
						Attributes: &app.FeatureAttributes{
							Description:     releasedFeature.Description,
							Enabled:         true,
							UserEnabled:     true,
							EnablementLevel: &enablementLevel,
						},
					}
					assert.Equal(t, expectedFeatureData, appFeature.Data)
				})

				t.Run("released level", func(t *testing.T) {
					// given
					ctx, err := createValidContext("../test/private_key.pem", "user_released_level", time.Now().Add(1*time.Hour))
					require.NoError(t, err)
					// when
					_, appFeature := test.ShowFeaturesOK(t, ctx, svc, ctrl, releasedFeature.Name)
					// then
					require.NotNil(t, appFeature)
					enablementLevel := featuretoggles.ReleasedLevel
					expectedFeatureData := &app.Feature{
						ID:   releasedFeature.Name,
						Type: "features",
						Attributes: &app.FeatureAttributes{
							Description:     releasedFeature.Description,
							Enabled:         true,
							UserEnabled:     true,
							EnablementLevel: &enablementLevel,
						},
					}
					assert.Equal(t, expectedFeatureData, appFeature.Data)
				})

				t.Run("nopreproduction level", func(t *testing.T) {
					// given
					ctx, err := createValidContext("../test/private_key.pem", "user_nopreproduction_level", time.Now().Add(1*time.Hour))
					require.NoError(t, err)
					// when
					_, appFeature := test.ShowFeaturesOK(t, ctx, svc, ctrl, releasedFeature.Name)
					// then
					require.NotNil(t, appFeature)
					enablementLevel := featuretoggles.ReleasedLevel
					expectedFeatureData := &app.Feature{
						ID:   releasedFeature.Name,
						Type: "features",
						Attributes: &app.FeatureAttributes{
							Description:     releasedFeature.Description,
							Enabled:         true,
							UserEnabled:     true,
							EnablementLevel: &enablementLevel,
						},
					}
					assert.Equal(t, expectedFeatureData, appFeature.Data)
				})
			})
		})
	})

	t.Run("fail", func(t *testing.T) {
		t.Run("not found", func(t *testing.T) {
			// given
			ctx, err := createValidContext("../test/private_key.pem", "user_beta_level", time.Now().Add(1*time.Hour))
			require.NoError(t, err)
			// when/then
			test.ShowFeaturesNotFound(t, ctx, svc, ctrl, "UnknownFeature")
		})
	})

	t.Run("invalid", func(t *testing.T) {

		t.Run("invalid token", func(t *testing.T) {
			// given
			ctx, err := createValidContext("../test/private_key2.pem", "user_beta_level", time.Now().Add(1*time.Hour))
			require.NoError(t, err)
			// when/then
			test.ShowFeaturesUnauthorized(t, ctx, svc, ctrl, releasedFeature.Name)
		})

		t.Run("expired token", func(t *testing.T) {
			// given
			ctx, err := createValidContext("../test/private_key.pem", "user_beta_level", time.Now().Add(-1*time.Hour))
			require.NoError(t, err)
			// when/then
			test.ShowFeaturesUnauthorized(t, ctx, svc, ctrl, releasedFeature.Name)
		})
	})
}

func TestListFeatures(t *testing.T) {
	// given
	r1, err := recorder.New("../test/data/controller/auth_get_user", recorder.WithMatcher(JWTMatcher()))
	require.NoError(t, err)
	defer r1.Stop()
	r2, err := recorder.New("../test/data/token/auth_get_keys")
	require.NoError(t, err)
	c, err := auth.NewClient(
		context.Background(),
		"http://authservice",
		auth.WithHTTPClient(
			&http.Client{
				Transport: r2.Transport,
			}),
	)
	require.NoError(t, err)
	p, err := token.NewParser(c)
	require.NoError(t, err)
	svc, ctrl := NewFeaturesController(p, &http.Client{Transport: r1.Transport}, &MockTogglesClient{})

	t.Run("ok", func(t *testing.T) {
		t.Run("2 matches", func(t *testing.T) {
			// when
			ctx, err := createValidContext("../test/private_key.pem", "user_beta_level", time.Now().Add(1*time.Hour))
			require.NoError(t, err)
			_, featuresList := test.ListFeaturesOK(t, ctx, svc, ctrl, []string{disabledFeature.Name, multiStrategiesFeature.Name})
			// then
			experimentalLevel := featuretoggles.BetaLevel
			expectedData := []*app.Feature{
				{
					ID:   disabledFeature.Name,
					Type: "features",
					Attributes: &app.FeatureAttributes{
						Description:     disabledFeature.Description,
						Enabled:         false,
						UserEnabled:     false,
						EnablementLevel: nil,
					},
				},
				{
					ID:   multiStrategiesFeature.Name,
					Type: "features",
					Attributes: &app.FeatureAttributes{
						Description:     multiStrategiesFeature.Description,
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
			ctx, err := createValidContext("../test/private_key.pem", "user_beta_level", time.Now().Add(1*time.Hour))
			require.NoError(t, err)
			_, featuresList := test.ListFeaturesOK(t, ctx, svc, ctrl, []string{"FeatureX", "FeatureY", "FeatureZ"})
			// then
			expectedData := []*app.Feature{}
			assert.Equal(t, expectedData, featuresList.Data)
		})

		t.Run("no user provided", func(t *testing.T) {
			// when
			_, featuresList := test.ListFeaturesOK(t, context.Background(), svc, ctrl, []string{"FeatureX", "FeatureY", "FeatureZ"})
			// then
			expectedData := []*app.Feature{}
			assert.Equal(t, expectedData, featuresList.Data)
		})

		t.Run("no user provided only released features matches", func(t *testing.T) {
			// when
			_, featuresList := test.ListFeaturesOK(t, context.Background(), svc, ctrl, []string{releasedFeature.Name, disabledFeature.Name, multiStrategiesFeature.Name})
			// then
			releasedLevel := featuretoggles.ReleasedLevel
			betaLevel := featuretoggles.BetaLevel
			expectedData := []*app.Feature{
				{
					ID:   releasedFeature.Name,
					Type: "features",
					Attributes: &app.FeatureAttributes{
						Description:     releasedFeature.Description,
						Enabled:         true,
						UserEnabled:     true,
						EnablementLevel: &releasedLevel,
					},
				},
				{
					ID:   disabledFeature.Name,
					Type: "features",
					Attributes: &app.FeatureAttributes{
						Description:     disabledFeature.Description,
						Enabled:         false,
						UserEnabled:     false,
						EnablementLevel: nil,
					},
				},
				{
					ID:   multiStrategiesFeature.Name,
					Type: "features",
					Attributes: &app.FeatureAttributes{
						Description:     multiStrategiesFeature.Description,
						Enabled:         true,
						UserEnabled:     false,
						EnablementLevel: &betaLevel,
					},
				},
			}
			assert.Equal(t, expectedData, featuresList.Data)
		})
	})

	t.Run("invalid", func(t *testing.T) {

		t.Run("invalid token", func(t *testing.T) {
			// given
			ctx, err := createValidContext("../test/private_key2.pem", "user_beta_level", time.Now().Add(1*time.Hour))
			require.NoError(t, err)
			// when/then
			test.ListFeaturesUnauthorized(t, ctx, svc, ctrl, []string{"FeatureX", "FeatureY", "FeatureZ"})
		})

		t.Run("expired token", func(t *testing.T) {
			// given
			ctx, err := createValidContext("../test/private_key.pem", "user_beta_level", time.Now().Add(-1*time.Hour))
			require.NoError(t, err)
			// when/then
			test.ListFeaturesUnauthorized(t, ctx, svc, ctrl, []string{"FeatureX", "FeatureY", "FeatureZ"})
		})
	})

}

// JWTMatcher a cassette matcher that verifies the request method/URL and the subject of the token in the "Authorization" header.
func JWTMatcher() cassette.Matcher {
	return func(httpRequest *http.Request, cassetteRequest cassette.Request) bool {
		// check the request URI and method
		if httpRequest.Method != cassetteRequest.Method ||
			(httpRequest.URL != nil && httpRequest.URL.String() != cassetteRequest.URL) {
			log.Debug(nil, map[string]interface{}{
				"httpRequest_method":     httpRequest.Method,
				"cassetteRequest_method": cassetteRequest.Method,
				"httpRequest_url":        httpRequest.URL,
				"cassetteRequest_url":    cassetteRequest.URL,
			}, "Cassette method/url doesn't match with the current request")
			return false
		}

		// look-up the JWT's "sub" claim and compare with the request
		token, err := parseFromRequest(httpRequest, jwtrequest.AuthorizationHeaderExtractor, func(*jwt.Token) (interface{}, error) {
			return testsupport.PublicKey("../test/public_key.pem")
		})
		if err != nil {
			log.Error(nil, map[string]interface{}{"error": err.Error()}, "failed to parse token from request")
			return false
		}
		claims := token.Claims.(jwt.MapClaims)
		if sub, found := cassetteRequest.Headers["sub"]; found {
			return sub[0] == claims["sub"]
		}
		return false
	}
}

func parseFromRequest(req *http.Request, extractor jwtrequest.Extractor, keyFunc jwt.Keyfunc) (token *jwt.Token, err error) {
	// Extract token from request
	if tokStr, err := extractor.ExtractToken(req); err == nil {
		p := new(jwt.Parser)
		p.SkipClaimsValidation = true // skip claims validation here to allow for expired tokens in the tests
		return p.ParseWithClaims(tokStr, jwt.MapClaims{}, keyFunc)
	} else {
		return nil, err
	}
}

func createValidContext(filename, userID string, exp time.Time) (context.Context, error) {
	claims := jwt.MapClaims{}
	if userID != "" {
		claims["sub"] = userID
	}
	claims["exp"] = exp.Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodRS512, claims)
	token.Header["kid"] = "test_key"
	// use the test private key to sign the token
	key, err := testsupport.PrivateKey(filename)
	if err != nil {
		return nil, err
	}
	signed, err := token.SignedString(key)
	if err != nil {
		return nil, err
	}
	token.Raw = signed
	log.Debug(nil, map[string]interface{}{"signed_token": token, "subject": userID}, "generated token")
	return goajwt.WithJWT(context.Background(), token), nil
}
