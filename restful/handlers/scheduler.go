package handlers

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

// StartScheduler starts the backup scheduler service
func (h *Handler) StartScheduler(c echo.Context) error {
	ctx := c.Request().Context()

	err := h.storage.StartScheduler(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to start scheduler: %v", err))
	}

	return c.NoContent(http.StatusOK)
}

// StopScheduler stops the backup scheduler service
func (h *Handler) StopScheduler(c echo.Context) error {
	ctx := c.Request().Context()

	err := h.storage.StopScheduler(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to stop scheduler: %v", err))
	}

	return c.NoContent(http.StatusOK)
}

// GetSchedulerStatus returns the current scheduler status
func (h *Handler) GetSchedulerStatus(c echo.Context) error {
	ctx := c.Request().Context()

	status, err := h.storage.GetSchedulerStatus(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to get scheduler status: %v", err))
	}

	return c.JSON(http.StatusOK, status)
}