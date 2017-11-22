package controller

import (
	"github.com/fabric8-services/fabric8-toggles-service/app"
	"github.com/goadesign/goa"
)

// FeatureController implements the feature resource.
type FeatureController struct {
	*goa.Controller
}

// NewFeatureController creates a feature controller.
func NewFeatureController(service *goa.Service) *FeatureController {
	return &FeatureController{Controller: service.NewController("FeatureController")}
}

// Show runs the show action.
func (c *FeatureController) Show(ctx *app.ShowFeatureContext) error {
	// FeatureController_Show: start_implement

	// Put your logic here

	// FeatureController_Show: end_implement
	return nil
}
