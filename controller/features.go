package controller

import (
	"fmt"
	jwtgo "github.com/dgrijalva/jwt-go"
	"github.com/fabric8-services/fabric8-toggles-service/app"
	"github.com/fabric8-services/fabric8-toggles-service/configuration"
	"github.com/fabric8-services/fabric8-toggles-service/featuretoggles"
	"github.com/fabric8-services/fabric8-wit/log"
	"github.com/goadesign/goa"
	goajwt "github.com/goadesign/goa/middleware/security/jwt"
	uuid "github.com/satori/go.uuid"
	"strings"
)

// FeaturesController implements the features resource.
type FeaturesController struct {
	*goa.Controller
	config *configuration.Data
}

// NewFeaturesController creates a features controller.
func NewFeaturesController(service *goa.Service, config *configuration.Data) *FeaturesController {
	return &FeaturesController{
		Controller: service.NewController("FeaturesController"),
		config:     config,
	}
}

// List runs the list action.
func (c *FeaturesController) List(ctx *app.ListFeaturesContext) error {
	// FeaturesController_List: start_implement

	jwtToken := goajwt.ContextJWT(ctx)
	groupId := ""
	if jwtToken != nil {
		groupId = jwtToken.Claims.(jwtgo.MapClaims)["company"].(string) // TODO replace with a custom claim group_id
	}
	var enableFeatures *app.FeatureList
	if c.config != nil {
		featuresListFromUnleash, err := getFeatureListFromUnleashServer(ctx, c.config, groupId)
		if err != nil {
			log.Error(ctx.Context, map[string]interface{}{
				"err": err,
			}, "Unable to connect to Unleash server")
			return ctx.Err()
		}
		log.Debug(ctx, nil, "FEATURES: %s", featuresListFromUnleash)
		enableFeatures = featuresListFromUnleash
	} else { // TEST Data
		enabledFeaturesList := app.FeatureList{Data: []*app.Feature{}}
		featuresList := buildFeaturesList(5)
		for _, elt := range featuresList.Data {
			if elt.Attributes.GroupID != nil && strings.ToLower(*elt.Attributes.GroupID) == strings.ToLower(groupId) {
				enabledFeaturesList.Data = append(enabledFeaturesList.Data, elt)
			}
		}
		enableFeatures = &enabledFeaturesList
	}

	// FeaturesController_List: end_implement
	return ctx.OK(enableFeatures)
}

func getFeatureListFromUnleashServer(ctx *app.ListFeaturesContext, config *configuration.Data, groupId string) (*app.FeatureList, error) {
	res := app.FeatureList{}
	toggleClient, err := featuretoggles.NewFeatureToggleClient(ctx.Context, config)
	if err != nil {
		log.Error(ctx.Context, map[string]interface{}{
			"addr": config.GetHTTPAddress(),
			"err":  err,
		}, "Unable to connect to Unleash server")
		return nil, err
	}

	listOfFeatures := toggleClient.GetEnabledFeatures(groupId)
	res = convert(listOfFeatures, groupId)
	return &res, nil
}

func convert(list []string, groupId string) app.FeatureList {
	res := app.FeatureList{}
	for i := 0; i < len(list); i++ {
		// TODO remove ID, make unleash client return description
		ID := uuid.NewV4()
		descriptionFeature := "Description of the feature"
		enabledFeature := true
		nameFeature := list[i]

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
	return res
}

func buildFeaturesList(length int) *app.FeatureList {
	res := app.FeatureList{}
	for i := 0; i < length; i++ {
		ID := uuid.NewV4()
		// TODO call unleash SDK to retrieve features/strategy
		descriptionFeature := "Description of the feature"
		enabledFeature := true
		nameFeature := fmt.Sprintf("Feature %d", i)
		var groupId string
		if i%2 == 0 {
			groupId = "BETA"
		} else {
			groupId = "RED HAT"
		}

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
