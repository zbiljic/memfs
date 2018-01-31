package cmd

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var unmountCmd = &cobra.Command{
	Use:   "unmount",
	Short: "Unmount the file system",
	Long:  `Unmount the file system.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return checkUnmountSyntax(cmd, args)
	},
	Run: func(cmd *cobra.Command, args []string) {
		defer timeTrack(time.Now(), "unmount")
		unmountMain()
	},
}

var unmountFlagsBind = []string{
	"force",
}

func init() {
	// add 'unmount' command to root command
	rootCmd.AddCommand(unmountCmd)
	initFlags(unmountFlags, unmountFlagsBind, unmountCmd)
}

// unmountArgs represents all possible arguments that can be specified while
// executing 'unmount' command.
type unmountArgs struct {
	MountPoint string
	Force      bool
}

var unmountArgsHolder = &unmountArgs{}

func populateArgsHolderUnmount(args []string) {
	// from command arguments
	unmountArgsHolder.MountPoint = valueOrEmptyString(args, 0)

	// from config
	argsSection := configSections["args"]

	unmountArgsHolder.Force = viper.GetBool(argsSection("force"))

}

func checkUnmountSyntax(cmd *cobra.Command, args []string) error {

	populateArgsHolderUnmount(args)

	if unmountArgsHolder.MountPoint == "" {
		fatalIf(errDummy(),
			"Mount point not provided.")
	}

	return nil
}

func unmountMain() error {
	err := unmount(
		unmountArgsHolder.MountPoint,
		unmountArgsHolder.Force,
	)

	errorIf(errors.WithStack(err), "Unmount failed:")

	return err
}

func unmount(mountPoint string, force bool) error {
	var err error

	// Linux flag for forcint unmount, according to Linux's fs.h http://tinyurl.com/hph4zpt
	const mntForceFlagLinux = 0x00000001
	// macOS flag for forcing unmount according to Darwin's mount.h http://tinyurl.com/zblb4hl
	const mntForceFlagMacOs = 0x00080000

	unmountFlags := 0

	if force {
		if runtime.GOOS == `linux` {
			unmountFlags = mntForceFlagLinux
		}
		if runtime.GOOS == `darwin` {
			unmountFlags = mntForceFlagMacOs
		}
		// We will not try to force umount for other systems.
	}

	if runtime.GOOS == `linux` {
		// Linux umount sequence

		// If it's forced unmount, first try system's unmount with 'force' option.
		if force {
			err = syscall.Unmount(mountPoint, unmountFlags)
			if err == nil {
				// Done! This unmounted the file system.
				return nil
			}
		}

		// Run 'fusermount -u' by defaul, or after forces system umount.
		fusermount := exec.Command("fusermount", "-u", mountPoint)
		var output []byte
		output, err = fusermount.CombinedOutput()
		if err != nil {
			if len(output) > 0 {
				err = fmt.Errorf(strings.TrimRight(string(output), "\n"))
			}
		}
	} else {
		// On macOS we will simply run system's unmount
		err = syscall.Unmount(mountPoint, unmountFlags)
	}

	// Error should be syscal.Errno. If so, we will try to give more descriptive
	// error message for commonly expected errors on Unmount.
	errno, ok := err.(syscall.Errno)
	if ok {
		switch errno {
		case syscall.EPERM:
			err = errUnmountNotPermitted(mountPoint)
		case syscall.ENOENT:
			err = errMountPointNotExist(mountPoint)
		case syscall.EBUSY:
			err = errUnmountBusyFilesystem(mountPoint)
		case syscall.EINVAL:
			err = errUnmountNotPossible(mountPoint)
		default:
			// We do not have a replacement, using the error as is.
		}
	}

	return err
}
