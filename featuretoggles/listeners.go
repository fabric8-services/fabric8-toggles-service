package featuretoggles

import (
	unleash "github.com/Unleash/unleash-client-go"
	"github.com/fabric8-services/fabric8-wit/log"
)

// UnleashClientListener a listener to the unleash client. Retains the `ready` state of the client it is registered to.
type UnleashClientListener struct {
	ready bool
}

// OnError prints out errors.
func (l *UnleashClientListener) OnError(err error) {
	log.Error(nil, map[string]interface{}{
		"err": err.Error(),
	}, "toggles error")
}

// OnWarning prints out warning.
func (l *UnleashClientListener) OnWarning(warning error) {
	log.Warn(nil, map[string]interface{}{
		"err": warning.Error(),
	}, "toggles warning")
}

// OnReady prints to the console when the repository is ready.
func (l *UnleashClientListener) OnReady() {
	l.ready = true
	log.Info(nil, map[string]interface{}{}, "toggles ready")
}

// OnCount prints to the console when the feature is queried.
func (l *UnleashClientListener) OnCount(name string, enabled bool) {
	log.Info(nil, map[string]interface{}{
		"name":    name,
		"enabled": enabled,
	}, "toggles count")
}

// OnSent prints to the console when the server has uploaded metrics.
func (l *UnleashClientListener) OnSent(payload unleash.MetricsData) {
	log.Info(nil, map[string]interface{}{
		"payload": payload,
	}, "toggles sent")
}

// OnRegistered prints to the console when the client has registered.
func (l *UnleashClientListener) OnRegistered(payload unleash.ClientData) {
	log.Info(nil, map[string]interface{}{
		"payload": payload,
	}, "toggles registered")
}
