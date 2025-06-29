package seeders

import (
	"log/slog"

	core "goodin/internal"
)

type SeederAction interface {
	Name() string
	Initialize() error
}

func InitSeeder(core core.App) {

	// Insert seeder struct here
	seeders := []SeederAction{}

	for _, seeder := range seeders {
		if err := seeder.Initialize(); err != nil {
			core.Logger().Error(err.Error(), slog.String("seeder", seeder.Name()))
		} else {
			core.Logger().Info("Success seeding data", slog.String("seeder", seeder.Name()))
		}
	}
}
