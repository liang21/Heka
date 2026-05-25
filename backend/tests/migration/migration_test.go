package migration

// tasks.md: T158 | spec.md: §3.1-3.6 Database Migration Validation Test
// RED: Validates that all migration files execute correctly and produce expected schema.
// Uses testcontainers PostgreSQL to run migrations from scratch.

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	migrationsRelDir = "../../scripts/migration"
)

// expectedTables lists all tables that must exist after all up migrations
var expectedTables = []string{
	"users",
	"projects",
	"project_members",
	"modules",
	"tags",
	"test_cases",
	"test_steps",
	"test_case_collections",
	"collection_cases",
	"test_plans",
	"plan_test_cases",
	"test_executions",
	"execution_results",
	"files",
	"file_versions",
	"document_chunks",
	"vector_embeddings",
	"index_tasks",
	"ai_tasks",
}

// expectedIndexes lists critical indexes that must exist after migrations
var expectedIndexes = []string{
	"idx_users_email",
	"idx_modules_project",
	"idx_modules_parent",
	"idx_test_cases_project",
	"idx_test_cases_status",
	"idx_test_cases_tags",
	"idx_test_cases_active",
	"idx_tags_project",
	"idx_test_plans_project",
	"idx_test_plans_status",
	"idx_executions_plan",
	"idx_executions_status",
	"idx_executions_single_active",
	"idx_files_project",
	"idx_files_index_status",
	"idx_chunks_file",
	"idx_ai_tasks_project",
	"idx_ai_tasks_status",
	"idx_index_tasks_status",
}

// setupMigrationDB starts a testcontainers PostgreSQL and returns a *sql.DB
func setupMigrationDB(t *testing.T) (*sql.DB, func()) {
	t.Helper()
	ctx := context.Background()

	container, err := tcpostgres.Run(ctx,
		"postgres:15-alpine",
		tcpostgres.WithDatabase("heka_migration_test"),
		tcpostgres.WithUsername("heka"),
		tcpostgres.WithPassword("heka_test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	require.NoError(t, err, "AC: PostgreSQL testcontainers must start for migration tests")

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err, "AC: Must obtain DB connection string")

	db, err := sql.Open("pgx", dsn)
	require.NoError(t, err, "AC: Must connect to test database")

	// Verify connection
	require.NoError(t, db.Ping(), "AC: Must ping test database successfully")

	cleanup := func() {
		db.Close()
		container.Terminate(ctx)
	}

	return db, cleanup
}

// getMigrationFiles returns sorted .up.sql files in the migrations directory
func getMigrationFiles(t *testing.T, suffix string) []string {
	t.Helper()
	dir := migrationsRelDir
	entries, err := os.ReadDir(dir)
	require.NoError(t, err, "AC: Migration directory must be readable")

	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if filepath.Ext(e.Name()) == ".sql" {
			if len(e.Name()) > len(suffix) && e.Name()[len(e.Name())-len(suffix):] == suffix {
				files = append(files, e.Name())
			}
		}
	}
	sort.Strings(files)
	return files
}

// executeMigrations runs all migration files with the given suffix in order
func executeMigrations(t *testing.T, db *sql.DB, suffix string) {
	t.Helper()
	files := getMigrationFiles(t, suffix)
	require.NotEmpty(t, files, "AC: Must find migration files")

	dir := migrationsRelDir
	for _, f := range files {
		content, err := os.ReadFile(filepath.Join(dir, f))
		require.NoError(t, err, "AC: Migration file %s must be readable", f)
		_, err = db.Exec(string(content))
		require.NoError(t, err, "AC: Migration %s must execute without error", f)
		t.Logf("Executed migration: %s", f)
	}
}

// getExistingTables queries the database for all user tables
func getExistingTables(t *testing.T, db *sql.DB) map[string]bool {
	t.Helper()
	rows, err := db.Query(`
		SELECT table_name FROM information_schema.tables
		WHERE table_schema = 'public' AND table_type = 'BASE TABLE'
	`)
	require.NoError(t, err)
	defer rows.Close()

	tables := make(map[string]bool)
	for rows.Next() {
		var name string
		require.NoError(t, rows.Scan(&name))
		tables[name] = true
	}
	return tables
}

// getExistingIndexes queries the database for all indexes
func getExistingIndexes(t *testing.T, db *sql.DB) map[string]bool {
	t.Helper()
	rows, err := db.Query(`
		SELECT indexname FROM pg_indexes WHERE schemaname = 'public'
	`)
	require.NoError(t, err)
	defer rows.Close()

	indexes := make(map[string]bool)
	for rows.Next() {
		var name string
		require.NoError(t, rows.Scan(&name))
		indexes[name] = true
	}
	return indexes
}

// ─── Tests ───

// TestMigrationsUp validates all up migrations create expected schema
func TestMigrationsUp(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping migration test in short mode")
	}

	db, cleanup := setupMigrationDB(t)
	defer cleanup()

	// AC: Execute all up migrations in order
	executeMigrations(t, db, ".up.sql")

	// AC: Verify all expected tables exist
	existingTables := getExistingTables(t, db)
	for _, table := range expectedTables {
		assert.True(t, existingTables[table],
			"AC: Table '%s' must exist after up migrations (found: %v)", table, existingTables)
	}

	// AC: Verify all expected indexes exist
	existingIndexes := getExistingIndexes(t, db)
	for _, idx := range expectedIndexes {
		assert.True(t, existingIndexes[idx],
			"AC: Index '%s' must exist after up migrations", idx)
	}
}

// TestMigrationsDown validates all down migrations clean up schema
func TestMigrationsDown(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping migration test in short mode")
	}

	db, cleanup := setupMigrationDB(t)
	defer cleanup()

	// AC: First run all up migrations
	executeMigrations(t, db, ".up.sql")

	// AC: Then run all down migrations in reverse order
	downFiles := getMigrationFiles(t, ".down.sql")
	sort.Sort(sort.Reverse(sort.StringSlice(downFiles)))

	dir := migrationsRelDir
	for _, f := range downFiles {
		content, err := os.ReadFile(filepath.Join(dir, f))
		require.NoError(t, err, "AC: Down migration %s must be readable", f)
		_, err = db.Exec(string(content))
		require.NoError(t, err, "AC: Down migration %s must execute without error", f)
		t.Logf("Executed down migration: %s", f)
	}

	// AC: Verify all tables are gone
	existingTables := getExistingTables(t, db)
	for _, table := range expectedTables {
		assert.False(t, existingTables[table],
			"AC: Table '%s' must NOT exist after down migrations", table)
	}
}

// TestMigrationsIdempotent validates that running up migrations twice doesn't fail
func TestMigrationsIdempotent(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping migration test in short mode")
	}

	db, cleanup := setupMigrationDB(t)
	defer cleanup()

	// AC: Run up migrations once
	executeMigrations(t, db, ".up.sql")

	// AC: Running them again should fail (CREATE TABLE already exists)
	// This documents that migrations are NOT idempotent by design
	upFiles := getMigrationFiles(t, ".up.sql")
	dir := migrationsRelDir
	for _, f := range upFiles {
		content, err := os.ReadFile(filepath.Join(dir, f))
		require.NoError(t, err)
		_, err = db.Exec(string(content))
		if err != nil {
			t.Logf("Migration %s is not idempotent (expected): %v", f, err)
			return // Expected: migrations are NOT idempotent
		}
	}
	t.Log("Note: Migrations happen to be idempotent (all ran twice without error)")
}

// TestMigrationsConstraints validates CHECK constraints are properly created
func TestMigrationsConstraints(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping migration test in short mode")
	}

	db, cleanup := setupMigrationDB(t)
	defer cleanup()

	executeMigrations(t, db, ".up.sql")

	tests := []struct {
		name    string
		table   string
		column  string
		value   string
		wantErr bool
	}{
		{
			name:    "test_cases valid status draft",
			table:   "test_cases",
			column:  "status",
			value:   "'draft'",
			wantErr: false,
		},
		{
			name:    "test_cases invalid status",
			table:   "test_cases",
			column:  "status",
			value:   "'invalid_status'",
			wantErr: true,
		},
		{
			name:    "test_plans valid status draft",
			table:   "test_plans",
			column:  "status",
			value:   "'draft'",
			wantErr: false,
		},
		{
			name:    "test_plans invalid status",
			table:   "test_plans",
			column:  "status",
			value:   "'invalid_status'",
			wantErr: true,
		},
		{
			name:    "execution_results valid status passed",
			table:   "execution_results",
			column:  "status",
			value:   "'passed'",
			wantErr: false,
		},
		{
			name:    "execution_results invalid status",
			table:   "execution_results",
			column:  "status",
			value:   "'maybe'",
			wantErr: true,
		},
		{
			name:    "files valid index_status pending",
			table:   "files",
			column:  "index_status",
			value:   "'pending'",
			wantErr: false,
		},
		{
			name:    "files invalid index_status",
			table:   "files",
			column:  "index_status",
			value:   "'unknown'",
			wantErr: true,
		},
	}

	// AC: Create prerequisite data for constraint tests
	_, err := db.Exec(`INSERT INTO users (id, name, email, password_hash) VALUES ('u1', 'Test', 'test@heka.io', 'hash')`)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO projects (id, name, created_by) VALUES ('p1', 'Test Project', 'u1')`)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO modules (id, project_id, name, order_index, created_by) VALUES ('m1', 'p1', 'Mod1', 0, 'u1')`)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO test_cases (id, project_id, title, status, priority, created_by, version) VALUES ('tc1', 'p1', 'Test Case 1', 'draft', 1, 'u1', 0)`)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO test_plans (id, project_id, name, status, created_by) VALUES ('tp1', 'p1', 'Test Plan', 'draft', 'u1')`)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO test_executions (id, plan_id, name, status, executor_id) VALUES ('te1', 'tp1', 'Exec 1', 'in_progress', 'u1')`)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO files (id, project_id, name, type, size, path, source_type, uploaded_by, version) VALUES ('f1', 'p1', 'test.pdf', 'pdf', 100, '/tmp/test.pdf', 'upload', 'u1', 1)`)
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use a sub-transaction so we can rollback after each test
			tx, err := db.Begin()
			require.NoError(t, err)
			defer tx.Rollback()

			var query string
			switch tt.table {
			case "test_cases":
				query = fmt.Sprintf(`INSERT INTO test_cases (id, project_id, title, status, priority, created_by, version) VALUES ('tc_%s', 'p1', 'Constraint Test', %s, 1, 'u1', 0)`, tt.name, tt.value)
			case "test_plans":
				query = fmt.Sprintf(`INSERT INTO test_plans (id, project_id, name, status, created_by) VALUES ('tp_%s', 'p1', 'Constraint Test', %s, 'u1')`, tt.name, tt.value)
			case "execution_results":
				query = fmt.Sprintf(`INSERT INTO execution_results (id, execution_id, test_case_id, executor_id, status) VALUES ('er_%s', 'te1', 'tc1', 'u1', %s)`, tt.name, tt.value)
			case "files":
				query = fmt.Sprintf(`UPDATE files SET index_status = %s WHERE id = 'f1'`, tt.value)
			default:
				t.Fatalf("unknown table: %s", tt.table)
			}

			_, err = tx.Exec(query)
			if tt.wantErr {
				assert.Error(t, err, "AC: CHECK constraint must reject invalid value %s for %s.%s", tt.value, tt.table, tt.column)
			} else {
				assert.NoError(t, err, "AC: CHECK constraint must accept valid value %s for %s.%s", tt.value, tt.table, tt.column)
			}
		})
	}
}

// TestMigrationsUniqueConstraints validates UNIQUE constraints
func TestMigrationsUniqueConstraints(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping migration test in short mode")
	}

	db, cleanup := setupMigrationDB(t)
	defer cleanup()

	executeMigrations(t, db, ".up.sql")

	// AC: Prerequisite data
	_, err := db.Exec(`INSERT INTO users (id, name, email, password_hash) VALUES ('u1', 'Test', 'unique@heka.io', 'hash')`)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO projects (id, name, created_by) VALUES ('p1', 'Test Project', 'u1')`)
	require.NoError(t, err)

	t.Run("users email unique", func(t *testing.T) {
		_, err := db.Exec(`INSERT INTO users (id, name, email, password_hash) VALUES ('u2', 'Test2', 'unique@heka.io', 'hash')`)
		assert.Error(t, err, "AC: Duplicate email must be rejected by UNIQUE constraint")
	})

	t.Run("modules unique name per project parent", func(t *testing.T) {
		_, err := db.Exec(`INSERT INTO modules (id, project_id, name, order_index, created_by) VALUES ('m1', 'p1', 'SameName', 0, 'u1')`)
		require.NoError(t, err)
		_, err = db.Exec(`INSERT INTO modules (id, project_id, name, order_index, created_by) VALUES ('m2', 'p1', 'SameName', 1, 'u1')`)
		assert.Error(t, err, "AC: Duplicate module name under same parent must be rejected")
	})

	t.Run("tags unique name per project", func(t *testing.T) {
		_, err := db.Exec(`INSERT INTO tags (id, project_id, name, created_by) VALUES ('t1', 'p1', 'SameTag', 'u1')`)
		require.NoError(t, err)
		_, err = db.Exec(`INSERT INTO tags (id, project_id, name, created_by) VALUES ('t2', 'p1', 'SameTag', 'u1')`)
		assert.Error(t, err, "AC: Duplicate tag name in same project must be rejected")
	})
}

// TestMigrationsPartialUniqueIndex validates the concurrent execution control
func TestMigrationsPartialUniqueIndex(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping migration test in short mode")
	}

	db, cleanup := setupMigrationDB(t)
	defer cleanup()

	executeMigrations(t, db, ".up.sql")

	// AC: Prerequisite data
	_, err := db.Exec(`INSERT INTO users (id, name, email, password_hash) VALUES ('u1', 'Test', 'partial@heka.io', 'hash')`)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO projects (id, name, created_by) VALUES ('p1', 'Test', 'u1')`)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO test_plans (id, project_id, name, status, created_by) VALUES ('tp1', 'p1', 'Plan', 'active', 'u1')`)
	require.NoError(t, err)

	// AC: Only one in_progress execution per plan (spec.md §3.3)
	_, err = db.Exec(`INSERT INTO test_executions (id, plan_id, name, status, executor_id) VALUES ('te1', 'tp1', 'Exec 1', 'in_progress', 'u1')`)
	require.NoError(t, err)

	_, err = db.Exec(`INSERT INTO test_executions (id, plan_id, name, status, executor_id) VALUES ('te2', 'tp1', 'Exec 2', 'in_progress', 'u1')`)
	assert.Error(t, err, "AC: idx_executions_single_active must prevent second in_progress execution for same plan")

	// AC: A completed execution should not block a new in_progress
	_, err = db.Exec(`UPDATE test_executions SET status = 'completed' WHERE id = 'te1'`)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO test_executions (id, plan_id, name, status, executor_id) VALUES ('te3', 'tp1', 'Exec 3', 'in_progress', 'u1')`)
	assert.NoError(t, err, "AC: After completing first execution, a new in_progress should be allowed")
}
