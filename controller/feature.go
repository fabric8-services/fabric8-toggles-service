package controller

import (
	jwtgo "github.com/dgrijalva/jwt-go"
	"github.com/fabric8-services/fabric8-toggles-service/app"
	"github.com/fabric8-services/fabric8-toggles-service/errorhandler"
	"github.com/fabric8-services/fabric8-toggles-service/errors"
	"github.com/fabric8-services/fabric8-toggles-service/featuretoggles"
	"github.com/fabric8-services/fabric8-wit/log"
	"github.com/goadesign/goa"
	goajwt "github.com/goadesign/goa/middleware/security/jwt"
)

// FeatureController implements the feature resource.
type FeatureController struct {
	*goa.Controller
	client *featuretoggles.Client
}

// NewFeatureController creates a feature controller.
func NewFeatureController(service *goa.Service, client *featuretoggles.Client) *FeatureController {
	return &FeatureController{
		Controller: service.NewController("FeatureController"),
		client:     client,
	}
}

// Show runs the show action.
func (c *FeatureController) Show(ctx *app.ShowFeatureContext) error {
	// FeatureController_Show: start_implement
	jwtToken := goajwt.ContextJWT(ctx)
	if jwtToken == nil {
		log.Error(ctx.Context, map[string]interface{}{}, "Unable to retrieve token")
		return errorhandler.JSONErrorResponse(ctx, errors.NewUnauthorizedError("Missing JWT token"))
	}
	featureID := ctx.ID
	var enabledFeature *app.Feature
	if groupID, ok := jwtToken.Claims.(jwtgo.MapClaims)["company"].(string); ok {
		log.Info(ctx, nil, "Is feature id: %s enabled? ", featureID)
		enabled := c.client.IsFeatureEnabled(featureID, groupID)
		enabledFeature = c.convert(featureID, groupID, enabled)
		return ctx.OK(enabledFeature)
	}
	return errorhandler.JSONErrorResponse(ctx, errors.NewUnauthorizedError("Incomplete JWT token"))
}

func (c *FeatureController) convert(featureID string, groupID string, enabledFeature bool) *app.Feature {
	descriptionFeature := "Description of the feature"
	nameFeature := featureID
	feature := app.Feature{
		ID: nameFeature,
		Attributes: &app.FeatureAttributes{
			Name:        &nameFeature,
			Description: &descriptionFeature,
			Enabled:     &enabledFeature,
			GroupID:     &groupID,
		},
	}
	return &feature
}
