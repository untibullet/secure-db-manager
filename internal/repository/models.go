package repository

import (
	"database/sql"
	"time"
)

// TestPlan представляет данные из v_public_test_plans
type TestPlan struct {
	ID          int          `db:"test_plan_id"`
	Name        string       `db:"test_plan_name"`
	Description string       `db:"description"`
	StartDate   time.Time    `db:"start_date"`
	EndDate     time.Time    `db:"end_date"`
	Status      string       `db:"status"`
	Owner       string       `db:"owner"`
	Priority    string       `db:"priority"`
}

// TestCase представляет данные из v_lead_all_cases
type TestCase struct {
	ID          int          `db:"test_case_id"`
	Name        string       `db:"test_case_name"`
	Description string       `db:"description"`
	IsAutomated bool         `db:"is_automated"`
	IsActive    bool         `db:"is_active"`
	Priority    string       `db:"priority"`
	Owner       string       `db:"owner"`
}

// Autotest представляет данные из v_engineer_autotests
type Autotest struct {
	ID            int          `db:"autotest_id"`
	Name          string       `db:"autotest_name"`
	Description   string       `db:"description"`
	IsActive      bool         `db:"is_active"`
	VersionString string       `db:"version_string"`
	CommitHash    string       `db:"commit_hash"`
	Author        string       `db:"author"`
}

// Environment представляет данные из v_shared_environments
type Environment struct {
	ID          int    `db:"env_config_id"`
	Name        string `db:"environment_name"`
	Description string `db:"description"`
	IsActive    bool   `db:"is_active"`
	ParamsCount int    `db:"params_count"`
}

// PublicResult представляет данные из v_public_results
type PublicResult struct {
	ResultID      int       `db:"test_result_id"`
	RunID         int       `db:"test_run_id"`
	TestCaseName  string    `db:"test_case_name"`
	Status        string    `db:"status"`
	ExecutionDate time.Time `db:"execution_date"`
}

// TestResult представляет данные из v_lead_all_results
type TestResult struct {
	ResultID       int             `db:"test_result_id"`
	RunName        string          `db:"test_run_name"`
	TestCaseName   string          `db:"test_case_name"`
	Status         string          `db:"status"`
	Executor       string          `db:"executor"`
	ExecutionDate  time.Time       `db:"execution_date"`
	ErrorMessage   sql.NullString  `db:"error_message"`
}

// Report представляет данные из v_shared_reports
type Report struct {
	ID           int       `db:"report_id"`
	Name         string    `db:"report_name"`
	CreationDate time.Time `db:"creation_date"`
	Template     string    `db:"template"`
	Author       string    `db:"author"`
}

// UserAdminView представляет данные из v_admin_users_and_roles
type UserAdminView struct {
	ID        int          `db:"user_id"`
	Username  string       `db:"username"`
	FullName  string       `db:"full_name"`
	Email     string       `db:"email"`
	IsActive  bool         `db:"is_active"`
	LastLogin sql.NullTime `db:"last_login"`
	Roles     string       `db:"roles"`
}

// ActiveRun представляет данные из v_active_test_runs
type ActiveRun struct {
	RunID       int    `db:"test_run_id"`
	RunName     string `db:"run_name"`
	PlanName    string `db:"plan_name"`
	Status      string `db:"status"`
	PassedTests int    `db:"passed_tests"`
	FailedTests int    `db:"failed_tests"`
}

// RunSummary представляет данные из v_test_execution_summary
type RunSummary struct {
	RunID      int     `db:"test_run_id"`
	RunName    string  `db:"test_run_name"`
	TotalCases int     `db:"total_cases"`
	Passed     int     `db:"passed"`
	Failed     int     `db:"failed"`
	PassRate   float64 `db:"pass_rate"`
}

// CaseStats представляет данные из v_test_case_statistics
type CaseStats struct {
	TestCaseID      int     `db:"test_case_id"`
	Name            string  `db:"name"`
	PassRatePercent float64 `db:"pass_rate_percent"`
	AvgDurationMins float64 `db:"avg_duration_minutes"`
}

// TestCaseDTO для создания/обновления тест-кейса
type TestCaseDTO struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	PriorityID  int    `json:"priorityId"`
	OwnerID     int    `json:"ownerId"`
	IsAutomated bool   `json:"isAutomated"`
}

// StepDTO для создания/обновления шага тест-кейса
type StepDTO struct {
	Order    int    `json:"order"`
	Action   string `json:"action"`
	Expected string `json:"expected"`
}

// RunDTO для создания тестового прогона
type RunDTO struct {
	Name      string `json:"name"`
	PlanID    int    `json:"planId"`
	VersionID int    `json:"versionId"`
	EnvID     int    `json:"envId"`
	UserID    int    `json:"userId"`
}

// ResultDTO для создания результата теста
type ResultDTO struct {
	RunItemID  int    `json:"runItemId"`
	StatusID   int    `json:"statusId"`
	Summary    string `json:"summary"`
	ExecutorID int    `json:"executorId"`
}

// ArtifactDTO для добавления артефакта к результату
type ArtifactDTO struct {
	Path string `json:"path"`
	Name string `json:"name"`
}

// AutotestDTO для создания/обновления автотеста
type AutotestDTO struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	OwnerID     int    `json:"ownerId"`
	IsActive    bool   `json:"isActive"`
}

// VersionDTO для создания версии автотеста
type VersionDTO struct {
	VersionString string `json:"versionString"`
	CommitHash    string `json:"commitHash"`
}

