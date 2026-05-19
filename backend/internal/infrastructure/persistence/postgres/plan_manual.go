//go:build ignore
// +build ignore

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/liang21/heka/internal/domain/plan"
	"github.com/liang21/heka/internal/domain/shared"
	"github.com/liang21/heka/internal/infrastructure/persistence/postgres"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	dsn := "host=localhost port=5432 user=test_user password=test_pass dbname=test_db sslmode=disable"

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	// Create tables
	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS test_plans (
			id VARCHAR(36) PRIMARY KEY,
			project_id VARCHAR(36) NOT NULL,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			status VARCHAR(50) NOT NULL,
			current_execution_id VARCHAR(36),
			started_at TIMESTAMP,
			paused_at TIMESTAMP,
			ended_at TIMESTAMP,
			created_by VARCHAR(36) NOT NULL,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL,
			deleted_at TIMESTAMP
		);
		CREATE TABLE IF NOT EXISTS plan_test_cases (
			plan_id VARCHAR(36) NOT NULL,
			test_case_id VARCHAR(36) NOT NULL,
			assigned_to VARCHAR(36),
			order_index INTEGER NOT NULL DEFAULT 0,
			PRIMARY KEY (plan_id, test_case_id)
		);
		CREATE INDEX IF NOT EXISTS idx_test_plans_project_id ON test_plans(project_id);
		CREATE INDEX IF NOT EXISTS idx_test_plans_status ON test_plans(status);
	`).Error
	if err != nil {
		log.Fatalf("Failed to create tables: %v", err)
	}

	// Clean tables
	db.Exec("DELETE FROM plan_test_cases")
	db.Exec("DELETE FROM test_plans")

	// Create repository
	repo := postgres.NewPlanRepository(db)
	ctx := context.Background()

	// Test Create
	projectID := shared.NewID()
	createdBy := shared.NewID()
	testPlan := &plan.TestPlan{
		ID:          shared.NewID(),
		ProjectID:   projectID,
		Name:        "Test Plan",
		Description: "Test Description",
		Status:      plan.PlanDraft,
		CreatedBy:   createdBy,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err = repo.Create(ctx, testPlan)
	if err != nil {
		log.Fatalf("Create failed: %v", err)
	}
	fmt.Println("✓ Create successful")

	// Test FindByID
	found, err := repo.FindByID(ctx, testPlan.ID)
	if err != nil {
		log.Fatalf("FindByID failed: %v", err)
	}
	if found.ID != testPlan.ID {
		log.Fatalf("FindByID returned wrong ID")
	}
	fmt.Println("✓ FindByID successful")

	// Test List
	plans, total, err := repo.List(ctx, projectID, nil, 1, 10)
	if err != nil {
		log.Fatalf("List failed: %v", err)
	}
	if len(plans) != 1 || total != 1 {
		log.Fatalf("List returned wrong count: %d, %d", len(plans), total)
	}
	fmt.Println("✓ List successful")

	// Test Update
	testPlan.Name = "Updated Test Plan"
	err = repo.Update(ctx, testPlan)
	if err != nil {
		log.Fatalf("Update failed: %v", err)
	}
	fmt.Println("✓ Update successful")

	// Test AddCases
	case1ID := shared.NewID()
	case2ID := shared.NewID()
	cases := []plan.PlanTestCase{
		{PlanID: testPlan.ID, TestCaseID: case1ID, AssignedTo: &createdBy, OrderIndex: 0},
		{PlanID: testPlan.ID, TestCaseID: case2ID, AssignedTo: nil, OrderIndex: 1},
	}
	err = repo.AddCases(ctx, testPlan.ID, cases)
	if err != nil {
		log.Fatalf("AddCases failed: %v", err)
	}
	fmt.Println("✓ AddCases successful")

	// Test FindByID with cases
	foundWithCases, err := repo.FindByID(ctx, testPlan.ID)
	if err != nil {
		log.Fatalf("FindByID with cases failed: %v", err)
	}
	if len(foundWithCases.Cases) != 2 {
		log.Fatalf("FindByID with cases returned wrong count: %d", len(foundWithCases.Cases))
	}
	fmt.Println("✓ FindByID with cases successful")

	// Test RemoveCases
	err = repo.RemoveCases(ctx, testPlan.ID, []shared.ID{case1ID})
	if err != nil {
		log.Fatalf("RemoveCases failed: %v", err)
	}
	fmt.Println("✓ RemoveCases successful")

	// Verify removal
	foundAfterRemoval, err := repo.FindByID(ctx, testPlan.ID)
	if err != nil {
		log.Fatalf("FindByID after removal failed: %v", err)
	}
	if len(foundAfterRemoval.Cases) != 1 {
		log.Fatalf("FindByID after removal returned wrong count: %d", len(foundAfterRemoval.Cases))
	}
	fmt.Println("✓ RemoveCases verification successful")

	fmt.Println("\n✅ All manual tests passed!")
}
