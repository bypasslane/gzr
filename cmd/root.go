package cmd

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/bypasslane/gzr/comms"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	// DefaultLogLevel maps to a valid value argument for logrus.ParseLevel
	DefaultLogLevel = "info"
)

var cfgFile string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "gzr",
	Short: "A toolkit for managing Kubernetes Deployments",
	Long:  `Create, interrogate, and annotate container-based Deployment resources`,
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		er(fmt.Sprintf(err.Error()))
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	registeredInterfaces["etcd"] = comms.NewEtcdStorage
	registeredInterfaces["bolt"] = comms.NewBoltStorage
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	viper.SetConfigName(".gzr")  // name of config file (without extension)
	viper.AddConfigPath("$HOME") // adding home directory as first search path
	viper.AutomaticEnv()         // read in environment variables that match
	flag.String("log-level", DefaultLogLevel, "the log level to use")
	viper.BindPFlag("log-level", flag.Lookup("log-level"))

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("Error using config file:", viper.ConfigFileUsed())
	}

	setupLogging()
}

func setupLogging() {
	lvl, err := log.ParseLevel(viper.GetString("log-level"))
	if err != nil {
		er("Invalid logging level specified")
	}
	log.SetLevel(lvl)
}
