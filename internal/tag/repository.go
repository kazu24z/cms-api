package tag

import (
	"database/sql"
	"time"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetAll() ([]Tag, error) {
	rows, err := r.db.Query("SELECT id, name, slug, created_at FROM tags ORDER BY id DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []Tag
	for rows.Next() {
		var t Tag
		if err := rows.Scan(&t.ID, &t.Name, &t.Slug, &t.CreatedAt); err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	return tags, nil
}

func (r *Repository) GetByID(id int64) (*Tag, error) {
	var t Tag
	err := r.db.QueryRow("SELECT id, name, slug, created_at FROM tags WHERE id = ?", id).
		Scan(&t.ID, &t.Name, &t.Slug, &t.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *Repository) Create(name, slug string) (*Tag, error) {
	result, err := r.db.Exec(
		"INSERT INTO tags (name, slug, created_at) VALUES (?, ?, ?)",
		name, slug, time.Now(),
	)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return r.GetByID(id)
}

func (r *Repository) Update(id int64, name, slug string) (*Tag, error) {
	_, err := r.db.Exec(
		"UPDATE tags SET name = ?, slug = ? WHERE id = ?",
		name, slug, id,
	)
	if err != nil {
		return nil, err
	}

	return r.GetByID(id)
}

func (r *Repository) Delete(id int64) error {
	_, err := r.db.Exec("DELETE FROM tags WHERE id = ?", id)
	return err
}
