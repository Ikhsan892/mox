package cmd

import (
	"mox/drivers/daemon"
	"mox/drivers/worker"
	core "mox/internal"

	"github.com/spf13/cobra"
)

func NewWorkerCommand(app core.App) *cobra.Command {
	var configPath string

	command := &cobra.Command{
		Use:   "worker",
		Short: "Start Worker",
		RunE: func(cmd *cobra.Command, args []string) error {
			app.OnAfterApplicationBootstrapped().ExecuteWithExclude(core.AfterApplicationBootstrapped{App: app, ConfigPath: configPath}, []string{"b_bootstrap"})

			if err := app.Driver().RunDriver(worker.NewWorkerAdapter(app)); err != nil {
				return err
			}

			if err := app.Driver().RunDriver(daemon.NewDaemonAdapter(app)); err != nil {
				return err
			}

			<-cmd.Context().Done()

			return nil
		},
	}

	command.Flags().StringVarP(&configPath, "config", "c", "", "Configuration file location")

	return command
}
