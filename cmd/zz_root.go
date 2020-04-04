//	Package cmd is the CLI handler
package cmd

import (
	goflag "flag"
	"fmt"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/pgillich/chat-bot/config"
)

// RootCmd is the root command
var RootCmd = &cobra.Command{ // nolint:gochecknoglobals
	Use:   "chat-bot",
	Short: "Sample chat bot",
	Long:  `Sample chat bot`,
}

// nolint:gochecknoglobals
var logger *log.Logger

// Execute is the main function
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Printf("Runtime error: %s\n", err)
		os.Exit(1)
	}
}

func getEnvReplacer() *strings.Replacer {
	return strings.NewReplacer("-", "_", ".", "_")
}

func init() { // nolint:gochecknoinits
	logger = log.New()
	logger.SetReportCaller(true)

	cobra.OnInitialize(initConfig)

	cobra.OnInitialize()

	registerStringOption(RootCmd, config.OptLogLevel, config.DefaultLogLevel, "log level")

	registerStringOption(RootCmd, config.OptServiceHostPort, config.DefaultServiceHostPort, "host:port listening on")

	registerStringOption(RootCmd, config.OptRedisHost, config.DefaultRedisHost, "URL to Redis server")
	registerStringOption(RootCmd, config.OptRedisUser, config.DefaultRedisUser, "Redis user name")
	registerStringOption(RootCmd, config.OptRedisKey, config.DefaultRedisKey, "Redis queue key")
	registerStringOption(RootCmd, config.OptRedisRequestChannel, config.DefaultRedisRequestChannel,
		"Redis channel name for sending message to worker")

	goflag.CommandLine.Usage = func() {
		RootCmd.Usage() // nolint:gosec,errcheck
	}
	goflag.Parse()

	logLevel, err := log.ParseLevel(viper.GetString(config.OptLogLevel))
	if err != nil {
		logger.Panic("invalid log level, ", err)
	}

	logger.SetLevel(logLevel)
}

func initConfig() {
	viper.AutomaticEnv() // read in environment variables that match
	viper.SetEnvKeyReplacer(getEnvReplacer())
}

func registerStringOption(command *cobra.Command, name string, value string, usage string) {
	envName := getEnvReplacer().Replace(name)
	command.PersistentFlags().String(name, value, strings.ToUpper(envName)+", "+usage)
	viper.BindPFlag(name, command.PersistentFlags().Lookup(name)) // nolint:errcheck,gosec
}
