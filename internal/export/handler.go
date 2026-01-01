package export

import (
	"database/sql"
	"net/http"

	"cms/internal/settings"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service         *Service
	settingsService *settings.Service
}

func NewHandler(db *sql.DB) *Handler {
	return &Handler{
		service:         NewService(db),
		settingsService: settings.NewService(),
	}
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	r.POST("/export", h.Export)
}

func (h *Handler) Export(c *gin.Context) {
	// 設定からexport_dirを取得
	s, err := h.settingsService.Get()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load settings: " + err.Error()})
		return
	}

	cfg := Config{
		ExportDir: s.ExportDir,
		UploadDir: "./uploads",
	}
	if err := h.service.Export(cfg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "export completed", "export_dir": s.ExportDir})
}
