package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
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
	pkg          = debug.NewPackage("config")
	functionOpen = debug.NewFunction(pkg, "Open")
)

// Open returns the configuration
func Open(configFileName string) (*Config, error) {
	f := functionOpen
	f.DebugVerbose("configFileName: %s", configFileName)

	bytearray, err := ioutil.ReadFile(configFileName)
	if err != nil {
		d := f.DumpError(err, "could not read config file: %s", configFileName)

		stats, err := os.Stat(configFileName)
		if err != nil {
			d.AddString("fileinfo.txt", fmt.Sprintf("%s\n", err))
		} else {
			d.AddObject("fileinfo.json", stats)
		}

		return nil, err
	}

	var configFile ConfigFile
	err = json.Unmarshal(bytearray, &configFile)
	if err != nil {
		f.DumpError(err, "could not Unmarshal configuration")
		return nil, err
	}

	return configFile.toConfig()
}
