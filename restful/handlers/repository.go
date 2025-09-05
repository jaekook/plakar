package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/PlakarKorp/plakar/restful/models"
)

// CreateRepository creates a new backup repository
func (h *Handler) CreateRepository(c echo.Context) error {
	var req models.CreateRepositoryRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Create repository using Plakar
	repoID, err := h.storage.CreateRepository(c.Request().Context(), &req)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return echo.NewHTTPError(http.StatusConflict, "Repository already exists")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to create repository: %v", err))
	}

	response := models.CreateRepositoryResponse{
		RepositoryID: repoID,
		Location:     req.Location,
	}

	return c.JSON(http.StatusCreated, response)
}

// GetRepositoryInfo returns repository information and statistics
func (h *Handler) GetRepositoryInfo(c echo.Context) error {
	ctx := c.Request().Context()

	info, err := h.storage.GetRepositoryInfo(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to get repository info: %v", err))
	}

	response := models.ItemWrapper{
		Item: info,
	}

	return c.JSON(http.StatusOK, response)
}

// ListSnapshots returns a paginated list of snapshots
func (h *Handler) ListSnapshots(c echo.Context) error {
	ctx := c.Request().Context()

	// Parse query parameters
	offset, err := parseUint32(c.QueryParam("offset"), 0)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid offset parameter")
	}

	limit, err := parseUint32(c.QueryParam("limit"), 50)
	if err != nil || limit > 50 {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid limit parameter (max 50)")
	}

	importer := c.QueryParam("importer")
	since := c.QueryParam("since")
	sort := c.QueryParam("sort")
	if sort == "" {
		sort = "Timestamp"
	}

	var sinceTime *time.Time
	if since != "" {
		t, err := time.Parse(time.RFC3339, since)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid since parameter format")
		}
		sinceTime = &t
	}

	req := models.ListSnapshotsRequest{
		Offset:   offset,
		Limit:    limit,
		Importer: importer,
		Since:    sinceTime,
		Sort:     sort,
	}

	snapshots, total, err := h.storage.ListSnapshots(ctx, &req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to list snapshots: %v", err))
	}

	response := models.ItemsWrapper{
		Total: total,
		Items: snapshots,
	}

	return c.JSON(http.StatusOK, response)
}

// LocatePathname finds a file across multiple snapshots
func (h *Handler) LocatePathname(c echo.Context) error {
	ctx := c.Request().Context()

	resource := c.QueryParam("resource")
	if resource == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "resource parameter is required")
	}

	offset, err := parseUint32(c.QueryParam("offset"), 0)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid offset parameter")
	}

	limit, err := parseUint32(c.QueryParam("limit"), 50)
	if err != nil || limit > 50 {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid limit parameter (max 50)")
	}

	req := models.LocatePathnameRequest{
		Resource:          resource,
		ImporterType:      c.QueryParam("importerType"),
		ImporterOrigin:    c.QueryParam("importerOrigin"),
		ImporterDirectory: c.QueryParam("importerDirectory"),
		Offset:            offset,
		Limit:             limit,
		Sort:              c.QueryParam("sort"),
	}

	locations, total, err := h.storage.LocatePathname(ctx, &req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to locate pathname: %v", err))
	}

	response := models.ItemsWrapper{
		Total: total,
		Items: locations,
	}

	return c.JSON(http.StatusOK, response)
}

// GetImporterTypes returns available importer types
func (h *Handler) GetImporterTypes(c echo.Context) error {
	ctx := c.Request().Context()

	types, err := h.storage.GetImporterTypes(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to get importer types: %v", err))
	}

	items := make([]models.ImporterType, len(types))
	for i, t := range types {
		items[i] = models.ImporterType{Name: t}
	}

	response := models.ItemsWrapper{
		Total: len(items),
		Items: items,
	}

	return c.JSON(http.StatusOK, response)
}

// GetRepositoryStates returns repository state objects
func (h *Handler) GetRepositoryStates(c echo.Context) error {
	ctx := c.Request().Context()

	states, err := h.storage.GetRepositoryStates(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to get repository states: %v", err))
	}

	response := models.ItemsWrapper{
		Total: len(states),
		Items: states,
	}

	return c.JSON(http.StatusOK, response)
}

// GetRepositoryState returns a specific repository state
func (h *Handler) GetRepositoryState(c echo.Context) error {
	ctx := c.Request().Context()
	stateID := c.Param("state")

	if len(stateID) != 64 {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid state ID format")
	}

	data, err := h.storage.GetRepositoryState(ctx, stateID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return echo.NewHTTPError(http.StatusNotFound, "State not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to get repository state: %v", err))
	}

	c.Response().Header().Set("Content-Type", "application/octet-stream")
	return c.Blob(http.StatusOK, "application/octet-stream", data)
}

// RunMaintenance performs repository maintenance operations
func (h *Handler) RunMaintenance(c echo.Context) error {
	ctx := c.Request().Context()

	var req models.MaintenanceRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if len(req.Operations) == 0 {
		req.Operations = []string{"cleanup"}
	}

	result, err := h.storage.RunMaintenance(ctx, &req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to run maintenance: %v", err))
	}

	return c.JSON(http.StatusOK, result)
}

// PruneRepository removes unreferenced data
func (h *Handler) PruneRepository(c echo.Context) error {
	ctx := c.Request().Context()

	var req models.PruneRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	result, err := h.storage.PruneRepository(ctx, &req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to prune repository: %v", err))
	}

	return c.JSON(http.StatusOK, result)
}

// SyncRepository synchronizes repositories
func (h *Handler) SyncRepository(c echo.Context) error {
	ctx := c.Request().Context()

	var req models.SyncRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	result, err := h.storage.SyncRepository(ctx, &req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to sync repository: %v", err))
	}

	return c.JSON(http.StatusOK, result)
}

// Helper function to parse uint32 parameters
func parseUint32(s string, defaultValue uint32) (uint32, error) {
	if s == "" {
		return defaultValue, nil
	}
	
	val, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0, err
	}
	
	return uint32(val), nil
}