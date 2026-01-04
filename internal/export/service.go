package export

import (
	"bytes"
	"database/sql"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"cms/internal/article"
	"cms/internal/category"
	"cms/internal/tag"
	tmpl "cms/internal/template"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/renderer/html"
)

type Service struct {
	articleRepo  *article.Repository
	categoryRepo *category.Repository
	tagRepo      *tag.Repository
	templateRepo *tmpl.Repository
	md           goldmark.Markdown
}

func NewService(db *sql.DB) *Service {
	return &Service{
		articleRepo:  article.NewRepository(db),
		categoryRepo: category.NewRepository(db),
		tagRepo:      tag.NewRepository(db),
		templateRepo: tmpl.NewRepository(db),
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
	// テンプレートをDBからロード
	t, err := s.loadTemplates()
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
		if err := s.exportArticle(cfg, t, a); err != nil {
			return err
		}
	}

	// 一覧ページ生成
	if err := s.exportIndex(cfg, t, articles); err != nil {
		return err
	}

	// カテゴリ別一覧ページ生成
	if err := s.exportCategories(cfg, t); err != nil {
		return err
	}

	// タグ別一覧ページ生成
	if err := s.exportTags(cfg, t); err != nil {
		return err
	}

	// 不要ファイル削除（下書きに戻した記事のHTMLなど）
	if err := s.cleanupOrphanedFiles(cfg, articles); err != nil {
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

func (s *Service) loadTemplates() (*template.Template, error) {
	// DBからテンプレートを取得
	templates, err := s.templateRepo.GetAll()
	if err != nil {
		return nil, err
	}

	// template.Templateを構築
	t := template.New("")
	for _, tmplData := range templates {
		_, err := t.New(tmplData.Name + ".html").Parse(tmplData.Content)
		if err != nil {
			return nil, err
		}
	}

	return t, nil
}

func (s *Service) exportArticle(cfg Config, t *template.Template, a article.Article) error {
	// 画像パスを変換: http://localhost:8080/api/images/ → ../images/ (postsフォルダからの相対パス)
	content := strings.ReplaceAll(a.Content, "http://localhost:8080/api/images/", "../images/")

	// Markdown → HTML
	var contentBuf bytes.Buffer
	if err := s.md.Convert([]byte(content), &contentBuf); err != nil {
		return err
	}

	// 記事テンプレート
	var articleBuf bytes.Buffer
	err := t.ExecuteTemplate(&articleBuf, "article.html", map[string]interface{}{
		"Article": a,
		"Content": template.HTML(contentBuf.String()),
	})
	if err != nil {
		return err
	}

	// ベーステンプレート
	var finalBuf bytes.Buffer
	err = t.ExecuteTemplate(&finalBuf, "base.html", map[string]interface{}{
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

func (s *Service) exportIndex(cfg Config, t *template.Template, articles []article.Article) error {
	// 一覧テンプレート
	var indexBuf bytes.Buffer
	err := t.ExecuteTemplate(&indexBuf, "index.html", map[string]interface{}{
		"Articles": articles,
	})
	if err != nil {
		return err
	}

	// ベーステンプレート
	var finalBuf bytes.Buffer
	err = t.ExecuteTemplate(&finalBuf, "base.html", map[string]interface{}{
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

func (s *Service) exportCategories(cfg Config, t *template.Template) error {
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
		err = t.ExecuteTemplate(&categoryBuf, "category.html", map[string]interface{}{
			"Category": c,
			"Articles": articles,
		})
		if err != nil {
			return err
		}

		// ベーステンプレート
		var finalBuf bytes.Buffer
		err = t.ExecuteTemplate(&finalBuf, "base.html", map[string]interface{}{
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

func (s *Service) exportTags(cfg Config, t *template.Template) error {
	tags, err := s.tagRepo.GetAll()
	if err != nil {
		return err
	}

	for _, tg := range tags {
		articles, err := s.articleRepo.GetByTag(tg.ID)
		if err != nil {
			return err
		}

		// 記事がなければスキップ
		if len(articles) == 0 {
			continue
		}

		// タグテンプレート
		var tagBuf bytes.Buffer
		err = t.ExecuteTemplate(&tagBuf, "tag.html", map[string]interface{}{
			"Tag":      tg,
			"Articles": articles,
		})
		if err != nil {
			return err
		}

		// ベーステンプレート
		var finalBuf bytes.Buffer
		err = t.ExecuteTemplate(&finalBuf, "base.html", map[string]interface{}{
			"Title":     "タグ: " + tg.Name,
			"SiteTitle": cfg.SiteTitle,
			"Content":   template.HTML(tagBuf.String()),
		})
		if err != nil {
			return err
		}

		// ファイル書き出し
		path := filepath.Join(cfg.ExportDir, "tags", tg.Slug+".html")
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

func (s *Service) cleanupOrphanedFiles(cfg Config, publishedArticles []article.Article) error {
	// 公開済み記事のslugセットを作成
	publishedSlugs := make(map[string]bool)
	for _, a := range publishedArticles {
		publishedSlugs[a.Slug] = true
	}

	// 公開済み記事が含まれるカテゴリ・タグのslugセットを作成
	publishedCategorySlugs := make(map[string]bool)
	publishedTagSlugs := make(map[string]bool)

	for _, a := range publishedArticles {
		if a.Category != nil {
			publishedCategorySlugs[a.Category.Slug] = true
		}
		for _, tag := range a.Tags {
			publishedTagSlugs[tag.Slug] = true
		}
	}

	// posts/ディレクトリの不要ファイルを削除
	if err := s.cleanupDirectory(cfg.ExportDir, "posts", publishedSlugs); err != nil {
		return err
	}

	// categories/ディレクトリの不要ファイルを削除
	if err := s.cleanupDirectory(cfg.ExportDir, "categories", publishedCategorySlugs); err != nil {
		return err
	}

	// tags/ディレクトリの不要ファイルを削除
	if err := s.cleanupDirectory(cfg.ExportDir, "tags", publishedTagSlugs); err != nil {
		return err
	}

	return nil
}

func (s *Service) cleanupDirectory(exportDir, subdir string, validSlugs map[string]bool) error {
	dir := filepath.Join(exportDir, subdir)

	// ディレクトリが存在しない場合はスキップ
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// .htmlファイルのみ処理
		if !strings.HasSuffix(entry.Name(), ".html") {
			continue
		}

		// slugを抽出（例: "article-slug.html" → "article-slug"）
		slug := strings.TrimSuffix(entry.Name(), ".html")

		// 公開済みリストに含まれていない場合は削除
		if !validSlugs[slug] {
			filePath := filepath.Join(dir, entry.Name())
			if err := os.Remove(filePath); err != nil {
				return err
			}
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
