package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "cms",
	Short: "CMS - Content Management System",
	Long:  `CMSはMarkdownベースのコンテンツ管理システムです。`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
