package cmd

import (
	"log"

	"cms/db"
	"cms/internal/article"
	"cms/internal/category"
	"cms/internal/export"
	"cms/internal/image"
	"cms/internal/settings"
	"cms/internal/tag"
	"cms/internal/template"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

var servePort string

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "APIサーバーを起動",
	Long:  `HTTPサーバーを起動してAPIを提供します。`,
	Run:   runServe,
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().StringVarP(&servePort, "port", "p", "8080", "サーバーポート")
}

func runServe(cmd *cobra.Command, args []string) {
	// DB初期化
	if err := db.Init(); err != nil {
		log.Fatal("Failed to connect database:", err)
	}
	defer db.Close()

	// マイグレーション
	if err := db.Migrate(); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// テンプレート初期化
	templateService := template.NewService(db.DB)
	if err := templateService.InitializeDefaults(); err != nil {
		log.Fatal("Failed to initialize templates:", err)
	}

	r := gin.Default()

	// CORS設定
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type"},
		AllowCredentials: true,
	}))

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

		exportHandler := export.NewHandler(db.DB)
		exportHandler.RegisterRoutes(api)

		imageHandler := image.NewHandler("./uploads")
		imageHandler.RegisterRoutes(api)

		settingsHandler := settings.NewHandler()
		settingsHandler.RegisterRoutes(api)

		templateHandler := template.NewHandler(db.DB)
		templateHandler.RegisterRoutes(api)
	}

	r.Run(":" + servePort)
}
