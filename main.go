package main

import (
	"fmt"
	"github.com/fabric8-services/fabric8-toggles-service/app"
	"github.com/fabric8-services/fabric8-toggles-service/configuration"
	"github.com/fabric8-services/fabric8-toggles-service/controller"
	witmiddleware "github.com/fabric8-services/fabric8-wit/goamiddleware"
	"github.com/fabric8-services/fabric8-wit/jsonapi"
	"github.com/fabric8-services/fabric8-wit/log"
	"github.com/fabric8-services/fabric8-wit/login"
	"github.com/fabric8-services/fabric8-wit/token"
	"github.com/goadesign/goa"
	"github.com/goadesign/goa/logging/logrus"
	"github.com/goadesign/goa/middleware"
	"github.com/goadesign/goa/middleware/gzip"
	goajwt "github.com/goadesign/goa/middleware/security/jwt"
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
	service.Use(log.LogRequest(config.IsDeveloperModeEnabled()))

	//service.Use(witmiddleware.TokenContext(publicKeys, nil, app.NewJWTSecurity()))
	//app.UseJWTMiddleware(service, goajwt.New(publicKeys, nil, app.NewJWTSecurity()))

	tokenManager, err := token.NewManager(config)
	if err != nil {
		log.Panic(nil, map[string]interface{}{
			"err": err,
		}, "failed to create token manager")
	}
	// Middleware that extracts and stores the token in the context
	jwtMiddlewareTokenContext := witmiddleware.TokenContext(tokenManager.PublicKeys(), nil, app.NewJWTSecurity())
	service.Use(jwtMiddlewareTokenContext)
	service.Use(login.InjectTokenManager(tokenManager))
	app.UseJWTMiddleware(service, goajwt.New(tokenManager.PublicKeys(), nil, app.NewJWTSecurity()))

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
