package cmd

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/pgillich/chat-bot/config"
	"github.com/pgillich/chat-bot/internal/queue"
	"github.com/pgillich/chat-bot/pkg/frontend"
)

// nolint:gochecknoglobals
var frontendCmd = &cobra.Command{
	Use:   "frontend",
	Short: "Frontend",
	Long:  `Start chat bot frontend service.`,
	Run: func(cmd *cobra.Command, args []string) {
		startFrontend()
	},
}

func init() { // nolint:gochecknoinits
	RootCmd.AddCommand(frontendCmd)

	registerStringOption(frontendCmd, config.OptChatPath, config.DefaultChatPath, "path to chat bot service")
}

func startFrontend() {
	dir, _ := os.Getwd() // nolint:errcheck
	logger.Infof("PWD %s", dir)

	publisher := &queue.RealRedisPublisher{
		Host:           viper.GetString(config.OptRedisHost),
		Key:            viper.GetString(config.OptRedisKey),
		User:           viper.GetString(config.OptRedisUser),
		RequestChannel: viper.GetString(config.OptRedisRequestChannel),
	}
	defer publisher.Close()

	idleConnsClosed := make(chan struct{})
	defer close(idleConnsClosed)

	server := &http.Server{
		Addr: viper.GetString(config.OptServiceHostPort),
		Handler: frontend.App(idleConnsClosed, publisher,
			viper.GetString(config.OptChatPath),
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
