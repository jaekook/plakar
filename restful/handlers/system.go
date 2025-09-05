package handlers

import (
	"net/http"
	"runtime"

	"github.com/labstack/echo/v4"

	"github.com/PlakarKorp/plakar/restful/models"
)

// GetInfo returns API information
func (h *Handler) GetInfo(c echo.Context) error {
	ctx := c.Request().Context()

	info, err := h.storage.GetAPIInfo(ctx)
	if err != nil {
		// Return basic info even if storage fails
		info = &models.APIInfo{
			RepositoryID:  "",
			Authenticated: false,
			Version:       "1.0.2",
			Browsable:     false,
			DemoMode:      false,
		}
	}

	return c.JSON(http.StatusOK, info)
}

// ProxyRequest proxies requests to external services
func (h *Handler) ProxyRequest(c echo.Context) error {
	ctx := c.Request().Context()
	path := c.Param("*")

	result, err := h.storage.ProxyRequest(ctx, c.Request().Method, path, c.Request().Body)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// Copy headers and status from proxied response
	for key, values := range result.Headers {
		for _, value := range values {
			c.Response().Header().Add(key, value)
		}
	}

	return c.Blob(result.StatusCode, result.ContentType, result.Body)
}