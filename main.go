package main

import (
	"fmt"
	"net/http"

	"github.com/fabric8-services/fabric8-toggles-service/featuretoggles"

	"github.com/fabric8-services/fabric8-toggles-service/app"
	"github.com/fabric8-services/fabric8-toggles-service/configuration"
	"github.com/fabric8-services/fabric8-toggles-service/controller"
	"github.com/fabric8-services/fabric8-toggles-service/errorhandler"
	"github.com/fabric8-services/fabric8-toggles-service/token"
	witmiddleware "github.com/fabric8-services/fabric8-wit/goamiddleware"
	"github.com/fabric8-services/fabric8-wit/log"
	"github.com/goadesign/goa"
	"github.com/goadesign/goa/logging/logrus"
	"github.com/goadesign/goa/middleware"
	"github.com/goadesign/goa/middleware/gzip"
	goajwt "github.com/goadesign/goa/middleware/security/jwt"
)

func main() {

	// Initialized configuration
	config, err := configuration.GetData()
	if err != nil || config == nil {
		log.Panic(nil, map[string]interface{}{
			"err": err,
		}, "failed to setup the configuration")
	}

	fmt.Printf("%s", config)

	// Create service
	service := goa.New("feature")

	// Initialize log
	log.InitializeLogger(config.IsLogJSON(), config.GetLogLevel())

	service.WithLogger(goalogrus.New(log.Logger()))

	// Mount middleware
	service.Use(middleware.RequestID())
	service.Use(gzip.Middleware(9))
	service.Use(errorhandler.ErrorHandler(service, true))
	service.Use(middleware.Recover())

	tokenManager, err := token.NewManager(config)
	if err != nil {
		log.Panic(nil, map[string]interface{}{
			"err": err,
		}, "failed to create token manager")
	}
	// Middleware that extracts and stores the token in the context
	jwtMiddlewareTokenContext := witmiddleware.TokenContext(tokenManager.PublicKeys(), nil, app.NewJWTSecurity())
	service.Use(jwtMiddlewareTokenContext)
	app.UseJWTMiddleware(service, goajwt.New(tokenManager.PublicKeys(), nil, app.NewJWTSecurity()))
	service.Use(log.LogRequest(config.IsDeveloperModeEnabled()))

	// init the toggle client
	toggleClient, err := featuretoggles.NewClient(config)
	if err != nil {
		log.Panic(nil, map[string]interface{}{
			"err": err,
		}, "failed to create toogle client")

	}
	// Mount "features" controller
	featuresCtrl := controller.NewFeaturesController(service, toggleClient)
	app.MountFeaturesController(service, featuresCtrl)

	// Mount "feature" controller
	featureCtrl := controller.NewFeatureController(service, toggleClient)
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
