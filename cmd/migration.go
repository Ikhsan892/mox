package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"

	migrationinstance "github.com/golang-migrate/migrate/v4"
	app "goodin/internal"
	core "goodin/internal"
	"github.com/spf13/cobra"
)

func NewMigration(core core.App) *cobra.Command {
	var configPath string

	command := &cobra.Command{
		Use:   "migration",
		Short: "Run Migration",
		Run: func(cmd *cobra.Command, args []string) {
			core.OnAfterApplicationBootstrapped().Execute(app.AfterApplicationBootstrapped{App: core, ConfigPath: configPath})

			// core.Data().Get("sql", "gorm").(*gorm.DB).Exec("DELETE FROM schema_migrations")
			dir, err := os.Getwd()
			if err != nil {
				core.Logger().Error(err.Error(), slog.Any("context", "os.GetWd"))
			}

			// Replace backslashes with forward slashes
			normalizedDir := strings.ReplaceAll(dir, "\\", "/")
			migrationPath := fmt.Sprintf("%s/migrations", normalizedDir)
			core.Logger().Debug(migrationPath)

			defer func() {
				os.Exit(0)
			}()

			core.Logger().Info("Running Migration", slog.Any("args", args))

			migrate := core.Migration("./migrations")
			if args[0] == "up" {
				if len(args) > 1 {
					if args[1] != "" {
						steps, err := strconv.Atoi(args[1])
						if err != nil {
							core.Logger().Error(err.Error())
							return
						}
						if err := migrate.Steps(steps); err != nil {
							core.Logger().Error(err.Error())
							return
						}
					}
				} else {
					err := migrate.Up()
					if err != nil {
						if err != migrationinstance.ErrNoChange {
							core.Logger().Error(err.Error())
							return
						}
					}
				}
				core.Logger().Info("Migration Done")
			}

			if args[0] == "down" {
				err := migrate.Down()
				if err != nil {
					core.Logger().Error(err.Error())
					return
				}
			}
		},
	}

	command.Flags().StringVarP(&configPath, "config", "c", "", "Configuration file location")

	return command
}
