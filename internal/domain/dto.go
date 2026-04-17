package domain

import "time"

type TestPlanDTO struct {
	Name               string     `json:"name"`
	Description        string     `json:"description"`
	PriorityID         int        `json:"priority_id"`
	StartDate          time.Time  `json:"start_date"`
	EndDate            time.Time  `json:"end_date"`
	AcceptanceCriteria string     `json:"acceptance_criteria"`
	OwnerID            int        `json:"owner_id"`
	Status             string     `json:"status"`
}

type TestCaseDTO struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	PriorityID  int    `json:"priority_id"`
	OwnerID     int    `json:"owner_id"`
	IsAutomated bool   `json:"is_automated"`
}

type StepDTO struct {
	Order    int    `json:"order"`
	Action   string `json:"action"`
	Expected string `json:"expected"`
}

type RunDTO struct {
	Name         string `json:"name"`
	PlanID       int    `json:"plan_id"`
	VersionID    int    `json:"version_id"`
	EnvID        int    `json:"env_id"`
	ToolConfigID int    `json:"tool_config_id"`
	UserID       int    `json:"user_id"`
}

type ResultDTO struct {
	RunItemID  int    `json:"run_item_id"`
	StatusID   int    `json:"status_id"`
	Summary    string `json:"summary"`
	ExecutorID int    `json:"executor_id"`
}

type ArtifactDTO struct {
	Kind string `json:"kind"`
	Path string `json:"path"`
}

type AutotestDTO struct {
	TestCaseID  int    `json:"test_case_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	OwnerID     int    `json:"owner_id"`
	IsActive    bool   `json:"is_active"`
}

type VersionDTO struct {
	VersionString string `json:"version_string"`
	CommitHash    string `json:"commit_hash"`
}

type CreateUserDTO struct {
	Username string `json:"username"`
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}
