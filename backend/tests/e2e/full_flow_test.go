package e2e

// tasks.md: T156 | spec.md: §4.1-4.13 E2E Full Flow Test
// RED: This test requires a fully wired application with DI, router, and real DB.
// The main.go DI wiring (T150) and router (T148) are not yet complete.

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/liang21/heka/internal/application/ai"
	appexecution "github.com/liang21/heka/internal/application/execution"
	appfile "github.com/liang21/heka/internal/application/file"
	appplan "github.com/liang21/heka/internal/application/plan"
	appproject "github.com/liang21/heka/internal/application/project"
	apprag "github.com/liang21/heka/internal/application/rag"
	appreport "github.com/liang21/heka/internal/application/report"
	apptestcase "github.com/liang21/heka/internal/application/testcase"
	appuser "github.com/liang21/heka/internal/application/user"
	"github.com/liang21/heka/internal/domain/shared"
	"github.com/liang21/heka/internal/infrastructure/auth"
	"github.com/liang21/heka/internal/infrastructure/event"
	"github.com/liang21/heka/internal/infrastructure/persistence/postgres"
	"github.com/liang21/heka/internal/interface/http/handler"
)

// apiResponse is the unified JSON response structure (spec.md §4.1)
type apiResponse struct {
	Code     interface{}    `json:"code"`
	Data     json.RawMessage `json:"data"`
	Message  string         `json:"message"`
	Total    int64          `json:"total,omitempty"`
	Page     int            `json:"page,omitempty"`
	PageSize int            `json:"page_size,omitempty"`
}

// testEnv holds all wired dependencies for E2E tests
type testEnv struct {
	db      *gorm.DB
	router  *chi.Mux
	server  *httptest.Server
	cleanup func()
}

// authMiddleware validates JWT and injects user_id into context
func e2eAuthMiddleware(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenStr := r.Header.Get("Authorization")
			if len(tokenStr) > 7 && tokenStr[:7] == "Bearer " {
				tokenStr = tokenStr[7:]
			}
			if tokenStr == "" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			claims, err := auth.ParseToken(secret, tokenStr)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// setupE2E wires the entire application with testcontainers PostgreSQL
func setupE2E(t *testing.T) *testEnv {
	t.Helper()
	ctx := context.Background()

	// AC: Start testcontainers PostgreSQL
	pgContainer, err := tcpostgres.Run(ctx,
		"postgres:15-alpine",
		tcpostgres.WithDatabase("heka_e2e_test"),
		tcpostgres.WithUsername("heka"),
		tcpostgres.WithPassword("heka_test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	require.NoError(t, err, "AC: PostgreSQL testcontainers must start")

	dsn, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err, "AC: Must obtain DB connection string")

	db, err := gorm.Open(gormpostgres.Open(dsn), &gorm.Config{})
	require.NoError(t, err, "AC: Must connect to test database")

	// AC: Run all migrations (T029-T038)
	migrationsDir := filepath.Join("..", "..", "scripts", "migration")
	migrationFiles := []string{
		"000001_init_schema.up.sql",
		"000002_modules.up.sql",
		"000003_tags.up.sql",
		"000004_testcases.up.sql",
		"000005_collections.up.sql",
		"000006_plans.up.sql",
		"000007_executions.up.sql",
		"000008_files.up.sql",
		"000009_rag.up.sql",
		"000010_ai_tasks.up.sql",
	}
	sqlDB, err := db.DB()
	require.NoError(t, err)
	for _, mf := range migrationFiles {
		content, err := os.ReadFile(filepath.Join(migrationsDir, mf))
		require.NoError(t, err, "AC: Migration file %s must exist", mf)
		_, err = sqlDB.Exec(string(content))
		require.NoError(t, err, "AC: Migration %s must execute successfully", mf)
	}

	// AC: Wire DI — Repos
	userRepo := postgres.NewUserRepository(db)
	projectRepo := postgres.NewProjectRepository(db)
	moduleRepo := postgres.NewModuleRepository(db)
	tagRepo := postgres.NewTagRepository(db)
	testcaseRepo := postgres.NewTestCaseRepository(db)
	collectionRepo := postgres.NewCollectionRepository(db)
	planRepo := postgres.NewPlanRepository(db)
	executionRepo := postgres.NewExecutionRepository(db)
	fileRepo := postgres.NewFileRepository(db)
	chunkRepo := postgres.NewChunkRepository(db)
	asyncTaskRepo := postgres.NewAsyncTaskRepository(db)

	// AC: Wire DI — EventBus
	eventBus := event.NewEventBus(4)

	// AC: Wire DI — Services
	jwtSecret := "e2e-test-jwt-secret-at-least-32chars"
	// Create auth service for auth handler (infrastructure layer)
	passwordHasher := auth.NewBCryptPasswordHasher()
	jwtMaker := auth.NewJWTMaker(jwtSecret, auth.TokenTTL{Duration: int((24 * time.Hour).Seconds())}, auth.TokenTTL{Duration: int((7 * 24 * time.Hour).Seconds())})
	authSvc := auth.NewService(userRepo, passwordHasher, jwtMaker)
	// Application services
	userSvc := appuser.NewService(userRepo, jwtSecret, 24*time.Hour, 7*24*time.Hour)
	projectSvc := appproject.NewService(projectRepo)
	testcaseSvc := apptestcase.NewService(testcaseRepo, moduleRepo, tagRepo, collectionRepo, eventBus)
	planSvc := appplan.NewService(planRepo)
	executionSvc := appexecution.NewService(executionRepo)
	fileSvc := appfile.NewService(fileRepo, nil, nil, eventBus, 100*1024*1024)
	ragSvc := apprag.NewService(nil, nil)
	aiSvc := ai.NewService(nil, ragSvc, asyncTaskRepo, eventBus)
	reportSvc := appreport.NewService(planRepo, executionRepo, testcaseRepo, userRepo)

	// AC: Wire DI — Handlers
	authHandler := handler.NewAuthHandler(authSvc)
	projectHandler := handler.NewProjectHandler(projectSvc)
	testcaseHandler := handler.NewTestCaseHandler(testcaseSvc)
	planHandler := handler.NewPlanHandler(planSvc)
	executionHandler := handler.NewExecutionHandler(executionSvc)
	fileHandler := handler.NewFileHandler(fileSvc, 100*1024*1024)
	aiHandler := handler.NewAIHandler(aiSvc)
	reportHandler := handler.NewReportHandler(reportSvc)

	// AC: Wire DI — Router (spec.md §4.1-4.13)
	r := chi.NewRouter()

	// Auth routes (no auth required)
	r.Route("/api/v1/auth", func(r chi.Router) {
		r.Post("/login", authHandler.Login)
		r.Post("/register", authHandler.Register)
		r.Post("/refresh", authHandler.RefreshToken)
		r.Post("/logout", authHandler.Logout)
		r.Get("/me", authHandler.GetMe)
	})

	// Authenticated routes
	r.Route("/api/v1", func(r chi.Router) {
		r.Use(e2eAuthMiddleware(jwtSecret))

		// Projects (spec.md §4.3)
		r.Route("/projects", func(r chi.Router) {
			r.Post("/", projectHandler.Create)
			r.Get("/", projectHandler.List)
			r.Get("/{id}", projectHandler.GetByID)
			r.Post("/{id}/members", projectHandler.AddMember)
		})

		// Modules (spec.md §4.4)
		r.Route("/modules", func(r chi.Router) {
			r.Post("/", testcaseHandler.CreateModule)
			r.Get("/", testcaseHandler.GetModuleTree)
			r.Put("/{id}", testcaseHandler.UpdateModule)
			r.Delete("/{id}", testcaseHandler.DeleteModule)
		})

		// Tags (spec.md §4.6)
		r.Route("/tags", func(r chi.Router) {
			r.Get("/", testcaseHandler.ListTags)
			r.Post("/", testcaseHandler.CreateTag)
		})

		// Test Cases (spec.md §4.5)
		r.Route("/testcases", func(r chi.Router) {
			r.Post("/", testcaseHandler.CreateTestCase)
			r.Get("/", testcaseHandler.ListTestCases)
			r.Get("/{id}", testcaseHandler.GetTestCase)
			r.Put("/{id}", testcaseHandler.UpdateTestCase)
			r.Delete("/{id}", testcaseHandler.DeleteTestCase)
		})

		// Collections (spec.md §4.7)
		r.Route("/collections", func(r chi.Router) {
			r.Post("/", testcaseHandler.CreateCollection)
			r.Post("/{id}/cases", testcaseHandler.AddToCollection)
			r.Get("/{id}/cases", testcaseHandler.ListCollections)
			r.Delete("/{id}/cases", testcaseHandler.RemoveFromCollection)
		})

		// Plans (spec.md §4.8)
		r.Route("/testplans", func(r chi.Router) {
			r.Post("/", planHandler.Create)
			r.Get("/", planHandler.List)
			r.Get("/{id}", planHandler.GetByID)
			r.Post("/{id}/start", planHandler.Start)
			r.Post("/{id}/pause", planHandler.Pause)
			r.Post("/{id}/resume", planHandler.Resume)
			r.Post("/{id}/complete", planHandler.Complete)
			r.Post("/{id}/cancel", planHandler.Cancel)
		})

		// Executions (spec.md §4.9)
		r.Route("/executions", func(r chi.Router) {
			r.Post("/", executionHandler.Create)
			r.Get("/{id}", executionHandler.GetByID)
			r.Get("/{id}/summary", executionHandler.GetSummary)
			r.Post("/{id}/results", executionHandler.SubmitResult)
			r.Post("/{id}/results/batch", executionHandler.BatchSubmit)
		})

		// Files (spec.md §4.10)
		r.Route("/files", func(r chi.Router) {
			r.Post("/upload", fileHandler.Upload)
			r.Get("/", fileHandler.List)
			r.Get("/{id}", fileHandler.GetByID)
			r.Post("/{id}/reindex", fileHandler.Reindex)
			r.Get("/{id}/index-status", fileHandler.GetIndexStatus)
			r.Delete("/{id}", fileHandler.Delete)
		})

		// AI (spec.md §4.11)
		r.Route("/ai", func(r chi.Router) {
			r.Post("/generate-testcases", aiHandler.GenerateTestCases)
			r.Get("/tasks/{id}", aiHandler.GetTask)
			r.Post("/analyze", aiHandler.Analyze)
		})

		// Reports (spec.md §4.12)
		r.Route("/reports", func(r chi.Router) {
			r.Get("/plan/{id}", reportHandler.PlanReport)
			r.Get("/coverage", reportHandler.Coverage)
			r.Get("/trend", reportHandler.Trend)
			r.Get("/bugs", reportHandler.BugDistribution)
			r.Get("/workload", reportHandler.Workload)
		})
	})

	server := httptest.NewServer(r)

	cleanup := func() {
		server.Close()
		// EventBus goroutines will be cleaned up on process exit
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
		pgContainer.Terminate(ctx)
	}

	_ = chunkRepo // used by workers, not directly in handlers

	return &testEnv{
		db:      db,
		router:  r,
		server:  server,
		cleanup: cleanup,
	}
}

// doRequest is a helper to make HTTP requests and parse responses
func doRequest(t *testing.T, env *testEnv, method, path string, body interface{}, token string) (*http.Response, *apiResponse) {
	t.Helper()
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		require.NoError(t, err)
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, env.server.URL+path, bodyReader)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	t.Cleanup(func() { resp.Body.Close() })

	var apiResp apiResponse
	if resp.StatusCode != http.StatusNoContent {
		err = json.NewDecoder(resp.Body).Decode(&apiResp)
		require.NoError(t, err)
	}
	return resp, &apiResp
}

// TestE2EFullFlow exercises the complete user journey through the API
func TestE2EFullFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	env := setupE2E(t)
	defer env.cleanup()

	var (
		token      string
		projectID  shared.ID
		moduleID   shared.ID
		tagID      shared.ID
		testcaseID shared.ID
		planID     shared.ID
		execID     shared.ID
	)

	// ── Step 1: Seed a test user directly into DB ──
	// AC (spec.md §7.1): Users stored with bcrypt hashed passwords
	t.Run("Step1_SeedUser", func(t *testing.T) {
		hash, err := auth.HashPassword("test-password-123")
		require.NoError(t, err)

		userID := shared.NewID()
		sqlDB, err := env.db.DB()
		require.NoError(t, err)
		_, err = sqlDB.Exec(
			`INSERT INTO users (id, name, email, password_hash) VALUES ($1, $2, $3, $4)`,
			userID.String(), "E2E Test User", "e2e@test.heka.io", hash,
		)
		require.NoError(t, err, "AC: Must seed test user into database")
	})

	// ── Step 2: Login (spec.md §4.2) ──
	// AC: POST /api/v1/auth/login returns access_token + refresh_token
	t.Run("Step2_Login", func(t *testing.T) {
		resp, apiResp := doRequest(t, env, "POST", "/api/v1/auth/login", map[string]string{
			"email":    "e2e@test.heka.io",
			"password": "test-password-123",
		}, "")

		assert.Equal(t, http.StatusOK, resp.StatusCode, "AC: Login should return 200")
		assert.Equal(t, float64(0), apiResp.Code, "AC: Success code is 0")

		var tokenResp struct {
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
			ExpiresAt    int64  `json:"expires_at"`
		}
		require.NoError(t, json.Unmarshal(apiResp.Data, &tokenResp))
		assert.NotEmpty(t, tokenResp.AccessToken, "AC: Must return access_token")
		assert.NotEmpty(t, tokenResp.RefreshToken, "AC: Must return refresh_token")
		assert.True(t, tokenResp.ExpiresAt > time.Now().Unix(), "AC: ExpiresAt must be in the future")
		token = tokenResp.AccessToken
	})

	// ── Step 3: GetMe (spec.md §4.2) ──
	// AC: GET /api/v1/auth/me returns user info
	t.Run("Step3_GetMe", func(t *testing.T) {
		require.NotEmpty(t, token, "Depends on Step 2 login")
		resp, apiResp := doRequest(t, env, "GET", "/api/v1/auth/me", nil, token)

		assert.Equal(t, http.StatusOK, resp.StatusCode, "AC: /me should return 200")
		var userResp struct {
			ID    string `json:"id"`
			Name  string `json:"name"`
			Email string `json:"email"`
		}
		require.NoError(t, json.Unmarshal(apiResp.Data, &userResp))
		assert.Equal(t, "e2e@test.heka.io", userResp.Email, "AC: Must return logged-in user's email")
	})

	// ── Step 4: Create Project (spec.md §4.3) ──
	// AC: POST /api/v1/projects returns 201 with project data
	t.Run("Step4_CreateProject", func(t *testing.T) {
		require.NotEmpty(t, token)
		resp, apiResp := doRequest(t, env, "POST", "/api/v1/projects/", map[string]string{
			"name":        "E2E Test Project",
			"description": "Project for E2E testing",
		}, token)

		assert.Equal(t, http.StatusCreated, resp.StatusCode, "AC: Create project should return 201")
		var projResp struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}
		require.NoError(t, json.Unmarshal(apiResp.Data, &projResp))
		assert.Equal(t, "E2E Test Project", projResp.Name)
		projectID = shared.ID(projResp.ID)
	})

	// ── Step 5: List Projects (spec.md §4.3) ──
	// AC: GET /api/v1/projects returns list containing created project
	t.Run("Step5_ListProjects", func(t *testing.T) {
		require.NotEmpty(t, token)
		resp, apiResp := doRequest(t, env, "GET", "/api/v1/projects/", nil, token)

		assert.Equal(t, http.StatusOK, resp.StatusCode, "AC: List projects should return 200")
		var projects []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}
		require.NoError(t, json.Unmarshal(apiResp.Data, &projects))
		assert.Len(t, projects, 1, "AC: Should list the created project")
	})

	// ── Step 6: Create Module (spec.md §4.4) ──
	// AC: POST /api/v1/modules returns 201 with module data
	t.Run("Step6_CreateModule", func(t *testing.T) {
		require.NotEmpty(t, token)
		require.False(t, projectID.IsEmpty())
		resp, apiResp := doRequest(t, env, "POST", "/api/v1/modules/", map[string]interface{}{
			"project_id":  projectID.String(),
			"name":        "Authentication Module",
			"description": "Login and auth module",
		}, token)

		assert.Equal(t, http.StatusCreated, resp.StatusCode, "AC: Create module should return 201")
		var modResp struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}
		require.NoError(t, json.Unmarshal(apiResp.Data, &modResp))
		assert.Equal(t, "Authentication Module", modResp.Name)
		moduleID = shared.ID(modResp.ID)
	})

	// ── Step 7: Get Module Tree (spec.md §4.4) ──
	// AC: GET /api/v1/modules?project_id={id} returns module tree
	t.Run("Step7_GetModuleTree", func(t *testing.T) {
		require.False(t, projectID.IsEmpty())
		resp, _ := doRequest(t, env, "GET",
			fmt.Sprintf("/api/v1/modules/?project_id=%s", projectID.String()), nil, token)

		assert.Equal(t, http.StatusOK, resp.StatusCode, "AC: Module tree should return 200")
	})

	// ── Step 8: Create Tag (spec.md §4.6) ──
	// AC: POST /api/v1/tags returns 201
	t.Run("Step8_CreateTag", func(t *testing.T) {
		require.False(t, projectID.IsEmpty())
		resp, apiResp := doRequest(t, env, "POST", "/api/v1/tags/", map[string]interface{}{
			"project_id": projectID.String(),
			"name":       "smoke",
			"color":      "#FF0000",
		}, token)

		assert.Equal(t, http.StatusCreated, resp.StatusCode, "AC: Create tag should return 201")
		var tagResp struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}
		require.NoError(t, json.Unmarshal(apiResp.Data, &tagResp))
		assert.Equal(t, "smoke", tagResp.Name)
		tagID = shared.ID(tagResp.ID)
	})

	// ── Step 9: Create Test Case (spec.md §4.5) ──
	// AC: POST /api/v1/testcases returns 201 with test case including steps
	t.Run("Step9_CreateTestCase", func(t *testing.T) {
		require.False(t, projectID.IsEmpty())
		resp, apiResp := doRequest(t, env, "POST",
			fmt.Sprintf("/api/v1/testcases/?project_id=%s", projectID.String()),
			map[string]interface{}{
				"title":       "User can login with valid credentials",
				"description": "Verify login flow works end to end",
				"module_id":   moduleID.String(),
				"priority":    1,
				"tags":        []string{"smoke"},
				"steps": []map[string]string{
					{"action": "Navigate to login page", "expected": "Login form is displayed"},
					{"action": "Enter valid email and password", "expected": "User is redirected to dashboard"},
				},
			}, token)

		assert.Equal(t, http.StatusCreated, resp.StatusCode, "AC: Create test case should return 201")
		var tcResp struct {
			ID     string `json:"id"`
			Title  string `json:"title"`
			Status string `json:"status"`
			Steps  []struct {
				Action   string `json:"action"`
				Expected string `json:"expected"`
			} `json:"steps"`
		}
		require.NoError(t, json.Unmarshal(apiResp.Data, &tcResp))
		assert.Equal(t, "User can login with valid credentials", tcResp.Title)
		assert.Equal(t, "draft", tcResp.Status, "AC: New test cases start as draft")
		assert.Len(t, tcResp.Steps, 2, "AC: Steps must be persisted with the test case")
		testcaseID = shared.ID(tcResp.ID)
	})

	// ── Step 10: Get Test Case (spec.md §4.5) ──
	// AC: GET /api/v1/testcases/{id} returns full case with steps
	t.Run("Step10_GetTestCase", func(t *testing.T) {
		require.False(t, testcaseID.IsEmpty())
		resp, apiResp := doRequest(t, env, "GET",
			fmt.Sprintf("/api/v1/testcases/%s", testcaseID.String()), nil, token)

		assert.Equal(t, http.StatusOK, resp.StatusCode, "AC: Get test case should return 200")
		var tcResp struct {
			ID    string `json:"id"`
			Steps []struct {
				Action   string `json:"action"`
				Expected string `json:"expected"`
			} `json:"steps"`
		}
		require.NoError(t, json.Unmarshal(apiResp.Data, &tcResp))
		assert.Len(t, tcResp.Steps, 2, "AC: Must return steps with test case")
	})

	// ── Step 11: List Test Cases (spec.md §4.5) ──
	// AC: GET /api/v1/testcases returns paginated list
	t.Run("Step11_ListTestCases", func(t *testing.T) {
		require.False(t, projectID.IsEmpty())
		resp, apiResp := doRequest(t, env, "GET",
			fmt.Sprintf("/api/v1/testcases/?project_id=%s&page=1&page_size=20", projectID.String()),
			nil, token)

		assert.Equal(t, http.StatusOK, resp.StatusCode, "AC: List should return 200")
		assert.Equal(t, float64(0), apiResp.Code, "AC: Success code is 0")
	})

	// ── Step 12: Update Test Case (spec.md §4.5) ──
	// AC: PUT /api/v1/testcases/{id} updates the case
	t.Run("Step12_UpdateTestCase", func(t *testing.T) {
		require.False(t, testcaseID.IsEmpty())
		resp, _ := doRequest(t, env, "PUT",
			fmt.Sprintf("/api/v1/testcases/%s", testcaseID.String()),
			map[string]interface{}{
				"title":       "User can login with valid credentials (updated)",
				"description": "Updated description",
				"priority":    0,
				"tags":        []string{"smoke", "regression"},
				"steps": []map[string]string{
					{"action": "Navigate to login page", "expected": "Login form is displayed"},
					{"action": "Enter valid credentials", "expected": "Dashboard shown"},
					{"action": "Click logout", "expected": "User is logged out"},
				},
				"version": 0,
			}, token)

		assert.Equal(t, http.StatusOK, resp.StatusCode, "AC: Update should return 200")
	})

	// ── Step 13: Create Plan (spec.md §4.8) ──
	// AC: POST /api/v1/testplans returns 201 with plan data
	t.Run("Step13_CreatePlan", func(t *testing.T) {
		require.False(t, projectID.IsEmpty())
		require.False(t, testcaseID.IsEmpty())
		resp, apiResp := doRequest(t, env, "POST", "/api/v1/testplans/", map[string]interface{}{
			"project_id":   projectID.String(),
			"name":         "Sprint 1 Regression Plan",
			"description":  "Full regression test for sprint 1",
			"cases": []map[string]interface{}{
				{"test_case_id": testcaseID.String()},
			},
		}, token)

		assert.Equal(t, http.StatusCreated, resp.StatusCode, "AC: Create plan should return 201")
		var planResp struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		}
		require.NoError(t, json.Unmarshal(apiResp.Data, &planResp))
		assert.Equal(t, "draft", planResp.Status, "AC: New plans start as draft")
		planID = shared.ID(planResp.ID)
	})

	// ── Step 14: Start Plan (spec.md §4.8) ──
	// AC: POST /api/v1/testplans/{id}/start transitions draft → active
	t.Run("Step14_StartPlan", func(t *testing.T) {
		require.False(t, planID.IsEmpty())
		resp, apiResp := doRequest(t, env, "POST",
			fmt.Sprintf("/api/v1/testplans/%s/start", planID.String()), nil, token)

		assert.Equal(t, http.StatusOK, resp.StatusCode, "AC: Start plan should return 200")
		var planResp struct {
			Status string `json:"status"`
		}
		require.NoError(t, json.Unmarshal(apiResp.Data, &planResp))
		assert.Equal(t, "active", planResp.Status, "AC: Plan status should be active after start")
	})

	// ── Step 15: Create Execution (spec.md §4.9) ──
	// AC: POST /api/v1/executions creates an execution record
	t.Run("Step15_CreateExecution", func(t *testing.T) {
		require.False(t, planID.IsEmpty())
		resp, apiResp := doRequest(t, env, "POST", "/api/v1/executions/", map[string]interface{}{
			"plan_id": planID.String(),
			"name":    "Execution Round 1",
		}, token)

		assert.Equal(t, http.StatusCreated, resp.StatusCode, "AC: Create execution should return 201")
		var execResp struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		}
		require.NoError(t, json.Unmarshal(apiResp.Data, &execResp))
		assert.Equal(t, "in_progress", execResp.Status, "AC: New executions start as in_progress")
		execID = shared.ID(execResp.ID)
	})

	// ── Step 16: Submit Execution Result (spec.md §4.9) ──
	// AC: POST /api/v1/executions/{id}/results submits a result
	t.Run("Step16_SubmitResult", func(t *testing.T) {
		require.False(t, execID.IsEmpty())
		require.False(t, testcaseID.IsEmpty())
		resp, _ := doRequest(t, env, "POST",
			fmt.Sprintf("/api/v1/executions/%s/results", execID.String()),
			map[string]interface{}{
				"test_case_id": testcaseID.String(),
				"status":       "passed",
				"notes":        "All steps passed successfully",
			}, token)

		assert.Equal(t, http.StatusOK, resp.StatusCode, "AC: Submit result should return 200")
	})

	// ── Step 17: Get Execution Summary (spec.md §4.9) ──
	// AC: GET /api/v1/executions/{id}/summary returns pass/fail counts
	t.Run("Step17_GetExecutionSummary", func(t *testing.T) {
		require.False(t, execID.IsEmpty())
		resp, apiResp := doRequest(t, env, "GET",
			fmt.Sprintf("/api/v1/executions/%s/summary", execID.String()), nil, token)

		assert.Equal(t, http.StatusOK, resp.StatusCode, "AC: Summary should return 200")
		var summary struct {
			Total   int `json:"total"`
			Passed  int `json:"passed"`
			Failed  int `json:"failed"`
			Blocked int `json:"blocked"`
			Skipped int `json:"skipped"`
		}
		require.NoError(t, json.Unmarshal(apiResp.Data, &summary))
		assert.Equal(t, 1, summary.Total, "AC: Total should be 1")
		assert.Equal(t, 1, summary.Passed, "AC: Passed should be 1")
	})

	// ── Step 18: Complete Plan (spec.md §4.8) ──
	// AC: POST /api/v1/testplans/{id}/complete transitions active → completed
	t.Run("Step18_CompletePlan", func(t *testing.T) {
		require.False(t, planID.IsEmpty())
		resp, apiResp := doRequest(t, env, "POST",
			fmt.Sprintf("/api/v1/testplans/%s/complete", planID.String()), nil, token)

		assert.Equal(t, http.StatusOK, resp.StatusCode, "AC: Complete plan should return 200")
		var planResp struct {
			Status string `json:"status"`
		}
		require.NoError(t, json.Unmarshal(apiResp.Data, &planResp))
		assert.Equal(t, "completed", planResp.Status, "AC: Plan status should be completed")
	})

	// ── Step 19: Create Collection (spec.md §4.7) ──
	// AC: POST /api/v1/collections returns 201
	t.Run("Step19_CreateCollection", func(t *testing.T) {
		require.False(t, projectID.IsEmpty())
		resp, apiResp := doRequest(t, env, "POST", "/api/v1/collections/", map[string]interface{}{
			"project_id":   projectID.String(),
			"name":         "Smoke Tests",
			"description":  "Quick smoke test suite",
		}, token)

		assert.Equal(t, http.StatusCreated, resp.StatusCode, "AC: Create collection should return 201")
		var collResp struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}
		require.NoError(t, json.Unmarshal(apiResp.Data, &collResp))
		assert.Equal(t, "Smoke Tests", collResp.Name)
	})

	// ── Step 20: Unauthorized access (spec.md §7.1) ──
	// AC: Requests without token return 401
	t.Run("Step20_UnauthorizedAccess", func(t *testing.T) {
		resp, _ := doRequest(t, env, "GET", "/api/v1/projects/", nil, "")
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "AC: Unauthorized requests must return 401")
	})

	// ── Step 21: Invalid credentials (spec.md §7.1) ──
	// AC: Wrong password does not reveal if user exists
	t.Run("Step21_InvalidCredentials", func(t *testing.T) {
		resp, apiResp := doRequest(t, env, "POST", "/api/v1/auth/login", map[string]string{
			"email":    "e2e@test.heka.io",
			"password": "wrong-password",
		}, "")

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "AC: Wrong password should return 401")
		assert.NotEqual(t, float64(0), apiResp.Code, "AC: Error response should have error code")
	})

	// ── Step 22: Upload File (spec.md §4.10) ──
	// AC: POST /api/v1/files/upload accepts multipart file
	t.Run("Step22_UploadFile", func(t *testing.T) {
		require.False(t, projectID.IsEmpty())
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile("file", "test.txt")
		require.NoError(t, err)
		_, err = part.Write([]byte("Test file content for E2E"))
		require.NoError(t, err)
		require.NoError(t, writer.WriteField("project_id", projectID.String()))
		require.NoError(t, writer.Close())

		req, err := http.NewRequest("POST",
			env.server.URL+"/api/v1/files/upload", body)
		require.NoError(t, err)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// AC: File upload should succeed (201 or 200 depending on implementation)
		assert.True(t, resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK,
			"AC: File upload should return 200 or 201, got %d", resp.StatusCode)
	})

	// ── Step 23: Reports (spec.md §4.12) ──
	// AC: GET /api/v1/reports/coverage returns coverage data
	t.Run("Step23_CoverageReport", func(t *testing.T) {
		require.False(t, projectID.IsEmpty())
		resp, _ := doRequest(t, env, "GET",
			fmt.Sprintf("/api/v1/reports/coverage?project_id=%s", projectID.String()), nil, token)

		assert.Equal(t, http.StatusOK, resp.StatusCode, "AC: Coverage report should return 200")
	})

	// ── Step 24: Register (spec.md §4.2) ──
	// AC: POST /api/v1/auth/register creates a new user account
	t.Run("Step24_Register", func(t *testing.T) {
		resp, apiResp := doRequest(t, env, "POST", "/api/v1/auth/register", map[string]string{
			"email":    "newuser@test.heka.io",
			"password": "new-user-password-123",
			"name":     "New E2E User",
		}, "")

		assert.True(t, resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK,
			"AC: Register should return 200 or 201, got %d", resp.StatusCode)
		assert.Equal(t, float64(0), apiResp.Code, "AC: Success code is 0")
	})

	// ── Step 25: Logout (spec.md §4.2) ──
	// AC: POST /api/v1/auth/logout invalidates the current token
	t.Run("Step25_Logout", func(t *testing.T) {
		require.NotEmpty(t, token)
		resp, _ := doRequest(t, env, "POST", "/api/v1/auth/logout", nil, token)

		assert.Equal(t, http.StatusOK, resp.StatusCode, "AC: Logout should return 200")

		// AC: Subsequent requests with the old token should fail
		resp2, _ := doRequest(t, env, "GET", "/api/v1/projects/", nil, token)
		assert.Equal(t, http.StatusUnauthorized, resp2.StatusCode,
			"AC: Token must be invalidated after logout")

		// Re-login to continue other tests
		resp3, apiResp3 := doRequest(t, env, "POST", "/api/v1/auth/login", map[string]string{
			"email":    "e2e@test.heka.io",
			"password": "test-password-123",
		}, "")
		require.Equal(t, http.StatusOK, resp3.StatusCode)
		var reLogin struct {
			AccessToken string `json:"access_token"`
		}
		require.NoError(t, json.Unmarshal(apiResp3.Data, &reLogin))
		token = reLogin.AccessToken
	})

	// ── Step 26: Add Project Member (spec.md §4.3) ──
	// AC: POST /api/v1/projects/{id}/members adds a member to the project
	t.Run("Step26_AddProjectMember", func(t *testing.T) {
		require.False(t, projectID.IsEmpty())
		require.NotEmpty(t, token)
		resp, _ := doRequest(t, env, "POST",
			fmt.Sprintf("/api/v1/projects/%s/members", projectID.String()),
			map[string]interface{}{
				"user_id": "member-user-id-placeholder",
				"role":    "tester",
			}, token)

		assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated,
			"AC: Add member should return 200 or 201, got %d", resp.StatusCode)
	})

	// ── Step 27: Pause Plan (spec.md §4.8) ──
	// AC: POST /api/v1/testplans/{id}/pause transitions active → paused
	t.Run("Step27_PausePlan", func(t *testing.T) {
		require.NotEmpty(t, token)
		// Create a new plan for lifecycle testing
		resp, apiResp := doRequest(t, env, "POST", "/api/v1/testplans/", map[string]interface{}{
			"project_id":  projectID.String(),
			"name":        "Lifecycle Test Plan",
			"description": "Plan for testing lifecycle transitions",
			"cases":       []map[string]interface{}{},
		}, token)
		require.Equal(t, http.StatusCreated, resp.StatusCode)
		var planResp struct {
			ID string `json:"id"`
		}
		require.NoError(t, json.Unmarshal(apiResp.Data, &planResp))
		lifecyclePlanID := shared.ID(planResp.ID)

		// Start the plan first
		resp, _ = doRequest(t, env, "POST",
			fmt.Sprintf("/api/v1/testplans/%s/start", lifecyclePlanID.String()), nil, token)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		// AC: Pause active plan
		resp, apiResp = doRequest(t, env, "POST",
			fmt.Sprintf("/api/v1/testplans/%s/pause", lifecyclePlanID.String()), nil, token)
		assert.Equal(t, http.StatusOK, resp.StatusCode, "AC: Pause should return 200")
		var pausedPlan struct {
			Status string `json:"status"`
		}
		require.NoError(t, json.Unmarshal(apiResp.Data, &pausedPlan))
		assert.Equal(t, "paused", pausedPlan.Status, "AC: Plan status should be paused")
	})

	// ── Step 28: Resume Plan (spec.md §4.8) ──
	// AC: POST /api/v1/testplans/{id}/resume transitions paused → active
	t.Run("Step28_ResumePlan", func(t *testing.T) {
		require.NotEmpty(t, token)
		// Create and pause a plan
		resp, apiResp := doRequest(t, env, "POST", "/api/v1/testplans/", map[string]interface{}{
			"project_id":  projectID.String(),
			"name":        "Resume Test Plan",
			"description": "Plan for testing resume",
			"cases":       []map[string]interface{}{},
		}, token)
		require.Equal(t, http.StatusCreated, resp.StatusCode)
		var planResp struct {
			ID string `json:"id"`
		}
		require.NoError(t, json.Unmarshal(apiResp.Data, &planResp))
		resumePlanID := shared.ID(planResp.ID)

		// Start then pause
		doRequest(t, env, "POST",
			fmt.Sprintf("/api/v1/testplans/%s/start", resumePlanID.String()), nil, token)
		doRequest(t, env, "POST",
			fmt.Sprintf("/api/v1/testplans/%s/pause", resumePlanID.String()), nil, token)

		// AC: Resume paused plan
		resp, apiResp = doRequest(t, env, "POST",
			fmt.Sprintf("/api/v1/testplans/%s/resume", resumePlanID.String()), nil, token)
		assert.Equal(t, http.StatusOK, resp.StatusCode, "AC: Resume should return 200")
		var resumedPlan struct {
			Status string `json:"status"`
		}
		require.NoError(t, json.Unmarshal(apiResp.Data, &resumedPlan))
		assert.Equal(t, "active", resumedPlan.Status, "AC: Plan status should be active after resume")
	})

	// ── Step 29: Cancel Plan (spec.md §4.8) ──
	// AC: POST /api/v1/testplans/{id}/cancel transitions draft → cancelled
	t.Run("Step29_CancelPlan", func(t *testing.T) {
		require.NotEmpty(t, token)
		// Create a plan to cancel
		resp, apiResp := doRequest(t, env, "POST", "/api/v1/testplans/", map[string]interface{}{
			"project_id":  projectID.String(),
			"name":        "Cancel Test Plan",
			"description": "Plan for testing cancel",
			"cases":       []map[string]interface{}{},
		}, token)
		require.Equal(t, http.StatusCreated, resp.StatusCode)
		var planResp struct {
			ID string `json:"id"`
		}
		require.NoError(t, json.Unmarshal(apiResp.Data, &planResp))
		cancelPlanID := shared.ID(planResp.ID)

		// AC: Cancel draft plan
		resp, apiResp = doRequest(t, env, "POST",
			fmt.Sprintf("/api/v1/testplans/%s/cancel", cancelPlanID.String()), nil, token)
		assert.Equal(t, http.StatusOK, resp.StatusCode, "AC: Cancel should return 200")
		var cancelledPlan struct {
			Status string `json:"status"`
		}
		require.NoError(t, json.Unmarshal(apiResp.Data, &cancelledPlan))
		assert.Equal(t, "cancelled", cancelledPlan.Status, "AC: Plan status should be cancelled")
	})

	// ── Step 30: Batch Submit Results (spec.md §4.9) ──
	// AC: POST /api/v1/executions/{id}/results/batch accepts array of results
	t.Run("Step30_BatchSubmitResults", func(t *testing.T) {
		require.NotEmpty(t, token)
		// Create a new plan + execution for batch testing
		resp, apiResp := doRequest(t, env, "POST", "/api/v1/testplans/", map[string]interface{}{
			"project_id":  projectID.String(),
			"name":        "Batch Submit Plan",
			"description": "Plan for batch submit testing",
			"cases": []map[string]interface{}{
				{"test_case_id": testcaseID.String()},
			},
		}, token)
		require.Equal(t, http.StatusCreated, resp.StatusCode)
		var batchPlanResp struct {
			ID string `json:"id"`
		}
		require.NoError(t, json.Unmarshal(apiResp.Data, &batchPlanResp))
		batchPlanID := shared.ID(batchPlanResp.ID)

		// Start plan
		doRequest(t, env, "POST",
			fmt.Sprintf("/api/v1/testplans/%s/start", batchPlanID.String()), nil, token)

		// Create execution
		resp, apiResp = doRequest(t, env, "POST", "/api/v1/executions/", map[string]interface{}{
			"plan_id": batchPlanID.String(),
			"name":    "Batch Execution Round",
		}, token)
		require.Equal(t, http.StatusCreated, resp.StatusCode)
		var batchExecResp struct {
			ID string `json:"id"`
		}
		require.NoError(t, json.Unmarshal(apiResp.Data, &batchExecResp))
		batchExecID := shared.ID(batchExecResp.ID)

		// AC: Batch submit
		resp, _ = doRequest(t, env, "POST",
			fmt.Sprintf("/api/v1/executions/%s/results/batch", batchExecID.String()),
			map[string]interface{}{
				"results": []map[string]interface{}{
					{
						"test_case_id": testcaseID.String(),
						"status":       "passed",
						"notes":        "Batch result 1",
					},
				},
			}, token)
		assert.Equal(t, http.StatusOK, resp.StatusCode, "AC: Batch submit should return 200")
	})

	// ── Step 31: Plan Report (spec.md §4.12) ──
	// AC: GET /api/v1/reports/plan/{id} returns plan report data
	t.Run("Step31_PlanReport", func(t *testing.T) {
		require.False(t, planID.IsEmpty())
		resp, _ := doRequest(t, env, "GET",
			fmt.Sprintf("/api/v1/reports/plan/%s", planID.String()), nil, token)
		assert.Equal(t, http.StatusOK, resp.StatusCode, "AC: Plan report should return 200")
	})

	// ── Step 32: Trend Report (spec.md §4.12) ──
	// AC: GET /api/v1/reports/trend returns trend data
	t.Run("Step32_TrendReport", func(t *testing.T) {
		require.False(t, projectID.IsEmpty())
		resp, _ := doRequest(t, env, "GET",
			fmt.Sprintf("/api/v1/reports/trend?project_id=%s", projectID.String()), nil, token)
		assert.Equal(t, http.StatusOK, resp.StatusCode, "AC: Trend report should return 200")
	})

	// ── Step 33: Bug Distribution (spec.md §4.12) ──
	// AC: GET /api/v1/reports/bugs returns bug distribution data
	t.Run("Step33_BugDistribution", func(t *testing.T) {
		require.False(t, projectID.IsEmpty())
		resp, _ := doRequest(t, env, "GET",
			fmt.Sprintf("/api/v1/reports/bugs?project_id=%s", projectID.String()), nil, token)
		assert.Equal(t, http.StatusOK, resp.StatusCode, "AC: Bug distribution should return 200")
	})

	// ── Step 34: Workload Report (spec.md §4.12) ──
	// AC: GET /api/v1/reports/workload returns workload distribution data
	t.Run("Step34_WorkloadReport", func(t *testing.T) {
		require.False(t, projectID.IsEmpty())
		resp, _ := doRequest(t, env, "GET",
			fmt.Sprintf("/api/v1/reports/workload?project_id=%s", projectID.String()), nil, token)
		assert.Equal(t, http.StatusOK, resp.StatusCode, "AC: Workload report should return 200")
	})

	// Suppress unused warnings
	_ = tagID
}

// TestE2ELoginRefreshFlow tests the token refresh cycle
func TestE2ELoginRefreshFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	env := setupE2E(t)
	defer env.cleanup()

	// Seed user
	hash, err := auth.HashPassword("refresh-test-pass")
	require.NoError(t, err)
	userID := shared.NewID()
	sqlDB, _ := env.db.DB()
	_, err = sqlDB.Exec(
		`INSERT INTO users (id, name, email, password_hash) VALUES ($1, $2, $3, $4)`,
		userID.String(), "Refresh Test User", "refresh@test.heka.io", hash,
	)
	require.NoError(t, err)

	// AC: Login returns tokens
	resp, apiResp := doRequest(t, env, "POST", "/api/v1/auth/login", map[string]string{
		"email":    "refresh@test.heka.io",
		"password": "refresh-test-pass",
	}, "")
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
	require.NoError(t, json.Unmarshal(apiResp.Data, &tokenResp))
	require.NotEmpty(t, tokenResp.RefreshToken)

	// AC: Refresh token returns new access token
	resp, apiResp = doRequest(t, env, "POST", "/api/v1/auth/refresh", map[string]string{
		"refresh_token": tokenResp.RefreshToken,
	}, "")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "AC: Refresh should return 200")

	var newTokenResp struct {
		AccessToken string `json:"access_token"`
	}
	require.NoError(t, json.Unmarshal(apiResp.Data, &newTokenResp))
	assert.NotEmpty(t, newTokenResp.AccessToken, "AC: Refresh must return new access token")
}
