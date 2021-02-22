package server

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func initAbout(e *echo.Echo) {
	e.GET("/about", aboutPage)
}

func aboutPage(c echo.Context) error {
	return c.Render(http.StatusOK, "aboutPage", nil)
}
