package cmdline

import (
	"flag"
	"path/filepath"

	"github.com/rsmaxwell/players-tt-api/internal/debug"
)

var (
	pkg                  = debug.NewPackage("cmdline")
	functionGetArguments = debug.NewFunction(pkg, "GetArguments")
)

// Config type
type CommandlineArguments struct {
	Configfile string `json:"configfile"`
}

func GetArguments() (CommandlineArguments, error) {
	f := functionGetArguments
	f.Verbosef("Get the comandline arguments")

	rootDir := debug.RootDir()
	defaultConfigFile := filepath.Join(rootDir, "config", "config.json")
	configfile := flag.String("config", defaultConfigFile, "configuration filename")

	flag.Parse()

	var commandlineArguments CommandlineArguments
	commandlineArguments.Configfile = *configfile

	return commandlineArguments, nil
}
