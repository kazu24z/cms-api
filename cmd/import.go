package cmd

import (
	"fmt"
	"log"

	"cms/db"
	"cms/internal/importer"

	"github.com/spf13/cobra"
)

var importCmd = &cobra.Command{
	Use:   "import <file.md> [file2.md ...]",
	Short: "MarkdownファイルをDBにインポート",
	Long: `Markdownファイルを解析してデータベースに記事として保存します。

フロントマター形式:
---
title: "記事タイトル"
slug: "article-slug"
category: "カテゴリ名"
tags: ["tag1", "tag2"]
status: "draft"  # draft または published
---

本文（Markdown）`,
	Args: cobra.MinimumNArgs(1),
	Run:  runImport,
}

func init() {
	rootCmd.AddCommand(importCmd)
}

func runImport(cmd *cobra.Command, args []string) {
	// DB初期化
	if err := db.Init(); err != nil {
		log.Fatal("Failed to connect database:", err)
	}
	defer db.Close()

	svc := importer.NewService(db.DB)

	for _, filePath := range args {
		article, err := svc.ImportMarkdown(filePath)
		if err != nil {
			log.Printf("✗ %s: %v\n", filePath, err)
			continue
		}
		fmt.Printf("✓ インポート完了: %s (ID: %d, Slug: %s)\n", filePath, article.ID, article.Slug)
	}
}
