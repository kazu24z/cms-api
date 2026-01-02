package template

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

func (r *Repository) GetAll() ([]Template, error) {
	rows, err := r.db.Query(queryGetAll)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var templates []Template
	for rows.Next() {
		var t Template
		err := rows.Scan(&t.ID, &t.Name, &t.Content, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			return nil, err
		}
		templates = append(templates, t)
	}
	return templates, nil
}

func (r *Repository) GetByName(name string) (*Template, error) {
	var t Template
	err := r.db.QueryRow(queryGetByName, name).Scan(
		&t.ID, &t.Name, &t.Content, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *Repository) Upsert(name, content string) (*Template, error) {
	now := time.Now()
	_, err := r.db.Exec(queryUpsert, name, content, now, now)
	if err != nil {
		return nil, err
	}
	return r.GetByName(name)
}

func (r *Repository) Count() (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM templates").Scan(&count)
	return count, err
}

