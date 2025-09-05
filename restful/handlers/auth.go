package handlers

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/PlakarKorp/plakar/restful/models"
)

// LoginGitHub initiates GitHub OAuth login flow
func (h *Handler) LoginGitHub(c echo.Context) error {
	ctx := c.Request().Context()

	var req models.LoginGitHubRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	result, err := h.storage.LoginGitHub(ctx, &req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to initiate GitHub login: %v", err))
	}

	return c.JSON(http.StatusOK, result)
}

// LoginEmail initiates email-based login flow
func (h *Handler) LoginEmail(c echo.Context) error {
	ctx := c.Request().Context()

	var req models.LoginEmailRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	result, err := h.storage.LoginEmail(ctx, &req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to initiate email login: %v", err))
	}

	return c.JSON(http.StatusOK, result)
}

// Logout logs out the current user
func (h *Handler) Logout(c echo.Context) error {
	ctx := c.Request().Context()

	err := h.storage.Logout(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to logout: %v", err))
	}

	return c.NoContent(http.StatusOK)
}