package featuretoggles

import (
	"strings"

	unleashcontext "github.com/Unleash/unleash-client-go/context"
	"github.com/fabric8-services/fabric8-auth/log"
)

const (
	// EnableByEmailsStrategyName the name of the strategy
	EnableByEmailsStrategyName string = "enableByEmails"
	// EmailsParameter the name of the 'email' parameter in the strategy
	EmailsParameter string = "emails"
)

// EnableByEmailsStrategy the strategy to roll out a feature if the user has the expected/configured email address
type EnableByEmailsStrategy struct {
}

// Name the name of the stragegy. Must match the name on the Unleash server.
func (s EnableByEmailsStrategy) Name() string {
	return EnableByEmailsStrategyName
}

// IsEnabled returns `true` if the given context is compatible with the settings configured on the Unleash server
func (s EnableByEmailsStrategy) IsEnabled(settings map[string]interface{}, ctx *unleashcontext.Context) bool {
	log.Debug(nil, map[string]interface{}{"settings_emails": settings[EmailsParameter], "context_email": ctx.Properties[EmailsParameter]}, "checking if feature is enabled for user, based on his/her email...")
	settingsEmails := settings[EmailsParameter]
	userEmail := ctx.Properties[EmailsParameter]
	if settingsEmails, ok := settingsEmails.(string); ok {
		emails := strings.Split(settingsEmails, ",")
		log.Debug(nil, map[string]interface{}{"emails": emails, "context_email": ctx.Properties[EmailsParameter]}, "checking if feature is enabled for user, based on his/her email...")
		for _, email := range emails {
			log.Debug(nil, map[string]interface{}{"email": email, "context_email": ctx.Properties[EmailsParameter]}, "checking if feature is enabled for user, based on his/her email...")
			if userEmail == strings.TrimSpace(email) {
				log.Debug(nil, map[string]interface{}{"settings_emails": settings[EmailsParameter], "context_email": ctx.Properties[EmailsParameter]}, "feature is enabled for user, based on his/her email.")
				return true
			}
		}
	}
	return false

}
