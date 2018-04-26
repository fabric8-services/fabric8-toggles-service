package featuretoggles_test

import (
	"sort"
	"testing"

	"github.com/fabric8-services/fabric8-toggles-service/app"
	"github.com/fabric8-services/fabric8-toggles-service/featuretoggles"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSortUserFeartures(t *testing.T) {
	// given
	features := []featuretoggles.UserFeature{
		{Name: "foo", Description: "foo"},
		{Name: "bar", Description: "bar"},
	}
	// when
	sort.Sort(featuretoggles.ByName(features))
	// then
	require.Len(t, features, 2)
	assert.Equal(t, "bar", features[0].Name)
	assert.Equal(t, "foo", features[1].Name)
}

func TestComputeUserFeatureEtag(t *testing.T) {
	// given
	feature := featuretoggles.UserFeature{
		Name:            "foo",
		Description:     "foo",
		UserEnabled:     false,
		Enabled:         false,
		EnablementLevel: featuretoggles.BetaLevel,
	}

	t.Run("change name", func(t *testing.T) {
		// given
		feature2 := duplicate(feature)
		feature2.Name = "bar"
		// when
		etag := app.GenerateEntityTag(feature)
		etag2 := app.GenerateEntityTag(feature2)
		// then
		assert.NotEqual(t, etag2, etag)
	})

	t.Run("change description", func(t *testing.T) {
		// given
		feature2 := duplicate(feature)
		feature2.Description = "bar"
		// when
		etag := app.GenerateEntityTag(feature)
		etag2 := app.GenerateEntityTag(feature2)
		// then
		assert.NotEqual(t, etag2, etag)
	})

	t.Run("change enabled", func(t *testing.T) {
		// given
		feature2 := duplicate(feature)
		feature2.Enabled = true
		// when
		etag := app.GenerateEntityTag(feature)
		etag2 := app.GenerateEntityTag(feature2)
		// then
		assert.NotEqual(t, etag2, etag)
	})

	t.Run("change enablement level", func(t *testing.T) {
		// given
		feature2 := duplicate(feature)
		feature2.EnablementLevel = featuretoggles.InternalLevel
		// when
		etag := app.GenerateEntityTag(feature)
		etag2 := app.GenerateEntityTag(feature2)
		// then
		assert.NotEqual(t, etag2, etag)
	})

	t.Run("change user-enabled", func(t *testing.T) {
		// given
		feature2 := duplicate(feature)
		feature2.UserEnabled = true
		// when
		etag := app.GenerateEntityTag(feature)
		etag2 := app.GenerateEntityTag(feature2)
		// then
		assert.NotEqual(t, etag2, etag)
	})

}

func duplicate(f featuretoggles.UserFeature) featuretoggles.UserFeature {
	return featuretoggles.UserFeature{
		Name:            f.Name,
		Description:     f.Description,
		Enabled:         f.Enabled,
		EnablementLevel: f.EnablementLevel,
		UserEnabled:     f.UserEnabled,
	}
}
