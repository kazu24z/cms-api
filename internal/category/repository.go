package category

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

func (r *Repository) GetAll() ([]Category, error) {
	rows, err := r.db.Query("SELECT id, name, slug, created_at FROM categories ORDER BY id DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []Category
	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug, &c.CreatedAt); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	return categories, nil
}

func (r *Repository) GetByID(id int64) (*Category, error) {
	var c Category
	err := r.db.QueryRow("SELECT id, name, slug, created_at FROM categories WHERE id = ?", id).
		Scan(&c.ID, &c.Name, &c.Slug, &c.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *Repository) Create(name, slug string) (*Category, error) {
	result, err := r.db.Exec(
		"INSERT INTO categories (name, slug, created_at) VALUES (?, ?, ?)",
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

func (r *Repository) Update(id int64, name, slug string) (*Category, error) {
	_, err := r.db.Exec(
		"UPDATE categories SET name = ?, slug = ? WHERE id = ?",
		name, slug, id,
	)
	if err != nil {
		return nil, err
	}

	return r.GetByID(id)
}

func (r *Repository) Delete(id int64) error {
	_, err := r.db.Exec("DELETE FROM categories WHERE id = ?", id)
	return err
}
