package cmd

import (
	"log/slog"
	"os"

	nats_messaging "goodin/drivers/messaging/nats"
	"goodin/drivers/monitoring"
	core "goodin/internal"
	"github.com/spf13/cobra"
)

func NewMessageBrokerCommand(app core.App) *cobra.Command {
	var configPath string

	command := &cobra.Command{
		Use:   "message-broker",
		Short: "Start Message Broker",
		RunE: func(cmd *cobra.Command, args []string) error {
			app.OnAfterApplicationBootstrapped().Execute(core.AfterApplicationBootstrapped{App: app, ConfigPath: configPath})

			if app.Config().Monitoring.EnableTelemetry {
				if err := app.Driver().RunDriver(monitoring.NewOtel(app)); err != nil {
					os.Exit(1)
				}
			} else {
				app.Logger().Info("Telemetry is disabled")
			}

			if err := app.Driver().RunDriver(nats_messaging.NewNatsMessaging(app, true)); err != nil {
				app.Logger().Error("Cannot run driver message-broker", slog.Any("err", err.Error()))
			}

			return nil
		},
	}

	command.Flags().StringVarP(&configPath, "config", "c", "", "Configuration file location")

	return command
}
