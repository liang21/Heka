package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"

	appai "github.com/liang21/heka/internal/application/ai"
	appexecution "github.com/liang21/heka/internal/application/execution"
	appfile "github.com/liang21/heka/internal/application/file"
	appplan "github.com/liang21/heka/internal/application/plan"
	appproject "github.com/liang21/heka/internal/application/project"
	apprag "github.com/liang21/heka/internal/application/rag"
	appreport "github.com/liang21/heka/internal/application/report"
	apptestcase "github.com/liang21/heka/internal/application/testcase"
	"github.com/liang21/heka/internal/domain/execution"
	"github.com/liang21/heka/internal/domain/file"
	"github.com/liang21/heka/internal/domain/plan"
	"github.com/liang21/heka/internal/domain/project"
	"github.com/liang21/heka/internal/domain/rag"
	"github.com/liang21/heka/internal/domain/shared"
	"github.com/liang21/heka/internal/domain/user"
	"github.com/liang21/heka/internal/domain/testcase"
	infraauth "github.com/liang21/heka/internal/infrastructure/auth"
	"github.com/liang21/heka/internal/infrastructure/cache"
	infraevent "github.com/liang21/heka/internal/infrastructure/event"
	"github.com/liang21/heka/internal/infrastructure/persistence/postgres"
	"github.com/liang21/heka/internal/infrastructure/storage"
	"github.com/liang21/heka/internal/interface/http/handler"
	httpmiddleware "github.com/liang21/heka/internal/interface/http/middleware"
	"github.com/liang21/heka/internal/shared/config"
	"github.com/liang21/heka/internal/shared/logger"
	"github.com/liang21/heka/scripts/migration"
)

func main() {
	logger.Init("info")
	log.Println("heka: starting...")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := initDB(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get sql.DB: %v", err)
	}

	redisClient, err := cache.NewCacheClient(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize Redis: %v", err)
	}

	eventBus := infraevent.NewEventBus(1000)

	repos := initRepositories(db, eventBus)

	services := initServices(cfg, repos, redisClient, eventBus, sqlDB)

	handlers := initHandlers(services, cfg, sqlDB, redisClient)

	router := setupRouter(handlers, cfg, services.Auth)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Printf("Server started on port %d", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	shutdown(srv, db, redisClient, eventBus)
}

func initDB(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
		cfg.Database.SSLMode,
	)

	db, err := gorm.Open(gormpostgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

	if err := migration.Up(db); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}

type repositories struct {
	User       user.UserRepository
	Project    project.ProjectRepository
	Module     testcase.ModuleRepository
	Tag        testcase.TagRepository
	TestCase   testcase.TestCaseRepository
	TestPlan   plan.TestPlanRepository
	Execution  execution.ExecutionRepository
	File       file.FileRepository
	Collection testcase.CollectionRepository
	Chunk      rag.ChunkRepository
	AsyncTask  shared.AsyncTaskRepository
}

func initRepositories(db *gorm.DB, eventBus shared.EventBus) *repositories {
	return &repositories{
		User:       postgres.NewUserRepository(db),
		Project:    postgres.NewProjectRepository(db),
		Module:     postgres.NewModuleRepository(db),
		Tag:        postgres.NewTagRepository(db),
		TestCase:   postgres.NewTestCaseRepository(db),
		TestPlan:   postgres.NewPlanRepository(db),
		Execution:  postgres.NewExecutionRepository(db),
		File:       postgres.NewFileRepository(db),
		Collection: postgres.NewCollectionRepository(db),
		Chunk:      postgres.NewChunkRepository(db),
		AsyncTask:  postgres.NewAsyncTaskRepository(db),
	}
}

type services struct {
	Project   *appproject.Service
	TestCase  *apptestcase.Service
	Plan      *appplan.Service
	Execution *appexecution.Service
	File      *appfile.Service
	Auth      *infraauth.Service
	AI        *appai.Service
	RAG       *apprag.Service
	Report    *appreport.Service
}

func initServices(cfg *config.Config, repos *repositories, redis *cache.CacheClient, eventBus shared.EventBus, sqlDB *sql.DB) *services {
	passwordHasher := infraauth.NewBCryptPasswordHasher()
	storage := storage.NewLocalStorage(cfg.Upload.StoragePath)

	projectSvc := appproject.NewService(repos.Project)
	testcaseSvc := apptestcase.NewService(repos.TestCase, repos.Module, repos.Tag, repos.Collection, eventBus)
	planSvc := appplan.NewService(repos.TestPlan)
	executionSvc := appexecution.NewService(repos.Execution)
	fileSvc := appfile.NewService(repos.File, storage, nil, eventBus, cfg.Upload.MaxSize)

	authJwtMaker := infraauth.NewJWTMaker(cfg.JWT.Secret, infraauth.TokenTTL{Duration: int(cfg.JWT.AccessTokenTTL.Seconds())}, infraauth.TokenTTL{Duration: int(cfg.JWT.RefreshTokenTTL.Seconds())})
	authSvc := infraauth.NewService(repos.User, passwordHasher, authJwtMaker)

	ragSvc := apprag.NewService(nil, nil)
	aiSvc := appai.NewService(nil, ragSvc, repos.AsyncTask, eventBus)
	reportSvc := appreport.NewService(repos.TestPlan, repos.Execution, repos.TestCase, repos.User)

	return &services{
		Project:   projectSvc,
		TestCase:  testcaseSvc,
		Plan:      planSvc,
		Execution: executionSvc,
		File:      fileSvc,
		Auth:      authSvc,
		AI:        aiSvc,
		RAG:       ragSvc,
		Report:    reportSvc,
	}
}

type handlers struct {
	Auth      *handler.AuthHandler
	Project   *handler.ProjectHandler
	TestCase  *handler.TestCaseHandler
	Plan      *handler.PlanHandler
	Execution *handler.ExecutionHandler
	File      *handler.FileHandler
	AI        *handler.AIHandler
	Report    *handler.ReportHandler
	Health    *handler.HealthHandler
}

func initHandlers(services *services, cfg *config.Config, sqlDB *sql.DB, redis *cache.CacheClient) *handlers {
	return &handlers{
		Auth:      handler.NewAuthHandler(services.Auth),
		Project:   handler.NewProjectHandler(services.Project),
		TestCase:  handler.NewTestCaseHandler(services.TestCase),
		Plan:      handler.NewPlanHandler(services.Plan),
		Execution: handler.NewExecutionHandler(services.Execution),
		File:      handler.NewFileHandler(services.File, cfg.Upload.MaxSize),
		AI:        handler.NewAIHandler(services.AI),
		Report:    handler.NewReportHandler(services.Report),
		Health:    handler.NewHealthHandler(sqlDB, redis),
	}
}

func setupRouter(h *handlers, cfg *config.Config, authSvc *infraauth.Service) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.Server.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	authMiddleware := httpmiddleware.NewAuthMiddleware(cfg.JWT.Secret)

	r.Route("/api/v1", func(r chi.Router) {
		// Auth routes (no auth required)
		r.Route("/auth", func(r chi.Router) {
			r.Post("/login", h.Auth.Login)
			r.Post("/register", h.Auth.Register)
			r.Post("/refresh", h.Auth.RefreshToken)
			r.Post("/logout", h.Auth.Logout)
			r.Get("/me", h.Auth.GetMe)
		})

		// Health check (no auth required)
		r.Get("/health", h.Health.Health)

		// Authenticated routes
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.Handler)

			r.Route("/projects", func(r chi.Router) {
				r.Get("/", h.Project.List)
				r.Post("/", h.Project.Create)
				r.Get("/{id}", h.Project.GetByID)
				r.Post("/{id}/members", h.Project.AddMember)
			})

			r.Route("/testcases", func(r chi.Router) {
				r.Get("/", h.TestCase.ListTestCases)
				r.Post("/", h.TestCase.CreateTestCase)
				r.Get("/{id}", h.TestCase.GetTestCase)
				r.Put("/{id}", h.TestCase.UpdateTestCase)
				r.Delete("/{id}", h.TestCase.DeleteTestCase)
			})

			r.Route("/modules", func(r chi.Router) {
				r.Post("/", h.TestCase.CreateModule)
				r.Get("/tree", h.TestCase.GetModuleTree)
				r.Put("/{id}", h.TestCase.UpdateModule)
				r.Delete("/{id}", h.TestCase.DeleteModule)
			})

			r.Route("/tags", func(r chi.Router) {
				r.Post("/", h.TestCase.CreateTag)
				r.Get("/", h.TestCase.ListTags)
				r.Delete("/{id}", h.TestCase.DeleteTag)
			})

			r.Route("/collections", func(r chi.Router) {
				r.Post("/", h.TestCase.CreateCollection)
				r.Get("/", h.TestCase.ListCollections)
				r.Post("/{id}/cases", h.TestCase.AddToCollection)
				r.Delete("/{id}/cases/{case_id}", h.TestCase.RemoveFromCollection)
			})

			r.Route("/plans", func(r chi.Router) {
				r.Get("/", h.Plan.List)
				r.Post("/", h.Plan.Create)
				r.Get("/{id}", h.Plan.GetByID)
				r.Post("/{id}/start", h.Plan.Start)
				r.Post("/{id}/pause", h.Plan.Pause)
				r.Post("/{id}/resume", h.Plan.Resume)
				r.Post("/{id}/complete", h.Plan.Complete)
				r.Post("/{id}/cancel", h.Plan.Cancel)
			})

			r.Route("/executions", func(r chi.Router) {
				r.Post("/", h.Execution.Create)
				r.Get("/{id}", h.Execution.GetByID)
				r.Get("/{id}/summary", h.Execution.GetSummary)
				r.Post("/{id}/results", h.Execution.SubmitResult)
				r.Post("/{id}/results/batch", h.Execution.BatchSubmit)
			})

			r.Route("/files", func(r chi.Router) {
				r.Post("/upload", h.File.Upload)
				r.Get("/", h.File.List)
				r.Get("/{id}", h.File.GetByID)
				r.Post("/{id}/reindex", h.File.Reindex)
				r.Get("/{id}/index-status", h.File.GetIndexStatus)
				r.Delete("/{id}", h.File.Delete)
			})

			r.Route("/ai", func(r chi.Router) {
				r.Post("/generate-testcases", h.AI.GenerateTestCases)
				r.Get("/tasks/{id}", h.AI.GetTask)
				r.Post("/analyze", h.AI.Analyze)
			})

			r.Route("/reports", func(r chi.Router) {
				r.Get("/plan/{id}", h.Report.PlanReport)
				r.Get("/coverage", h.Report.Coverage)
				r.Get("/trend", h.Report.Trend)
				r.Get("/bugs", h.Report.BugDistribution)
				r.Get("/workload", h.Report.Workload)
			})
		})
	})

	return r
}

func shutdown(srv *http.Server, db *gorm.DB, redis *cache.CacheClient, eventBus shared.EventBus) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	sqlDB, _ := db.DB()
	if sqlDB != nil {
		if err := sqlDB.Close(); err != nil {
			log.Printf("Database close error: %v", err)
		}
	}

	if err := redis.Close(); err != nil {
		log.Printf("Redis close error: %v", err)
	}

	log.Println("Server stopped")
}
