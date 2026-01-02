package template

import (
	"archive/zip"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(db *sql.DB) *Handler {
	return &Handler{service: NewService(db)}
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/templates", h.GetAll)
	r.GET("/templates/:name", h.GetByName)
	r.PUT("/templates/:name", h.Update)
	r.POST("/templates/:name/upload", h.Upload)
	r.POST("/templates/import", h.Import)
	r.POST("/templates/reset", h.Reset)
}

func (h *Handler) GetAll(c *gin.Context) {
	templates, err := h.service.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, templates)
}

func (h *Handler) GetByName(c *gin.Context) {
	name := c.Param("name")

	template, err := h.service.GetByName(name)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "template not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, template)
}

type UpdateRequest struct {
	Content string `json:"content" binding:"required"`
}

func (h *Handler) Update(c *gin.Context) {
	name := c.Param("name")

	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	template, err := h.service.Update(name, req.Content)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "invalid template name"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, template)
}

func (h *Handler) Reset(c *gin.Context) {
	if err := h.service.ResetToDefaults(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "templates reset to defaults"})
}

// Upload は単一テンプレートファイルをアップロード
func (h *Handler) Upload(c *gin.Context) {
	name := c.Param("name")

	// テンプレート名のバリデーション
	if !isValidTemplateName(name) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid template name"})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	// ファイルサイズチェック (1MB上限)
	if file.Size > 1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file size exceeds 1MB limit"})
		return
	}

	// ファイルを読み込み
	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer f.Close()

	content, err := io.ReadAll(f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// DB に保存
	template, err := h.service.Update(name, string(content))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, template)
}

// Import はZIPファイルから複数テンプレートを一括インポート
func (h *Handler) Import(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "zip file is required"})
		return
	}

	// ファイルサイズチェック (5MB上限)
	if file.Size > 5*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file size exceeds 5MB limit"})
		return
	}

	// ファイルを開く
	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer f.Close()

	// ZIPとして読み込み
	zipReader, err := zip.NewReader(f, file.Size)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid zip file"})
		return
	}

	// インポート結果を追跡
	imported := []string{}
	errors := []string{}

	for _, zipFile := range zipReader.File {
		// ディレクトリはスキップ
		if zipFile.FileInfo().IsDir() {
			continue
		}

		// ファイル名からテンプレート名を取得 (パス除去、拡張子除去)
		baseName := filepath.Base(zipFile.Name)
		templateName := strings.TrimSuffix(baseName, ".html")

		// 有効なテンプレート名かチェック
		if !isValidTemplateName(templateName) {
			continue // 無効なファイルはスキップ
		}

		// ファイルを読み込み
		rc, err := zipFile.Open()
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: %s", templateName, err.Error()))
			continue
		}

		content, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: %s", templateName, err.Error()))
			continue
		}

		// DB に保存
		_, err = h.service.Update(templateName, string(content))
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: %s", templateName, err.Error()))
			continue
		}

		imported = append(imported, templateName)
	}

	if len(imported) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "no valid templates found in zip",
			"errors": errors,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  fmt.Sprintf("imported %d templates", len(imported)),
		"imported": imported,
		"errors":   errors,
	})
}

func isValidTemplateName(name string) bool {
	for _, n := range AllTemplateNames {
		if n == name {
			return true
		}
	}
	return false
}

