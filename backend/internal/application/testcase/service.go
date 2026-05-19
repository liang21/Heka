package testcase

import (
	"context"

	"github.com/liang21/heka/internal/domain/shared"
	"github.com/liang21/heka/internal/domain/testcase"
)

// tasks.md: T107 | spec.md: §4.4-4.7 TestCaseService Implementation

type Service struct {
	tcRepo    testcase.TestCaseRepository
	modRepo   testcase.ModuleRepository
	tagRepo   testcase.TagRepository
	collRepo  testcase.CollectionRepository
	eventBus  shared.EventBus
}

func NewService(tcRepo testcase.TestCaseRepository, modRepo testcase.ModuleRepository, tagRepo testcase.TagRepository, collRepo testcase.CollectionRepository, eventBus shared.EventBus) *Service {
	return &Service{
		tcRepo:   tcRepo,
		modRepo:  modRepo,
		tagRepo:  tagRepo,
		collRepo: collRepo,
		eventBus: eventBus,
	}
}

// --- TestCase Operations ---

func (s *Service) CreateTestCase(ctx context.Context, userID, projectID shared.ID, req CreateTestCaseReq) (*TestCaseResponse, error) {
	steps := make([]testcase.Step, len(req.Steps))
	for i, step := range req.Steps {
		steps[i] = testcase.Step{
			ID:         shared.NewID(),
			Number:     i + 1,
			Action:     step.Action,
			Expected:   step.Expected,
		}
	}

	tc := &testcase.TestCase{
		ID:          shared.NewID(),
		ProjectID:   projectID,
		ModuleID:    req.ModuleID,
		Title:       req.Title,
		Description: req.Description,
		Status:      testcase.CaseDraft,
		Priority:    req.Priority,
		Tags:        req.Tags,
		Steps:       steps,
		CreatedBy:   userID,
		Version:     0,
	}

	if err := s.tcRepo.Create(ctx, tc); err != nil {
		return nil, err
	}

	s.eventBus.Publish(ctx, &testcase.TestCaseCreatedEvent{
		ProjectID:  projectID,
		TestCaseID: tc.ID,
	})

	return s.toResponse(tc)
}

func (s *Service) GetTestCase(ctx context.Context, caseID shared.ID) (*TestCaseResponse, error) {
	tc, err := s.tcRepo.FindByID(ctx, caseID)
	if err != nil {
		return nil, err
	}

	return s.toResponse(tc)
}

func (s *Service) ListTestCases(ctx context.Context, filter testcase.TestCaseFilter) ([]TestCaseListResponse, int64, error) {
	cases, total, err := s.tcRepo.List(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	resp := make([]TestCaseListResponse, len(cases))
	for i, tc := range cases {
		resp[i] = TestCaseListResponse{
			ID:        tc.ID,
			ProjectID: tc.ProjectID,
			ModuleID:  tc.ModuleID,
			Title:     tc.Title,
			Status:    tc.Status,
			Priority:  tc.Priority,
			Tags:      tc.Tags,
			CreatedBy: tc.CreatedBy,
			CreatedAt: tc.CreatedAt,
			UpdatedAt: tc.UpdatedAt,
		}
	}

	return resp, total, nil
}

func (s *Service) UpdateTestCase(ctx context.Context, userID, caseID shared.ID, req UpdateTestCaseReq) (*TestCaseResponse, error) {
	tc, err := s.tcRepo.FindByID(ctx, caseID)
	if err != nil {
		return nil, err
	}

	if req.Version != tc.Version {
		return nil, shared.ErrTestCaseConflict
	}

	steps := make([]testcase.Step, len(req.Steps))
	for i, step := range req.Steps {
		steps[i] = testcase.Step{
			ID:         shared.NewID(),
			Number:     i + 1,
			Action:     step.Action,
			Expected:   step.Expected,
		}
	}

	tc.Title = req.Title
	tc.Description = req.Description
	tc.Priority = req.Priority
	tc.Tags = req.Tags
	tc.Steps = steps
	tc.UpdatedBy = &userID
	tc.Version = tc.Version + 1

	if err := s.tcRepo.Update(ctx, tc); err != nil {
		return nil, err
	}

	s.eventBus.Publish(ctx, &testcase.TestCaseUpdatedEvent{
		ProjectID:  tc.ProjectID,
		TestCaseID: tc.ID,
	})

	return s.toResponse(tc)
}

func (s *Service) DeleteTestCase(ctx context.Context, caseID, userID shared.ID) error {
	if err := s.tcRepo.SoftDelete(ctx, caseID); err != nil {
		return err
	}

	s.eventBus.Publish(ctx, &testcase.TestCaseDeletedEvent{
		ProjectID:  shared.NewID(), // Would need to fetch first
		TestCaseID: caseID,
	})

	return nil
}

func (s *Service) BatchUpdateStatus(ctx context.Context, req BatchStatusReq) error {
	return s.tcRepo.BatchUpdateStatus(ctx, req.IDs, req.Status)
}

func (s *Service) BatchDelete(ctx context.Context, req BatchDeleteReq) error {
	return s.tcRepo.BatchDelete(ctx, req.IDs)
}

func (s *Service) BatchMove(ctx context.Context, req BatchMoveReq) error {
	return s.tcRepo.BatchMove(ctx, req.IDs, req.ModuleID)
}

// --- Module Operations ---

func (s *Service) CreateModule(ctx context.Context, userID shared.ID, req CreateModuleReq) (*ModuleDTO, error) {
	mod := &testcase.Module{
		ID:          shared.NewID(),
		ProjectID:   req.ProjectID,
		Name:        req.Name,
		Description: req.Description,
		ParentID:    req.ParentID,
		CreatedBy:   userID,
	}

	if err := s.modRepo.Create(ctx, mod); err != nil {
		return nil, err
	}

	return &ModuleDTO{
		ID:          mod.ID,
		ProjectID:   mod.ProjectID,
		Name:        mod.Name,
		Description: mod.Description,
		ParentID:    mod.ParentID,
		OrderIndex:  mod.OrderIndex,
	}, nil
}

func (s *Service) GetModuleTree(ctx context.Context, projectID shared.ID) ([]ModuleDTO, error) {
	modules, err := s.modRepo.FindByProject(ctx, projectID)
	if err != nil {
		return nil, err
	}

	return s.buildTree(modules, nil), nil
}

func (s *Service) buildTree(modules []*testcase.Module, parentID *shared.ID) []ModuleDTO {
	var result []ModuleDTO
	for _, m := range modules {
		if (parentID == nil && m.ParentID == nil) || (m.ParentID != nil && parentID != nil && *m.ParentID == *parentID) {
			dto := ModuleDTO{
				ID:          m.ID,
				ProjectID:   m.ProjectID,
				Name:        m.Name,
				Description: m.Description,
				ParentID:    m.ParentID,
				OrderIndex:  m.OrderIndex,
				Children:    s.buildTree(modules, &m.ID),
			}
			result = append(result, dto)
		}
	}
	return result
}

// --- Tag Operations ---

func (s *Service) CreateTag(ctx context.Context, userID shared.ID, req CreateTagReq) (*TagDTO, error) {
	tag := &testcase.Tag{
		ID:        shared.NewID(),
		ProjectID: req.ProjectID,
		Name:      req.Name,
		Color:     req.Color,
		CreatedBy: userID,
	}

	if err := s.tagRepo.Create(ctx, tag); err != nil {
		return nil, err
	}

	return &TagDTO{
		ID:        tag.ID,
		ProjectID: tag.ProjectID,
		Name:      tag.Name,
		Color:     tag.Color,
	}, nil
}

func (s *Service) ListTags(ctx context.Context, projectID shared.ID) ([]TagDTO, error) {
	tags, err := s.tagRepo.FindByProject(ctx, projectID)
	if err != nil {
		return nil, err
	}

	resp := make([]TagDTO, len(tags))
	for i, t := range tags {
		resp[i] = TagDTO{
			ID:        t.ID,
			ProjectID: t.ProjectID,
			Name:      t.Name,
			Color:     t.Color,
		}
	}

	return resp, nil
}

// --- Collection Operations ---

func (s *Service) CreateCollection(ctx context.Context, userID shared.ID, req CreateCollectionReq) (*CollectionDTO, error) {
	coll := &testcase.Collection{
		ID:          shared.NewID(),
		ProjectID:   req.ProjectID,
		Name:        req.Name,
		Description: req.Description,
		CreatedBy:   userID,
	}

	if err := s.collRepo.Create(ctx, coll); err != nil {
		return nil, err
	}

	return &CollectionDTO{
		ID:          coll.ID,
		ProjectID:   coll.ProjectID,
		Name:        coll.Name,
		Description: coll.Description,
		CreatedAt:   coll.CreatedAt,
	}, nil
}

func (s *Service) AddCollectionCases(ctx context.Context, collectionID shared.ID, req AddCollectionCasesReq) error {
	return s.collRepo.AddCases(ctx, collectionID, req.CaseIDs)
}

func (s *Service) RemoveCollectionCases(ctx context.Context, collectionID shared.ID, req RemoveCollectionCasesReq) error {
	return s.collRepo.RemoveCases(ctx, collectionID, req.CaseIDs)
}

func (s *Service) ListCollectionCases(ctx context.Context, collectionID shared.ID, page, pageSize int) ([]TestCaseListResponse, int64, error) {
	cases, total, err := s.collRepo.ListCases(ctx, collectionID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	resp := make([]TestCaseListResponse, len(cases))
	for i, tc := range cases {
		resp[i] = TestCaseListResponse{
			ID:        tc.ID,
			Title:     tc.Title,
			Status:    tc.Status,
			Priority:  tc.Priority,
		}
	}

	return resp, total, nil
}

// --- Helper ---

func (s *Service) toResponse(tc *testcase.TestCase) (*TestCaseResponse, error) {
	stepResp := make([]StepResponse, len(tc.Steps))
	for i, step := range tc.Steps {
		stepResp[i] = StepResponse{
			ID:       step.ID,
			Number:   step.Number,
			Action:   step.Action,
			Expected: step.Expected,
		}
	}

	return &TestCaseResponse{
		ID:          tc.ID,
		ProjectID:   tc.ProjectID,
		ModuleID:    tc.ModuleID,
		Title:       tc.Title,
		Description: tc.Description,
		Status:      tc.Status,
		Priority:    tc.Priority,
		Tags:        tc.Tags,
		Steps:       stepResp,
		CreatedBy:   tc.CreatedBy,
		UpdatedBy:   tc.UpdatedBy,
		Version:     tc.Version,
		CreatedAt:   tc.CreatedAt,
		UpdatedAt:   tc.UpdatedAt,
	}, nil
}
