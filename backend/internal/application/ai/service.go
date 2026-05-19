package ai

import (
	"context"
	"encoding/json"

	"github.com/liang21/heka/internal/application/rag"
	"github.com/liang21/heka/internal/domain/shared"
	"github.com/liang21/heka/internal/infrastructure/ai"
)

// tasks.md: T122 | spec.md: §4.11 AIService Implementation

type Service struct {
	manager  *ai.Manager
	ragSvc   *rag.Service
	taskRepo shared.AsyncTaskRepository
	eventBus shared.EventBus
}

func NewService(manager *ai.Manager, ragSvc *rag.Service, taskRepo shared.AsyncTaskRepository, eventBus shared.EventBus) *Service {
	return &Service{
		manager:  manager,
		ragSvc:   ragSvc,
		taskRepo: taskRepo,
		eventBus: eventBus,
	}
}

func (s *Service) GenerateTestCases(ctx context.Context, userID shared.ID, req GenerateRequest) (*TaskResponse, error) {
	input, _ := json.Marshal(map[string]interface{}{
		"file_id":  req.FileID,
		"count":    req.Count,
		"negative": req.IncludeNegative,
	})

	task := &shared.AsyncTask{
		ID:       shared.NewID(),
		ProjectID: req.ProjectID,
		Type:     "generate_testcases",
		Status:   "pending",
		Input:    json.RawMessage(input),
		CreatedBy: userID,
	}

	if err := s.taskRepo.Create(ctx, task); err != nil {
		return nil, err
	}

	return &TaskResponse{
		TaskID:  task.ID,
		Status:  task.Status,
		Progress: 0,
		Total:    1,
	}, nil
}

func (s *Service) Analyze(ctx context.Context, userID shared.ID, req AnalyzeRequest) (*TaskResponse, error) {
	input, _ := json.Marshal(map[string]interface{}{
		"project_id":  req.ProjectID,
		"description": req.Description,
	})

	task := &shared.AsyncTask{
		ID:       shared.NewID(),
		ProjectID: req.ProjectID,
		Type:     "analyze",
		Status:   "pending",
		Input:    json.RawMessage(input),
		CreatedBy: userID,
	}

	if err := s.taskRepo.Create(ctx, task); err != nil {
		return nil, err
	}

	return &TaskResponse{
		TaskID:  task.ID,
		Status:  task.Status,
		Progress: 0,
		Total:    1,
	}, nil
}

func (s *Service) GetTask(ctx context.Context, taskID shared.ID) (*TaskResponse, error) {
	task, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		return nil, err
	}

	return &TaskResponse{
		TaskID:  task.ID,
		Status:  task.Status,
		Progress: task.ProgressCurrent,
		Total:    task.ProgressTotal,
		Result:   task.Result,
		Error:    task.Error,
	}, nil
}
