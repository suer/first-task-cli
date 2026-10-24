package cmd

import (
	"fmt"
	"runtime/debug"

	"github.com/spf13/cobra"
)

var version = "dev"

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show the version",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(currentVersion())
		},
	}
}

func currentVersion() string {
	if version != "dev" {
		return version
	}
	// Fall back to the module version for `go install` builds.
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return version
}
