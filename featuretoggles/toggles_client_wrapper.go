package featuretoggles

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Unleash/unleash-client-go"
	unleashapi "github.com/Unleash/unleash-client-go/api"
	unleashcontext "github.com/Unleash/unleash-client-go/context"
	"github.com/fabric8-services/fabric8-auth/log"
	authclient "github.com/fabric8-services/fabric8-toggles-service/auth/client"
)

// UnleashClient the interface to the unleash client
type UnleashClient interface {
	Ready() <-chan bool
	GetFeature(name string) *unleashapi.Feature
	IsEnabled(feature string, options ...unleash.FeatureOption) (enabled bool)
	GetFeaturesByPattern(pattern string) []unleashapi.Feature
	Close() error
}

// Client the toggle client interface
type Client interface {
	GetFeature(ctx context.Context, name string, user *authclient.User) UserFeature
	GetFeaturesByName(ctx context.Context, names []string, user *authclient.User) []UserFeature
	GetFeaturesByPattern(ctx context.Context, pattern string, user *authclient.User) []UserFeature
	// IsFeatureEnabled(ctx context.Context, feature UserFeature, user *authclient.User) (bool, string)
	Close() error
}

// ClientImpl the toggle client default impl
type ClientImpl struct {
	UnleashClient  UnleashClient
	clientListener *UnleashClientListener
}

// verify that `ClientImpl`` is a valid impl of the `Client`` interface
var _ Client = &ClientImpl{}

// ToggleServiceConfiguration the configuration to the Toggle service
type ToggleServiceConfiguration interface {
	// GetToggleServiceAppName() string
	GetTogglesURL() string
}

// NewDefaultClient returns a new client to the toggle feature service including the default underlying unleash client initialized
func NewDefaultClient(serviceName string, config ToggleServiceConfiguration) (Client, error) {
	l := UnleashClientListener{ready: false}
	unleashclient, err := unleash.NewClient(
		unleash.WithAppName(serviceName),
		unleash.WithInstanceId(os.Getenv("HOSTNAME")),
		unleash.WithUrl(config.GetTogglesURL()),
		unleash.WithStrategies(EnableByLevelStrategy{}, EnableByEmailsStrategy{}),
		unleash.WithMetricsInterval(1*time.Minute),
		unleash.WithRefreshInterval(10*time.Second),
		unleash.WithListener(&l),
	)
	if err != nil {
		return nil, err
	}
	return &ClientImpl{
		UnleashClient:  unleashclient,
		clientListener: &l,
	}, nil
}

// NewClientWithState returns a new client to the toggle feature service with a pre-initialized unleash client listener
func NewClientWithState(unleashclient UnleashClient, ready bool) Client {
	return &ClientImpl{
		UnleashClient:  unleashclient,
		clientListener: &UnleashClientListener{ready: ready},
	}
}

// Close closes the underlying Unleash client
func (c *ClientImpl) Close() error {
	return c.UnleashClient.Close()
}

// GetFeature returns the feature given its name
func (c *ClientImpl) GetFeature(ctx context.Context, name string, user *authclient.User) UserFeature {
	if !c.clientListener.ready {
		log.Error(ctx, map[string]interface{}{"error": "client is not ready"}, "unable to list features by name")
		return UserFeature{}
	}
	f := c.UnleashClient.GetFeature(name)
	if f == nil {
		return UserFeature{}
	}
	return c.toUserFeature(ctx, *f, user)
}

func (c *ClientImpl) toUserFeature(ctx context.Context, f unleashapi.Feature, user *authclient.User) UserFeature {
	userEnabled, enablementLevel := c.isFeatureEnabled(ctx, f, user)
	return UserFeature{
		Name:            f.Name,
		Description:     f.Description,
		Enabled:         f.Enabled,
		UserEnabled:     userEnabled,
		EnablementLevel: enablementLevel,
	}
}

// GetFeaturesByName returns the features from their names
func (c *ClientImpl) GetFeaturesByName(ctx context.Context, names []string, user *authclient.User) []UserFeature {
	result := make([]UserFeature, 0)
	if !c.clientListener.ready {
		log.Error(ctx, map[string]interface{}{"error": "client is not ready"}, "unable to list features by name")
		return result
	}
	for _, name := range names {
		f := c.GetFeature(ctx, name, user)
		if f != ZeroUserFeature {
			result = append(result, f)
		}
	}
	return result
}

// GetFeaturesByPattern returns the features whose ID matches the given pattern
func (c *ClientImpl) GetFeaturesByPattern(ctx context.Context, pattern string, user *authclient.User) []UserFeature {
	result := make([]UserFeature, 0)
	if !c.clientListener.ready {
		log.Error(ctx, map[string]interface{}{"error": "client is not ready"}, "unable to list features by pattern")
		return result
	}
	feats := c.UnleashClient.GetFeaturesByPattern(fmt.Sprintf("^%[1]s$|^%[1]s\\.(.*)", pattern))
	for _, f := range feats {
		result = append(result, c.toUserFeature(ctx, f, user))
	}
	return result
}

// isFeatureEnabled returns a boolean to specify whether on feature is enabled for a given user level
func (c *ClientImpl) isFeatureEnabled(ctx context.Context, feature unleashapi.Feature, user *authclient.User) (bool, string) {
	if !c.clientListener.ready {
		log.Warn(ctx, nil, "unable to check if feature is enabled due to: client is not ready")
		return false, UnknownLevel
	}
	internalUser := false
	userLevel := ReleasedLevel // default level of features that the user can use
	userEmail := ""            // default email: empty
	if user != nil {
		if user.Data.Attributes.Email != nil {
			userEmail = *user.Data.Attributes.Email
		}
		// internal users have may be able to access the feature by opting-in to the `internal` level of features.
		if user.Data.Attributes.EmailVerified != nil && *user.Data.Attributes.EmailVerified && strings.HasSuffix(userEmail, "@redhat.com") {
			internalUser = true
		}
		// do not override the userLevel if the value is nil or empty. Any other value is accepted,
		// but will be converted (with a fallback to `unknown` if needed)
		if user.Data.Attributes.FeatureLevel != nil && *user.Data.Attributes.FeatureLevel != "" {
			userLevel = *user.Data.Attributes.FeatureLevel
		}
	}
	log.Debug(ctx, map[string]interface{}{"user_level": userLevel, "user_email": userEmail}, "checking if feature is enabled for user...")
	userEnabled := c.UnleashClient.IsEnabled(
		feature.Name,
		unleash.WithContext(unleashcontext.Context{
			Properties: map[string]string{
				LevelParameter:  userLevel,
				EmailsParameter: userEmail,
			},
		}),
	)
	enablementLevel := ComputeEnablementLevel(ctx, feature, internalUser)
	return userEnabled, enablementLevel
}
