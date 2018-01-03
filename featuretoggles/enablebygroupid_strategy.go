package featuretoggles

import (
	unleashcontext "github.com/Unleash/unleash-client-go/context"
	"github.com/fabric8-services/fabric8-wit/log"
)

const (
	EnableByGroupID string = "enableByGroupID"
	GroupID         string = "groupID"
)

// EnableByGroupIDStrategy the strategy to roll out a feature if the user belongs to a given group
type EnableByGroupIDStrategy struct {
}

// Name the name of the stragegy. Must match the name on the Unleash server.
func (s *EnableByGroupIDStrategy) Name() string {
	return EnableByGroupID
}

// IsEnabled returns `true` if the given context is compatible with the settings configured on the Unleash server
func (s *EnableByGroupIDStrategy) IsEnabled(settings map[string]interface{}, ctx *unleashcontext.Context) bool {
	log.Debug(nil, map[string]interface{}{"settings_group_id": settings[GroupID], "propeties_group_id": ctx.Properties[GroupID]}, "checking if feature is enabled for user...")
	return settings[GroupID] == ctx.Properties[GroupID]
}
