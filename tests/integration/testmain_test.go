//go:build integration

package integration_test

import (
	"context"
	"database/sql"
	"fmt"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"golang.org/x/crypto/bcrypt"

	"github.com/untibullet/secure-db-manager/internal/config"
	"github.com/untibullet/secure-db-manager/internal/handler"
	appmw "github.com/untibullet/secure-db-manager/internal/middleware"
	"github.com/untibullet/secure-db-manager/internal/repository"
	"github.com/untibullet/secure-db-manager/internal/seclog"
	"github.com/untibullet/secure-db-manager/internal/service"
	"github.com/untibullet/secure-db-manager/internal/session"
)

const (
	pgImage    = "postgres:16-alpine"
	pgDB       = "testdb"
	pgUser     = "postgres"
	pgPassword = "postgres"

	testJWTSecret = "integration-test-jwt-secret-key!"
)

// Package-level state shared across all integration test files.
var (
	superConn     *pgx.Conn
	containerDSN  string
	containerHost string
	containerPort string
	sysPool       *pgxpool.Pool
	testStore     *session.Store
	testServerURL string

	testLeadUserID int
	testerAUserID  int
	testerBUserID  int
	adminUserID    int
	guestUserID    int

	testPlanID       int
	testCaseID1      int
	testCaseID2      int
	testRunID        int
	runItemID1       int
	runItemID2       int
	resultAID        int
	resultBID        int
	nonFinalResultID int

	priorityID         int
	passedStatusID     int
	inProgressStatusID int
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true") //nolint:errcheck

	// 1. Start postgres container.
	ctr, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        pgImage,
			ExposedPorts: []string{"5432/tcp"},
			Env: map[string]string{
				"POSTGRES_USER":     pgUser,
				"POSTGRES_PASSWORD": pgPassword,
				"POSTGRES_DB":       pgDB,
			},
			WaitingFor: wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60 * time.Second),
		},
		Started: true,
	})
	if err != nil {
		panic("start container: " + err.Error())
	}
	defer ctr.Terminate(ctx) //nolint:errcheck

	host, err := ctr.Host(ctx)
	if err != nil {
		panic("container host: " + err.Error())
	}
	port, err := ctr.MappedPort(ctx, "5432")
	if err != nil {
		panic("container port: " + err.Error())
	}
	containerHost = host
	containerPort = port.Port()
	containerDSN = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		pgUser, pgPassword, host, port.Port(), pgDB)

	// 2. Apply migrations via goose.
	sqlDB, err := sql.Open("pgx", containerDSN)
	if err != nil {
		panic("sql.Open: " + err.Error())
	}
	_, thisFile, _, _ := runtime.Caller(0)
	migrationsDir := filepath.Join(filepath.Dir(thisFile), "../../db/migrations")
	if err := goose.SetDialect("postgres"); err != nil {
		panic("goose dialect: " + err.Error())
	}
	if err := goose.Up(sqlDB, migrationsDir); err != nil {
		panic("goose up: " + err.Error())
	}
	sqlDB.Close() //nolint:errcheck

	// 3. Open superuser connection and pool.
	superConn, err = pgx.Connect(ctx, containerDSN)
	if err != nil {
		panic("superuser connect: " + err.Error())
	}
	defer superConn.Close(ctx) //nolint:errcheck

	sysPool, err = pgxpool.New(ctx, containerDSN)
	if err != nil {
		panic("sysPool: " + err.Error())
	}
	defer sysPool.Close()

	// 4. Set up fixtures.
	if err := setupFixtures(ctx); err != nil {
		panic("setupFixtures: " + err.Error())
	}
	defer cleanup(ctx)

	// 5. Start test HTTP server.
	srv, err := setupTestServer()
	if err != nil {
		panic("setupTestServer: " + err.Error())
	}
	defer srv.Close()
	testServerURL = srv.URL

	os.Exit(m.Run())
}

// setupTestServer wires up the full application stack against the test container.
func setupTestServer() (*httptest.Server, error) {
	testStore = session.NewStore()
	cfg := &config.Config{
		DBHost:        containerHost,
		DBPort:        containerPort,
		DBName:        pgDB,
		SuperuserUser: pgUser,
		SuperuserPass: pgPassword,
		JWTSecret:     testJWTSecret,
		SessionTTL:    time.Hour,
	}
	roleRepo := repository.NewRoleRepository(sysPool)
	h := handler.New(
		service.NewAuthService(testStore, cfg, seclog.Noop()),
		&service.TestPlanService{},
		&service.TestCaseService{},
		&service.RunService{},
		&service.ResultService{},
		&service.AutotestService{},
		&service.ReferenceService{},
		&service.StatsService{},
		service.NewAdminService(roleRepo),
		roleRepo,
		cfg.SessionTTL,
		seclog.Noop(),
	)
	e := echo.New()
	e.HideBanner = true
	e.HTTPErrorHandler = handler.ErrorHandler
	h.Register(e, appmw.Auth(cfg.JWTSecret, testStore, seclog.Noop()))
	return httptest.NewServer(e), nil
}

// setupFixtures creates pg login roles and inserts all business data needed by tests.
func setupFixtures(ctx context.Context) error {
	type testUser struct {
		username string
		password string
		pgGroup  string
		idPtr    *int
	}
	users := []testUser{
		{"test_lead_user", "leadpass", "db_test_lead", &testLeadUserID},
		{"tester_a", "testerapass", "db_tester", &testerAUserID},
		{"tester_b", "testerbpass", "db_tester", &testerBUserID},
		{"admin_user", "adminpass", "db_admin", &adminUserID},
		{"guest_user", "guestpass", "db_guest", &guestUserID},
	}

	for _, u := range users {
		hash, err := bcrypt.GenerateFromPassword([]byte(u.password), bcrypt.MinCost)
		if err != nil {
			return fmt.Errorf("bcrypt %s: %w", u.username, err)
		}
		if _, err := superConn.Exec(ctx,
			fmt.Sprintf("CREATE ROLE %s LOGIN PASSWORD '%s'", u.username, u.password),
		); err != nil {
			return fmt.Errorf("CREATE ROLE %s: %w", u.username, err)
		}
		if _, err := superConn.Exec(ctx,
			fmt.Sprintf("GRANT %s TO %s", u.pgGroup, u.username),
		); err != nil {
			return fmt.Errorf("GRANT %s TO %s: %w", u.pgGroup, u.username, err)
		}
		if err := superConn.QueryRow(ctx,
			`INSERT INTO users (username, password_hash, email, full_name, is_active)
			 VALUES ($1, $2, $3, $4, TRUE) RETURNING user_id`,
			u.username, string(hash), u.username+"@test.local", u.username,
		).Scan(u.idPtr); err != nil {
			return fmt.Errorf("INSERT user %s: %w", u.username, err)
		}
	}

	// Assign application roles so GetUserForLogin returns the correct RoleCode.
	type roleAssign struct {
		userIDPtr *int
		code      string
	}
	for _, ra := range []roleAssign{
		{&testLeadUserID, "TEST_LEAD"},
		{&testerAUserID, "TESTER"},
		{&testerBUserID, "TESTER"},
		{&adminUserID, "ADMIN"},
	} {
		var roleID int
		if err := superConn.QueryRow(ctx,
			`SELECT role_id FROM roles WHERE code = $1`, ra.code,
		).Scan(&roleID); err != nil {
			return fmt.Errorf("query role %s: %w", ra.code, err)
		}
		if _, err := superConn.Exec(ctx,
			`INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2)`,
			*ra.userIDPtr, roleID,
		); err != nil {
			return fmt.Errorf("user_roles %s: %w", ra.code, err)
		}
	}

	// Reference data.
	if err := superConn.QueryRow(ctx,
		`SELECT priority_id FROM priorities WHERE code = 'CRITICAL'`,
	).Scan(&priorityID); err != nil {
		return fmt.Errorf("query priority: %w", err)
	}
	if err := superConn.QueryRow(ctx,
		`SELECT status_id FROM statuses WHERE code = 'PASSED'`,
	).Scan(&passedStatusID); err != nil {
		return fmt.Errorf("query PASSED status: %w", err)
	}
	if err := superConn.QueryRow(ctx,
		`SELECT status_id FROM statuses WHERE code = 'IN_PROGRESS'`,
	).Scan(&inProgressStatusID); err != nil {
		return fmt.Errorf("query IN_PROGRESS status: %w", err)
	}

	var versionID, envConfigID, toolConfigID int
	if err := superConn.QueryRow(ctx,
		`INSERT INTO app_versions (version_string) VALUES ('1.0.0') RETURNING version_id`,
	).Scan(&versionID); err != nil {
		return fmt.Errorf("INSERT app_versions: %w", err)
	}
	if err := superConn.QueryRow(ctx,
		`INSERT INTO env_configs (name) VALUES ('test-env') RETURNING env_config_id`,
	).Scan(&envConfigID); err != nil {
		return fmt.Errorf("INSERT env_configs: %w", err)
	}
	if err := superConn.QueryRow(ctx,
		`INSERT INTO tool_configs (name) VALUES ('test-tool') RETURNING tool_config_id`,
	).Scan(&toolConfigID); err != nil {
		return fmt.Errorf("INSERT tool_configs: %w", err)
	}

	if err := superConn.QueryRow(ctx,
		`INSERT INTO test_plans (name, priority_id, start_date, end_date, owner_user_id, status)
		 VALUES ('Integration Test Plan', $1, CURRENT_DATE, CURRENT_DATE+1, $2, 'DRAFT')
		 RETURNING test_plan_id`,
		priorityID, testLeadUserID,
	).Scan(&testPlanID); err != nil {
		return fmt.Errorf("INSERT test_plans: %w", err)
	}

	if err := superConn.QueryRow(ctx,
		`INSERT INTO test_cases (name, priority_id, owner_user_id)
		 VALUES ('Test Case Alpha', $1, $2) RETURNING test_case_id`,
		priorityID, testLeadUserID,
	).Scan(&testCaseID1); err != nil {
		return fmt.Errorf("INSERT test_case 1: %w", err)
	}
	if err := superConn.QueryRow(ctx,
		`INSERT INTO test_cases (name, priority_id, owner_user_id)
		 VALUES ('Test Case Beta', $1, $2) RETURNING test_case_id`,
		priorityID, testLeadUserID,
	).Scan(&testCaseID2); err != nil {
		return fmt.Errorf("INSERT test_case 2: %w", err)
	}

	if err := superConn.QueryRow(ctx,
		`INSERT INTO test_runs (test_plan_id, env_config_id, tool_config_id, version_id,
		  name, start_date, created_by)
		 VALUES ($1, $2, $3, $4, 'Integration Test Run', NOW(), $5) RETURNING test_run_id`,
		testPlanID, envConfigID, toolConfigID, versionID, testLeadUserID,
	).Scan(&testRunID); err != nil {
		return fmt.Errorf("INSERT test_runs: %w", err)
	}

	if err := superConn.QueryRow(ctx,
		`INSERT INTO test_run_items (test_run_id, test_case_id, execution_order)
		 VALUES ($1, $2, 1) RETURNING run_item_id`,
		testRunID, testCaseID1,
	).Scan(&runItemID1); err != nil {
		return fmt.Errorf("INSERT run_item 1: %w", err)
	}
	if err := superConn.QueryRow(ctx,
		`INSERT INTO test_run_items (test_run_id, test_case_id, execution_order)
		 VALUES ($1, $2, 2) RETURNING run_item_id`,
		testRunID, testCaseID2,
	).Scan(&runItemID2); err != nil {
		return fmt.Errorf("INSERT run_item 2: %w", err)
	}

	// resultAID: tester_a, PASSED (final)
	if err := superConn.QueryRow(ctx,
		`INSERT INTO test_results (run_item_id, status_id, executor_user_id)
		 VALUES ($1, $2, $3) RETURNING test_result_id`,
		runItemID1, passedStatusID, testerAUserID,
	).Scan(&resultAID); err != nil {
		return fmt.Errorf("INSERT resultA: %w", err)
	}
	// resultBID: tester_b, PASSED (final)
	if err := superConn.QueryRow(ctx,
		`INSERT INTO test_results (run_item_id, status_id, executor_user_id)
		 VALUES ($1, $2, $3) RETURNING test_result_id`,
		runItemID2, passedStatusID, testerBUserID,
	).Scan(&resultBID); err != nil {
		return fmt.Errorf("INSERT resultB: %w", err)
	}
	// nonFinalResultID: tester_a, IN_PROGRESS (non-final)
	if err := superConn.QueryRow(ctx,
		`INSERT INTO test_results (run_item_id, status_id, executor_user_id)
		 VALUES ($1, $2, $3) RETURNING test_result_id`,
		runItemID1, inProgressStatusID, testerAUserID,
	).Scan(&nonFinalResultID); err != nil {
		return fmt.Errorf("INSERT nonFinalResult: %w", err)
	}

	return nil
}

// cleanup removes all fixture data in reverse FK order.
func cleanup(ctx context.Context) {
	superConn.Exec(ctx, //nolint:errcheck
		`DELETE FROM test_results WHERE test_result_id = ANY($1)`,
		[]int{resultAID, resultBID, nonFinalResultID},
	)
	superConn.Exec(ctx, //nolint:errcheck
		`DELETE FROM test_run_items WHERE run_item_id = ANY($1)`,
		[]int{runItemID1, runItemID2},
	)
	superConn.Exec(ctx, "DELETE FROM test_runs WHERE test_run_id = $1", testRunID)     //nolint:errcheck
	superConn.Exec(ctx, "DELETE FROM test_cases WHERE test_case_id = $1", testCaseID1) //nolint:errcheck
	superConn.Exec(ctx, "DELETE FROM test_cases WHERE test_case_id = $1", testCaseID2) //nolint:errcheck
	superConn.Exec(ctx, "DELETE FROM test_plans WHERE test_plan_id = $1", testPlanID)  //nolint:errcheck
	superConn.Exec(ctx, "DELETE FROM tool_configs WHERE name = 'test-tool'")           //nolint:errcheck
	superConn.Exec(ctx, "DELETE FROM env_configs WHERE name = 'test-env'")             //nolint:errcheck
	superConn.Exec(ctx, "DELETE FROM app_versions WHERE version_string = '1.0.0'")     //nolint:errcheck
	superConn.Exec(ctx, //nolint:errcheck
		`DELETE FROM user_roles WHERE user_id = ANY($1)`,
		[]int{testLeadUserID, testerAUserID, testerBUserID, adminUserID, guestUserID},
	)
	superConn.Exec(ctx, //nolint:errcheck
		`DELETE FROM users WHERE user_id = ANY($1)`,
		[]int{testLeadUserID, testerAUserID, testerBUserID, adminUserID, guestUserID},
	)
	for _, u := range []string{"test_lead_user", "tester_a", "tester_b", "admin_user", "guest_user"} {
		superConn.Exec(ctx, fmt.Sprintf("DROP ROLE IF EXISTS %s", u)) //nolint:errcheck
	}
}
