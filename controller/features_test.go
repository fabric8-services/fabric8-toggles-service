package controller

import (
	"context"
	"fmt"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/fabric8-services/fabric8-toggles-service/app"
	"github.com/fabric8-services/fabric8-toggles-service/app/test"
	"github.com/fabric8-services/fabric8-toggles-service/configuration"
	"github.com/fabric8-services/fabric8-toggles-service/featuretoggles"
	"github.com/goadesign/goa"
	goajwt "github.com/goadesign/goa/middleware/security/jwt"
	"github.com/magiconair/properties/assert"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"
	"testing"
)

type MockFeaturesController struct {
	FeaturesController
}
type MockClient struct {
}

func (c *MockClient) GetEnabledFeatures(groupId string) []string {
	return buildFeaturesList(5)
}

func (c *MockFeaturesController) GetClient(ctx *app.ListFeaturesContext, config *configuration.Data) (featuretoggles.Client, error) {
	return &MockClient{}, nil
}

func buildFeaturesList(length int) []string {
	res := make([]string, length)
	for i := 0; i < length; i++ {
		nameFeature := fmt.Sprintf("Feature %d", i)
		res = append(res, nameFeature)
	}
	return res
}

func TestListFeatures(t *testing.T) {
	svc := goa.New("feature")
	ctl := FeaturesController{
		Controller: svc.NewController("FeaturesController"),
		config:     nil,
	}
	mockCtrl := MockFeaturesController{ctl}

	expectedFeaturesList := buildExpectedFeaturesList(5)

	t.Run("OK with jwt token without goup id claim", func(t *testing.T) {
		_, featuresList := test.ListFeaturesOK(t, context.Background(), svc, &mockCtrl)
		assert.Equal(t, 0, len(featuresList.Data))
	})
	t.Run("OK with jwt token containing goup id", func(t *testing.T) {
		_, featuresList := test.ListFeaturesOK(t, createValidContext(), svc, &mockCtrl)
		require.Equal(t, 2, len(featuresList.Data))
		assert.Equal(t, featuresList.Data[0].Attributes.Name, expectedFeaturesList.Data[1].Attributes.Name)
		assert.Equal(t, featuresList.Data[1].Attributes.Name, expectedFeaturesList.Data[3].Attributes.Name)
	})
	//t.Run("Unauhorized - no token", func(t *testing.T) {
	//	test.ListFeaturesUnauthorized(t, context.Background(), svc, ctrl)
	//})
	//t.Run("Not found", func(t *testing.T) {
	//	test.ListFeaturesNotFound(t, context.Background(), svc, ctrl)
	//})
}
func createValidContext() context.Context {
	claims := jwt.MapClaims{}
	claims["company"] = "Red Hat" // TODO replace by BETA
	token := jwt.NewWithClaims(jwt.SigningMethodRS512, claims)
	return goajwt.WithJWT(context.Background(), token)
}

func buildExpectedFeaturesList(length int) *app.FeatureList {
	res := app.FeatureList{}
	for i := 0; i < length; i++ {
		ID := uuid.NewV4()
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
