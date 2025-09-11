package http

import (
	"log/slog"
	"net/http"

	db "order/internal/infrastructure/database"

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
	r.Static("/static", "./static")
	r.GET("/", h.serveHome)
	r.GET("/api/orders/:uid", h.getOrderByUID)
}

func (h *Handler) serveHome(c *gin.Context) {
	slog.Info("Serving home page")
	c.File("./static/index.html")
}

func (h *Handler) getOrderByUID(c *gin.Context) {
	uid := c.Param("uid")
	slog.Info("Fetching order", slog.String("uid", uid))

	order, err := h.repo.GetOrderWithCache(uid)
	if err != nil {
		slog.Error("Failed to get order",
			slog.String("uid", uid),
			slog.String("error", err.Error()),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get order"})
		return
	}
	if order == nil {
		slog.Warn("Order not found", slog.String("uid", uid))
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}

	slog.Info("Order fetched successfully", slog.String("uid", uid))
	c.JSON(http.StatusOK, order)
}
