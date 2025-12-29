package export

import (
	"bytes"
	"database/sql"
	"embed"
	"html/template"
	"os"
	"path/filepath"

	"cms/internal/article"

	"github.com/yuin/goldmark"
)

//go:embed templates/*.html
var defaultTemplateFS embed.FS

type Service struct {
	articleRepo *article.Repository
	md          goldmark.Markdown
}

func NewService(db *sql.DB) *Service {
	return &Service{
		articleRepo: article.NewRepository(db),
		md:          goldmark.New(),
	}
}

type Config struct {
	ExportDir string `json:"export_dir"`
}

func (s *Service) Export(cfg Config) error {
	// テンプレートをロード（カスタム優先）
	tmpl, err := s.loadTemplates(cfg.ExportDir)
	if err != nil {
		return err
	}

	// ディレクトリ作成
	if err := os.MkdirAll(cfg.ExportDir, 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(cfg.ExportDir, "posts"), 0755); err != nil {
		return err
	}

	// 公開済み記事を取得
	articles, err := s.articleRepo.GetPublished()
	if err != nil {
		return err
	}

	// 記事個別ページ生成
	for _, a := range articles {
		if err := s.exportArticle(cfg.ExportDir, tmpl, a); err != nil {
			return err
		}
	}

	// 一覧ページ生成
	if err := s.exportIndex(cfg.ExportDir, tmpl, articles); err != nil {
		return err
	}

	return nil
}

func (s *Service) loadTemplates(exportDir string) (*template.Template, error) {
	customDir := filepath.Join(exportDir, "_templates")

	// カスタムテンプレートが存在するかチェック
	if _, err := os.Stat(customDir); err == nil {
		// カスタムテンプレートを使用
		return template.ParseGlob(filepath.Join(customDir, "*.html"))
	}

	// デフォルトテンプレートを使用
	return template.ParseFS(defaultTemplateFS, "templates/*.html")
}

func (s *Service) exportArticle(exportDir string, tmpl *template.Template, a article.Article) error {
	// Markdown → HTML
	var contentBuf bytes.Buffer
	if err := s.md.Convert([]byte(a.Content), &contentBuf); err != nil {
		return err
	}

	// 記事テンプレート
	var articleBuf bytes.Buffer
	err := tmpl.ExecuteTemplate(&articleBuf, "article.html", map[string]interface{}{
		"Article": a,
		"Content": template.HTML(contentBuf.String()),
	})
	if err != nil {
		return err
	}

	// ベーステンプレート
	var finalBuf bytes.Buffer
	err = tmpl.ExecuteTemplate(&finalBuf, "base.html", map[string]interface{}{
		"Title":   a.Title,
		"Content": template.HTML(articleBuf.String()),
	})
	if err != nil {
		return err
	}

	// ファイル書き出し
	path := filepath.Join(exportDir, "posts", a.Slug+".html")
	return os.WriteFile(path, finalBuf.Bytes(), 0644)
}

func (s *Service) exportIndex(exportDir string, tmpl *template.Template, articles []article.Article) error {
	// 一覧テンプレート
	var indexBuf bytes.Buffer
	err := tmpl.ExecuteTemplate(&indexBuf, "index.html", map[string]interface{}{
		"Articles": articles,
	})
	if err != nil {
		return err
	}

	// ベーステンプレート
	var finalBuf bytes.Buffer
	err = tmpl.ExecuteTemplate(&finalBuf, "base.html", map[string]interface{}{
		"Title":   "Blog",
		"Content": template.HTML(indexBuf.String()),
	})
	if err != nil {
		return err
	}

	// ファイル書き出し
	path := filepath.Join(exportDir, "index.html")
	return os.WriteFile(path, finalBuf.Bytes(), 0644)
}
