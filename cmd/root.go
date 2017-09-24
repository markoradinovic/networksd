package cmd

import (
	log "github.com/Sirupsen/logrus"
	"github.com/markoradinovic/networksd/service"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

var cfgFile string
var debug bool
var cfg service.Conf

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "networksd",
	Short: "Docker Networks Utility",
	Long:  `Create Docker networks with predefined IP range.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
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

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	//RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
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

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Info("Using config file: ", viper.ConfigFileUsed())
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("Cannot unmarshal config: %s", err)
	}
	log.Info(cfg)
}
