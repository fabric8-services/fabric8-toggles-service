package design

import (
	d "github.com/goadesign/goa/design"
	a "github.com/goadesign/goa/design/apidsl"
)

var userFeatureSingle = JSONSingle(
	"UserFeature", "Holds a single user feature",
	userFeature,
	nil)

var userFeatureList = JSONList(
	"UserFeature", "Holds the list of user features",
	userFeature,
	nil,
	nil)

var userFeature = a.Type("UserFeature", func() {
	a.Description(`JSONAPI for the user feature object. See also http://jsonapi.org/format/#document-resource-object`)
	a.Attribute("id", d.String, "Id of feature", func() {
		a.Example("Feature name")
	})
	a.Attribute("type", d.String, "the 'features' type", func() {
		a.Example("features")
	})

	a.Attribute("attributes", userFeatureAttributes)
	a.Required("id", "type", "attributes")
})

var userFeatureAttributes = a.Type("UserFeatureAttributes", func() {
	a.Description(`JSONAPI store for all the "attributes" of a Feature. See also see http://jsonapi.org/format/#document-resource-object-attributes`)
	a.Attribute("description", d.String, "The description of the feature", func() {
		a.Example("Description of the feature")
	})
	a.Attribute("enabled", d.Boolean, "marks if the feature is globally enabled (prior to applying strategies)", func() {
		a.Example(true)
	})
	a.Attribute("user-enabled", d.Boolean, "marks if the feature is enabled for the current user", func() {
		a.Example(true)
	})
	a.Attribute("enablement-level", d.String, "The mimimum level of enablement for this feature. Empty/missing means that the feature is not accessible to the user", func() {
		a.Example("beta")
	})
	a.Required("description", "enabled", "user-enabled")
})

var _ = a.Resource("features", func() {
	a.BasePath("/features")

	a.Action("show", func() {
		a.Routing(
			a.GET("/:featureName"),
		)
		a.Params(func() {
			a.Param("featureName", d.String, "featureName")
		})
		a.Description("Show feature details.")
		a.UseTrait("conditional")
		a.Response(d.OK, userFeatureSingle)
		a.Response(d.NotModified)
		a.Response(d.BadRequest, JSONAPIErrors)
		a.Response(d.Unauthorized, JSONAPIErrors)
		a.Response(d.InternalServerError, JSONAPIErrors)
	})

	a.Action("list", func() {
		a.Routing(
			a.GET(""),
		)
		a.Params(func() {
			a.Param("names", a.ArrayOf(d.String), "names")
			a.Param("group", d.String, "group")
		})
		a.Description("Show a list of features by their names.")
		a.UseTrait("conditional")
		a.Response(d.OK, userFeatureList)
		a.Response(d.NotModified)
		a.Response(d.BadRequest, JSONAPIErrors)
		a.Response(d.Unauthorized, JSONAPIErrors)
		a.Response(d.InternalServerError, JSONAPIErrors)
	})
})
