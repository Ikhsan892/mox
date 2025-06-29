package api

import (
	"reflect"
	"time"

	"github.com/labstack/echo/v4"

	"goodin/gorm/models"
)

type ApiResponse struct {
	Path      string    `json:"path"`
	Time      time.Time `json:"time"`
	Data      any       `json:"data"`
	Code      int       `json:"code"`
	RequestId string    `json:"request_id"`
}

func NewApiResponse(data any, code int, c echo.Context) error {
	resp := ApiResponse{
		Path: c.Path(),
		Time: time.Now(),
		Data: data,
		Code: code,
	}

	return c.JSON(resp.Code, resp)
}

type ResponsePaginate struct {
	Key  string
	Meta *models.Paginate
	Data any
	Code int
}

func NewApiPaginationResponse(rt ResponsePaginate, c echo.Context) error {
	v := reflect.ValueOf(*rt.Meta)
	res := make(map[string]interface{})
	res[rt.Key] = rt.Data

	for i := 0; i < v.NumField(); i++ {
		res[v.Type().Field(i).Tag.Get("json")] = v.Field(i).Interface()
	}

	return c.JSON(rt.Code, res)
}
