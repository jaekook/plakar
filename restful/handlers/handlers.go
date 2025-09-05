package handlers

import (
	"github.com/PlakarKorp/plakar/restful/config"
	"github.com/PlakarKorp/plakar/restful/storage"
)

// Handler contains all HTTP handlers
type Handler struct {
	storage storage.Storage
	config  *config.Config
}

// New creates a new handler instance
func New(storage storage.Storage, config *config.Config) *Handler {
	return &Handler{
		storage: storage,
		config:  config,
	}
}