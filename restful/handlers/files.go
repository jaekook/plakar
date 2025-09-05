package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/PlakarKorp/plakar/restful/models"
)

// ReadFile reads file content from a snapshot
func (h *Handler) ReadFile(c echo.Context) error {
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

	download := c.QueryParam("download") == "true"
	render := c.QueryParam("render")
	if render == "" {
		render = "auto"
	}

	// Validate render parameter
	validRenders := []string{"auto", "text", "code", "text_styled"}
	isValidRender := false
	for _, v := range validRenders {
		if render == v {
			isValidRender = true
			break
		}
	}
	if !isValidRender {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid render parameter")
	}

	req := models.ReadFileRequest{
		SnapshotID: snapshotID,
		Path:       path,
		Download:   download,
		Render:     render,
	}

	content, contentType, err := h.storage.ReadFile(ctx, &req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return echo.NewHTTPError(http.StatusNotFound, "File not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to read file: %v", err))
	}

	// Set appropriate headers
	if download {
		filename := strings.Split(path, "/")
		if len(filename) > 0 {
			c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename[len(filename)-1]))
		}
	}

	return c.Blob(http.StatusOK, contentType, content)
}

// CreateSignedURL creates a signed URL for file access
func (h *Handler) CreateSignedURL(c echo.Context) error {
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

	req := models.CreateSignedURLRequest{
		SnapshotID: snapshotID,
		Path:       path,
	}

	result, err := h.storage.CreateSignedURL(ctx, &req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return echo.NewHTTPError(http.StatusNotFound, "File not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to create signed URL: %v", err))
	}

	return c.JSON(http.StatusOK, result)
}

// GetFileContent gets file content with processing options
func (h *Handler) GetFileContent(c echo.Context) error {
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

	decompress := c.QueryParam("decompress") == "true"
	highlight := c.QueryParam("highlight") == "true"

	req := models.GetFileContentRequest{
		SnapshotID:  snapshotID,
		Path:        path,
		Decompress:  decompress,
		Highlight:   highlight,
	}

	content, contentType, err := h.storage.GetFileContent(ctx, &req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return echo.NewHTTPError(http.StatusNotFound, "File not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to get file content: %v", err))
	}

	if highlight && contentType == "text/html" {
		return c.HTML(http.StatusOK, string(content))
	}

	return c.Blob(http.StatusOK, contentType, content)
}

// GetFileDigest calculates and returns file digest
func (h *Handler) GetFileDigest(c echo.Context) error {
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

	algorithm := c.QueryParam("algorithm")
	if algorithm == "" {
		algorithm = "SHA256"
	}

	// Validate algorithm
	validAlgorithms := []string{"SHA256", "SHA512", "BLAKE3", "MD5"}
	isValidAlgorithm := false
	for _, v := range validAlgorithms {
		if algorithm == v {
			isValidAlgorithm = true
			break
		}
	}
	if !isValidAlgorithm {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid algorithm parameter")
	}

	req := models.GetFileDigestRequest{
		SnapshotID: snapshotID,
		Path:       path,
		Algorithm:  algorithm,
	}

	result, err := h.storage.GetFileDigest(ctx, &req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return echo.NewHTTPError(http.StatusNotFound, "File not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to get file digest: %v", err))
	}

	return c.JSON(http.StatusOK, result)
}