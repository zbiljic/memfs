package cmd

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func initFlags(flagSetFun map[string]func(*pflag.FlagSet), flags []string, cmd *cobra.Command) {
	for _, k := range flags {
		flagSetFun[k](cmd.Flags())
	}
	bindFlags(flags, cmd)
}

func bindFlags(flags []string, cmd *cobra.Command) {
	argsSection := configSections["args"]
	for _, k := range flags {
		viper.BindPFlag(argsSection(k), cmd.Flags().Lookup(k))
	}
}

// Place for all the flags for all commands

var globalFlags = map[string]func(*pflag.FlagSet){
	"quiet": func(flags *pflag.FlagSet) {
		flags.Bool("quiet", false, "Suppress chatty console output.")
	},
	"debug": func(flags *pflag.FlagSet) {
		flags.Bool("debug", false, "Run the command with debug information in the output.")
	},
	"log-file": func(flags *pflag.FlagSet) {
		flags.String("log-file", "", "File to which to send logs to.")
	},
	"pprof-addr": func(flags *pflag.FlagSet) {
		flags.String("pprof-addr", "", "pprof address to listen on, format: localhost:6060 or :6060.")
	},
}

func defineFlagsGlobal(flags *pflag.FlagSet) {
	for _, ffn := range globalFlags {
		ffn(flags)
	}
}

func bindFlagsGlobal(cmd *cobra.Command) {
	globalSection := configSections["global"]
	for k := range globalFlags {
		key := globalSection(k)
		log.Printf("TRACE Binding global flag: '%s' to '%s'", k, key)
		viper.BindPFlag(key, cmd.Flag(k))
	}
}

////////////////////////////////////////////////////////////////////////
// Mount flags
////////////////////////////////////////////////////////////////////////

var mountFlags = map[string]func(*pflag.FlagSet){
	"foreground": func(flags *pflag.FlagSet) {
		flags.Bool("foreground", false, "Stay in the foreground after mounting.")
	},
	"o": func(flags *pflag.FlagSet) {
		flags.StringSliceP("o", "o", []string{}, "Additional system-specific mount options. Be careful!")
	},
	"dir-mode": func(flags *pflag.FlagSet) {
		flags.String("dir-mode", "755", "Permissions bits for directories, in octal.")
	},
	"file-mode": func(flags *pflag.FlagSet) {
		flags.String("file-mode", "644", "Permission bits for files, in octal.")
	},
	"uid": func(flags *pflag.FlagSet) {
		flags.Int("uid", -1, "UID owner of all inodes.")
	},
	"gid": func(flags *pflag.FlagSet) {
		flags.Int("gid", -1, "GID owner of all inodes.")
	},
	"debug_fuse": func(flags *pflag.FlagSet) {
		flags.Bool("debug_fuse", false, "Enable fuse-related debugging output.")
	},
	"debug_invariants": func(flags *pflag.FlagSet) {
		flags.Bool("debug_invariants", false, "Panic when internal invariants are violated.")
	},
}

////////////////////////////////////////////////////////////////////////
// Unmount flags
////////////////////////////////////////////////////////////////////////

var unmountFlags = map[string]func(*pflag.FlagSet){
	"force": func(flags *pflag.FlagSet) {
		flags.Bool("force", false, "Unmount file system even if it is busy.")
	},
}
