package cmd

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/pgillich/chat-bot/config"
	"github.com/pgillich/chat-bot/internal/db"
	"github.com/pgillich/chat-bot/internal/queue"
	"github.com/pgillich/chat-bot/pkg/engine"
)

// nolint:gochecknoglobals
var engineCmd = &cobra.Command{
	Use:   "engine",
	Short: "Engine",
	Long:  `Start chat bot engine service.`,
	Run: func(cmd *cobra.Command, args []string) {
		startEngine()
	},
}

func init() { // nolint:gochecknoinits
	RootCmd.AddCommand(engineCmd)

	registerStringOption(engineCmd, config.OptClientEndpoint, config.DefaultClientEndpoint, "client endpoint")
	registerStringOption(engineCmd, config.OptRsaKey, config.DefaultRsaKey, "RSA key for JWT")

	registerStringOption(engineCmd, config.OptDbHost, config.DefaultDbHost, "DB host")
	registerStringOption(engineCmd, config.OptDbName, config.DefaultDbName, "DB name")
	registerStringOption(engineCmd, config.OptDbUser, config.DefaultDbUser, "DB user")
	registerStringOption(engineCmd, config.OptDbPassword, config.DefaultDbPassword, "DB password")
}

func startEngine() {
	dir, _ := os.Getwd() // nolint:errcheck
	logger.Infof("PWD %s", dir)

	subscriber := &queue.RealRedisSubscriber{
		Host:           viper.GetString(config.OptRedisHost),
		Key:            viper.GetString(config.OptRedisKey),
		User:           viper.GetString(config.OptRedisUser),
		RequestChannel: viper.GetString(config.OptRedisRequestChannel),
	}
	defer subscriber.Close()

	dbHandler := &db.RealDbHandler{
		Host:     viper.GetString(config.OptDbHost),
		Database: viper.GetString(config.OptDbName),
		User:     viper.GetString(config.OptDbUser),
		Password: viper.GetString(config.OptDbPassword),
	}
	defer dbHandler.Close()

	httpClient := &http.Client{
		Timeout: time.Second * 10,
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout: 5 * time.Second,
			}).Dial,
			TLSHandshakeTimeout: 5 * time.Second,
		},
	}

	idleConnsClosed := make(chan struct{})
	defer close(idleConnsClosed)

	server := &http.Server{
		Addr: viper.GetString(config.OptServiceHostPort),
		Handler: engine.App(idleConnsClosed, subscriber, dbHandler, httpClient,
			viper.GetString(config.OptRsaKey),
			viper.GetString(config.OptClientEndpoint),
			viper.GetString(config.OptLogLevel),
		),
	}

	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint

		logger.Info("Shutdown server...")
		if err := server.Shutdown(context.Background()); err != nil {
			logger.Infof("HTTP server Shutdown: %v", err)
		}
	}()

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		logger.Panic("HTTP server ListenAndServe", err)
	}

	logger.Info("App closing...")
}
