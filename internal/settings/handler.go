package settings

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler() *Handler {
	return &Handler{service: NewService()}
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/settings", h.Get)
	r.POST("/settings", h.Update)
}

func (h *Handler) Get(c *gin.Context) {
	settings, err := h.service.Get()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, settings)
}

type UpdateRequest struct {
	ExportDir string `json:"export_dir" binding:"required"`
	SiteTitle string `json:"site_title"`
}

func (h *Handler) Update(c *gin.Context) {
	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	settings := &Settings{
		ExportDir: req.ExportDir,
		SiteTitle: req.SiteTitle,
	}

	if err := h.service.Update(settings); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, settings)
}
