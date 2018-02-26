package cmd

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"golang.org/x/sys/unix"

	"github.com/jacobsa/daemonize"
	"github.com/jacobsa/fuse"
	"github.com/jacobsa/syncutil"
	"github.com/kardianos/osext"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/zbiljic/pkg/logger"

	"github.com/zbiljic/memfs/filesystem"
	"github.com/zbiljic/memfs/pkg/console"
	mountpkg "github.com/zbiljic/memfs/pkg/mount"
	"github.com/zbiljic/memfs/pkg/user"
)

var mountCmd = &cobra.Command{
	Use:   "mount",
	Short: "Mount a In-Memory file system",
	Long:  `Mount a In-Memory file system.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return checkMountSyntax(cmd, args)
	},
	Run: func(cmd *cobra.Command, args []string) {
		mountMain()
	},
}

var mountFlagsBind = []string{
	"foreground",
	"o",
	"dir-mode",
	"file-mode",
	"uid",
	"gid",
	"debug_fuse",
	"debug_invariants",
}

func init() {
	// add 'mount' command to root command
	rootCmd.AddCommand(mountCmd)
	initFlags(mountFlags, mountFlagsBind, mountCmd)
}

// mountArgs represents all possible arguments that can be specified while
// executing 'mount' command.
type mountArgs struct {
	MountPoint string

	Foreground bool

	// File system
	MountOptions map[string]string
	DirMode      os.FileMode
	FileMode     os.FileMode
	Uid          int
	Gid          int

	// Debugging
	DebugFuse       bool
	DebugInvariants bool
}

var mountArgsHolder = &mountArgs{
	MountOptions: make(map[string]string),
}

func populateArgsHolderMount(args []string) {
	// from command arguments
	mountArgsHolder.MountPoint = valueOrEmptyString(args, 0)

	// from config
	argsSection := configSections["args"]

	mountArgsHolder.Foreground = viper.GetBool(argsSection("foreground"))

	optionsArray := viper.GetStringSlice(argsSection("o"))
	// Handle the repeated "-o" flag.
	for _, o := range optionsArray {
		mountpkg.ParseOptions(mountArgsHolder.MountOptions, o)
	}

	dirMode, err := parseOctalInt(viper.GetString(argsSection("dir-mode")))
	fatalIf(errors.WithStack(err), "Provided value for --dir-mode is not valid")
	mountArgsHolder.DirMode = os.FileMode(dirMode)

	fileMode, err := parseOctalInt(viper.GetString(argsSection("file-mode")))
	fatalIf(errors.WithStack(err), "Provided value for --file-mode is not valid")
	mountArgsHolder.FileMode = os.FileMode(fileMode)

	uid := viper.GetInt(argsSection("uid"))
	mountArgsHolder.Uid = uid

	gid := viper.GetInt(argsSection("gid"))
	mountArgsHolder.Gid = gid

	mountArgsHolder.DebugFuse = viper.GetBool(argsSection("debug_fuse"))
	mountArgsHolder.DebugInvariants = viper.GetBool(argsSection("debug_invariants"))

}

func checkMountSyntax(cmd *cobra.Command, args []string) error {

	populateArgsHolderMount(args)

	if mountArgsHolder.MountPoint == "" {
		fatalIf(errDummy(),
			"Mount point not provided.")
	} else {
		mpfi, err := os.Stat(mountArgsHolder.MountPoint)

		if err != nil {
			fatalIf(errors.WithStack(err),
				"Provided mountpoint not found")
		}

		if !mpfi.IsDir() {
			fatalIf(errDummy(),
				"Provided mountpoint is not a directory")
		}

		if unix.Access(mountArgsHolder.MountPoint, unix.W_OK|unix.R_OK|unix.X_OK) != nil {
			fatalIf(errDummy(),
				"Insufficient permissions to mount on specified directory")
		}

		if list, errList := ioutil.ReadDir(mountArgsHolder.MountPoint); errList != nil || len(list) != 0 {
			fatalIf(errDummy(),
				"Provided mountpoint is not an empty directory")
		}

		absMountPoint, err := filepath.Abs(mountArgsHolder.MountPoint)
		if err != nil {
			fatalIf(errors.WithStack(err),
				"Could not convert filepath:")
		}
		mountArgsHolder.MountPoint = absMountPoint

	}

	if uint32(mountArgsHolder.Uid) > maxUint32 {
		fatalIf(errDummy(),
			"Provided value for --uid is not valid.")
	}
	if uint32(mountArgsHolder.Gid) > maxUint32 {
		fatalIf(errDummy(),
			"Provided value for --gid is not valid.")
	}

	return nil
}

func mountMain() {
	var err error

	// Enable invariant checking if requested.
	if mountArgsHolder.DebugInvariants {
		syncutil.EnableInvariantChecking()
	}

	// If we haven't been asked to run in foreground mode, we should run a daemon
	// and wait for it to mount.
	if mountArgsHolder.Foreground || globalIsDaemon {
		err = foregroundMount()
	} else {
		err = daemonMount()
	}

	// Many errors (for example in daemonizer) will stack messages splitted
	// by colon (:). By default we want to show just the innermost error.
	if err != nil {
		message := err.Error()
		if globalIsDaemon {
			parts := strings.Split(message, ": ")
			message = parts[len(parts)-1]
			message = strings.TrimSpace(message)
		} else {
			// If message came from the demonized process, it may be prefixed with
			// "readFromProcess: sub-process: "
			// We don't want to print this.
			message = strings.TrimPrefix(message, "readFromProcess: sub-process: ")
		}

		errorIf(errDummy(), message)
		exitStatus(globalErrorExitStatus)
		exit()
	}

}

func printWarningForRoot(uid uint32) {
	if uid == 0 && !isInDocker() {
		fmt.Fprintf(os.Stdout,
			`
WARNING: %[1]s invoked as root. This will cause all files to be owned
by root. If this is not what you intended, invoke %[1]s as the user
that will be interacting with the file system.

`,
			AppName)
	}
}

func isInDocker() bool {
	return os.Getenv("DOCKERIMAGE") == "1"
}

func foregroundMount() error {
	if !globalIsDaemon {
		logger.SetupLogging(globalDebug, globalQuiet, "") // will use os.Stderr
	}

	startProfilerServerIfConfigured()

	// Print a warning if we run process as root except when in docker container.
	uid, gid, err := user.MyUserAndGroup()
	if err != nil {
		daemonize.SignalOutcome(err)
		return err
	}
	if mountArgsHolder.Uid > -1 {
		uid = uint32(mountArgsHolder.Uid)
	}
	if mountArgsHolder.Gid > -1 {
		gid = uint32(mountArgsHolder.Gid)
	}
	printWarningForRoot(uid)

	// Create a file system server.
	serverCfg := &filesystem.ServerConfig{
		Uid:       uid,
		Gid:       gid,
		FilePerms: mountArgsHolder.FileMode,
		DirPerms:  mountArgsHolder.DirMode,
	}

	server, err := filesystem.NewServer(serverCfg)
	if err != nil {
		err = errors.Errorf("filesystem.NewServer: %v", err)
		daemonize.SignalOutcome(err)
		return err
	}

	// Mount the file system.
	console.Println("Mounting file system...")

	var mfs *fuse.MountedFileSystem

	options := make(map[string]string)
	if runtime.GOOS == `darwin` {
		options["daemon_timeout"] = "600" // 10 minutes for FS operation timeout.
	}

	mountCfg := &fuse.MountConfig{
		FSName:                  AppName,
		VolumeName:              AppName,
		Options:                 options,
		DisableWritebackCaching: true,
		ErrorLogger:             logger.NewLogger("fuse: "),
	}

	if globalDebug || mountArgsHolder.DebugFuse {
		mountCfg.DebugLogger = logger.NewLogger("fuse_debug: ")
	}

	mfs, err = fuse.Mount(mountArgsHolder.MountPoint, server, mountCfg)
	if err != nil {
		err = errors.Errorf("Failed to mount file system: %v", err)
		daemonize.SignalOutcome(err)
		return err
	}

	// Let the user unmount with Ctrl-C (SIGINT).
	trapCh := signalTrap(os.Interrupt, syscall.SIGINT)
	go func() {
		select {
		case sig := <-trapCh:
			log.Printf("INFO Received %s, attempting to unmount...", sig.String())

			err := fuse.Unmount(mountArgsHolder.MountPoint)
			if err != nil {
				log.Printf("ERROR Failed to unmount file system in response to %s. Error: %v", sig.String(), err)
			} else {
				log.Printf("INFO Successfully unmounted in response to %s.", sig.String())
				return
			}
		}
	}()

	console.Println("File system mounted successfully.")

	daemonize.SignalOutcome(nil)

	// Wait for the file system to be unmounted.
	err = mfs.Join(context.Background())
	if err != nil {
		err = errors.Errorf("MountedFileSystem.Join: %v", err)
		daemonize.SignalOutcome(err)
		return err
	}

	return nil
}

func daemonMount() error {

	// Print a warning if we run process as root except when in docker container.
	uid, _, err := user.MyUserAndGroup()
	if err != nil {
		return err
	}
	if mountArgsHolder.Uid > -1 {
		uid = uint32(mountArgsHolder.Uid)
	}
	printWarningForRoot(uid)

	// Find the executable.
	path, err := osext.Executable()
	if err != nil {
		return errors.Errorf("osext.Executable: %v", err)
	}

	// Set up arguments.
	args := os.Args[1:]

	// Convert all relative file paths to absolute.
	//
	// This is important when daemonizing, since the daemon will change its
	// working directory before running this code again.

	flagsWithPaths := []string{
		"--log-file",
	}

	for i, arg := range args {
		for _, flagsWithPath := range flagsWithPaths {
			if arg == flagsWithPath {
				pos := i + 1
				path := args[pos]
				absPath, err := filepath.Abs(path)
				if err != nil {
					return errors.Errorf("Could not convert filepath: %v", err)
				}
				args[pos] = absPath
			}
		}
	}

	// Set up environment variables.
	env := []string{}

	envVars := []string{
		"PATH", // Pass along PATH so that the daemon can find fusermount on Linux.
		"HOME",
		logger.LogLevelEnvVar,
		logger.LogFileEnvVar,
	}

	if isInDocker() {
		envVars = append(envVars, "DOCKERIMAGE")
	}

	// Copy environment variables.
	for _, key := range envVars {
		envVarValue, ok := os.LookupEnv(key)
		if ok {
			env = append(env, fmt.Sprintf("%s=%s", key, envVarValue))
		}
	}

	// Run.
	err = daemonize.Run(path, args, env, os.Stdout)
	if err != nil {
		return err
	}

	return nil
}
