package audit

import (
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type Service interface {
	Record(userID string, module string, action string, resourceID string, description string) error
	RecordTrail(trail *AuditTrail) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Record(userID string, module string, action string, resourceID string, description string) error {
	var userIDUUID *uuid.UUID
	if userID != "" {
		if parsed, err := uuid.Parse(userID); err == nil {
			userIDUUID = &parsed
		}
	}

	trail := &AuditTrail{
		UserID:      userIDUUID,
		Module:      module,
		Action:      action,
		ResourceID:  resourceID,
		Description: description,
		CreatedAt:   time.Now(),
		Status:      "success",
	}
	return s.repo.Create(trail)
}

func (s *service) RecordTrail(trail *AuditTrail) error {
	if trail.ID == uuid.Nil {
		trail.ID = uuid.New()
	}
	if trail.CreatedAt.IsZero() {
		trail.CreatedAt = time.Now()
	}

	go func(t *AuditTrail) {
		if err := s.repo.Create(t); err != nil {
			log.Error().Err(err).Msg("Failed to persist async audit trail")
		}
	}(trail)

	return nil
}
