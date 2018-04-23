package featuretoggles_test

import (
	"sort"
	"testing"

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
