package tag

import "database/sql"

type Service struct {
	repo *Repository
}

func NewService(db *sql.DB) *Service {
	return &Service{repo: NewRepository(db)}
}

func (s *Service) GetAll() ([]Tag, error) {
	return s.repo.GetAll()
}

func (s *Service) GetByID(id int64) (*Tag, error) {
	return s.repo.GetByID(id)
}

func (s *Service) Create(name, slug string) (*Tag, error) {
	return s.repo.Create(name, slug)
}

func (s *Service) Update(id int64, name, slug string) (*Tag, error) {
	return s.repo.Update(id, name, slug)
}

func (s *Service) Delete(id int64) error {
	return s.repo.Delete(id)
}
