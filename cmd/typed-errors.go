package cmd

import "github.com/pkg/errors"

var (
	errDummy = func() error {
		return errors.New("")
	}

	errInvalidCommandCall = func(cmdName string) error {
		return errors.Errorf("Run '%s help %s' for usage.", AppName, cmdName)
	}

	//////////////////////
	// Unmount
	//////////////////////

	errUnmountNotPermitted = func(mountPoint string) error {
		return errors.Errorf("Insufficient permissions to unmount specified mountpoint '%s'.", mountPoint)
	}

	errMountPointNotExist = func(mountPoint string) error {
		return errors.Errorf("Provided mountpoint '%s' does not exist.", mountPoint)
	}

	errUnmountBusyFilesystem = func(mountPoint string) error {
		return errors.Errorf("Unable to unmount specified mountpoint '%s' because it is still in use.", mountPoint)
	}

	errUnmountNotPossible = func(mountPoint string) error {
		return errors.Errorf("Provided mountpoint '%s' is not a file system and cannot be unmounted.", mountPoint)
	}
)
