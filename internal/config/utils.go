package config

import (
	"fmt"
	"time"

	"github.com/rsmaxwell/players-tt-api/internal/basic"
	"github.com/rsmaxwell/players-tt-api/internal/debug"
)

var (
	functionGetDuration = debug.NewFunction(pkg, "GetDuration")
)

func (c *ConfigFile) toConfig() (*Config, error) {
	config := Config{Database: c.Database, Server: c.Server, Mqtt: c.Mqtt}

	var err error
	config.AccessTokenExpiry, err = GetDuration("AccessTokenExpiry", c.AccessTokenExpiry, "10m")
	if err != nil {
		return nil, err
	}

	config.RefreshTokenExpiry, err = GetDuration("RefreshTokenExpiry", c.RefreshTokenExpiry, "2h")
	if err != nil {
		return nil, err
	}

	config.ClientRefreshDelta, err = GetDuration("ClientRefreshDelta", c.ClientRefreshDelta, "30s")
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func GetDuration(envvar string, def1 string, def2 string) (time.Duration, error) {
	f := functionGetDuration

	str, err := basic.GetEnvString(envvar, "")
	if err != nil {
		f.DumpError(err, "could get the environment variable [%s]", envvar)
		return 0, err
	}
	if str == "" {
		str = def1
	}
	if str == "" {
		str = def2
	}

	duration, err := time.ParseDuration(str)
	if err != nil {
		f.DumpError(err, "could not parse the AccessTokenExpiry: [%s]", str)
		return 0, err
	}

	return duration, nil
}

// DriverName returns the driver name for the configured database
func (c *Config) DriverName() string {
	return c.Database.DriverName
}

// ConnectionString returns the string used to connect to the database
func (c *Config) ConnectionString() string {
	return fmt.Sprintf("%s://%s:%s@%s/%s", c.Database.Scheme, c.Database.UserName, c.Database.Password, c.Database.Host, c.Database.DatabaseName)
}

// ConnectionStringBasic returns the string used to connect to the database
func (c *Config) ConnectionStringBasic() string {
	return fmt.Sprintf("%s://%s:%s@%s", c.Database.Scheme, c.Database.UserName, c.Database.Password, c.Database.Host)
}
