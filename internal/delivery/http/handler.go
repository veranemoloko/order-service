package http

import (
	"net/http"
	db "order/internal/infrastructure/postgres"

	"github.com/gin-gonic/gin"
)

// Handler handles HTTP requests and interacts with the database repository.
type Handler struct {
	repo *db.Repository
}

// NewHandler creates a new Handler with the given repository.
func NewHandler(repo *db.Repository) *Handler {
	return &Handler{repo: repo}
}

// RegisterRoutes registers all HTTP endpoints to the given Gin engine.
func (h *Handler) RegisterRoutes(r *gin.Engine) {
	r.GET("/orders/:uid", h.getOrderByUID)
}

// getOrderByUID returns an order by its UID.
func (h *Handler) getOrderByUID(c *gin.Context) {
	uid := c.Param("uid")

	order, err := h.repo.GetOrderWithCache(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get order"})
		return
	}
	if order == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}

	c.JSON(http.StatusOK, order)
}
