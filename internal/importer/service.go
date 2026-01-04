package importer

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"regexp"
	"strings"

	"cms/internal/article"
	"cms/internal/category"
	"cms/internal/tag"

	"github.com/goccy/go-yaml"
)

// FrontMatter はMarkdownファイルのフロントマター
type FrontMatter struct {
	Title    string   `yaml:"title"`
	Slug     string   `yaml:"slug"`
	Category string   `yaml:"category"`
	Tags     []string `yaml:"tags"`
	Status   string   `yaml:"status"` // draft or published
}

type Service struct {
	articleService *article.Service
	categoryRepo   *category.Repository
	tagRepo        *tag.Repository
}

func NewService(db *sql.DB) *Service {
	return &Service{
		articleService: article.NewService(db),
		categoryRepo:   category.NewRepository(db),
		tagRepo:        tag.NewRepository(db),
	}
}

// ImportMarkdown はMarkdownファイルを解析してDBに保存
func (s *Service) ImportMarkdown(filePath string) (*article.Article, error) {
	// ファイル読み込み
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("ファイル読み込みエラー: %w", err)
	}

	// フロントマターと本文を分離
	frontMatter, body, err := parseFrontMatter(string(content))
	if err != nil {
		return nil, fmt.Errorf("フロントマター解析エラー: %w", err)
	}

	// バリデーション
	if frontMatter.Title == "" {
		return nil, fmt.Errorf("titleは必須です")
	}
	if frontMatter.Slug == "" {
		// タイトルからslugを生成
		frontMatter.Slug = generateSlug(frontMatter.Title)
	}
	if frontMatter.Status == "" {
		frontMatter.Status = "draft"
	}

	// カテゴリ解決（名前からIDを取得、なければ作成）
	var categoryID *int64
	if frontMatter.Category != "" {
		cat, err := s.findOrCreateCategory(frontMatter.Category)
		if err != nil {
			return nil, fmt.Errorf("カテゴリ解決エラー: %w", err)
		}
		categoryID = &cat.ID
	}

	// タグ解決（名前からIDを取得、なければ作成）
	var tagIDs []int64
	for _, tagName := range frontMatter.Tags {
		t, err := s.findOrCreateTag(tagName)
		if err != nil {
			return nil, fmt.Errorf("タグ解決エラー: %w", err)
		}
		tagIDs = append(tagIDs, t.ID)
	}

	// 記事作成（authorID=1をデフォルトとする）
	article, err := s.articleService.Create(
		frontMatter.Title,
		frontMatter.Slug,
		body,
		frontMatter.Status,
		1, // authorID (デフォルト)
		categoryID,
		tagIDs,
	)
	if err != nil {
		return nil, err
	}

	// statusが"published"の場合は、published_atを設定するためにPublishを呼ぶ
	if frontMatter.Status == "published" {
		return s.articleService.Publish(article.ID)
	}

	return article, nil
}

// parseFrontMatter はフロントマターと本文を分離
func parseFrontMatter(content string) (*FrontMatter, string, error) {
	scanner := bufio.NewScanner(strings.NewReader(content))

	// 最初の---を探す
	if !scanner.Scan() || strings.TrimSpace(scanner.Text()) != "---" {
		return nil, "", fmt.Errorf("フロントマターが見つかりません（---で開始してください）")
	}

	// ---で終わるまでフロントマター部分を収集
	var yamlLines []string
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "---" {
			break
		}
		yamlLines = append(yamlLines, line)
	}

	// 残りを本文として収集
	var bodyLines []string
	for scanner.Scan() {
		bodyLines = append(bodyLines, scanner.Text())
	}

	// YAML解析
	var fm FrontMatter
	yamlContent := strings.Join(yamlLines, "\n")
	if err := yaml.Unmarshal([]byte(yamlContent), &fm); err != nil {
		return nil, "", fmt.Errorf("YAML解析エラー: %w", err)
	}

	body := strings.TrimSpace(strings.Join(bodyLines, "\n"))
	return &fm, body, nil
}

// findOrCreateCategory はカテゴリを名前で検索し、なければ作成
func (s *Service) findOrCreateCategory(name string) (*category.Category, error) {
	categories, err := s.categoryRepo.GetAll()
	if err != nil {
		return nil, err
	}

	// 名前で検索
	for _, c := range categories {
		if c.Name == name {
			return &c, nil
		}
	}

	// なければ作成
	slug := generateSlug(name)
	return s.categoryRepo.Create(name, slug)
}

// findOrCreateTag はタグを名前で検索し、なければ作成
func (s *Service) findOrCreateTag(name string) (*tag.Tag, error) {
	tags, err := s.tagRepo.GetAll()
	if err != nil {
		return nil, err
	}

	// 名前で検索
	for _, t := range tags {
		if t.Name == name {
			return &t, nil
		}
	}

	// なければ作成
	slug := generateSlug(name)
	return s.tagRepo.Create(name, slug)
}

// generateSlug は名前からスラッグを生成
func generateSlug(name string) string {
	// 小文字に変換
	slug := strings.ToLower(name)
	// 空白をハイフンに置換
	slug = strings.ReplaceAll(slug, " ", "-")
	// 英数字とハイフン以外を除去（日本語はそのまま）
	reg := regexp.MustCompile(`[^\p{L}\p{N}-]`)
	slug = reg.ReplaceAllString(slug, "")
	// 連続ハイフンを1つに
	reg = regexp.MustCompile(`-+`)
	slug = reg.ReplaceAllString(slug, "-")
	// 先頭・末尾のハイフンを除去
	slug = strings.Trim(slug, "-")
	return slug
}

