package controller

import (
	"context"
	"net/http"
	"sort"

	"github.com/fabric8-services/fabric8-auth/goasupport"
	"github.com/fabric8-services/fabric8-auth/log"
	"github.com/fabric8-services/fabric8-auth/rest"
	"github.com/fabric8-services/fabric8-auth/token"
	"github.com/fabric8-services/fabric8-toggles-service/app"
	"github.com/fabric8-services/fabric8-toggles-service/auth"
	authclient "github.com/fabric8-services/fabric8-toggles-service/auth/client"
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
	GetFeaturesCacheControl() string
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

	log.Error(ctx, nil, "LOG::GETTING.... List")
	log.Error(ctx, nil, "LOG::GETTING.... List3")

	var user *authclient.User
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
	var features []featuretoggles.UserFeature
	// look-up by pattern
	if ctx.Group != nil {
		features = c.togglesClient.GetFeaturesByPattern(ctx, *ctx.Group, user)
	} else if ctx.Names != nil {
		features = c.togglesClient.GetFeaturesByName(ctx, ctx.Names, user)
	} else if ctx.Strategy != nil { // all features with strategy enableByLevel
		features = c.togglesClient.GetFeaturesByStrategy(ctx, *ctx.Strategy, user)
	}
	if features == nil {
		log.Info(ctx, nil, "missing query params in request")
		// default, empty response
		return ctx.OK(&app.UserFeatureList{
			Data: []*app.UserFeature{},
		})
	}
	// sort features by name to make sure that the same result list is returned between 2 calls (assuming nothing changed in the settings)
	// so that ETag comparison works
	sort.Sort(featuretoggles.ByName(features))
	return ctx.ConditionalEntities(features, c.config.GetFeaturesCacheControl, func() error {
		appFeatures := c.convertFeatures(ctx, features)
		return ctx.OK(appFeatures)
	})
}

// Show runs the show action.
func (c *FeaturesController) Show(ctx *app.ShowFeaturesContext) error {
	jwtToken := goajwt.ContextJWT(ctx)
	var user *authclient.User
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
	feature := c.togglesClient.GetFeature(ctx, featureName, user)
	return ctx.ConditionalRequest(feature, c.config.GetFeaturesCacheControl, func() error {
		appFeature := c.convertFeature(ctx, featureName, feature)
		return ctx.OK(appFeature)
	})
}

// getUserProfile retrieves the user's profile from the auth service, by forwarding the current JWT token
func (c *FeaturesController) getUserProfile(ctx context.Context) (*authclient.User, error) {
	authClient, err := auth.NewClient(ctx, c.config.GetAuthServiceURL(), auth.WithHTTPClient(c.httpClient))
	if err != nil {
		log.Error(ctx, map[string]interface{}{
			"err": err.Error(),
		}, "unable to initialize auth service client")
		return nil, errs.Wrap(err, "unable to initialize auth service client")
	}
	res, err := authClient.ShowUser(goasupport.ForwardContextRequestID(ctx), authclient.ShowUserPath(), nil, nil)
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

func (c *FeaturesController) convertFeatures(ctx context.Context, features []featuretoggles.UserFeature) *app.UserFeatureList {
	result := make([]*app.UserFeature, 0)
	for _, feature := range features {
		result = append(result, c.convertFeatureData(ctx, feature.Name, feature))
	}
	return &app.UserFeatureList{
		Data: result,
	}
}

func (c *FeaturesController) convertFeature(ctx context.Context, name string, feature featuretoggles.UserFeature) *app.UserFeatureSingle {
	// unknown feature has no description and is not enabled at all
	if feature == featuretoggles.ZeroUserFeature {
		log.Warn(ctx, map[string]interface{}{"feature_name": name}, "feature not found")
		return &app.UserFeatureSingle{
			Data: &app.UserFeature{
				ID:   name,
				Type: "features",
				Attributes: &app.UserFeatureAttributes{
					Description: "unknown feature",
					Enabled:     false,
					UserEnabled: false,
				},
			},
		}
	}
	return &app.UserFeatureSingle{
		Data: c.convertFeatureData(ctx, name, feature),
	}
}

func (c *FeaturesController) convertFeatureData(ctx context.Context, name string, feature featuretoggles.UserFeature) *app.UserFeature {
	var enablementLevel *string
	if feature.EnablementLevel != featuretoggles.UnknownLevel {
		enablementLevel = &feature.EnablementLevel
	}
	return &app.UserFeature{
		ID:   feature.Name,
		Type: "features",
		Attributes: &app.UserFeatureAttributes{
			Description:     feature.Description,
			Enabled:         feature.Enabled,
			EnablementLevel: enablementLevel,
			UserEnabled:     feature.UserEnabled,
		},
	}
}
