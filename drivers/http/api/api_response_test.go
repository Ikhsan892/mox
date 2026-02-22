package api_test

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"mox/drivers/http/api"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestNewApiResponse(t *testing.T) {
	testTables :=
		[]struct {
			data    any
			message string
			wantErr bool
			code    int
		}{
			{
				data:    "success",
				message: "Return String",
				wantErr: false,
				code:    200,
			},
			{
				data:    12345567,
				message: "Return Integer",
				wantErr: false,
				code:    200,
			},
			{
				data:    "{\"browsers\":{\"firefox\":{\"name\":\"Firefox\",\"pref_url\":\"about:config\",\"releases\":{\"1\":{\"release_date\":\"2004-11-09\",\"status\":\"retired\",\"engine\":\"Gecko\",\"engine_version\":\"1.7\"}}}}}",
				message: "Return JSON",
				wantErr: false,
				code:    200,
			},
			{
				data:    "browsers\":{\"firefox\":{\"name\":\"Firefox\",\"pref_url\":\"about:config\",\"releases\":{\"1\":{\"release_date\":\"2004-11-09\",\"status\":\"retired\",\"engine\":\"Gecko\",\"engine_version\":\"1.7\"}}}}}",
				message: "Return Invalid JSON",
				wantErr: false,
				code:    200,
			},
		}

	for _, table := range testTables {
		t.Run(table.message, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			res := httptest.NewRecorder()

			var body []byte
			switch v := table.data.(type) {
			case string:
				body = []byte(v)
			case int:
				body = []byte(strconv.Itoa(v))
			}

			res.Write(body)
			res.WriteHeader(400)
			c := e.NewContext(req, res)

			err := api.NewApiResponse(table.data, table.code, c)
			if !table.wantErr {
				assert.Nil(t, err)
			} else {
				assert.Error(t, err)
			}
		})

	}

}
