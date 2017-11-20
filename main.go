package main

import (
	"fmt"
	"github.com/fabric8-services/fabric8-toggles-service/configuration"
	"github.com/fabric8-services/fabric8-wit/log"
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
}
