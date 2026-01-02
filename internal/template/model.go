package template

import "time"

type Template struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// テンプレート名の定数
const (
	TemplateBase     = "base"
	TemplateArticle  = "article"
	TemplateIndex    = "index"
	TemplateCategory = "category"
	TemplateTag      = "tag"
)

// AllTemplateNames は全テンプレート名のリスト
var AllTemplateNames = []string{
	TemplateBase,
	TemplateArticle,
	TemplateIndex,
	TemplateCategory,
	TemplateTag,
}

