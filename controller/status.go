package controller

import (
	"time"

	"github.com/fabric8-services/fabric8-toggles-service/app"
	"github.com/goadesign/goa"
)

var (
	// Commit current build commit set by build script
	Commit = "0"
	// BuildTime set by build script in ISO 8601 (UTC) format: YYYY-MM-DDThh:mm:ssTZD (see https://www.w3.org/TR/NOTE-datetime for details)
	BuildTime = "0"
	// StartTime in ISO 8601 (UTC) format
	StartTime = time.Now().UTC().Format("2006-01-02T15:04:05Z")
)

// StatusConfiguration the status configuration
type StatusConfiguration interface {
	IsDeveloperModeEnabled() bool
}

// StatusController implements the status resource.
type StatusController struct {
	*goa.Controller
	config StatusConfiguration
}

// NewStatusController creates a status controller.
func NewStatusController(service *goa.Service, config StatusConfiguration) *StatusController {
	return &StatusController{
		Controller: service.NewController("StatusController"),
		config:     config,
	}
}

// Show runs the show action.
func (c *StatusController) Show(ctx *app.ShowStatusContext) error {
	res := &app.Status{
		Commit:    Commit,
		BuildTime: BuildTime,
		StartTime: StartTime,
	}
	devMode := c.config.IsDeveloperModeEnabled()
	if devMode {
		res.DevMode = &devMode
	}
	return ctx.OK(res)
}
