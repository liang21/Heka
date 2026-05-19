package project

import (
	"time"

	"github.com/liang21/heka/internal/domain/shared"
)

// tasks.md: T102 | spec.md: §4.3 项目 DTO

type CreateProjectReq struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
}

type ProjectResponse struct {
	ID          shared.ID        `json:"id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Members     []MemberResponse `json:"members"`
	Stats       ProjectStats     `json:"stats"`
	CreatedAt   time.Time        `json:"created_at"`
}

type MemberResponse struct {
	UserID    shared.ID `json:"user_id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	JoinedAt  time.Time `json:"joined_at"`
}

type ProjectStats struct {
	TestCaseCount int64 `json:"test_case_count"`
	PlanCount     int64 `json:"plan_count"`
	MemberCount   int   `json:"member_count"`
}

type AddMemberReq struct {
	UserID shared.ID `json:"user_id" validate:"required,uuid"`
}
