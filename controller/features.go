package controller

import (
	"fmt"
	"github.com/fabric8-services/fabric8-toggles-service/app"
	"github.com/goadesign/goa"
	uuid "github.com/satori/go.uuid"
)

// FeaturesController implements the features resource.
type FeaturesController struct {
	*goa.Controller
}

// NewFeaturesController creates a features controller.
func NewFeaturesController(service *goa.Service) *FeaturesController {
	return &FeaturesController{Controller: service.NewController("FeaturesController")}
}

// List runs the list action.
func (c *FeaturesController) List(ctx *app.ListFeaturesContext) error {
	// FeaturesController_List: start_implement

	// Put your logic here

	// FeaturesController_List: end_implement
	return ctx.OK(buildFeaturesList(5))
}

func buildFeaturesList(length int) *app.FeatureList {
	res := app.FeatureList{}
	for i := 0; i < length; i++ {
		ID := uuid.NewV4()
		// TODO call unleash SDK to retrieve features/strategy
		descriptionFeature := "Description of the feature"
		enabledFeature := true
		nameFeature := fmt.Sprintf("Feature %d", i)
		groupId := "BETA"

		feature := app.Feature{
			ID: ID,
			Attributes: &app.FeatureAttributes{
				Description: &descriptionFeature,
				Enabled:     &enabledFeature,
				Name:        &nameFeature,
				GroupID:     &groupId,
			},
		}
		res.Data = append(res.Data, &feature)
	}
	return &res
}
