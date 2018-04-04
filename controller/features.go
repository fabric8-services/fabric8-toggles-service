package controller

import (
	"context"
	"net/http"
	"strings"

	unleashapi "github.com/Unleash/unleash-client-go/api"
	"github.com/fabric8-services/fabric8-auth/goasupport"
	"github.com/fabric8-services/fabric8-auth/log"
	"github.com/fabric8-services/fabric8-auth/rest"
	"github.com/fabric8-services/fabric8-auth/token"
	"github.com/fabric8-services/fabric8-toggles-service/app"
	"github.com/fabric8-services/fabric8-toggles-service/auth"
	"github.com/fabric8-services/fabric8-toggles-service/auth/client"
	"github.com/fabric8-services/fabric8-toggles-service/errors"
	"github.com/fabric8-services/fabric8-toggles-service/featuretoggles"
	"github.com/fabric8-services/fabric8-toggles-service/jsonapi"
	"github.com/goadesign/goa"
	goajwt "github.com/goadesign/goa/middleware/security/jwt"
	errs "github.com/pkg/errors"
)

// FeaturesController implements the features resource.
type FeaturesController struct {
	*goa.Controller
	config        FeaturesControllerConfig
	togglesClient featuretoggles.Client
	httpClient    *http.Client
	tokenParser   token.Parser
}

// FeaturesControllerConfig the configuration required for the FeaturesController
type FeaturesControllerConfig interface {
	featuretoggles.ToggleServiceConfiguration
	GetAuthServiceURL() string
}

// NewFeaturesController creates a FeaturesController.
func NewFeaturesController(service *goa.Service, tokenParser token.Parser, config FeaturesControllerConfig, options ...FeaturesControllerOption) *FeaturesController {
	// init the toggle client
	ctrl := FeaturesController{
		httpClient:  http.DefaultClient,
		Controller:  service.NewController("FeaturesController"),
		tokenParser: tokenParser,
		config:      config,
	}
	// apply options
	for _, opt := range options {
		opt(&ctrl)
	}
	if ctrl.togglesClient == nil {
		togglesClient, err := featuretoggles.NewDefaultClient("fabric8-toggle-service", config)
		if err != nil {
			log.Panic(nil, map[string]interface{}{
				"err": err,
			}, "failed to create toogle client")
		}
		ctrl.togglesClient = togglesClient
	}

	return &ctrl
}

// FeaturesControllerOption a function to customize the FeaturesController during its initialization
type FeaturesControllerOption func(*FeaturesController)

// WithHTTPClient configure the FeatureController with a custom HTTP client
func WithHTTPClient(client *http.Client) FeaturesControllerOption {
	return func(ctrl *FeaturesController) {
		ctrl.httpClient = client
	}
}

// WithTogglesClient configure the FeatureController with a custom Toggles client
func WithTogglesClient(client featuretoggles.Client) FeaturesControllerOption {
	return func(ctrl *FeaturesController) {
		ctrl.togglesClient = client
	}
}

// List runs the list action.
func (c *FeaturesController) List(ctx *app.ListFeaturesContext) error {
	var user *client.User
	jwtToken := goajwt.ContextJWT(ctx)
	if jwtToken != nil {
		_, err := c.tokenParser.Parse(ctx, jwtToken.Raw)
		if err != nil {
			log.Error(ctx, map[string]interface{}{"error": err.Error()}, "error while parsing the user's token")
			return jsonapi.JSONErrorResponse(ctx, errors.NewUnauthorizedError("invalid token"))
		}
		if user, err = c.getUserProfile(ctx); err != nil {
			return jsonapi.JSONErrorResponse(ctx, err)
		}
	} else {
		log.Warn(ctx, map[string]interface{}{}, "No JWT found in the request.")
	}
	// look-up by pattern
	if ctx.Page != nil {
		features := c.togglesClient.GetFeaturesByPattern(ctx, *ctx.Page)
		appFeatures := c.convertFeatures(ctx, features, user)
		return ctx.OK(appFeatures)

	} else if ctx.Names != nil {
		features := c.togglesClient.GetFeaturesByName(ctx, ctx.Names)
		appFeatures := c.convertFeatures(ctx, features, user)
		return ctx.OK(appFeatures)
	}
	// default, empty response
	return ctx.OK(&app.FeatureList{
		Data: []*app.Feature{},
	})
}

// Show runs the show action.
func (c *FeaturesController) Show(ctx *app.ShowFeaturesContext) error {
	jwtToken := goajwt.ContextJWT(ctx)
	var user *client.User
	if jwtToken != nil {
		_, err := c.tokenParser.Parse(ctx, jwtToken.Raw)
		if err != nil {
			log.Error(ctx, map[string]interface{}{"error": err.Error()}, "error while parsing the user's token")
			return jsonapi.JSONErrorResponse(ctx, errors.NewUnauthorizedError("invalid token"))
		}
		if user, err = c.getUserProfile(ctx); err != nil {
			return jsonapi.JSONErrorResponse(ctx, err)
		}
	} else {
		log.Warn(ctx, map[string]interface{}{}, "No JWT found in the request.")
	}
	featureName := ctx.FeatureName
	feature := c.togglesClient.GetFeature(ctx, featureName)
	appFeature := c.convertFeature(ctx, featureName, feature, user)
	return ctx.OK(appFeature)
}

// getUserProfile retrieves the user's profile from the auth service, by forwarding the current JWT token
func (c *FeaturesController) getUserProfile(ctx context.Context) (*client.User, error) {
	authClient, err := auth.NewClient(ctx, c.config.GetAuthServiceURL(), auth.WithHTTPClient(c.httpClient))
	if err != nil {
		log.Error(ctx, map[string]interface{}{
			"err": err.Error(),
		}, "unable to initialize auth service client")
		return nil, errs.Wrap(err, "unable to initialize auth service client")
	}
	res, err := authClient.ShowUser(goasupport.ForwardContextRequestID(ctx), client.ShowUserPath(), nil, nil)
	if err != nil {
		log.Error(ctx, map[string]interface{}{
			"err": err.Error(),
		}, "unable to get user from the auth service")
		return nil, errs.Wrap(err, "unable to get user from the auth service")
	}
	defer res.Body.Close()
	switch res.StatusCode {
	case 200:
	// OK
	case 401:
		return nil, errors.NewUnauthorizedError(rest.ReadBody(res.Body))
	default:
		return nil, errs.Errorf("status: %s, body: %s", res.Status, rest.ReadBody(res.Body))
	}
	return authClient.DecodeUser(res)
}

func (c *FeaturesController) convertFeatures(ctx context.Context, features []unleashapi.Feature, user *client.User) *app.FeatureList {
	result := make([]*app.Feature, 0)
	for _, feature := range features {
		result = append(result, c.convertFeatureData(ctx, feature.Name, feature, user))
	}
	return &app.FeatureList{
		Data: result,
	}
}

func (c *FeaturesController) convertFeature(ctx context.Context, name string, feature *unleashapi.Feature, user *client.User) *app.FeatureSingle {
	// unknown feature has no description and is not enabled at all
	if feature == nil {
		log.Warn(ctx, map[string]interface{}{"feature_name": name}, "feature not found")
		return &app.FeatureSingle{
			Data: &app.Feature{
				ID:   name,
				Type: "features",
				Attributes: &app.FeatureAttributes{
					Description: "unknown feature",
					Enabled:     false,
					UserEnabled: false,
				},
			},
		}
	}
	return &app.FeatureSingle{
		Data: c.convertFeatureData(ctx, name, *feature, user),
	}
}

func (c *FeaturesController) convertFeatureData(ctx context.Context, name string, feature unleashapi.Feature, user *client.User) *app.Feature {
	internalUser := false
	userLevel := featuretoggles.ReleasedLevel // default level of features that the user can use
	if user != nil {
		userEmail := user.Data.Attributes.Email
		userEmailVerified := user.Data.Attributes.EmailVerified
		// internal users have may be able to access the feature by opting-in to the `internal` level of features.
		if userEmailVerified != nil && *userEmailVerified && userEmail != nil && strings.HasSuffix(*userEmail, "@redhat.com") {
			internalUser = true
		}
		// do not override the userLevel if the value is nil or empty. Any other value is accepted,
		// but will be converted (with a fallback to `unknown` if needed)
		if user.Data.Attributes.FeatureLevel != nil && *user.Data.Attributes.FeatureLevel != "" {
			userLevel = *user.Data.Attributes.FeatureLevel
		}
	}
	enabledForUser := c.togglesClient.IsFeatureEnabled(ctx, feature, userLevel)
	log.Debug(ctx, map[string]interface{}{"internal_user": internalUser}, "converting feature")
	enablementLevel := featuretoggles.ComputeEnablementLevel(ctx, feature, internalUser)
	var enablement *string
	if enablementLevel != featuretoggles.UnknownLevel { // skip value in response if enablement level is "unknown"
		enablement = &enablementLevel
	}
	return &app.Feature{
		ID:   feature.Name,
		Type: "features",
		Attributes: &app.FeatureAttributes{
			Description:     feature.Description,
			Enabled:         feature.Enabled,
			EnablementLevel: enablement,
			UserEnabled:     enabledForUser,
		},
	}
}
