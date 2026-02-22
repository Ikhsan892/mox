package cmd

import (
	"mox/drivers/http"
	"mox/drivers/master"
	core "mox/internal"

	"github.com/spf13/cobra"
)

func NewMasterCommand(app core.App) *cobra.Command {
	var configPath string

	command := &cobra.Command{
		Use:   "master",
		Short: "Start Master Worker",
		RunE: func(cmd *cobra.Command, args []string) error {
			app.OnAfterApplicationBootstrapped().ExecuteWithExclude(core.AfterApplicationBootstrapped{App: app, ConfigPath: configPath}, []string{"b_bootstrap"})

			if err := app.Driver().RunDriver(master.NewMasterAdapter(cmd.Context(), app)); err != nil {
				return err
			}

			if err := app.Driver().RunDriver(http.NewEcho(app)); err != nil {
				return err
			}

			<-cmd.Context().Done()

			return nil
		},
	}

	command.Flags().StringVarP(&configPath, "config", "c", "", "Configuration file location")

	return command
}
