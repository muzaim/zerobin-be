package audit

import (
	"time"
)

type Service interface {
	Record(userID string, module string, action string, description string) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Record(userID string, module string, action string, description string) error {
	trail := &AuditTrail{
		UserID:      userID,
		Module:      module,
		Action:      action,
		Description: description,
		CreatedAt:   time.Now(),
	}
	return s.repo.Create(trail)
}
