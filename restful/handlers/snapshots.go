package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/PlakarKorp/plakar/restful/models"
)

// CreateSnapshot creates a new backup snapshot
func (h *Handler) CreateSnapshot(c echo.Context) error {
	ctx := c.Request().Context()

	var req models.CreateSnapshotRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Set defaults
	if req.Concurrency == 0 {
		req.Concurrency = 4
	}

	result, err := h.storage.CreateSnapshot(ctx, &req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to create snapshot: %v", err))
	}

	return c.JSON(http.StatusCreated, result)
}

// GetSnapshotHeader returns snapshot header information
func (h *Handler) GetSnapshotHeader(c echo.Context) error {
	ctx := c.Request().Context()
	snapshotID := c.Param("snapshot")

	if len(snapshotID) != 64 {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid snapshot ID format")
	}

	header, err := h.storage.GetSnapshotHeader(ctx, snapshotID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return echo.NewHTTPError(http.StatusNotFound, "Snapshot not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to get snapshot header: %v", err))
	}

	response := models.ItemWrapper{
		Item: header,
	}

	return c.JSON(http.StatusOK, response)
}

// RestoreSnapshot restores files from a snapshot
func (h *Handler) RestoreSnapshot(c echo.Context) error {
	ctx := c.Request().Context()
	snapshotID := c.Param("snapshot")

	if len(snapshotID) != 64 {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid snapshot ID format")
	}

	var req models.RestoreSnapshotRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Set defaults
	if req.Concurrency == 0 {
		req.Concurrency = 4
	}

	req.SnapshotID = snapshotID

	result, err := h.storage.RestoreSnapshot(ctx, &req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return echo.NewHTTPError(http.StatusNotFound, "Snapshot not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to restore snapshot: %v", err))
	}

	return c.JSON(http.StatusOK, result)
}

// CheckSnapshot verifies snapshot integrity
func (h *Handler) CheckSnapshot(c echo.Context) error {
	ctx := c.Request().Context()
	snapshotID := c.Param("snapshot")

	if len(snapshotID) != 64 {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid snapshot ID format")
	}

	var req models.CheckSnapshotRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Set defaults
	if req.Concurrency == 0 {
		req.Concurrency = 4
	}

	req.SnapshotID = snapshotID

	result, err := h.storage.CheckSnapshot(ctx, &req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return echo.NewHTTPError(http.StatusNotFound, "Snapshot not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to check snapshot: %v", err))
	}

	return c.JSON(http.StatusOK, result)
}

// DiffSnapshots compares two snapshots
func (h *Handler) DiffSnapshots(c echo.Context) error {
	ctx := c.Request().Context()
	snapshotID := c.Param("snapshot")
	targetSnapshotID := c.Param("target_snapshot")

	if len(snapshotID) != 64 || len(targetSnapshotID) != 64 {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid snapshot ID format")
	}

	path := c.QueryParam("path")
	recursive := c.QueryParam("recursive") == "true"
	highlight := c.QueryParam("highlight") == "true"

	req := models.DiffSnapshotsRequest{
		SnapshotID:       snapshotID,
		TargetSnapshotID: targetSnapshotID,
		Path:             path,
		Recursive:        recursive,
		Highlight:        highlight,
	}

	result, err := h.storage.DiffSnapshots(ctx, &req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return echo.NewHTTPError(http.StatusNotFound, "Snapshot not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to diff snapshots: %v", err))
	}

	return c.JSON(http.StatusOK, result)
}

// MountSnapshot mounts a snapshot as a filesystem
func (h *Handler) MountSnapshot(c echo.Context) error {
	ctx := c.Request().Context()
	snapshotID := c.Param("snapshot")

	if len(snapshotID) != 64 {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid snapshot ID format")
	}

	var req models.MountSnapshotRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	req.SnapshotID = snapshotID

	result, err := h.storage.MountSnapshot(ctx, &req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return echo.NewHTTPError(http.StatusNotFound, "Snapshot not found")
		}
		if strings.Contains(err.Error(), "already in use") {
			return echo.NewHTTPError(http.StatusConflict, "Mountpoint already in use")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to mount snapshot: %v", err))
	}

	return c.JSON(http.StatusOK, result)
}

// UnmountSnapshot unmounts a snapshot filesystem
func (h *Handler) UnmountSnapshot(c echo.Context) error {
	ctx := c.Request().Context()
	snapshotID := c.Param("snapshot")

	if len(snapshotID) != 64 {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid snapshot ID format")
	}

	var req models.UnmountSnapshotRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	req.SnapshotID = snapshotID

	err := h.storage.UnmountSnapshot(ctx, &req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return echo.NewHTTPError(http.StatusNotFound, "Snapshot or mountpoint not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to unmount snapshot: %v", err))
	}

	return c.NoContent(http.StatusOK)
}

// RemoveSnapshots removes one or more snapshots
func (h *Handler) RemoveSnapshots(c echo.Context) error {
	ctx := c.Request().Context()

	var req models.RemoveSnapshotsRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	result, err := h.storage.RemoveSnapshots(ctx, &req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to remove snapshots: %v", err))
	}

	return c.JSON(http.StatusOK, result)
}

// BrowseVFS browses the virtual filesystem of a snapshot
func (h *Handler) BrowseVFS(c echo.Context) error {
	ctx := c.Request().Context()
	snapshotPath := c.Param("*")

	parts := strings.SplitN(snapshotPath, ":", 2)
	if len(parts) < 1 {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid snapshot path format")
	}

	snapshotID := parts[0]
	path := ""
	if len(parts) > 1 {
		path = parts[1]
	}

	if len(snapshotID) != 64 {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid snapshot ID format")
	}

	entry, err := h.storage.BrowseVFS(ctx, snapshotID, path)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return echo.NewHTTPError(http.StatusNotFound, "Path not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to browse VFS: %v", err))
	}

	return c.JSON(http.StatusOK, entry)
}

// ListVFSChildren lists directory children in VFS
func (h *Handler) ListVFSChildren(c echo.Context) error {
	ctx := c.Request().Context()
	snapshotPath := c.Param("*")

	parts := strings.SplitN(snapshotPath, ":", 2)
	if len(parts) < 1 {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid snapshot path format")
	}

	snapshotID := parts[0]
	path := ""
	if len(parts) > 1 {
		path = parts[1]
	}

	if len(snapshotID) != 64 {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid snapshot ID format")
	}

	offset, err := parseUint32(c.QueryParam("offset"), 0)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid offset parameter")
	}

	limit, err := parseUint32(c.QueryParam("limit"), 100)
	if err != nil || limit > 1000 {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid limit parameter (max 1000)")
	}

	sort := c.QueryParam("sort")
	if sort == "" {
		sort = "Name"
	}

	req := models.ListVFSChildrenRequest{
		SnapshotID: snapshotID,
		Path:       path,
		Offset:     offset,
		Limit:      limit,
		Sort:       sort,
	}

	result, err := h.storage.ListVFSChildren(ctx, &req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return echo.NewHTTPError(http.StatusNotFound, "Path not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to list VFS children: %v", err))
	}

	return c.JSON(http.StatusOK, result)
}

// SearchVFS searches files in a snapshot
func (h *Handler) SearchVFS(c echo.Context) error {
	ctx := c.Request().Context()
	snapshotPath := c.Param("*")

	parts := strings.SplitN(snapshotPath, ":", 2)
	if len(parts) < 1 {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid snapshot path format")
	}

	snapshotID := parts[0]
	path := ""
	if len(parts) > 1 {
		path = parts[1]
	}

	if len(snapshotID) != 64 {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid snapshot ID format")
	}

	pattern := c.QueryParam("pattern")
	if pattern == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "pattern parameter is required")
	}

	offset, err := parseUint32(c.QueryParam("offset"), 0)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid offset parameter")
	}

	limit, err := parseUint32(c.QueryParam("limit"), 100)
	if err != nil || limit > 1000 {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid limit parameter (max 1000)")
	}

	req := models.SearchVFSRequest{
		SnapshotID: snapshotID,
		Path:       path,
		Pattern:    pattern,
		Offset:     offset,
		Limit:      limit,
	}

	result, err := h.storage.SearchVFS(ctx, &req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return echo.NewHTTPError(http.StatusNotFound, "Snapshot not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to search VFS: %v", err))
	}

	return c.JSON(http.StatusOK, result)
}

// GetVFSChunks returns chunk information for a file
func (h *Handler) GetVFSChunks(c echo.Context) error {
	ctx := c.Request().Context()
	snapshotPath := c.Param("*")

	parts := strings.SplitN(snapshotPath, ":", 2)
	if len(parts) != 2 {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid snapshot path format")
	}

	snapshotID := parts[0]
	path := parts[1]

	if len(snapshotID) != 64 {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid snapshot ID format")
	}

	chunks, err := h.storage.GetVFSChunks(ctx, snapshotID, path)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return echo.NewHTTPError(http.StatusNotFound, "File not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to get VFS chunks: %v", err))
	}

	return c.JSON(http.StatusOK, chunks)
}

// GetVFSErrors returns errors for a snapshot path
func (h *Handler) GetVFSErrors(c echo.Context) error {
	ctx := c.Request().Context()
	snapshotPath := c.Param("*")

	parts := strings.SplitN(snapshotPath, ":", 2)
	if len(parts) < 1 {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid snapshot path format")
	}

	snapshotID := parts[0]
	path := ""
	if len(parts) > 1 {
		path = parts[1]
	}

	if len(snapshotID) != 64 {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid snapshot ID format")
	}

	errors, err := h.storage.GetVFSErrors(ctx, snapshotID, path)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return echo.NewHTTPError(http.StatusNotFound, "Snapshot not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to get VFS errors: %v", err))
	}

	return c.JSON(http.StatusOK, errors)
}

// CreateDownloadPackage creates a downloadable package
func (h *Handler) CreateDownloadPackage(c echo.Context) error {
	ctx := c.Request().Context()
	snapshotPath := c.Param("*")

	parts := strings.SplitN(snapshotPath, ":", 2)
	if len(parts) < 1 {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid snapshot path format")
	}

	snapshotID := parts[0]
	path := ""
	if len(parts) > 1 {
		path = parts[1]
	}

	if len(snapshotID) != 64 {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid snapshot ID format")
	}

	var req models.CreateDownloadPackageRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	req.SnapshotID = snapshotID
	req.Path = path

	// Set defaults
	if req.Format == "" {
		req.Format = "zip"
	}

	result, err := h.storage.CreateDownloadPackage(ctx, &req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return echo.NewHTTPError(http.StatusNotFound, "Snapshot not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to create download package: %v", err))
	}

	return c.JSON(http.StatusOK, result)
}

// GetSignedDownloadURL returns a signed URL for download
func (h *Handler) GetSignedDownloadURL(c echo.Context) error {
	ctx := c.Request().Context()
	downloadID := c.Param("id")

	if downloadID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Download ID is required")
	}

	url, err := h.storage.GetSignedDownloadURL(ctx, downloadID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return echo.NewHTTPError(http.StatusNotFound, "Download not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to get signed download URL: %v", err))
	}

	response := models.SignedURLResponse{
		URL: url,
	}

	return c.JSON(http.StatusOK, response)
}