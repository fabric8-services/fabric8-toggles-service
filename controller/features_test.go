package controller

import (
	"context"
	"github.com/fabric8-services/fabric8-toggles-service/app/test"
	"github.com/goadesign/goa"
	"github.com/magiconair/properties/assert"
	"testing"
)

func TestFeatures(t *testing.T) {
	svc := goa.New("feature")
	ctrl := NewFeaturesController(svc)
	expectedFeaturesList := buildFeaturesList(5)

	t.Run("OK", func(t *testing.T) {
		_, featuresList := test.ListFeaturesOK(t, context.Background(), svc, ctrl)
		assert.Equal(t, 5, len(featuresList.Data))
		assert.Equal(t, featuresList.Data[0].Attributes.Name, expectedFeaturesList.Data[0].Attributes.Name)
	})
	//t.Run("Unauhorized - no token", func(t *testing.T) {
	//	test.ListFeaturesUnauthorized(t, context.Background(), svc, ctrl)
	//})
	//t.Run("Not found", func(t *testing.T) {
	//	test.ListFeaturesNotFound(t, context.Background(), svc, ctrl)
	//})
}
