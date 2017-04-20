package cmd

import (
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/bypasslane/gzr/comms"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	// DefaultLogLevel maps to a valid value argument for logrus.ParseLevel
	DefaultLogLevel  = "info"
	DefaultLogFormat = "text"
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
		erWithDetails(err, "run time error")
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	RootCmd.PersistentFlags().StringVar(&logLevel, "log-level", DefaultLogLevel, "the log level to use. (debug, info, warn(ing), error, fatal, panic)")
	RootCmd.PersistentFlags().StringVar(&logFormat, "log-format", DefaultLogFormat, "The log formatter to use - (json | text)")
	viper.BindPFlag("log-level", RootCmd.PersistentFlags().Lookup("log-level"))
	viper.BindPFlag("log-format", RootCmd.PersistentFlags().Lookup("log-format"))
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


	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("Error using config file:", viper.ConfigFileUsed())
	}
	setupLogging()
}

func parseLogFormat(value string) (log.Formatter, error) {
	val := strings.ToLower(value)
	switch val {
	case "json":
		return &log.JSONFormatter{}, nil
	case "text":
		return &log.TextFormatter{FullTimestamp: true}, nil
	default:
		return nil, fmt.Errorf("Not a valid formatter : %q", value)
	}
}

func setupLogging() {
	lvl, err := log.ParseLevel(logLevel)
	if err != nil {
		erWithDetails(err, "Invalid logging level specified")
	}
	log.SetLevel(lvl)

	formatter, err := parseLogFormat(logFormat)
	if err != nil {
		erWithDetails(err, "Invalid formatter specified")
	}
	log.SetFormatter(formatter)
}
