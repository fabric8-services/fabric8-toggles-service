package featuretoggles

import (
	"context"
	"strings"

	unleashapi "github.com/Unleash/unleash-client-go/api"
	"github.com/fabric8-services/fabric8-auth/log"
)

// FeatureLevel custom type for feature level constants as integers, for comparisons
type FeatureLevel int

const (
	unknown FeatureLevel = iota
	internal
	experimental
	beta
	released
)

// FeatureLevelStr custom type for feature level constants as strings
const (
	// UnknownLevel the unknown level for feature toggles (only used if the backend config does not match any level here)
	UnknownLevel = "unknown"
	// InternalLevel the Internal level for feature toggles
	InternalLevel = "internal"
	// ExperimentalLevel the Experimental level for feature toggles
	ExperimentalLevel = "experimental"
	// BetaLevel the Beta level for feature toggles
	BetaLevel = "beta"
	// ReleasedLevel the Released level for feature toggles
	ReleasedLevel = "released"
)

// ComputeEnablementLevel computes the enablement level required to be able to use the given feature (if it is enabled at all)
func ComputeEnablementLevel(ctx context.Context, feature *unleashapi.Feature, internalUser bool) *string {
	enablementLevel := unknown
	// iterate on feature's strategies
	for _, s := range feature.Strategies {
		// log.Debug(ctx, map[string]interface{}{"feature_name": feature.Name, "enablement_level": enablementLevel, "strategy_name": s.Name}, "computing enablement level")
		if s.Name == EnableByLevelStrategyName {
			if level, found := s.Parameters[LevelParameter]; found {
				if levelStr, ok := level.(string); ok {
					featureLevel := toFeatureLevel(levelStr, internalUser)
					// log.Debug(ctx, map[string]interface{}{"feature_name": feature.Name, "enablement_level": enablementLevel, "strategy_group": featureLevel}, "computing enablement level")
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
	case InternalLevel:
		if internalUser {
			return internal
		}
		return unknown
	case ExperimentalLevel:
		return experimental
	case BetaLevel:
		return beta
	case ReleasedLevel:
		return released
	default:
		return unknown
	}
}

func fromFeatureLevel(level FeatureLevel) *string {
	var result string
	switch level {
	case internal:
		result = InternalLevel
	case experimental:
		result = ExperimentalLevel
	case beta:
		result = BetaLevel
	case released:
		result = ReleasedLevel
	default:
		return nil
	}

	return &result
}

func (featureLevel FeatureLevel) IsEnabled(userLevel string) bool {
	userLevelInt := toFeatureLevel(userLevel, userLevel == InternalLevel)
	return featureLevel >= userLevelInt
}
