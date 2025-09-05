package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/PlakarKorp/plakar/restful/models"
)

// LocateFiles searches for files matching patterns across snapshots
func (h *Handler) LocateFiles(c echo.Context) error {
	ctx := c.Request().Context()

	patterns := c.QueryParams()["patterns"]
	if len(patterns) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "patterns parameter is required")
	}

	snapshot := c.QueryParam("snapshot")
	filters := c.QueryParam("filters")

	limit, err := parseUint32(c.QueryParam("limit"), 100)
	if err != nil || limit > 1000 {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid limit parameter (max 1000)")
	}

	req := models.LocateFilesRequest{
		Patterns: patterns,
		Snapshot: snapshot,
		Filters:  filters,
		Limit:    limit,
	}

	result, err := h.storage.LocateFiles(ctx, &req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to locate files: %v", err))
	}

	return c.JSON(http.StatusOK, result)
}