package template

import _ "embed"

//go:embed defaults/base.html
var defaultBase string

//go:embed defaults/article.html
var defaultArticle string

//go:embed defaults/index.html
var defaultIndex string

//go:embed defaults/category.html
var defaultCategory string

//go:embed defaults/tag.html
var defaultTag string

// DefaultTemplates はデフォルトテンプレートのマップ
var DefaultTemplates = map[string]string{
	TemplateBase:     defaultBase,
	TemplateArticle:  defaultArticle,
	TemplateIndex:    defaultIndex,
	TemplateCategory: defaultCategory,
	TemplateTag:      defaultTag,
}

