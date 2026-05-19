package project

import (
	"context"

	"github.com/liang21/heka/internal/domain/project"
	"github.com/liang21/heka/internal/domain/shared"
)

// tasks.md: T104 | spec.md: §4.3 ProjectService Implementation

type Service struct {
	repo project.ProjectRepository
}

func NewService(repo project.ProjectRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, userID shared.ID, req CreateProjectReq) (*ProjectResponse, error) {
	p := &project.Project{
		ID:          shared.NewID(),
		Name:        req.Name,
		Description: req.Description,
		CreatedBy:   userID,
	}

	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}

	return s.GetByID(ctx, p.ID)
}

func (s *Service) GetByID(ctx context.Context, projectID shared.ID) (*ProjectResponse, error) {
	p, err := s.repo.FindByID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	count, err := s.repo.CountMembers(ctx, projectID)
	if err != nil {
		return nil, err
	}

	return &ProjectResponse{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Stats: ProjectStats{
			MemberCount: int(count),
		},
		CreatedAt:   p.CreatedAt,
	}, nil
}

func (s *Service) ListByUser(ctx context.Context, userID shared.ID) ([]*ProjectResponse, error) {
	projects, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	resp := make([]*ProjectResponse, len(projects))
	for i, p := range projects {
		count, _ := s.repo.CountMembers(ctx, p.ID)
		resp[i] = &ProjectResponse{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Stats: ProjectStats{
				MemberCount: int(count),
			},
			CreatedAt:   p.CreatedAt,
		}
	}

	return resp, nil
}

func (s *Service) AddMember(ctx context.Context, projectID, creatorID shared.ID, req AddMemberReq) error {
	isMember, err := s.repo.IsMember(ctx, projectID, req.UserID)
	if err != nil {
		return err
	}
	if isMember {
		return shared.ErrProjectNotMember // Already a member
	}

	member := &project.ProjectMember{
		ProjectID: projectID,
		UserID:    req.UserID,
	}

	return s.repo.AddMember(ctx, member)
}
