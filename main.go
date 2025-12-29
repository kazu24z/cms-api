package main

import (
	"log"

	"cms/db"
	"cms/internal/article"
	"cms/internal/category"
	"cms/internal/tag"

	"github.com/gin-gonic/gin"
)

func main() {
	// DB初期化
	if err := db.Init("cms.db"); err != nil {
		log.Fatal("Failed to connect database:", err)
	}
	defer db.Close()

	// マイグレーション
	if err := db.Migrate(); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API routes
	api := r.Group("/api")
	{
		categoryHandler := category.NewHandler(db.DB)
		categoryHandler.RegisterRoutes(api)

		tagHandler := tag.NewHandler(db.DB)
		tagHandler.RegisterRoutes(api)

		articleHandler := article.NewHandler(db.DB)
		articleHandler.RegisterRoutes(api)
	}

	r.Run(":8080")
}
