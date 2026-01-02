package template

import (
	"database/sql"
)

type Service struct {
	repo *Repository
}

func NewService(db *sql.DB) *Service {
	return &Service{repo: NewRepository(db)}
}

func (s *Service) GetAll() ([]Template, error) {
	return s.repo.GetAll()
}

func (s *Service) GetByName(name string) (*Template, error) {
	return s.repo.GetByName(name)
}

func (s *Service) Update(name, content string) (*Template, error) {
	// バリデーション: 有効なテンプレート名かチェック
	valid := false
	for _, n := range AllTemplateNames {
		if n == name {
			valid = true
			break
		}
	}
	if !valid {
		return nil, sql.ErrNoRows
	}

	return s.repo.Upsert(name, content)
}

func (s *Service) ResetToDefaults() error {
	for _, name := range AllTemplateNames {
		content, ok := DefaultTemplates[name]
		if !ok {
			continue
		}
		if _, err := s.repo.Upsert(name, content); err != nil {
			return err
		}
	}
	return nil
}

// InitializeDefaults はDBにテンプレートがなければデフォルトを投入
func (s *Service) InitializeDefaults() error {
	count, err := s.repo.Count()
	if err != nil {
		return err
	}
	if count == 0 {
		return s.ResetToDefaults()
	}
	return nil
}

