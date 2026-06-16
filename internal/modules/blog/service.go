package blog

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"zerobin-be/internal/modules/audit"

	"github.com/google/uuid"
)

type Service interface {
	Create(ctx context.Context, req *CreateBlogRequest, authorID string) (*BlogResponse, error)
	GetAll(ctx context.Context, category string, authorID string, search string) ([]BlogResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*BlogResponse, error)
	GetBySlug(ctx context.Context, slug string) (*BlogResponse, error)
	Update(ctx context.Context, id uuid.UUID, req *UpdateBlogRequest, operatorID string) (*BlogResponse, error)
	Delete(ctx context.Context, id uuid.UUID, operatorID string) error
}

type service struct {
	repo      Repository
	auditServ audit.Service
}

func NewService(repo Repository, auditServ audit.Service) Service {
	return &service{
		repo:      repo,
		auditServ: auditServ,
	}
}

func (s *service) Create(ctx context.Context, req *CreateBlogRequest, authorID string) (*BlogResponse, error) {
	authorUUID, err := uuid.Parse(authorID)
	if err != nil {
		return nil, fmt.Errorf("invalid author ID: %w", err)
	}

	baseSlug := slugify(req.Title)
	slug := baseSlug
	counter := 1

	for {
		existing, _ := s.repo.FindBySlug(ctx, slug)
		if existing == nil {
			break
		}
		slug = fmt.Sprintf("%s-%d", baseSlug, counter)
		counter++
	}

	b := &Blog{
		UserID:   &authorUUID,
		Title:    req.Title,
		Slug:     slug,
		Desc:     req.Desc,
		Category: req.Category,
	}

	if err := s.repo.Create(ctx, b); err != nil {
		return nil, err
	}

	loadedBlog, err := s.repo.FindByID(ctx, b.ID)
	if err == nil {
		b = loadedBlog
	}

	res := ToBlogResponse(b)
	return &res, nil
}

func (s *service) GetAll(ctx context.Context, category string, authorID string, search string) ([]BlogResponse, error) {
	blogs, err := s.repo.FindAll(ctx, category, authorID, search)
	if err != nil {
		return nil, err
	}
	return ToBlogResponses(blogs), nil
}

func (s *service) GetByID(ctx context.Context, id uuid.UUID) (*BlogResponse, error) {
	b, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("blog post not found")
	}
	res := ToBlogResponse(b)
	return &res, nil
}

func (s *service) GetBySlug(ctx context.Context, slug string) (*BlogResponse, error) {
	b, err := s.repo.FindBySlug(ctx, slug)
	if err != nil {
		return nil, errors.New("blog post not found")
	}
	res := ToBlogResponse(b)
	return &res, nil
}

func (s *service) Update(ctx context.Context, id uuid.UUID, req *UpdateBlogRequest, operatorID string) (*BlogResponse, error) {
	b, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("blog post not found")
	}

	if req.Title != "" && req.Title != b.Title {
		b.Title = req.Title
		baseSlug := slugify(req.Title)
		slug := baseSlug
		counter := 1

		for {
			existing, _ := s.repo.FindBySlug(ctx, slug)
			if existing == nil || existing.ID == b.ID {
				break
			}
			slug = fmt.Sprintf("%s-%d", baseSlug, counter)
			counter++
		}
		b.Slug = slug
	}

	if req.Desc != "" {
		b.Desc = req.Desc
	}

	if req.Category != "" {
		b.Category = req.Category
	}

	if err := s.repo.Update(ctx, b); err != nil {
		return nil, err
	}

	res := ToBlogResponse(b)
	return &res, nil
}

func (s *service) Delete(ctx context.Context, id uuid.UUID, operatorID string) error {
	_, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return errors.New("blog post not found")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	return nil
}

func slugify(s string) string {
	s = strings.ToLower(s)
	var sb strings.Builder
	lastWasHyphen := false

	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			sb.WriteRune(r)
			lastWasHyphen = false
		} else if r == ' ' || r == '-' || r == '_' {
			if !lastWasHyphen && sb.Len() > 0 {
				sb.WriteRune('-')
				lastWasHyphen = true
			}
		}
	}

	res := sb.String()
	res = strings.TrimSuffix(res, "-")
	if res == "" {
		return "post"
	}
	return res
}
