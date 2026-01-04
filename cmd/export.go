package cmd

import (
	"fmt"
	"log"

	"cms/db"
	"cms/internal/export"

	"github.com/spf13/cobra"
)

var (
	exportDir string
	uploadDir string
	siteTitle string
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "記事をHTMLにエクスポート",
	Long:  `データベースの公開済み記事を静的HTMLファイルとして出力します。`,
	Run:   runExport,
}

func init() {
	rootCmd.AddCommand(exportCmd)
	exportCmd.Flags().StringVarP(&exportDir, "output", "o", "./dist", "出力ディレクトリ")
	exportCmd.Flags().StringVarP(&uploadDir, "uploads", "u", "./uploads", "画像ディレクトリ")
	exportCmd.Flags().StringVarP(&siteTitle, "title", "t", "My Blog", "サイトタイトル")
}

func runExport(cmd *cobra.Command, args []string) {
	// DB初期化
	if err := db.Init(); err != nil {
		log.Fatal("Failed to connect database:", err)
	}
	defer db.Close()

	svc := export.NewService(db.DB)
	err := svc.Export(export.Config{
		ExportDir: exportDir,
		UploadDir: uploadDir,
		SiteTitle: siteTitle,
	})
	if err != nil {
		log.Fatal("Export failed:", err)
	}

	fmt.Printf("✓ エクスポート完了: %s\n", exportDir)
}
