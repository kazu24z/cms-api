package article

import (
	"time"

	"cms/internal/category"
	"cms/internal/tag"
	"cms/internal/user"
)

type Article struct {
	ID          int64      `json:"id"`
	Title       string     `json:"title"`
	Slug        string     `json:"slug"`
	Content     string     `json:"content"`
	Status      string     `json:"status"`
	AuthorID    *int64     `json:"author_id"`
	CategoryID  *int64     `json:"category_id"`
	PublishedAt *time.Time `json:"published_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

	// Relations (for response)
	Author   *user.User         `json:"author,omitempty"`
	Category *category.Category `json:"category,omitempty"`
	Tags     []tag.Tag          `json:"tags,omitempty"`
}
