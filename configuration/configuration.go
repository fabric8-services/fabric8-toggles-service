package configuration

import (
	"fmt"
	"github.com/fabric8-services/fabric8-wit/log"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
	"strings"
)

const (
	defaultLogLevel   = "info"
	defaultWitURL     = "https://api.prod-preview.openshift.io/api/"
	defaultTogglesURL = "http://f8toggles/api"
)

const (
	// Constants for viper variable names. Will be used to set
	// default values as well as to get each value

	varHTTPAddress                    = "http.address"
	varDeveloperModeEnabled           = "developer.mode.enabled"
	varTogglesURL                     = "toggles.url"
	varAPIServerInsecureSkipTLSVerify = "api.server.insecure.skip.tls.verify"
	varLogLevel                       = "log.level"
	varLogJSON                        = "log.json"
)

// Data encapsulates the Viper configuration object which stores the configuration data in-memory.
type Data struct {
	v *viper.Viper
}

// NewData creates a configuration reader object using a configurable configuration file path
func NewData() (*Data, error) {
	c := Data{
		v: viper.New(),
	}
	c.v.SetEnvPrefix("F8")
	c.v.AutomaticEnv()
	c.v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	c.v.SetTypeByDefaultValue(true)
	c.setConfigDefaults()

	return &c, nil
}

// String returns the current configuration as a string
func (c *Data) String() string {
	allSettings := c.v.AllSettings()
	y, err := yaml.Marshal(&allSettings)
	if err != nil {
		log.Panic(nil, map[string]interface{}{
			"settings": allSettings,
			"err":      err,
		}, "Failed to marshall config to string")
	}
	return fmt.Sprintf("%s\n", y)
}

// GetData is a wrapper over NewData which reads configuration file path
// from the environment variable.
func GetData() (*Data, error) {
	cd, err := NewData()
	return cd, err
}

func (c *Data) setConfigDefaults() {

	//-----
	// HTTP
	//-----
	c.v.SetDefault(varHTTPAddress, "0.0.0.0:8080")

	//-----
	// Misc
	//-----
	c.v.SetDefault(varAPIServerInsecureSkipTLSVerify, false)
	c.v.SetDefault(varTogglesURL, defaultTogglesURL)

	//-----
	// Enable development related features, e.g. token generation endpoint
	//-----
	c.v.SetDefault(varDeveloperModeEnabled, false)
	c.v.SetDefault(varLogLevel, defaultLogLevel)

}

// GetHTTPAddress returns the HTTP address (as set via default, config file, or environment variable)
// that the alm server binds to (e.g. "0.0.0.0:8080")
func (c *Data) GetHTTPAddress() string {
	return c.v.GetString(varHTTPAddress)
}

// IsDeveloperModeEnabled returns if development related features (as set via default, config file, or environment variable),
// e.g. token generation endpoint are enabled
func (c *Data) IsDeveloperModeEnabled() bool {
	return c.v.GetBool(varDeveloperModeEnabled)
}

// GetTogglesURL returns Toggle service URL
func (c *Data) GetTogglesURL() string {
	return c.v.GetString(varTogglesURL)
}

// APIServerInsecureSkipTLSVerify returns if the server's certificate should be checked for validity. This will make your HTTPS connections insecure.
func (c *Data) APIServerInsecureSkipTLSVerify() bool {
	return c.v.GetBool(varAPIServerInsecureSkipTLSVerify)
}

// GetLogLevel returns the loggging level (as set via config file or environment variable)
func (c *Data) GetLogLevel() string {
	return c.v.GetString(varLogLevel)
}

// IsLogJSON returns if we should log json format (as set via config file or environment variable)
func (c *Data) IsLogJSON() bool {
	if c.v.IsSet(varLogJSON) {
		return c.v.GetBool(varLogJSON)
	}
	if c.IsDeveloperModeEnabled() {
		return false
	}
	return true
}
