package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const (
	VersionMajor = 1
	VersionMinor = 0
	VersionPatch = 0
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long:  "Semantic versioning in the form MAJOR.MINOR.PATCH - http://semver.org",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("v%d.%d.%d\n", VersionMajor, VersionMinor, VersionPatch)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
