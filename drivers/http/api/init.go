package api

import (
	"context"

	"github.com/labstack/echo/v4"

	core "goodin/internal"
)

func InitRoutes(e *echo.Echo, app core.App) {
	ctx := context.Background()

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

	bindOrderApi(prefix, ctx, app)

}
