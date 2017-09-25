package cmd

import (
	log "github.com/Sirupsen/logrus"
	"github.com/markoradinovic/networksd/service"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

var cfgFile string
var debug bool
var cfg service.Conf
var unixSocket string
var tcpPort int

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "networksd",
	Short: "Docker Networks Utility",
	Long:  `Create Docker networks with predefined IP range.`,
	Run: func(cmd *cobra.Command, args []string) {
		service.StartDaemon(cfg)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		log.Errorln(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/networksd.yaml and networksd current folder)")
	RootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "enable debug log")
	RootCmd.PersistentFlags().StringVarP(&unixSocket, "unix-socket", "u", "networksd.sock", "Unix socket (default ./networksd.sock)")
	RootCmd.PersistentFlags().IntVarP(&tcpPort, "port", "p", 4444, "HPP interface listening port (default: 4444)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			log.Errorln(err)
			os.Exit(1)
		}

		// Search config in home directory with name "networksd" (without extension).
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigName("networksd")
	}

	viper.AutomaticEnv() // read in environment variables that match

	//bind to debug flag
	viper.BindPFlag("debug", RootCmd.Flags().Lookup("debug"))
	viper.BindPFlag("unix-socket", RootCmd.Flags().Lookup("unix-socket"))
	viper.BindPFlag("port", RootCmd.Flags().Lookup("port"))

	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		log.Panicf("Fatal error config file: %s \n", err)
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("Cannot unmarshal config: %s", err)
	}
}
