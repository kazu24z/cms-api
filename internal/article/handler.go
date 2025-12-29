package article

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(db *sql.DB) *Handler {
	return &Handler{service: NewService(db)}
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/articles", h.GetAll)
	r.GET("/articles/:id", h.GetByID)
	r.POST("/articles", h.Create)
	r.PUT("/articles/:id", h.Update)
	r.POST("/articles/:id/publish", h.Publish)
	r.DELETE("/articles/:id", h.Delete)
}

type CreateRequest struct {
	Title      string  `json:"title" binding:"required"`
	Slug       string  `json:"slug" binding:"required"`
	Content    string  `json:"content"`
	Status     string  `json:"status"`
	AuthorID   int64   `json:"author_id" binding:"required"`
	CategoryID *int64  `json:"category_id"`
	TagIDs     []int64 `json:"tag_ids"`
}

type UpdateRequest struct {
	Title      string  `json:"title" binding:"required"`
	Slug       string  `json:"slug" binding:"required"`
	Content    string  `json:"content"`
	Status     string  `json:"status"`
	CategoryID *int64  `json:"category_id"`
	TagIDs     []int64 `json:"tag_ids"`
}

func (h *Handler) GetAll(c *gin.Context) {
	articles, err := h.service.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, articles)
}

func (h *Handler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	article, err := h.service.GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "article not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, article)
}

func (h *Handler) Create(c *gin.Context) {
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	article, err := h.service.Create(req.Title, req.Slug, req.Content, req.Status, req.AuthorID, req.CategoryID, req.TagIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, article)
}

func (h *Handler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	article, err := h.service.Update(id, req.Title, req.Slug, req.Content, req.Status, req.CategoryID, req.TagIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, article)
}

func (h *Handler) Publish(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	article, err := h.service.Publish(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, article)
}

func (h *Handler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.service.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}
