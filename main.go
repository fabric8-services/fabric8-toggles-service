package main

import (
	"fmt"
	"github.com/fabric8-services/fabric8-toggles-service/app"
	"github.com/fabric8-services/fabric8-toggles-service/configuration"
	"github.com/fabric8-services/fabric8-toggles-service/controller"
	"github.com/fabric8-services/fabric8-wit/jsonapi"
	"github.com/fabric8-services/fabric8-wit/log"
	"github.com/goadesign/goa"
	"github.com/goadesign/goa/logging/logrus"
	"github.com/goadesign/goa/middleware"
	"github.com/goadesign/goa/middleware/gzip"
	"net/http"
)

func main() {

	// Initialized configuration
	config, err := configuration.GetData()
	if err != nil {
		log.Panic(nil, map[string]interface{}{
			"err": err,
		}, "failed to setup the configuration")
	}

	// Initialized developer mode flag for the logger
	log.InitializeLogger(config.IsLogJSON(), config.GetLogLevel())
	fmt.Printf("%s", config)

	// Create service
	service := goa.New("feature")

	// Mount middleware
	service.WithLogger(goalogrus.New(log.Logger()))
	service.Use(middleware.RequestID())
	service.Use(gzip.Middleware(9))
	service.Use(jsonapi.ErrorHandler(service, true))
	service.Use(middleware.Recover())

	//service.Use(witmiddleware.TokenContext(publicKeys, nil, app.NewJWTSecurity()))
	service.Use(log.LogRequest(config.IsDeveloperModeEnabled()))
	//app.UseJWTMiddleware(service, goajwt.New(publicKeys, nil, app.NewJWTSecurity()))

	// Mount "features" controller
	featuresCtrl := controller.NewFeaturesController(service)
	app.MountFeaturesController(service, featuresCtrl)

	// Mount "feature" controller
	featureCtrl := controller.NewFeatureController(service)
	app.MountFeatureController(service, featureCtrl)

	http.Handle("/favicon.ico", http.NotFoundHandler())
	http.Handle("/", service.Mux)

	// Start http
	if err := http.ListenAndServe(config.GetHTTPAddress(), nil); err != nil {
		log.Error(nil, map[string]interface{}{
			"addr": config.GetHTTPAddress(),
			"err":  err,
		}, "unable to connect to server")
		service.LogError("startup", "err", err)
	}

}
