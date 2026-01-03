package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/nerdneilsfield/RSSHub-Gateway/internal/logging"
	"github.com/nerdneilsfield/RSSHub-Gateway/internal/metrics"
	"github.com/nerdneilsfield/RSSHub-Gateway/internal/proxy"
	"github.com/nerdneilsfield/RSSHub-Gateway/internal/runtime"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func newServeCmd() *cobra.Command {
	var configPath string

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the rsshub-gateway server",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger, err := logging.NewLogger()
			if err != nil {
				return fmt.Errorf("init logger: %w", err)
			}
			defer func() { _ = logger.Sync() }()

			m := metrics.New()
			mgr, err := runtime.NewManager(configPath, m, logger)
			if err != nil {
				return err
			}

			app := fiber.New()
			proxyHandler := proxy.New(mgr, m, logger)
			app.All("/*", proxyHandler.Serve)

			signalChan := make(chan os.Signal, 1)
			signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
			go func() {
				for sig := range signalChan {
					switch sig {
					case syscall.SIGHUP:
						if err := mgr.Reload(); err != nil {
							logger.Error("reload failed", zap.Error(err))
						}
					default:
						_ = app.Shutdown()
						return
					}
				}
			}()

			listenAddr := mgr.Get().Server.Listen
			logger.Info("server start", zap.String("listen", listenAddr))
			return app.Listen(listenAddr)
		},
	}

	cmd.Flags().StringVarP(&configPath, "config", "c", "config.yaml", "Path to config file")
	return cmd
}
