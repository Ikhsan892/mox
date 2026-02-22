package cmd

import (
	"mox/drivers/http"
	core "mox/internal"
	"github.com/spf13/cobra"
)

func NewHttpCommand(app core.App) *cobra.Command {
	var configPath string

	command := &cobra.Command{
		Use:   "http",
		Short: "Start http Application",
		RunE: func(cmd *cobra.Command, args []string) error {
			app.OnAfterApplicationBootstrapped().Execute(core.AfterApplicationBootstrapped{App: app, ConfigPath: configPath})

			app.Driver().RunDriver(http.NewEcho(app))

			return nil
		},
	}

	command.Flags().StringVarP(&configPath, "config", "c", "", "Configuration file location")

	return command
}
