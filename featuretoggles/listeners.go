package featuretoggles

import (
	unleash "github.com/Unleash/unleash-client-go"
	"github.com/fabric8-services/fabric8-wit/log"
)

var ready = false

type MetricsListener struct {
}

func (m *MetricsListener) OnCount(name string, enabled bool) {
	log.Debug(nil, map[string]interface{}{
		"name":    name,
		"enabled": enabled,
	}, "toggles count")
}

// OnSent prints to the console when the server has uploaded metrics.
func (m *MetricsListener) OnSent(payload unleash.MetricsData) {
	log.Info(nil, map[string]interface{}{
		"payload": payload,
	}, "toggles sent")
}

// OnRegistered prints to the console when the client has registered.
func (m *MetricsListener) OnRegistered(payload unleash.ClientData) {
	log.Info(nil, map[string]interface{}{
		"payload": payload,
	}, "toggles registered")
}

// OnError prints out errors.
func (m *MetricsListener) OnError(err error) {
	log.Error(nil, map[string]interface{}{
		"err": err.Error(),
	}, "toggles error")
}

// OnWarning prints out warning.
func (m *MetricsListener) OnWarning(warning error) {
	log.Warn(nil, map[string]interface{}{
		"err": warning.Error(),
	}, "toggles warning")
}

// OnReady prints to the console when the repository is ready.
func (m *MetricsListener) OnReady() {
	ready = true
	log.Info(nil, map[string]interface{}{}, "toggles ready")
}
