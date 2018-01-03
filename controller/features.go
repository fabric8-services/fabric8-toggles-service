package controller

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	unleashapi "github.com/Unleash/unleash-client-go/api"
	jwtgo "github.com/dgrijalva/jwt-go"
	"github.com/fabric8-services/fabric8-toggles-service/app"
	"github.com/fabric8-services/fabric8-toggles-service/auth/authservice"
	"github.com/fabric8-services/fabric8-toggles-service/errorhandler"
	"github.com/fabric8-services/fabric8-toggles-service/errors"
	"github.com/fabric8-services/fabric8-toggles-service/featuretoggles"
	"github.com/fabric8-services/fabric8-wit/goasupport"
	"github.com/fabric8-services/fabric8-wit/log"
	"github.com/fabric8-services/fabric8-wit/rest"
	"github.com/goadesign/goa"
	goaclient "github.com/goadesign/goa/client"
	goajwt "github.com/goadesign/goa/middleware/security/jwt"
	errs "github.com/pkg/errors"
)

// FeaturesControllerConfig the configuration required for the FeaturesController
type FeaturesControllerConfig interface {
	GetAuthServiceURL() string
}

// NewFeaturesController creates a FeaturesController.
func NewFeaturesController(service *goa.Service, togglesClient *featuretoggles.Client, httpClient *http.Client, config FeaturesControllerConfig) *FeaturesController {
	return &FeaturesController{
		Controller:    service.NewController("FeaturesController"),
		togglesClient: togglesClient,
		httpClient:    httpClient,
		authURL:       config.GetAuthServiceURL(),
	}
}

// FeaturesController implements the features resource.
type FeaturesController struct {
	*goa.Controller
	togglesClient *featuretoggles.Client
	httpClient    *http.Client
	authURL       string
}

// List runs the list action.
func (c *FeaturesController) List(ctx *app.ListFeaturesContext) error {
	jwtToken := goajwt.ContextJWT(ctx)
	if jwtToken == nil {
		log.Error(ctx.Context, map[string]interface{}{}, "Unable to retrieve token")
		return errorhandler.JSONErrorResponse(ctx, errors.NewUnauthorizedError("Missing JSON Web Token in request header"))
	}
	if groupID, ok := jwtToken.Claims.(jwtgo.MapClaims)["company"].(string); ok {
		enableFeatures := c.getEnabledFeatures(ctx, groupID)
		log.Debug(ctx, nil, "FEATURES: %s", enableFeatures)
		return ctx.OK(enableFeatures)
	}
	return errorhandler.JSONErrorResponse(ctx, errors.NewUnauthorizedError("Incomplete JWT token"))
}

func (c *FeaturesController) getEnabledFeatures(ctx *app.ListFeaturesContext, groupID string) *app.FeatureList {
	listOfFeatures := c.togglesClient.GetEnabledFeatures(groupID)
	return convertFeatures(ctx, listOfFeatures, groupID)
}

// Show runs the show action.
func (c *FeaturesController) Show(ctx *app.ShowFeaturesContext) error {
	jwtToken := goajwt.ContextJWT(ctx)
	if jwtToken == nil {
		log.Error(ctx.Context, map[string]interface{}{}, "Unable to retrieve token")
		return errorhandler.JSONErrorResponse(ctx, errors.NewUnauthorizedError("Missing JSON Web Token in request header"))
	}
	user, err := c.getUserProfile(ctx)
	if err != nil {
		return errorhandler.JSONErrorResponse(ctx, err)
	}
	featureName := ctx.FeatureName
	feature := c.togglesClient.GetFeature(featureName)
	if feature == nil {
		log.Warn(ctx, map[string]interface{}{"feature_name": featureName}, "feature not found")
		return errorhandler.JSONErrorResponse(ctx, errors.NewNotFoundError("feature", featureName))
	}
	log.Debug(ctx, map[string]interface{}{"user_name": *user.Data.Attributes.Username, "feature_level": *user.Data.Attributes.FeatureLevel, "feature_name": featureName}, "checking if feature is enabled for user...")
	enabledForUser := c.togglesClient.IsFeatureEnabled(*feature, user.Data.Attributes.FeatureLevel)
	appFeature := convertFeature(ctx, feature, user, enabledForUser)
	return ctx.OK(appFeature)
}

// getUserProfile retrieves the user's profile from the auth service, by forwarding the current JWT token
func (c *FeaturesController) getUserProfile(ctx context.Context) (*authservice.User, error) {
	authClient, err := c.newAuthClient(ctx)
	if err != nil {
		log.Error(ctx, map[string]interface{}{
			"err": err.Error(),
		}, "unable to initialize auth service client")
		return nil, errs.Wrap(err, "unable to initialize auth service client")
	}
	res, err := authClient.ShowUser(goasupport.ForwardContextRequestID(ctx), authservice.ShowUserPath(), nil, nil)
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

// NewAuthClient initializes a new client to the `auth` service
func (c *FeaturesController) newAuthClient(ctx context.Context) (*authservice.Client, error) {
	u, err := url.Parse(c.authURL)
	if err != nil {
		return nil, err
	}
	authClient := authservice.New(goaclient.HTTPClientDoer(c.httpClient))
	authClient.Host = u.Host
	authClient.Scheme = u.Scheme
	authClient.SetJWTSigner(goasupport.NewForwardSigner(ctx))
	return authClient, nil
}

func convertFeature(ctx context.Context, feature *unleashapi.Feature, user *authservice.User, enabledForUser bool) *app.FeatureSingle {
	userEmail := user.Data.Attributes.Email
	internalUser := false
	// internal users have may be able to access the feature by opting-in to the `internal` level of features.
	if strings.HasSuffix(*userEmail, "@redhat.com") {
		internalUser = true
	}
	log.Debug(ctx, map[string]interface{}{"internal_user": internalUser}, "converting feature")
	// TODO include the `email verified` field
	return &app.FeatureSingle{
		Data: &app.Feature{
			ID:   feature.Name,
			Type: "features",
			Attributes: &app.FeatureAttributes{
				Description:     feature.Description,
				Enabled:         feature.Enabled,
				EnablementLevel: featuretoggles.ComputeEnablementLevel(ctx, feature, internalUser),
				UserEnabled:     enabledForUser,
			},
		},
	}
}

func convertFeatures(ctx context.Context, list []string, groupID string) *app.FeatureList {
	res := app.FeatureList{}
	for _, name := range list {
		// TODO remove ID, make unleash client return description
		descriptionFeature := "Description of the feature"
		enabledFeature := true
		feature := app.Feature{
			ID: name,
			Attributes: &app.FeatureAttributes{
				Description: descriptionFeature,
				Enabled:     enabledFeature,
			},
		}
		log.Info(ctx, map[string]interface{}{"feature_name": name}, "found enabled feature for user")
		res.Data = append(res.Data, &feature)
	}
	return &res
}
