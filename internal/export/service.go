package export

import (
	"bytes"
	"database/sql"
	"embed"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"cms/internal/article"
	"cms/internal/category"
	"cms/internal/tag"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/renderer/html"
)

//go:embed templates/*.html
var defaultTemplateFS embed.FS

type Service struct {
	articleRepo  *article.Repository
	categoryRepo *category.Repository
	tagRepo      *tag.Repository
	md           goldmark.Markdown
}

func NewService(db *sql.DB) *Service {
	return &Service{
		articleRepo:  article.NewRepository(db),
		categoryRepo: category.NewRepository(db),
		tagRepo:      tag.NewRepository(db),
		md: goldmark.New(
			goldmark.WithRendererOptions(html.WithUnsafe()),
		),
	}
}

type Config struct {
	ExportDir string `json:"export_dir"`
	UploadDir string `json:"upload_dir"`
	SiteTitle string `json:"site_title"`
}

func (s *Service) Export(cfg Config) error {
	// テンプレートをロード（カスタム優先）
	tmpl, err := s.loadTemplates(cfg.ExportDir)
	if err != nil {
		return err
	}

	// ディレクトリ作成
	dirs := []string{
		cfg.ExportDir,
		filepath.Join(cfg.ExportDir, "posts"),
		filepath.Join(cfg.ExportDir, "categories"),
		filepath.Join(cfg.ExportDir, "tags"),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	// 公開済み記事を取得
	articles, err := s.articleRepo.GetPublished()
	if err != nil {
		return err
	}

	// 記事個別ページ生成
	for _, a := range articles {
		if err := s.exportArticle(cfg, tmpl, a); err != nil {
			return err
		}
	}

	// 一覧ページ生成
	if err := s.exportIndex(cfg, tmpl, articles); err != nil {
		return err
	}

	// カテゴリ別一覧ページ生成
	if err := s.exportCategories(cfg, tmpl); err != nil {
		return err
	}

	// タグ別一覧ページ生成
	if err := s.exportTags(cfg, tmpl); err != nil {
		return err
	}

	// 画像ファイルをコピー
	if cfg.UploadDir != "" {
		if err := s.copyImages(cfg.UploadDir, cfg.ExportDir); err != nil {
			return err
		}
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

func (s *Service) exportArticle(cfg Config, tmpl *template.Template, a article.Article) error {
	// 画像パスを変換: http://localhost:8080/api/images/ → ../images/ (postsフォルダからの相対パス)
	content := strings.ReplaceAll(a.Content, "http://localhost:8080/api/images/", "../images/")

	// Markdown → HTML
	var contentBuf bytes.Buffer
	if err := s.md.Convert([]byte(content), &contentBuf); err != nil {
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
		"Title":     a.Title,
		"SiteTitle": cfg.SiteTitle,
		"Content":   template.HTML(articleBuf.String()),
	})
	if err != nil {
		return err
	}

	// ファイル書き出し
	path := filepath.Join(cfg.ExportDir, "posts", a.Slug+".html")
	return os.WriteFile(path, finalBuf.Bytes(), 0644)
}

func (s *Service) exportIndex(cfg Config, tmpl *template.Template, articles []article.Article) error {
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
		"Title":     cfg.SiteTitle,
		"SiteTitle": cfg.SiteTitle,
		"Content":   template.HTML(indexBuf.String()),
	})
	if err != nil {
		return err
	}

	// ファイル書き出し
	path := filepath.Join(cfg.ExportDir, "index.html")
	return os.WriteFile(path, finalBuf.Bytes(), 0644)
}

func (s *Service) exportCategories(cfg Config, tmpl *template.Template) error {
	categories, err := s.categoryRepo.GetAll()
	if err != nil {
		return err
	}

	for _, c := range categories {
		articles, err := s.articleRepo.GetByCategory(c.ID)
		if err != nil {
			return err
		}

		// 記事がなければスキップ
		if len(articles) == 0 {
			continue
		}

		// カテゴリテンプレート
		var categoryBuf bytes.Buffer
		err = tmpl.ExecuteTemplate(&categoryBuf, "category.html", map[string]interface{}{
			"Category": c,
			"Articles": articles,
		})
		if err != nil {
			return err
		}

		// ベーステンプレート
		var finalBuf bytes.Buffer
		err = tmpl.ExecuteTemplate(&finalBuf, "base.html", map[string]interface{}{
			"Title":     "カテゴリ: " + c.Name,
			"SiteTitle": cfg.SiteTitle,
			"Content":   template.HTML(categoryBuf.String()),
		})
		if err != nil {
			return err
		}

		// ファイル書き出し
		path := filepath.Join(cfg.ExportDir, "categories", c.Slug+".html")
		if err := os.WriteFile(path, finalBuf.Bytes(), 0644); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) exportTags(cfg Config, tmpl *template.Template) error {
	tags, err := s.tagRepo.GetAll()
	if err != nil {
		return err
	}

	for _, t := range tags {
		articles, err := s.articleRepo.GetByTag(t.ID)
		if err != nil {
			return err
		}

		// 記事がなければスキップ
		if len(articles) == 0 {
			continue
		}

		// タグテンプレート
		var tagBuf bytes.Buffer
		err = tmpl.ExecuteTemplate(&tagBuf, "tag.html", map[string]interface{}{
			"Tag":      t,
			"Articles": articles,
		})
		if err != nil {
			return err
		}

		// ベーステンプレート
		var finalBuf bytes.Buffer
		err = tmpl.ExecuteTemplate(&finalBuf, "base.html", map[string]interface{}{
			"Title":     "タグ: " + t.Name,
			"SiteTitle": cfg.SiteTitle,
			"Content":   template.HTML(tagBuf.String()),
		})
		if err != nil {
			return err
		}

		// ファイル書き出し
		path := filepath.Join(cfg.ExportDir, "tags", t.Slug+".html")
		if err := os.WriteFile(path, finalBuf.Bytes(), 0644); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) copyImages(uploadDir, exportDir string) error {
	// uploadsディレクトリが存在しない場合はスキップ
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		return nil
	}

	// 出力先のimagesディレクトリを作成
	imagesDir := filepath.Join(exportDir, "images")
	if err := os.MkdirAll(imagesDir, 0755); err != nil {
		return err
	}

	// uploadsディレクトリ内のファイルをコピー
	entries, err := os.ReadDir(uploadDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		srcPath := filepath.Join(uploadDir, entry.Name())
		dstPath := filepath.Join(imagesDir, entry.Name())

		if err := copyFile(srcPath, dstPath); err != nil {
			return err
		}
	}

	return nil
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = dstFile.ReadFrom(srcFile)
	return err
}
