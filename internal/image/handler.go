package image

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	uploadDir string
}

func NewHandler(uploadDir string) *Handler {
	// ディレクトリが存在しなければ作成
	os.MkdirAll(uploadDir, 0755)
	return &Handler{uploadDir: uploadDir}
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	r.POST("/images", h.Upload)
	r.GET("/images/:filename", h.Serve)
}

// Upload handles image upload
func (h *Handler) Upload(c *gin.Context) {
	file, header, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "画像ファイルが必要です"})
		return
	}
	defer file.Close()

	// 拡張子を取得・検証
	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowedExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true}
	if !allowedExts[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "許可されていないファイル形式です"})
		return
	}

	// ユニークなファイル名を生成
	randBytes := make([]byte, 4)
	rand.Read(randBytes)
	filename := fmt.Sprintf("%d_%s%s", time.Now().Unix(), hex.EncodeToString(randBytes), ext)
	savePath := filepath.Join(h.uploadDir, filename)

	// ファイルを保存
	out, err := os.Create(savePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ファイルの保存に失敗しました"})
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ファイルの保存に失敗しました"})
		return
	}

	// URLパスを返す
	c.JSON(http.StatusOK, gin.H{
		"filename": filename,
		"url":      "/api/images/" + filename,
	})
}

// Serve serves uploaded images
func (h *Handler) Serve(c *gin.Context) {
	filename := c.Param("filename")

	// パストラバーサル対策
	if strings.Contains(filename, "..") || strings.Contains(filename, "/") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不正なファイル名です"})
		return
	}

	filePath := filepath.Join(h.uploadDir, filename)

	// ファイルが存在するか確認
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "ファイルが見つかりません"})
		return
	}

	c.File(filePath)
}

// GetUploadDir returns the upload directory path
func (h *Handler) GetUploadDir() string {
	return h.uploadDir
}
