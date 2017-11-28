package controller

import (
	"context"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/fabric8-services/fabric8-toggles-service/app/test"
	"github.com/goadesign/goa"
	goajwt "github.com/goadesign/goa/middleware/security/jwt"
	"github.com/magiconair/properties/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestListFeatures(t *testing.T) {
	svc := goa.New("feature")
	ctrl := NewFeaturesController(svc, nil)
	expectedFeaturesList := buildFeaturesList(5)

	t.Run("OK with jwt token without goup id claim", func(t *testing.T) {
		_, featuresList := test.ListFeaturesOK(t, context.Background(), svc, ctrl)
		assert.Equal(t, 0, len(featuresList.Data))
	})
	t.Run("OK with jwt token containing goup id", func(t *testing.T) {
		_, featuresList := test.ListFeaturesOK(t, createValidContext(), svc, ctrl)
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
