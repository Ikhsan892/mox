package api

import (
	"strconv"

	"github.com/labstack/echo/v4"

	"mox/drivers/master"
	core "mox/internal"
	"mox/pkg/driver/v2"
	"mox/use_cases/mastercore"
)

func InitRoutes(e *echo.Echo, app core.App) {
	// ctx := context.Background()

	prefix := e.Group("/api")

	//prefix.Use(echojwt.WithConfig(echojwt.Config{
	//	SigningKey:  []byte(app.Settings().App.AppSecret),
	//	TokenLookup: "header:Authorization,cookie:Token",
	//	ErrorHandler: func(c echo.Context, err error) error {
	//		return NewUnauthorizedError("", err)
	//	},
	//}))
	//{

	e.Static("/excel", "/writable").Name = "STATIC"

	prefix.GET("/health", func(c echo.Context) error {
		master, err := driver.Get[*mastercore.Master](app.Driver(), master.MasterAdapterName)
		if err != nil {
			return c.String(200, "NOT HEALTHY")
		}

		return c.String(200, master.Orchestrator.CheckHealth())
	})

	prefix.GET("/workers", func(c echo.Context) error {
		master, err := driver.Get[*mastercore.Master](app.Driver(), master.MasterAdapterName)
		if err != nil {
			return c.String(400, "WORKER NOT READY")
		}

		return c.String(200, strconv.Itoa(int(master.Orchestrator.GetTotalWorkers())))
	})
}
