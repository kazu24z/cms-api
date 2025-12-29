package export

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(db *sql.DB) *Handler {
	return &Handler{service: NewService(db)}
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	r.POST("/export", h.Export)
}

type ExportRequest struct {
	ExportDir string `json:"export_dir" binding:"required"`
}

func (h *Handler) Export(c *gin.Context) {
	var req ExportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cfg := Config{ExportDir: req.ExportDir}
	if err := h.service.Export(cfg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "export completed", "export_dir": req.ExportDir})
}
