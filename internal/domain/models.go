package domain

import (
	"database/sql"
	"time"
)

type TestPlan struct {
	ID          int            `json:"test_plan_id"`
	Name        string         `json:"test_plan_name"`
	Description sql.NullString `json:"description"`
	StartDate   time.Time      `json:"start_date"`
	EndDate     time.Time      `json:"end_date"`
	Status      string         `json:"status"`
	Owner       sql.NullString `json:"owner"`
	Priority    string         `json:"priority"`
}

type TestCase struct {
	ID          int            `json:"test_case_id"`
	Name        string         `json:"test_case_name"`
	Description sql.NullString `json:"description"`
	IsAutomated bool           `json:"is_automated"`
	IsActive    bool           `json:"is_active"`
	Priority    string         `json:"priority"`
	Owner       sql.NullString `json:"owner"`
}

type Step struct {
	ID       int    `json:"step_id"`
	CaseID   int    `json:"test_case_id"`
	Order    int    `json:"step_order"`
	Action   string `json:"action_text"`
	Expected string `json:"expected_result"`
}

type Autotest struct {
	ID            int            `json:"autotest_id"`
	Name          string         `json:"autotest_name"`
	Description   sql.NullString `json:"description"`
	IsActive      bool           `json:"is_active"`
	VersionString string         `json:"version_string"`
	CommitHash    sql.NullString `json:"commit_hash"`
	Author        sql.NullString `json:"author"`
}

type AutotestVersion struct {
	ID            int            `json:"version_id"`
	AutotestID    int            `json:"autotest_id"`
	VersionString string         `json:"version_string"`
	CommitHash    sql.NullString `json:"commit_hash"`
	Description   sql.NullString `json:"change_description"`
	CreatedAt     time.Time      `json:"created_at"`
}

type ActiveRun struct {
	RunID        int    `json:"test_run_id"`
	RunName      string `json:"run_name"`
	PlanName     string `json:"plan_name"`
	Status       string `json:"status"`
	PassedTests  int    `json:"passed_tests"`
	FailedTests  int    `json:"failed_tests"`
	BlockedTests int    `json:"blocked_tests"`
}

type RunItem struct {
	ItemID    int `json:"run_item_id"`
	RunID     int `json:"test_run_id"`
	CaseID    int `json:"test_case_id"`
	ExecOrder int `json:"execution_order"`
}

type TestResult struct {
	ResultID      int            `json:"test_result_id"`
	RunName       string         `json:"test_run_name"`
	TestCaseName  string         `json:"test_case_name"`
	Status        string         `json:"status"`
	Executor      string         `json:"executor"`
	ExecutionDate time.Time      `json:"execution_date"`
	ErrorMessage  sql.NullString `json:"error_message"`
}

type PublicResult struct {
	ResultID      int       `json:"test_result_id"`
	RunID         int       `json:"test_run_id"`
	TestCaseName  string    `json:"test_case_name"`
	Status        string    `json:"status"`
	ExecutionDate time.Time `json:"execution_date"`
}

type Artifact struct {
	ID        int            `json:"artifact_id"`
	ResultID  int            `json:"test_result_id"`
	Kind      string         `json:"kind"`
	FilePath  string         `json:"file_path"`
	FileSize  sql.NullInt64  `json:"file_size_bytes"`
	MimeType  sql.NullString `json:"mime_type"`
	CreatedAt time.Time      `json:"created_at"`
}

type Environment struct {
	ID          int            `json:"env_config_id"`
	Name        string         `json:"environment_name"`
	Description sql.NullString `json:"description"`
	IsActive    bool           `json:"is_active"`
	ParamsCount int            `json:"params_count"`
}

type Report struct {
	ID           int            `json:"report_id"`
	Name         string         `json:"report_name"`
	CreationDate time.Time      `json:"creation_date"`
	Template     string         `json:"template"`
	Author       sql.NullString `json:"author"`
}

type UserAdminView struct {
	ID        int            `json:"user_id"`
	Username  string         `json:"username"`
	FullName  sql.NullString `json:"full_name"`
	Email     sql.NullString `json:"email"`
	IsActive  bool           `json:"is_active"`
	LastLogin sql.NullTime   `json:"last_login"`
	Roles     sql.NullString `json:"roles"`
}

type AuditEntry struct {
	AuditID   int            `json:"audit_id"`
	TableName string         `json:"table_name"`
	Operation string         `json:"operation"`
	RecordID  int            `json:"record_id"`
	UserID    sql.NullInt64  `json:"user_id"`
	ChangedAt time.Time      `json:"changed_at"`
	OldValues sql.NullString `json:"old_values"`
	NewValues sql.NullString `json:"new_values"`
	IPAddress sql.NullString `json:"ip_address"`
}

type RunSummary struct {
	RunID      int             `json:"test_run_id"`
	RunName    string          `json:"test_run_name"`
	TotalCases int             `json:"total_cases"`
	Passed     int             `json:"passed"`
	Failed     int             `json:"failed"`
	PassRate   sql.NullFloat64 `json:"pass_rate"`
}

type CaseStats struct {
	TestCaseID      int             `json:"test_case_id"`
	Name            string          `json:"name"`
	PassRatePercent sql.NullFloat64 `json:"pass_rate_percent"`
	AvgDurationMins sql.NullFloat64 `json:"avg_duration_minutes"`
}

type Role struct {
	ID   int    `json:"role_id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

// UserAuth используется исключительно при логине (AD-11).
// Не передавать в логи — содержит password_hash.
type UserAuth struct {
	UserID             int
	Username           string
	PasswordHash       string
	IsActive           bool
	AccountLockedUntil *time.Time
	RoleCode           string
}
