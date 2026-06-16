package asset

import (
	"context"
	"io"

	"zerobin-be/pkg/storage"

	"github.com/google/uuid"
)

type Service interface {
	Upload(ctx context.Context, filename string, folder string, reader io.Reader, size int64, contentType string, uploadedByID *uuid.UUID) (*Asset, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Asset, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type service struct {
	repo    Repository
	storage storage.Storage
}

func NewService(repo Repository, storage storage.Storage) Service {
	return &service{
		repo:    repo,
		storage: storage,
	}
}

func (s *service) Upload(ctx context.Context, filename string, folder string, reader io.Reader, size int64, contentType string, uploadedByID *uuid.UUID) (*Asset, error) {
	fileURL, err := s.storage.UploadFile(ctx, folder, filename, reader, size, contentType)
	if err != nil {
		return nil, err
	}

	a := &Asset{
		FileName:     filename,
		OriginalName: filename,
		FilePath:     folder + "/" + filename,
		FileURL:      fileURL,
		FileSize:     size,
		ContentType:  contentType,
		UploadedByID: uploadedByID,
	}

	if err := s.repo.Create(ctx, a); err != nil {
		return nil, err
	}

	return a, nil
}

func (s *service) GetByID(ctx context.Context, id uuid.UUID) (*Asset, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *service) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
