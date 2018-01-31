package cmd

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/spf13/viper"

	"github.com/zbiljic/memfs/pkg/console"
)

const (
	globalErrorExitStatus = 1 // Global error exit status.

	daemonizeEnvVar = "DAEMONIZE_STATUS_FD"
)

var (
	globalQuiet     = false // Quiet flag set via command line
	globalDebug     = false // Debug flag set via command line
	globalLogFile   = ""    // LogFile flag set via command line
	globalPprofAddr = ""    // pprof address flag set via command line
	// WHEN YOU ADD NEXT GLOBAL FLAG, MAKE SURE TO ALSO UPDATE PERSISTENT FLAGS, FLAG CONSTANTS AND UPDATE FUNC.
)

// Special variables
var (
	globalIsDaemon = false // If the running application was run daemonized
)

var (
	// Terminal width
	globalTermWidth int
)

var (
	configSections = map[string]func(string) string{
		"global": nil,
		"args":   nil,
	}
	// WHEN YOU ADD NEXT GLOBAL CONFIG SECTION, MAKE SURE TO ALSO UPDATE TESTS, ETC.
)

var _ error = initConfigSections()

func init() {
	// Checking if daemonizer pipe exists
	_, ok := os.LookupEnv(daemonizeEnvVar)
	if ok {
		globalIsDaemon = true
	}
}

func initConfigSections() error {
	// DO NOT EDIT - builds section functions
	for k := range configSections {
		localK := k
		configSections[localK] = func(key string) string {
			return localK + "." + key
		}
	}
	return nil
}

func configureGlobals() {
	// Enable debug messages if requested.
	if globalDebug {
		console.DebugPrint = true
	}

	if globalLogFile != "" {
		absLogfile, err := filepath.Abs(globalLogFile)
		if err != nil {
			fatalIf(errors.WithStack(err),
				"Could not convert filepath:")
		}
		globalLogFile = absLogfile
	}
}

func updateGlobals() {
	globalSection := configSections["global"]
	globalQuiet = viper.GetBool(globalSection("quiet"))
	globalDebug = viper.GetBool(globalSection("debug"))
	globalLogFile = viper.GetString(globalSection("log-file"))
	globalPprofAddr = viper.GetString(globalSection("pprof-addr"))
}
