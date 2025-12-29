package article

import (
	"database/sql"
	"time"

	"cms/internal/tag"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetAll() ([]Article, error) {
	rows, err := r.db.Query(queryGetAll)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanArticlesWithTags(rows)
}

func (r *Repository) GetByID(id int64) (*Article, error) {
	rows, err := r.db.Query(queryGetByID, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	articles, err := r.scanArticlesWithTags(rows)
	if err != nil {
		return nil, err
	}
	if len(articles) == 0 {
		return nil, sql.ErrNoRows
	}
	return &articles[0], nil
}

func (r *Repository) scanArticlesWithTags(rows *sql.Rows) ([]Article, error) {
	articleMap := make(map[int64]*Article)
	var articleOrder []int64

	for rows.Next() {
		var a Article
		var tagID sql.NullInt64
		var tagName, tagSlug sql.NullString
		var tagCreatedAt sql.NullTime

		err := rows.Scan(
			&a.ID, &a.Title, &a.Slug, &a.Content, &a.Status,
			&a.AuthorID, &a.CategoryID, &a.PublishedAt, &a.CreatedAt, &a.UpdatedAt,
			&tagID, &tagName, &tagSlug, &tagCreatedAt,
		)
		if err != nil {
			return nil, err
		}

		existing, ok := articleMap[a.ID]
		if !ok {
			a.Tags = []tag.Tag{}
			articleMap[a.ID] = &a
			articleOrder = append(articleOrder, a.ID)
			existing = &a
		}

		if tagID.Valid {
			existing.Tags = append(existing.Tags, tag.Tag{
				ID:        tagID.Int64,
				Name:      tagName.String,
				Slug:      tagSlug.String,
				CreatedAt: tagCreatedAt.Time,
			})
		}
	}

	articles := make([]Article, 0, len(articleOrder))
	for _, id := range articleOrder {
		articles = append(articles, *articleMap[id])
	}
	return articles, nil
}

func (r *Repository) Create(title, slug, content, status string, authorID int64, categoryID *int64, tagIDs []int64) (*Article, error) {
	now := time.Now()
	result, err := r.db.Exec(queryCreate, title, slug, content, status, authorID, categoryID, now, now)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	if err := r.SetArticleTags(id, tagIDs); err != nil {
		return nil, err
	}

	return r.GetByID(id)
}

func (r *Repository) Update(id int64, title, slug, content, status string, categoryID *int64, tagIDs []int64) (*Article, error) {
	_, err := r.db.Exec(queryUpdate, title, slug, content, status, categoryID, time.Now(), id)
	if err != nil {
		return nil, err
	}

	if err := r.SetArticleTags(id, tagIDs); err != nil {
		return nil, err
	}

	return r.GetByID(id)
}

func (r *Repository) SetArticleTags(articleID int64, tagIDs []int64) error {
	_, err := r.db.Exec(queryDeleteTags, articleID)
	if err != nil {
		return err
	}

	for _, tagID := range tagIDs {
		_, err := r.db.Exec(queryInsertTag, articleID, tagID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) Publish(id int64) (*Article, error) {
	now := time.Now()
	_, err := r.db.Exec(queryPublish, now, now, id)
	if err != nil {
		return nil, err
	}
	return r.GetByID(id)
}

func (r *Repository) Delete(id int64) error {
	_, err := r.db.Exec(queryDelete, id)
	return err
}
