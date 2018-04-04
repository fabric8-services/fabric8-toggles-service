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
	internal FeatureLevel = iota
	experimental
	beta
	released
	unknown
)

// FeatureLevelStr custom type for feature level constants as strings
const (
	// InternalLevel the Internal level for feature toggles
	InternalLevel = "internal"
	// ExperimentalLevel the Experimental level for feature toggles
	ExperimentalLevel = "experimental"
	// BetaLevel the Beta level for feature toggles
	BetaLevel = "beta"
	// ReleasedLevel the Released level for feature toggles
	ReleasedLevel = "released"
	// UnknownLevel the unknown level for feature toggles (only used if the backend config does not match any level here)
	UnknownLevel = "unknown"
)

// ComputeEnablementLevel computes the enablement level required to be able to use the given feature (if it is enabled at all)
func ComputeEnablementLevel(ctx context.Context, feature unleashapi.Feature, internalUser bool) string {
	if feature.Enabled == false || len(feature.Strategies) == 0 {
		return UnknownLevel
	}
	enablementLevel := internal
	// iterate on feature's strategies
	for _, s := range feature.Strategies {
		// log.Debug(ctx, map[string]interface{}{"feature_name": feature.Name, "enablement_level": enablementLevel, "strategy_name": s.Name}, "computing enablement level")
		if s.Name == EnableByLevelStrategyName {
			if level, found := s.Parameters[LevelParameter]; found {
				if levelStr, ok := level.(string); ok {
					featureLevel := toFeatureLevel(levelStr, unknown)
					// log.Debug(ctx, map[string]interface{}{"feature_name": feature.Name, "enablement_level": enablementLevel, "strategy_group": featureLevel}, "computing enablement level")
					// beta > experimental > internal (if user is a RH internal)
					if featureLevel > enablementLevel {
						enablementLevel = featureLevel
					}
				}
			}
		}
	}
	// need to re-adjust the level if the user is "external" and the enablement level is "internal"
	// i.e., the feature is internal-only, so the external user is not allowed to use it
	if !internalUser && enablementLevel == internal {
		enablementLevel = unknown
	}
	result := fromFeatureLevel(enablementLevel)
	log.Debug(ctx, map[string]interface{}{"internal_user": internalUser, "feature_name": feature.Name, "enablement_level": result}, "computed enablement level")
	return result
}

func toFeatureLevel(level string, defaultLevel FeatureLevel) (result FeatureLevel) {
	defer log.Debug(nil,
		map[string]interface{}{
			"feature_level": result,
			"level":         level},
		"converted feature level")
	switch strings.ToLower(level) {
	case InternalLevel:
		return internal
	case ExperimentalLevel:
		return experimental
	case BetaLevel:
		return beta
	case ReleasedLevel:
		return released
	default:
		return defaultLevel
	}
}

func fromFeatureLevel(level FeatureLevel) string {
	switch level {
	case internal:
		return InternalLevel
	case experimental:
		return ExperimentalLevel
	case beta:
		return BetaLevel
	case released:
		return ReleasedLevel
	default:
		return UnknownLevel
	}
}

// IsEnabled verifies if this feature is enabled for the given userLevel
func (featureLevel FeatureLevel) IsEnabled(userLevel string) bool {
	userLevelInt := toFeatureLevel(userLevel, released)
	return featureLevel >= userLevelInt
}
