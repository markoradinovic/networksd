package cmd

import (
	"fmt"
	"github.com/markoradinovic/networksd/buildinfo"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of networksd",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Version: ", buildinfo.VERSION)
		fmt.Println("Commit: ", buildinfo.COMMIT)
		fmt.Println("Branch: ", buildinfo.BRANCH)
	},
}
