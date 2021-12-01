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
	Configdir string
	Version   bool
}

func GetArguments() (CommandlineArguments, error) {
	f := functionGetArguments
	f.DebugVerbose("Get the comandline arguments")

	rootDir := debug.RootDir()
	defaultConfigDir := filepath.Join(rootDir, "config")
	configdir := flag.String("config", defaultConfigDir, "configuration directory")

	version := flag.Bool("version", false, "display the version")

	flag.Parse()

	var args CommandlineArguments
	args.Configdir = *configdir
	args.Version = *version

	f.DebugVerbose("args.Configdir: %s", args.Configdir)
	f.DebugVerbose("args.Version: %t", args.Version)

	return args, nil
}
