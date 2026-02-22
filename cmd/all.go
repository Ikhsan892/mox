package cmd

import (
	"os"

	"mox/drivers/http"
	nats_messaging "mox/drivers/messaging/nats"
	"mox/drivers/monitoring"
	core "mox/internal"
	"github.com/spf13/cobra"
)

func NewAllCommand(app core.App) *cobra.Command {
	var configPath string
	command := &cobra.Command{
		Use:   "all",
		Short: "Start Application",
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
				os.Exit(1)
			}

			if err := app.Driver().RunDriver(http.NewEcho(app)); err != nil {
				os.Exit(1)
			}

			return nil
		},
	}

	command.Flags().StringVarP(&configPath, "config", "c", "", "Configuration file location")

	return command
}
