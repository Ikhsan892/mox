package http

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"mox/drivers/http/api"
	core "mox/internal"
	"mox/pkg/driver"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"

	"github.com/fatih/color"

	_ "mox/docs"
)

var _ (driver.IDriver) = (*EchoWebAdapter)(nil)

const ECHO_ADAPTER = "ECHO_ADAPTER"

type EchoWebAdapter struct {
	ec  *echo.Echo
	app core.App
}

// @title        TiulTemplate Documentation
// @version      1.0
// @description  Application Service for Microservices Architecture
// @contact.name Muhammad Fatihul Ikhsan
// @license.name Private License
// @schemes      http
// @BasePath  	/api
func NewEcho(app core.App) *EchoWebAdapter {
	e := echo.New()

	e.Debug = false
	e.HideBanner = true
	e.HidePort = true

	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:       true,
		LogStatus:    true,
		LogMethod:    true,
		LogHost:      true,
		LogLatency:   true,
		LogError:     true,
		LogRequestID: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			if v.Status > 100 && v.Status < 400 {
				app.Logger().Info(v.Host,
					slog.Any("method", v.Method),
					slog.Any("status", v.Status),
					slog.Any("latency", v.Latency.String()),
					slog.Any("uri", v.URI),
				)
			} else if v.Status >= 400 && v.Status < 500 {
				app.Logger().Warn(v.Error.Error(),
					slog.Any("method", v.Method),
					slog.Any("status", v.Status),
					slog.Any("latency", v.Latency.String()),
					slog.Any("uri", v.URI),
				)
			} else if v.Status >= 500 {
				app.Logger().Error(v.Error.Error(),
					slog.Any("method", v.Method),
					slog.Any("status", v.Status),
					slog.Any("latency", v.Latency.String()),
					slog.Any("uri", v.URI),
				)
			} else {
				app.Logger().Error(v.Error.Error(),
					slog.Any("method", v.Method),
					slog.Any("status", v.Status),
					slog.Any("latency", v.Latency.String()),
					slog.Any("uri", v.URI),
				)
			}

			return nil
		},
	}))
	e.HTTPErrorHandler = HttpErrorHandler(app)
	e.Use(middleware.RequestID())
	// e.Use(middleware.Gzip())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:  []string{"*"},
		ExposeHeaders: []string{"Content-Disposition"},
		AllowHeaders:  []string{echo.HeaderOrigin, echo.HeaderAuthorization, echo.HeaderContentType, "module", "Content-Range", "Accept-Language"},
	}))

	if app.IsDev() {
		e.GET("/swagger/*", echoSwagger.WrapHandler).Name = "SWAGGER"
	}
	api.InitRoutes(e, app)

	return &EchoWebAdapter{ec: e, app: app}
}

// Instance implements driver.IDriver.
func (e *EchoWebAdapter) Instance() interface{} {
	return e.ec
}

func (e *EchoWebAdapter) Name() string {
	return ECHO_ADAPTER
}

func (e *EchoWebAdapter) Close() error {
	e.app.Logger().Info("echo web adapter closed")
	return e.ec.Close()
}

func (e *EchoWebAdapter) Init() error {
	// s := http.Server{
	// 	Addr:    fmt.Sprintf(":%d", e.app.Config().App.WebServerPort),
	// 	Handler: e.ec, // set Echo as handler
	// 	//ReadTimeout: 30 * time.Second, // use custom timeouts
	// }

	schema := "http"

	bold := color.New(color.Bold).Add(color.FgGreen).SprintfFunc()

	go func(e *echo.Echo, app core.App) {
		if err := e.Start(fmt.Sprintf(":%d", app.Config().Api.Port)); err != http.ErrServerClosed {
			app.Logger().Error("err serve echo", slog.String("err", err.Error()))
			os.Exit(1)
		}
	}(e.ec, e.app)

	cyan := color.New(color.FgCyan).SprintfFunc()

	e.app.Logger().Info(bold("> REST API Server started at: %s", cyan("%s://localhost:%d", schema, e.app.Config().Api.Port)))

	return nil
}
