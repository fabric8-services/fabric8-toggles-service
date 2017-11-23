package design

import (
	d "github.com/goadesign/goa/design"
	a "github.com/goadesign/goa/design/apidsl"
	"strings"
)

/* TODO reuse it from fabric8-services/fabric8-tenant */
//#############################################################################
//
// 			JSONAPI common
//
//#############################################################################

// JSONAPILink represents a JSONAPI link object (see http://jsonapi.org/format/#document-links)
var JSONAPILink = a.Type("JSONAPILink", func() {
	a.Description(`See also http://jsonapi.org/format/#document-links.`)
	a.Attribute("href", d.String, "a string containing the link's URL.", func() {
		a.Example("http://example.com/articles/1/comments")
	})
	a.Attribute("meta", a.HashOf(d.String, d.Any), "a meta object containing non-standard meta-information about the link.")
})

// JSONAPIError represents a JSONAPI error object (see http://jsonapi.org/format/#error-objects)
var JSONAPIError = a.Type("JSONAPIError", func() {
	a.Description(`Error objects provide additional information about problems encountered while
performing an operation. Error objects MUST be returned as an array keyed by errors in the
top level of a JSON API document.
See. also http://jsonapi.org/format/#error-objects.`)
	a.Attribute("id", d.String, "a unique identifier for this particular occurrence of the problem.")
	a.Attribute("links", a.HashOf(d.String, JSONAPILink), `a links object containing the following members:
* about: a link that leads to further details about this particular occurrence of the problem.`)
	a.Attribute("status", d.String, "the HTTP status code applicable to this problem, expressed as a string value.")
	a.Attribute("code", d.String, "an application-specific error code, expressed as a string value.")
	a.Attribute("title", d.String, `a short, human-readable summary of the problem that SHOULD NOT
change from occurrence to occurrence of the problem, except for purposes of localization.`)
	a.Attribute("detail", d.String, `a human-readable explanation specific to this occurrence of the problem.
Like title, this field’s value can be localized.`)
	a.Attribute("source", a.HashOf(d.String, d.Any), `an object containing references to the source of the error,
optionally including any of the following members
* pointer: a JSON Pointer [RFC6901] to the associated entity in the request document [e.g. "/data" for a primary data object,
           or "/data/attributes/title" for a specific attribute].
* parameter: a string indicating which URI query parameter caused the error.`)
	a.Attribute("meta", a.HashOf(d.String, d.Any), "a meta object containing non-standard meta-information about the error")
	a.Required("detail")
})

// JSONAPIErrors is an array of JSONAPI error objects
var JSONAPIErrors = a.MediaType("application/vnd.jsonapierrors+json", func() {
	a.UseTrait("jsonapi-media-type")
	a.TypeName("JSONAPIErrors")
	a.Description(``)
	a.Attributes(func() {
		a.Attribute("errors", a.ArrayOf(JSONAPIError))
		a.Required("errors")
	})
	a.View("default", func() {
		a.Attribute("errors")
		a.Required("errors")
	})
})

// JSONList creates a UserTypeDefinition
func JSONList(name, description string, data *d.UserTypeDefinition, links *d.UserTypeDefinition, meta *d.UserTypeDefinition) *d.MediaTypeDefinition {
	return a.MediaType("application/vnd."+strings.ToLower(name)+"list+json", func() {
		a.UseTrait("jsonapi-media-type")
		a.TypeName(name + "List")
		a.Description(description)
		if links != nil {
			a.Attribute("links", links)
		}
		if meta != nil {
			a.Attribute("meta", meta)
		}
		a.Attribute("data", a.ArrayOf(data))
		a.Attribute("included", a.ArrayOf(d.Any), "An array of mixed types")
		a.Required("data")

		a.View("default", func() {
			if links != nil {
				a.Attribute("links")
			}
			if meta != nil {
				a.Attribute("meta")
			}
			a.Attribute("data")
			a.Attribute("included")
			a.Required("data")
		})
	})
}

/* End of TODO */

var feature = a.Type("Feature", func() {
	a.Description(`JSONAPI for the feature object. See also http://jsonapi.org/format/#document-resource-object`)
	a.Attribute("id", d.UUID, "Id of feature", func() {
		a.Example("40bbdd3d-8b5d-4fd6-ac90-7236b669af04")
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
		a.Description("Show feature details.")
		a.Response(d.OK, feature)
		a.Response(d.BadRequest, JSONAPIErrors)
		a.Response(d.NotFound, JSONAPIErrors)
		a.Response(d.InternalServerError, JSONAPIErrors)
		a.Response(d.Unauthorized, JSONAPIErrors)
	})
})

var featureList = JSONList(
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
		a.Description("Show a single tenant environment.")
		a.Response(d.OK, featureList)
		a.Response(d.BadRequest, JSONAPIErrors)
		a.Response(d.NotFound, JSONAPIErrors)
		a.Response(d.InternalServerError, JSONAPIErrors)
		a.Response(d.Unauthorized, JSONAPIErrors)
	})
})
