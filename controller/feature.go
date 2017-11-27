package controller

import (
	"github.com/fabric8-services/fabric8-toggles-service/app"
	"github.com/fabric8-services/fabric8-toggles-service/configuration"
	"github.com/fabric8-services/fabric8-toggles-service/errorhandler"
	"github.com/goadesign/goa"
	uuid "github.com/satori/go.uuid"
)

// FeatureController implements the feature resource.
type FeatureController struct {
	*goa.Controller
	config *configuration.Data
}

// NewFeatureController creates a feature controller.
func NewFeatureController(service *goa.Service, config *configuration.Data) *FeatureController {
	return &FeatureController{
		Controller: service.NewController("FeatureController"),
		config:     config,
	}
}

// Show runs the show action.
func (c *FeatureController) Show(ctx *app.ShowFeatureContext) error {
	// FeatureController_Show: start_implement
	featureID, err := uuid.FromString(ctx.ID)
	if err != nil {
		return errorhandler.JSONErrorResponse(ctx, goa.ErrNotFound(err.Error()))
	}
	// TODO call unleash SDK to retrieve features/strategy
	descriptionFeature := "Description of the feature"
	enabledFeature := true
	nameFeature := "Feature A"
	groupId := "BETA"

	feature := app.Feature{
		ID: featureID,
		Attributes: &app.FeatureAttributes{
			Description: &descriptionFeature,
			Enabled:     &enabledFeature,
			Name:        &nameFeature,
			GroupID:     &groupId,
		},
	}
	// FeatureController_Show: end_implement
	return ctx.OK(&feature)
}
