package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/PlakarKorp/plakar/restful/models"
)

// InstallIntegration installs a new integration plugin
func (h *Handler) InstallIntegration(c echo.Context) error {
	ctx := c.Request().Context()

	var req models.InstallIntegrationRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	err := h.storage.InstallIntegration(ctx, &req)
	if err != nil {
		if strings.Contains(err.Error(), "invalid configuration") {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid integration configuration")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to install integration: %v", err))
	}

	return c.NoContent(http.StatusOK)
}

// UninstallIntegration removes an installed integration
func (h *Handler) UninstallIntegration(c echo.Context) error {
	ctx := c.Request().Context()
	integrationID := c.Param("id")

	if integrationID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Integration ID is required")
	}

	err := h.storage.UninstallIntegration(ctx, integrationID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return echo.NewHTTPError(http.StatusNotFound, "Integration not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to uninstall integration: %v", err))
	}

	return c.NoContent(http.StatusOK)
}