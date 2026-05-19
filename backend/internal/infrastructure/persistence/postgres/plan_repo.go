// tasks.md: T056 | TDD GREEN phase - TestPlanRepository implementation
package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/liang21/heka/internal/domain/plan"
	"github.com/liang21/heka/internal/domain/shared"
	"gorm.io/gorm"
)

// PlanRepository implements plan.TestPlanRepository using GORM
type PlanRepository struct {
	db *gorm.DB
}

// NewPlanRepository creates a new PlanRepository instance
func NewPlanRepository(db *gorm.DB) plan.TestPlanRepository {
	return &PlanRepository{db: db}
}

// Create creates a new test plan in the database
func (r *PlanRepository) Create(ctx context.Context, p *plan.TestPlan) error {
	db := DBOrTx(ctx, r.db)

	now := time.Now()
	if p.CreatedAt.IsZero() {
		p.CreatedAt = now
	}
	p.UpdatedAt = now

	type dbPlan struct {
		ID                string
		ProjectID         string
		Name              string
		Description       string
		Status            string
		CurrentExecutionID *string
		StartedAt         *time.Time
		PausedAt          *time.Time
		EndedAt           *time.Time
		CreatedBy         string
		CreatedAt         time.Time
		UpdatedAt         time.Time
		DeletedAt         *time.Time
	}

	newDbPlan := &dbPlan{
		ID:                p.ID.String(),
		ProjectID:         p.ProjectID.String(),
		Name:              p.Name,
		Description:       p.Description,
		Status:            string(p.Status),
		CurrentExecutionID: idPtrToString(p.CurrentExecutionID),
		StartedAt:         p.StartedAt,
		PausedAt:          p.PausedAt,
		EndedAt:           p.EndedAt,
		CreatedBy:         p.CreatedBy.String(),
		CreatedAt:         p.CreatedAt,
		UpdatedAt:         p.UpdatedAt,
		DeletedAt:         p.DeletedAt,
	}

	if err := db.Table("test_plans").Create(newDbPlan).Error; err != nil {
		return err
	}

	return nil
}

// FindByID retrieves a test plan by its ID, preloading associated test cases
func (r *PlanRepository) FindByID(ctx context.Context, id shared.ID) (*plan.TestPlan, error) {
	db := DBOrTx(ctx, r.db)

	type dbPlan struct {
		ID                string
		ProjectID         string
		Name              string
		Description       string
		Status            string
		CurrentExecutionID *string
		StartedAt         *time.Time
		PausedAt          *time.Time
		EndedAt           *time.Time
		CreatedBy         string
		CreatedAt         time.Time
		UpdatedAt         time.Time
		DeletedAt         *time.Time
	}

	var dbP dbPlan
	err := db.Table("test_plans").Where("id = ?", id.String()).First(&dbP).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, shared.ErrPlanNotFound
		}
		return nil, err
	}

	// Load associated cases
	type dbCase struct {
		PlanID     string
		TestCaseID string
		AssignedTo *string
		OrderIndex int
	}

	var dbCases []dbCase
	err = db.Table("plan_test_cases").
		Where("plan_id = ?", id.String()).
		Order("order_index ASC").
		Find(&dbCases).Error
	if err != nil {
		return nil, err
	}

	cases := make([]plan.PlanTestCase, len(dbCases))
	for i, c := range dbCases {
		planID, err := shared.ParseID(c.PlanID)
		if err != nil {
			return nil, err
		}
		testCaseID, err := shared.ParseID(c.TestCaseID)
		if err != nil {
			return nil, err
		}
		cases[i] = plan.PlanTestCase{
			PlanID:     planID,
			TestCaseID: testCaseID,
			AssignedTo: stringPtrToID(c.AssignedTo),
			OrderIndex: c.OrderIndex,
		}
	}

	planID, err := shared.ParseID(dbP.ID)
	if err != nil {
		return nil, err
	}
	projectID, err := shared.ParseID(dbP.ProjectID)
	if err != nil {
		return nil, err
	}
	createdBy, err := shared.ParseID(dbP.CreatedBy)
	if err != nil {
		return nil, err
	}

	return &plan.TestPlan{
		ID:                planID,
		ProjectID:         projectID,
		Name:              dbP.Name,
		Description:       dbP.Description,
		Status:            plan.PlanStatus(dbP.Status),
		CurrentExecutionID: stringPtrToID(dbP.CurrentExecutionID),
		StartedAt:         dbP.StartedAt,
		PausedAt:          dbP.PausedAt,
		EndedAt:           dbP.EndedAt,
		CreatedBy:         createdBy,
		CreatedAt:         dbP.CreatedAt,
		UpdatedAt:         dbP.UpdatedAt,
		DeletedAt:         dbP.DeletedAt,
		Cases:             cases,
	}, nil
}

// List retrieves test plans for a project with optional status filter and pagination
func (r *PlanRepository) List(ctx context.Context, projectID shared.ID, status *plan.PlanStatus, page, pageSize int) ([]*plan.TestPlan, int64, error) {
	db := DBOrTx(ctx, r.db)

	query := db.Table("test_plans").Where("project_id = ?", projectID.String())

	if status != nil {
		query = query.Where("status = ?", string(*status))
	}

	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Calculate offset
	offset := (page - 1) * pageSize

	// Fetch paginated results ordered by updated_at desc
	type dbPlan struct {
		ID                string
		ProjectID         string
		Name              string
		Description       string
		Status            string
		CurrentExecutionID *string
		StartedAt         *time.Time
		PausedAt          *time.Time
		EndedAt           *time.Time
		CreatedBy         string
		CreatedAt         time.Time
		UpdatedAt         time.Time
		DeletedAt         *time.Time
	}

	var dbPlans []dbPlan
	err := query.Order("updated_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&dbPlans).Error
	if err != nil {
		return nil, 0, err
	}

	plans := make([]*plan.TestPlan, len(dbPlans))
	for i, dbP := range dbPlans {
		planID, err := shared.ParseID(dbP.ID)
		if err != nil {
			return nil, 0, err
		}
		projectID, err := shared.ParseID(dbP.ProjectID)
		if err != nil {
			return nil, 0, err
		}
		createdBy, err := shared.ParseID(dbP.CreatedBy)
		if err != nil {
			return nil, 0, err
		}

		plans[i] = &plan.TestPlan{
			ID:                planID,
			ProjectID:         projectID,
			Name:              dbP.Name,
			Description:       dbP.Description,
			Status:            plan.PlanStatus(dbP.Status),
			CurrentExecutionID: stringPtrToID(dbP.CurrentExecutionID),
			StartedAt:         dbP.StartedAt,
			PausedAt:          dbP.PausedAt,
			EndedAt:           dbP.EndedAt,
			CreatedBy:         createdBy,
			CreatedAt:         dbP.CreatedAt,
			UpdatedAt:         dbP.UpdatedAt,
			DeletedAt:         dbP.DeletedAt,
			Cases:             []plan.PlanTestCase{}, // Empty for list, preload with FindByID if needed
		}
	}

	return plans, total, nil
}

// Update updates an existing test plan
func (r *PlanRepository) Update(ctx context.Context, p *plan.TestPlan) error {
	db := DBOrTx(ctx, r.db)

	p.UpdatedAt = time.Now()

	type dbPlan struct {
		Name              string
		Description       string
		Status            string
		CurrentExecutionID *string
		StartedAt         *time.Time
		PausedAt          *time.Time
		EndedAt           *time.Time
		UpdatedAt         time.Time
	}

	updates := &dbPlan{
		Name:              p.Name,
		Description:       p.Description,
		Status:            string(p.Status),
		CurrentExecutionID: idPtrToString(p.CurrentExecutionID),
		StartedAt:         p.StartedAt,
		PausedAt:          p.PausedAt,
		EndedAt:           p.EndedAt,
		UpdatedAt:         p.UpdatedAt,
	}

	result := db.Table("test_plans").
		Where("id = ?", p.ID.String()).
		Updates(updates)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return shared.ErrPlanNotFound
	}

	return nil
}

// AddCases adds test cases to a plan, handling duplicates and auto-increment order
func (r *PlanRepository) AddCases(ctx context.Context, planID shared.ID, cases []plan.PlanTestCase) error {
	db := DBOrTx(ctx, r.db)

	if len(cases) == 0 {
		return nil
	}

	// Verify plan exists
	var count int64
	if err := db.Table("test_plans").Where("id = ?", planID.String()).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return shared.ErrPlanNotFound
	}

	// Get current max order_index for this plan
	var maxOrder int
	type maxResult struct {
		Max int
	}
	var maxRes maxResult
	db.Table("plan_test_cases").
		Where("plan_id = ?", planID.String()).
		Select("COALESCE(MAX(order_index), -1) as max").
		Scan(&maxRes)
	maxOrder = maxRes.Max

	// Prepare records for insertion, handling duplicates
	type dbCase struct {
		PlanID     string
		TestCaseID string
		AssignedTo *string
		OrderIndex int
	}

	var dbCases []dbCase
	addedCases := make(map[shared.ID]bool)

	for _, c := range cases {
		// Skip if this test case is already added to this plan
		if addedCases[c.TestCaseID] {
			continue
		}

		// Check if case already exists in database for this plan
		var existingCount int64
		err := db.Table("plan_test_cases").
			Where("plan_id = ? AND test_case_id = ?", planID.String(), c.TestCaseID.String()).
			Count(&existingCount).Error
		if err != nil {
			return err
		}

		if existingCount > 0 {
			addedCases[c.TestCaseID] = true
			continue
		}

		// Auto-increment order_index if not explicitly set
		orderIndex := c.OrderIndex
		if orderIndex == 0 && maxOrder >= -1 {
			maxOrder++
			orderIndex = maxOrder
		}

		dbCases = append(dbCases, dbCase{
			PlanID:     planID.String(),
			TestCaseID: c.TestCaseID.String(),
			AssignedTo: idPtrToString(c.AssignedTo),
			OrderIndex: orderIndex,
		})

		addedCases[c.TestCaseID] = true
	}

	if len(dbCases) == 0 {
		return nil
	}

	// Batch insert
	if err := db.Table("plan_test_cases").Create(&dbCases).Error; err != nil {
		return err
	}

	return nil
}

// RemoveCases removes test cases from a plan
func (r *PlanRepository) RemoveCases(ctx context.Context, planID shared.ID, caseIDs []shared.ID) error {
	db := DBOrTx(ctx, r.db)

	if len(caseIDs) == 0 {
		return nil
	}

	// Verify plan exists
	var count int64
	if err := db.Table("test_plans").Where("id = ?", planID.String()).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return shared.ErrPlanNotFound
	}

	// Convert case IDs to strings
	idStrs := make([]string, len(caseIDs))
	for i, id := range caseIDs {
		idStrs[i] = id.String()
	}

	// Delete cases
	result := db.Table("plan_test_cases").
		Where("plan_id = ? AND test_case_id IN ?", planID.String(), idStrs).
		Delete(nil)

	if result.Error != nil {
		return result.Error
	}

	return nil
}

// Helper functions for ID pointer conversions

func idPtrToString(id *shared.ID) *string {
	if id == nil {
		return nil
	}
	s := id.String()
	return &s
}

func stringPtrToID(s *string) *shared.ID {
	if s == nil {
		return nil
	}
	id, err := shared.ParseID(*s)
	if err != nil {
		return nil
	}
	return &id
}
