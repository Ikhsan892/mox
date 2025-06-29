package http

import (
	"errors"
	"log"
	"net/http"

	"goodin/drivers/http/api"

	"github.com/labstack/echo/v4"

	core "goodin/internal"
)

func GetRequestId(c echo.Context) string {
	requestId := c.Request().Header.Get(echo.HeaderXRequestID)
	if requestId == "" {
		requestId = c.Response().Header().Get(echo.HeaderXRequestID)
	}
	return requestId
}

func HttpErrorHandler(app core.App) echo.HTTPErrorHandler {
	msg := "Oops something went wrong, Contact Developer for the issue"
	return func(err error, c echo.Context) {
		report, ok := err.(*echo.HTTPError)

		var errorApp *api.ApiError
		if !ok {
			if errors.As(err, &errorApp) {
				errorApp.RequestId = GetRequestId(c)

				if app.Config().App.Mode == "production" && errorApp.Code == http.StatusInternalServerError {
					errorApp.Message = msg
				}

				report = echo.NewHTTPError(errorApp.Code, errorApp.Message)
			} else {
				report = echo.NewHTTPError(http.StatusInternalServerError, err.Error())
			}
		}

		err2 := c.JSON(report.Code, errorApp)
		if err2 != nil {
			log.Fatal("Cannot return anything ", err2.Error())
		}
	}
}
