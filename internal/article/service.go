package article

import "database/sql"

type Service struct {
	repo *Repository
}

func NewService(db *sql.DB) *Service {
	return &Service{repo: NewRepository(db)}
}

func (s *Service) GetAll() ([]Article, error) {
	return s.repo.GetAll()
}

func (s *Service) GetByID(id int64) (*Article, error) {
	return s.repo.GetByID(id)
}

func (s *Service) Create(title, slug, content, status string, authorID, categoryID *int64, tagIDs []int64) (*Article, error) {
	if status == "" {
		status = "draft"
	}
	return s.repo.Create(title, slug, content, status, authorID, categoryID, tagIDs)
}

func (s *Service) Update(id int64, title, slug, content, status string, categoryID *int64, tagIDs []int64) (*Article, error) {
	return s.repo.Update(id, title, slug, content, status, categoryID, tagIDs)
}

func (s *Service) Publish(id int64) (*Article, error) {
	return s.repo.Publish(id)
}

func (s *Service) Delete(id int64) error {
	return s.repo.Delete(id)
}
