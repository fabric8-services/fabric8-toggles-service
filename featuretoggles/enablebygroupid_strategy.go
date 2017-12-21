package featuretoggles

import (
	unleashcontext "github.com/Unleash/unleash-client-go/context"
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
	return settings[GroupID] == ctx.Properties[GroupID]
}
