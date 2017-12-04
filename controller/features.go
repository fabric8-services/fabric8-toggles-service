package controller

import (
	"context"

	jwtgo "github.com/dgrijalva/jwt-go"
	"github.com/fabric8-services/fabric8-toggles-service/app"
	"github.com/fabric8-services/fabric8-toggles-service/errorhandler"
	"github.com/fabric8-services/fabric8-toggles-service/errors"
	"github.com/fabric8-services/fabric8-toggles-service/featuretoggles"
	"github.com/fabric8-services/fabric8-wit/log"
	"github.com/goadesign/goa"
	goajwt "github.com/goadesign/goa/middleware/security/jwt"
)

// FeaturesController implements the features resource.
type FeaturesController struct {
	*goa.Controller
	client *featuretoggles.Client
}

// NewFeaturesController creates a features controller.
func NewFeaturesController(service *goa.Service, client *featuretoggles.Client) *FeaturesController {
	return &FeaturesController{
		Controller: service.NewController("FeaturesController"),
		client:     client,
	}
}

// List runs the list action.
func (c *FeaturesController) List(ctx *app.ListFeaturesContext) error {
	jwtToken := goajwt.ContextJWT(ctx)
	if jwtToken == nil {
		log.Error(ctx.Context, map[string]interface{}{}, "Unable to retrieve token")
		return errorhandler.JSONErrorResponse(ctx, errors.NewUnauthorizedError("Missing JWT token"))
	}
	if groupID, ok := jwtToken.Claims.(jwtgo.MapClaims)["company"].(string); ok {
		enableFeatures := c.getEnabledFeatures(ctx, groupID)
		log.Debug(ctx, nil, "FEATURES: %s", enableFeatures)
		return ctx.OK(enableFeatures)
	}
	return errorhandler.JSONErrorResponse(ctx, errors.NewUnauthorizedError("Incomplete JWT token"))
}

func (c *FeaturesController) getEnabledFeatures(ctx *app.ListFeaturesContext, groupID string) *app.FeatureList {
	listOfFeatures := c.client.GetEnabledFeatures(groupID)
	return convert(ctx, listOfFeatures, groupID)
}

func convert(ctx context.Context, list []string, groupID string) *app.FeatureList {
	res := app.FeatureList{}
	for i := 0; i < len(list); i++ {
		// TODO remove ID, make unleash client return description
		ID := list[i]
		descriptionFeature := "Description of the feature"
		enabledFeature := true
		nameFeature := list[i]
		feature := app.Feature{
			ID: ID,
			Attributes: &app.FeatureAttributes{
				Description: &descriptionFeature,
				Enabled:     &enabledFeature,
				Name:        &nameFeature,
				GroupID:     &groupID,
			},
		}
		log.Info(ctx, map[string]interface{}{"feature_name": nameFeature}, "found enabled feature for user")
		res.Data = append(res.Data, &feature)
	}
	return &res
}
