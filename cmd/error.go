package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/jacobsa/daemonize"

	"github.com/zbiljic/memfs/pkg/console"
	"github.com/zbiljic/memfs/pkg/sysinfo"
)

var (
	// Exit status code that will be used if the application returned an error.
	errorExitStatusCode = globalErrorExitStatus
)

// fatalIf wrapper function which takes error and selectively prints stack
// frames if available on debug
func fatalIf(err error, msg string) {
	if err == nil {
		return
	}
	log.Println("ERROR", msg, fmt.Sprintf("%+v", err))

	if globalIsDaemon {
		daemonize.SignalOutcome(fmt.Errorf("%s %s", msg, err.Error()))
		os.Exit(1)
		return
	}
	if !globalDebug {
		console.Fatalln(fmt.Sprintf("%s %s", msg, err.Error()))
	}
	sysInfo := sysinfo.GetSysInfo()
	console.Fatalln(fmt.Sprintf("%s %+v", msg, err), "\n", sysInfo)
}

// errorIf synonymous with fatalIf but doesn't exit on error != nil
func errorIf(err error, msg string) {
	if err == nil {
		return
	}
	log.Println("ERROR", msg, fmt.Sprintf("%+v", err))

	if globalIsDaemon {
		daemonize.SignalOutcome(fmt.Errorf("%s %s", msg, err.Error()))
		return
	}
	if !globalDebug {
		console.Errorln(fmt.Sprintf("%s %s", msg, err.Error()))
		return
	}
	sysInfo := sysinfo.GetSysInfo()
	console.Errorln(fmt.Sprintf("%s %+v", msg, err), "\n", sysInfo)
}

// exitStatus allows setting custom exitStatus number. It returns empty error.
func exitStatus(status int) error {
	errorExitStatusCode = status
	return errDummy()
}

func exit() {
	os.Exit(errorExitStatusCode)
}
