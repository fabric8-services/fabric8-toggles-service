package configuration

import (
	"fmt"
	"github.com/fabric8-services/fabric8-wit/log"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
	"strings"
	"time"
)

const (
	defaultLogLevel   = "info"
	defaultWitURL     = "https://api.prod-preview.openshift.io/api/"
	defaultTogglesURL = "http://f8toggles/api"
)

const (
	// Constants for viper variable names. Will be used to set
	// default values as well as to get each value

	varPostgresHost                   = "postgres.host"
	varPostgresPort                   = "postgres.port"
	varPostgresUser                   = "postgres.user"
	varPostgresDatabase               = "postgres.database"
	varPostgresPassword               = "postgres.password"
	varPostgresSSLMode                = "postgres.sslmode"
	varPostgresConnectionTimeout      = "postgres.connection.timeout"
	varPostgresConnectionRetrySleep   = "postgres.connection.retrysleep"
	varPostgresConnectionMaxIdle      = "postgres.connection.maxidle"
	varPostgresConnectionMaxOpen      = "postgres.connection.maxopen"
	varHTTPAddress                    = "http.address"
	varDeveloperModeEnabled           = "developer.mode.enabled"
	varTogglesURL                     = "toggles.url"
	varConsoleURL                     = "console.url"
	varWitURL                         = "wit.url"
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
	//---------
	// Postgres
	//---------
	c.v.SetTypeByDefaultValue(true)
	c.v.SetDefault(varPostgresHost, "localhost")
	c.v.SetDefault(varPostgresPort, 5432)
	c.v.SetDefault(varPostgresUser, "postgres")
	c.v.SetDefault(varPostgresDatabase, "tenant")
	c.v.SetDefault(varPostgresPassword, "mysecretpassword")
	c.v.SetDefault(varPostgresSSLMode, "disable")
	c.v.SetDefault(varPostgresConnectionTimeout, 5)
	c.v.SetDefault(varPostgresConnectionMaxIdle, -1)
	c.v.SetDefault(varPostgresConnectionMaxOpen, -1)
	// Number of seconds to wait before trying to connect again
	c.v.SetDefault(varPostgresConnectionRetrySleep, time.Duration(time.Second))

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

// GetPostgresHost returns the postgres host as set via default, config file, or environment variable
func (c *Data) GetPostgresHost() string {
	return c.v.GetString(varPostgresHost)
}

// GetPostgresPort returns the postgres port as set via default, config file, or environment variable
func (c *Data) GetPostgresPort() int64 {
	return c.v.GetInt64(varPostgresPort)
}

// GetPostgresUser returns the postgres user as set via default, config file, or environment variable
func (c *Data) GetPostgresUser() string {
	return c.v.GetString(varPostgresUser)
}

// GetPostgresDatabase returns the postgres database as set via default, config file, or environment variable
func (c *Data) GetPostgresDatabase() string {
	return c.v.GetString(varPostgresDatabase)
}

// GetPostgresPassword returns the postgres password as set via default, config file, or environment variable
func (c *Data) GetPostgresPassword() string {
	return c.v.GetString(varPostgresPassword)
}

// GetPostgresSSLMode returns the postgres sslmode as set via default, config file, or environment variable
func (c *Data) GetPostgresSSLMode() string {
	return c.v.GetString(varPostgresSSLMode)
}

// GetPostgresConnectionTimeout returns the postgres connection timeout as set via default, config file, or environment variable
func (c *Data) GetPostgresConnectionTimeout() int64 {
	return c.v.GetInt64(varPostgresConnectionTimeout)
}

// GetPostgresConnectionRetrySleep returns the number of seconds (as set via default, config file, or environment variable)
// to wait before trying to connect again
func (c *Data) GetPostgresConnectionRetrySleep() time.Duration {
	return c.v.GetDuration(varPostgresConnectionRetrySleep)
}

// GetPostgresConnectionMaxIdle returns the number of connections that should be keept alive in the database connection pool at
// any given time. -1 represents no restrictions/default behavior
func (c *Data) GetPostgresConnectionMaxIdle() int {
	return c.v.GetInt(varPostgresConnectionMaxIdle)
}

// GetPostgresConnectionMaxOpen returns the max number of open connections that should be open in the database connection pool.
// -1 represents no restrictions/default behavior
func (c *Data) GetPostgresConnectionMaxOpen() int {
	return c.v.GetInt(varPostgresConnectionMaxOpen)
}

// GetPostgresConfigString returns a ready to use string for usage in sql.Open()
func (c *Data) GetPostgresConfigString() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s connect_timeout=%d",
		c.GetPostgresHost(),
		c.GetPostgresPort(),
		c.GetPostgresUser(),
		c.GetPostgresPassword(),
		c.GetPostgresDatabase(),
		c.GetPostgresSSLMode(),
		c.GetPostgresConnectionTimeout(),
	)
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

// GetConsoleURL returns the fabric8-ui Console URL
func (c *Data) GetConsoleURL() string {
	if c.v.IsSet(varConsoleURL) {
		return c.v.GetString(varConsoleURL)
	}
	return ""
}

// GetWitURL returns WIT URL
func (c *Data) GetWitURL() string {
	if c.v.IsSet(varWitURL) {
		return c.v.GetString(varWitURL)
	}
	return defaultWitURL
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
