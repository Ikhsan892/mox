package cmd

import (
	"fmt"
	"os"

	core "mox/internal"
	"github.com/spf13/cobra"
)

func newVersionCmd(app core.App) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Check Version",
		RunE: func(cmd *cobra.Command, args []string) error {

			defer os.Exit(0)
			app.Logger().Info(fmt.Sprintf("Version template %d", app.Config().App.Version))

			return nil
		},
	}
}
