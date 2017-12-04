package design

import (
	"github.com/fabric8-services/fabric8-toggles-service/jsonapi"
	d "github.com/goadesign/goa/design"
	a "github.com/goadesign/goa/design/apidsl"
)

var feature = a.Type("Feature", func() {
	a.Description(`JSONAPI for the feature object. See also http://jsonapi.org/format/#document-resource-object`)
	a.Attribute("id", d.String, "Id of feature", func() {
		a.Example("Feature name")
	})
	a.Attribute("attributes", featureAttributes)
	a.Required("id", "attributes")
})

var featureAttributes = a.Type("FeatureAttributes", func() {
	a.Description(`JSONAPI store for all the "attributes" of a Feature. See also see http://jsonapi.org/format/#document-resource-object-attributes`)
	a.Attribute("name", d.String, "The feature name", func() {
		a.Example("Name of the feature")
	})
	a.Attribute("description", d.String, "The feature name", func() {
		a.Example("Description of the feature")
	})
	a.Attribute("enabled", d.Boolean, "User profile type", func() {
		a.Example(true)
	})
	a.Attribute("groupId", d.String, "The feature name", func() {
		a.Example("Id/name of the group, loosely coupled to user claim in auth token")
	})
})

var _ = a.Resource("feature", func() {
	a.BasePath("/api/features")

	a.Action("show", func() {
		a.Security("jwt")
		a.Routing(
			a.GET("/:id"),
		)
		a.Params(func() {
			a.Param("id", d.String, "id")
		})
		a.Description("Show feature details.")
		a.Response(d.OK, feature)
		a.Response(d.BadRequest, jsonapi.JSONAPIErrors)
		a.Response(d.NotFound, jsonapi.JSONAPIErrors)
		a.Response(d.InternalServerError, jsonapi.JSONAPIErrors)
		a.Response(d.Unauthorized, jsonapi.JSONAPIErrors)
	})
})

var featureList = jsonapi.JSONList(
	"Feature", "Holds the list of Features",
	feature,
	nil,
	nil)

var _ = a.Resource("features", func() {
	a.BasePath("/api/features")
	a.Action("list", func() {
		a.Security("jwt")
		a.Routing(
			a.GET(""),
		)
		a.Description("Show a list of features enabled.")
		a.Response(d.OK, featureList)
		a.Response(d.BadRequest, jsonapi.JSONAPIErrors)
		a.Response(d.NotFound, jsonapi.JSONAPIErrors)
		a.Response(d.InternalServerError, jsonapi.JSONAPIErrors)
		a.Response(d.Unauthorized, jsonapi.JSONAPIErrors)
	})
})
