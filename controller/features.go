package controller

import (
	"github.com/fabric8-services/fabric8-toggles-service/app"
	"github.com/goadesign/goa"
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
	res := &app.FeatureList{}
	return ctx.OK(res)
}
