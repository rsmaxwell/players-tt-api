package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
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
	pkg          = debug.NewPackage("config")
	functionOpen = debug.NewFunction(pkg, "Open")
)

const (
	DefaultConfigFile = "config.json"
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

		listFileInDir(d, configFileName)

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

func listFileInDir(d *debug.Dump, filename string) error {

	dir := filepath.Dir(filename)

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}

	var b bytes.Buffer
	line := fmt.Sprintf("directory: %s\n", dir)
	b.WriteString(line)

	length := 0
	for _, file := range files {
		line = fmt.Sprintf("%d", file.Size())
		if len(line) > length {
			length = len(line)
		}
	}

	for _, info := range files {
		dirString := "-"
		if info.IsDir() {
			dirString = "d"
		}
		unixPerms := info.Mode() & os.ModePerm
		permString := fmt.Sprintf("%v", unixPerms)
		line = fmt.Sprintf("%s%s  %*d %s\n", dirString, permString, length, info.Size(), info.Name())
		b.WriteString(line)
	}

	d.AddString("dirlist.txt", b.String())
	return nil
}
