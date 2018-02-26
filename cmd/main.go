package cmd

import (
	"fmt"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof" // Comment this line to disable pprof endpoint.
	"strings"
	"sync"

	"github.com/cheggaaa/pb"
	"github.com/spf13/cobra"
	"github.com/zbiljic/pkg/logger"

	"github.com/zbiljic/memfs/pkg/sysinfo"
)

const (
	// AppName - the name of the application.
	AppName = "memfs"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   AppName,
	Short: "A user-space file system whose contents lives in memory during program lifetime",
	Long:  `A user-space file system whose contents lives in memory during program lifetime.`,
}

// Main starts application
func Main() {
	// run exit function only once
	var once sync.Once
	defer once.Do(mainExitFn)

	// Fetch terminal size, if not available, automatically
	// set globalQuiet to true.
	if w, e := pb.GetTerminalWidth(); e != nil {
		globalQuiet = true
	} else {
		globalTermWidth = w
	}

	if err := rootCmd.Execute(); err != nil {
		once.Do(mainExitFn)
		exit()
	}
}

var mainExitFn = func() {
	log.Printf("DEBUG Version:%s, Build-time:%s, Commit:%s", Version, BuildTime, ShortCommitID)
	log.Println("DEBUG", sysinfo.GetSysInfo())
}

func init() {

	defineFlagsGlobal(rootCmd.PersistentFlags())

	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		log.Printf("DEBUG Running command '%s'", cmd.CommandPath())
		registerBefore(cmd)
	}
}

func registerBefore(cmd *cobra.Command) error {

	// Bind global flags to viper config.
	bindFlagsGlobal(cmd)

	// Update global flags (if anything changed from other sources).
	updateGlobals()

	// Configure global flags.
	configureGlobals()

	// Configure logger.
	logger.SetupLogging(globalDebug, globalQuiet, globalLogFile)

	return nil
}

// validateCommandCall prints an error if the precondition function returns
// false for a command
func validateCommandCall(cmd *cobra.Command, preconditionFunc func() bool) {
	if !preconditionFunc() {
		cmdName := cmd.Name()
		// prepend parent commands
		tempCmd := cmd
		for {
			if !tempCmd.HasParent() {
				break
			}
			if tempCmd.Parent() == cmd.Root() {
				break
			}
			tempCmd = tempCmd.Parent()
			cmdName = tempCmd.Name() + " " + cmdName
		}
		// The command line is wrong. Print help message and stop.
		fatalIf(errInvalidCommandCall(cmdName), "Invalid command call.")
	}
}

func startProfilerServerIfConfigured() {
	if globalPprofAddr != "" {
		go func() {

			addr, err := net.ResolveTCPAddr("tcp", globalPprofAddr)
			if err != nil {
				errorIf(err, "Could not resolve TCP address:")
			}

			pprofHostPort := addr.String()
			parts := strings.Split(pprofHostPort, ":")
			if len(parts) == 2 && parts[0] == "" {
				pprofHostPort = fmt.Sprintf("localhost:%s", parts[1])
			}
			pprofHostPort = "http://" + pprofHostPort + "/debug/pprof"

			log.Printf("INFO Starting pprof HTTP server at: %s", pprofHostPort)

			if err := http.ListenAndServe(globalPprofAddr, nil); err != nil {
				errorIf(err, "Failed to start pprof server:")
			}
		}()
	}
}
