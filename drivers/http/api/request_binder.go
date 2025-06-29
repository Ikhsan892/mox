package api

import (
	"reflect"
	"strings"

	"github.com/labstack/echo/v4"
)

func requestBinder(c echo.Context, param interface{}) error {
	if err := c.Bind(param); err != nil {
		return err
	}

	// Use reflection to dynamically process fields with special handling
	v := reflect.ValueOf(param).Elem()
	t := v.Type()

	queryParams := c.QueryParams()
	for j := 0; j < t.NumField(); j++ {
		field := t.Field(j)

		// Handle dynamic sorter parameters
		if field.Type.Kind() == reflect.Map && field.Name == "Sorter" {
			fieldMap := reflect.MakeMap(field.Type)
			fieldValue := v.Field(j)
			if fieldValue.CanSet() {
				for key, values := range queryParams {
					if strings.HasPrefix(key, "sorter[") && strings.HasSuffix(key, "]") {
						sortKey := key[7 : len(key)-1]

						fieldMap.SetMapIndex(reflect.ValueOf(sortKey), reflect.ValueOf(values[0]))
					}
				}
				fieldValue.Set(fieldMap)
			}
		}

		// Handle dynamic search parameters
		if field.Type.Kind() == reflect.Map && field.Name == "Search" {
			fieldMap := reflect.MakeMap(field.Type)
			fieldValue := v.Field(j)
			if fieldValue.CanSet() {
				for key, values := range queryParams {
					if strings.HasPrefix(key, "search[") && strings.HasSuffix(key, "]") {
						sortKey := key[7 : len(key)-1]

						fieldMap.SetMapIndex(reflect.ValueOf(sortKey), reflect.ValueOf(values[0]))
					}

				}
				fieldValue.Set(fieldMap)
			}
		}
	}

	return nil

}
