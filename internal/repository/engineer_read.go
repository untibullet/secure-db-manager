package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/untibullet/secure-db-manager/internal/domain"
)

func (r *AppRepository) ListEngineerAutotests(ctx context.Context, filter domain.Filter, paging domain.Paging) ([]domain.Autotest, error) {
	where, args, err := buildWhereClause(filter, &paging)
	if err != nil {
		return nil, fmt.Errorf("ListEngineerAutotests: %w", err)
	}
	sql := `SELECT autotest_id, autotest_name, description, is_active, version_string, commit_hash, author
	        FROM v_engineer_autotests` + where

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("ListEngineerAutotests: %w", err)
	}
	defer rows.Close()

	var autotests []domain.Autotest
	for rows.Next() {
		var t domain.Autotest
		if err := rows.Scan(&t.ID, &t.Name, &t.Description, &t.IsActive, &t.VersionString, &t.CommitHash, &t.Author); err != nil {
			return nil, fmt.Errorf("ListEngineerAutotests: %w", err)
		}
		autotests = append(autotests, t)
	}
	return autotests, rows.Err()
}

func (r *AppRepository) GetAutotestByID(ctx context.Context, id int) (*domain.Autotest, error) {
	sql := `SELECT autotest_id, autotest_name, description, is_active, version_string, commit_hash, author
	        FROM v_engineer_autotests WHERE autotest_id = $1`
	var t domain.Autotest
	err := r.db.QueryRow(ctx, sql, id).Scan(&t.ID, &t.Name, &t.Description, &t.IsActive, &t.VersionString, &t.CommitHash, &t.Author)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("GetAutotestByID: %w", err)
	}
	return &t, nil
}

func (r *AppRepository) ListAutotestVersions(ctx context.Context, autotestID int) ([]domain.AutotestVersion, error) {
	sql := `SELECT version_id, autotest_id, version_string, commit_hash, change_description, created_at
	        FROM v_autotest_versions WHERE autotest_id = $1 ORDER BY created_at DESC`

	rows, err := r.db.Query(ctx, sql, autotestID)
	if err != nil {
		return nil, fmt.Errorf("ListAutotestVersions: %w", err)
	}
	defer rows.Close()

	var versions []domain.AutotestVersion
	for rows.Next() {
		var v domain.AutotestVersion
		if err := rows.Scan(&v.ID, &v.AutotestID, &v.VersionString, &v.CommitHash, &v.Description, &v.CreatedAt); err != nil {
			return nil, fmt.Errorf("ListAutotestVersions: %w", err)
		}
		versions = append(versions, v)
	}
	return versions, rows.Err()
}

func (r *AppRepository) ListLeadAllCases(ctx context.Context, filter domain.Filter, paging domain.Paging) ([]domain.TestCase, error) {
	where, args, err := buildWhereClause(filter, &paging)
	if err != nil {
		return nil, fmt.Errorf("ListLeadAllCases: %w", err)
	}
	sql := `SELECT test_case_id, test_case_name, description, is_automated, is_active, priority, owner
	        FROM v_lead_all_cases` + where

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("ListLeadAllCases: %w", err)
	}
	defer rows.Close()

	var testcases []domain.TestCase
	for rows.Next() {
		var tc domain.TestCase
		if err := rows.Scan(&tc.ID, &tc.Name, &tc.Description, &tc.IsAutomated, &tc.IsActive, &tc.Priority, &tc.Owner); err != nil {
			return nil, fmt.Errorf("ListLeadAllCases: %w", err)
		}
		testcases = append(testcases, tc)
	}
	return testcases, rows.Err()
}

func (r *AppRepository) GetTestCaseByID(ctx context.Context, id int) (*domain.TestCase, error) {
	sql := `SELECT test_case_id, test_case_name, description, is_automated, is_active, priority, owner
	        FROM v_lead_all_cases WHERE test_case_id = $1`
	var tc domain.TestCase
	err := r.db.QueryRow(ctx, sql, id).Scan(&tc.ID, &tc.Name, &tc.Description, &tc.IsAutomated, &tc.IsActive, &tc.Priority, &tc.Owner)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("GetTestCaseByID: %w", err)
	}
	return &tc, nil
}

func (r *AppRepository) ListTestCaseSteps(ctx context.Context, testCaseID int) ([]domain.Step, error) {
	sql := `SELECT step_id, test_case_id, step_order, action_text, expected_result
	        FROM v_test_case_steps WHERE test_case_id = $1 ORDER BY step_order`

	rows, err := r.db.Query(ctx, sql, testCaseID)
	if err != nil {
		return nil, fmt.Errorf("ListTestCaseSteps: %w", err)
	}
	defer rows.Close()

	var steps []domain.Step
	for rows.Next() {
		var s domain.Step
		if err := rows.Scan(&s.ID, &s.CaseID, &s.Order, &s.Action, &s.Expected); err != nil {
			return nil, fmt.Errorf("ListTestCaseSteps: %w", err)
		}
		steps = append(steps, s)
	}
	return steps, rows.Err()
}

func (r *AppRepository) ListLeadAllResults(ctx context.Context, filter domain.Filter, paging domain.Paging) ([]domain.TestResult, error) {
	where, args, err := buildWhereClause(filter, &paging)
	if err != nil {
		return nil, fmt.Errorf("ListLeadAllResults: %w", err)
	}
	sql := `SELECT test_result_id, test_run_name, test_case_name, status, executor, execution_date, error_message
	        FROM v_lead_all_results` + where

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("ListLeadAllResults: %w", err)
	}
	defer rows.Close()

	var results []domain.TestResult
	for rows.Next() {
		var tr domain.TestResult
		if err := rows.Scan(&tr.ResultID, &tr.RunName, &tr.TestCaseName, &tr.Status, &tr.Executor, &tr.ExecutionDate, &tr.ErrorMessage); err != nil {
			return nil, fmt.Errorf("ListLeadAllResults: %w", err)
		}
		results = append(results, tr)
	}
	return results, rows.Err()
}

func (r *AppRepository) GetLeadResultByID(ctx context.Context, id int) (*domain.TestResult, error) {
	sql := `SELECT test_result_id, test_run_name, test_case_name, status, executor, execution_date, error_message
	        FROM v_lead_all_results WHERE test_result_id = $1`
	var tr domain.TestResult
	err := r.db.QueryRow(ctx, sql, id).Scan(&tr.ResultID, &tr.RunName, &tr.TestCaseName, &tr.Status, &tr.Executor, &tr.ExecutionDate, &tr.ErrorMessage)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("GetLeadResultByID: %w", err)
	}
	return &tr, nil
}

func (r *AppRepository) ListTesterResults(ctx context.Context, filter domain.Filter, paging domain.Paging) ([]domain.PublicResult, error) {
	where, args, err := buildWhereClause(filter, &paging)
	if err != nil {
		return nil, fmt.Errorf("ListTesterResults: %w", err)
	}
	sql := `SELECT test_result_id, test_run_id, test_case_name, status, execution_date
	        FROM v_tester_my_results` + where

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("ListTesterResults: %w", err)
	}
	defer rows.Close()

	var results []domain.PublicResult
	for rows.Next() {
		var tr domain.PublicResult
		if err := rows.Scan(&tr.ResultID, &tr.RunID, &tr.TestCaseName, &tr.Status, &tr.ExecutionDate); err != nil {
			return nil, fmt.Errorf("ListTesterResults: %w", err)
		}
		results = append(results, tr)
	}
	return results, rows.Err()
}

// GetTesterResultByID возвращает результат через v_tester_my_results (RLS: только свои данные).
func (r *AppRepository) GetTesterResultByID(ctx context.Context, id int) (*domain.PublicResult, error) {
	sql := `SELECT test_result_id, test_run_id, test_case_name, status, execution_date
	        FROM v_tester_my_results WHERE test_result_id = $1`
	var tr domain.PublicResult
	err := r.db.QueryRow(ctx, sql, id).Scan(&tr.ResultID, &tr.RunID, &tr.TestCaseName, &tr.Status, &tr.ExecutionDate)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("GetTesterResultByID: %w", err)
	}
	return &tr, nil
}

func (r *AppRepository) ListResultArtifacts(ctx context.Context, resultID int) ([]domain.Artifact, error) {
	sql := `SELECT artifact_id, test_result_id, kind, file_path, file_size_bytes, mime_type, created_at
	        FROM v_test_result_artifacts WHERE test_result_id = $1`

	rows, err := r.db.Query(ctx, sql, resultID)
	if err != nil {
		return nil, fmt.Errorf("ListResultArtifacts: %w", err)
	}
	defer rows.Close()

	var artifacts []domain.Artifact
	for rows.Next() {
		var a domain.Artifact
		if err := rows.Scan(&a.ID, &a.ResultID, &a.Kind, &a.FilePath, &a.FileSize, &a.MimeType, &a.CreatedAt); err != nil {
			return nil, fmt.Errorf("ListResultArtifacts: %w", err)
		}
		artifacts = append(artifacts, a)
	}
	return artifacts, rows.Err()
}

func (r *AppRepository) ListActiveRuns(ctx context.Context, filter domain.Filter, paging domain.Paging) ([]domain.ActiveRun, error) {
	where, args, err := buildWhereClause(filter, &paging)
	if err != nil {
		return nil, fmt.Errorf("ListActiveRuns: %w", err)
	}
	sql := `SELECT test_run_id, run_name, plan_name, status, passed_tests, failed_tests, blocked_tests
	        FROM v_active_test_runs` + where

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("ListActiveRuns: %w", err)
	}
	defer rows.Close()

	var runs []domain.ActiveRun
	for rows.Next() {
		var rn domain.ActiveRun
		if err := rows.Scan(&rn.RunID, &rn.RunName, &rn.PlanName, &rn.Status, &rn.PassedTests, &rn.FailedTests, &rn.BlockedTests); err != nil {
			return nil, fmt.Errorf("ListActiveRuns: %w", err)
		}
		runs = append(runs, rn)
	}
	return runs, rows.Err()
}

func (r *AppRepository) GetRunByID(ctx context.Context, id int) (*domain.ActiveRun, error) {
	sql := `SELECT test_run_id, run_name, plan_name, status, passed_tests, failed_tests, blocked_tests
	        FROM v_active_test_runs WHERE test_run_id = $1`
	var rn domain.ActiveRun
	err := r.db.QueryRow(ctx, sql, id).Scan(&rn.RunID, &rn.RunName, &rn.PlanName, &rn.Status, &rn.PassedTests, &rn.FailedTests, &rn.BlockedTests)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("GetRunByID: %w", err)
	}
	return &rn, nil
}

func (r *AppRepository) ListRunItems(ctx context.Context, runID int) ([]domain.RunItem, error) {
	sql := `SELECT run_item_id, test_run_id, test_case_id, execution_order
	        FROM v_test_run_items WHERE test_run_id = $1 ORDER BY execution_order`

	rows, err := r.db.Query(ctx, sql, runID)
	if err != nil {
		return nil, fmt.Errorf("ListRunItems: %w", err)
	}
	defer rows.Close()

	var items []domain.RunItem
	for rows.Next() {
		var it domain.RunItem
		if err := rows.Scan(&it.ItemID, &it.RunID, &it.CaseID, &it.ExecOrder); err != nil {
			return nil, fmt.Errorf("ListRunItems: %w", err)
		}
		items = append(items, it)
	}
	return items, rows.Err()
}

func (r *AppRepository) ListRunSummaries(ctx context.Context, filter domain.Filter, paging domain.Paging) ([]domain.RunSummary, error) {
	where, args, err := buildWhereClause(filter, &paging)
	if err != nil {
		return nil, fmt.Errorf("ListRunSummaries: %w", err)
	}
	sql := `SELECT test_run_id, test_run_name, total_cases, passed, failed, pass_rate
	        FROM v_test_execution_summary` + where

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("ListRunSummaries: %w", err)
	}
	defer rows.Close()

	var summs []domain.RunSummary
	for rows.Next() {
		var sm domain.RunSummary
		if err := rows.Scan(&sm.RunID, &sm.RunName, &sm.TotalCases, &sm.Passed, &sm.Failed, &sm.PassRate); err != nil {
			return nil, fmt.Errorf("ListRunSummaries: %w", err)
		}
		summs = append(summs, sm)
	}
	return summs, rows.Err()
}

func (r *AppRepository) ListTestCaseStatistics(ctx context.Context, filter domain.Filter, paging domain.Paging) ([]domain.CaseStats, error) {
	where, args, err := buildWhereClause(filter, &paging)
	if err != nil {
		return nil, fmt.Errorf("ListTestCaseStatistics: %w", err)
	}
	sql := `SELECT test_case_id, name, pass_rate_percent, avg_duration_minutes
	        FROM v_test_case_statistics` + where

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("ListTestCaseStatistics: %w", err)
	}
	defer rows.Close()

	var stats []domain.CaseStats
	for rows.Next() {
		var st domain.CaseStats
		if err := rows.Scan(&st.TestCaseID, &st.Name, &st.PassRatePercent, &st.AvgDurationMins); err != nil {
			return nil, fmt.Errorf("ListTestCaseStatistics: %w", err)
		}
		stats = append(stats, st)
	}
	return stats, rows.Err()
}
