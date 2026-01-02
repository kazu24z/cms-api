package template

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
	r.GET("/templates", h.GetAll)
	r.GET("/templates/:name", h.GetByName)
	r.PUT("/templates/:name", h.Update)
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

