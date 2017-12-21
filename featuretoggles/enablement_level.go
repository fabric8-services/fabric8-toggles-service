package featuretoggles

import (
	"context"
	"strings"

	unleashapi "github.com/Unleash/unleash-client-go/api"
	"github.com/fabric8-services/fabric8-wit/log"
)

type FeatureLevel int

const (
	unknown FeatureLevel = iota
	internal
	experimental
	beta
)

// ComputeEnablementLevel computes the enablement level required to be able to use the given feature (if it is enabled at all)
func ComputeEnablementLevel(ctx context.Context, feature *unleashapi.Feature, internalUser bool) *string {
	enablementLevel := unknown
	// iterate on feature's strategies
	for _, s := range feature.Strategies {
		log.Debug(ctx, map[string]interface{}{"feature_name": feature.Name, "enablement_level": enablementLevel, "strategy_name": s.Name}, "computing enablement level")
		if s.Name == EnableByGroupID {
			if groupID, found := s.Parameters["groupID"]; found {
				if groupIDStr, ok := groupID.(string); ok {
					featureLevel := toFeatureLevel(groupIDStr, internalUser)
					log.Debug(ctx, map[string]interface{}{"feature_name": feature.Name, "enablement_level": enablementLevel, "strategy_group": featureLevel}, "computing enablement level")
					// beta > experimental > internal (if user is a RH internal)
					if featureLevel > enablementLevel {
						enablementLevel = featureLevel
					}
				}
			}
		}
	}
	log.Debug(ctx, map[string]interface{}{"internal_user": internalUser, "feature_name": feature.Name, "enablement_level": enablementLevel}, "computed enablement level")
	return fromFeatureLevel(enablementLevel)
}

func toFeatureLevel(level string, internalUser bool) FeatureLevel {
	switch strings.ToLower(level) {
	case "internal":
		if internalUser {
			return internal
		}
		return unknown
	case "experimental":
		return experimental
	case "beta":
		return beta
	default:
		return unknown
	}
}

func fromFeatureLevel(level FeatureLevel) *string {
	var result string
	switch level {
	case internal:
		result = "internal"
	case experimental:
		result = "experimental"
	case beta:
		result = "beta"
	default:
		return nil
	}

	return &result
}
