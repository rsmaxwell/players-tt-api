package config

import (
	"database/sql"
	"encoding/json"
	"flag"
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/rsmaxwell/players-tt-api/internal/debug"

	_ "github.com/jackc/pgx/stdlib"
)

// Database type
type Database struct {
	DriverName   string `json:"driverName"`
	UserName     string `json:"userName"`
	Password     string `json:"password"`
	Scheme       string `json:"scheme"`
	Host         string `json:"host"`
	Path         string `json:"path"`
	DatabaseName string `json:"databaseName"`
}

// Server type
type Server struct {
	Port int `json:"port"`
}

// Mqtt type
type Mqtt struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	ClientID string `json:"clientId"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// Config type
type ConfigFile struct {
	Database           Database `json:"database"`
	Server             Server   `json:"server"`
	Mqtt               Mqtt     `json:"mqtt"`
	AccessTokenExpiry  string   `json:"accessToken_expiry"`
	RefreshTokenExpiry string   `json:"refreshToken_expiry"`
	ClientRefreshDelta string   `json:"clientRefreshDelta"`
}

// Config type
type Config struct {
	Database           Database
	Server             Server
	Mqtt               Mqtt
	AccessTokenExpiry  time.Duration
	RefreshTokenExpiry time.Duration
	ClientRefreshDelta time.Duration
}

var (
	pkg           = debug.NewPackage("config")
	functionOpen  = debug.NewFunction(pkg, "Open")
	functionSetup = debug.NewFunction(pkg, "Setup")
)

// Setup function
func Setup() (*sql.DB, *Config, error) {
	f := functionSetup

	// Find the configuration file
	rootDir := debug.RootDir()
	defaultConfigFile := filepath.Join(rootDir, "config", "config.json")
	configFileName := flag.String("config", defaultConfigFile, "configuration filename")

	flag.Parse()

	// Read the configuration file
	c, err := Open(*configFileName)
	if err != nil {
		message := "Could not open configuration"
		f.Errorf(message)
		f.DumpError(err, message)
		return nil, nil, err
	}

	// Connect to the database
	driverName := c.DriverName()
	connectionString := c.ConnectionString()
	f.DebugVerbose("driverName: %s", driverName)
	f.DebugVerbose("connectionString: %s", connectionString)

	db, err := sql.Open(driverName, connectionString)
	if err != nil {
		message := "Could not connect to the database"
		f.Errorf(message)
		f.DumpError(err, message)
		return nil, nil, err
	}

	return db, c, nil
}

// Open returns the configuration
func Open(configFileName string) (*Config, error) {
	f := functionOpen

	bytearray, err := ioutil.ReadFile(configFileName)
	if err != nil {
		f.Dump("could not read config file: %s", configFileName)
		return nil, err
	}

	var configFile ConfigFile
	err = json.Unmarshal(bytearray, &configFile)
	if err != nil {
		f.Dump("could not Unmarshal configuration")
		return nil, err
	}

	return configFile.toConfig()
}
