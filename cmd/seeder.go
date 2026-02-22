package cmd

import (
	"github.com/spf13/cobra"

	core "mox/internal"

	"mox/seeders"
)

func NewSeeder(core core.App) *cobra.Command {
	return &cobra.Command{
		Use:   "seeders",
		Short: "Run Seeder",
		Run: func(cmd *cobra.Command, args []string) {
			seeders.InitSeeder(core)
		},
	}
}
