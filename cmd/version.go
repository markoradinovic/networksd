package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/markoradinovic/networksd/buildinfo"
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
