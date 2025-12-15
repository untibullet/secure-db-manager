package repository

import "context"

func (r *Repository) ListEngineerAutotests(ctx context.Context, filter Filter, paging Paging) ([]Autotest, error) {
	where, args := buildWhereClause(filter, &paging)
	sql := `SELECT autotest_id, autotest_name, description, is_active, version_string, commit_hash, author 
            FROM v_engineer_autotests` + where

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var autotests []Autotest
	for rows.Next() {
		var t Autotest
		if err := rows.Scan(&t.ID, &t.Name, &t.Description, &t.IsActive, &t.VersionString, &t.CommitHash, &t.Author); err != nil {
			return nil, err
		}
		autotests = append(autotests, t)
	}
	return autotests, nil
}

func (r *Repository) ListLeadAllCases(ctx context.Context, filter Filter, paging Paging) ([]TestCase, error) {
	where, args := buildWhereClause(filter, &paging)
	sql := `SELECT test_case_id, test_case_name, description, is_automated, is_active, priority, owner 
            FROM v_lead_all_cases` + where

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var testcases []TestCase
	for rows.Next() {
		var tc TestCase
		if err := rows.Scan(&tc.ID, &tc.Name, &tc.Description, &tc.IsAutomated, &tc.IsActive, &tc.Priority, &tc.Owner); err != nil {
			return nil, err
		}
		testcases = append(testcases, tc)
	}
	return testcases, nil
}

func (r *Repository) ListLeadAllResults(ctx context.Context, filter Filter, paging Paging) ([]TestResult, error) {
	where, args := buildWhereClause(filter, &paging)
	sql := `SELECT test_result_id, test_run_name, test_case_name, status, executor, execution_date, error_message 
            FROM v_lead_all_results` + where
	
	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []TestResult
	for rows.Next() {
		var tr TestResult
		if err := rows.Scan(&tr.ResultID, &tr.RunName, &tr.TestCaseName, &tr.Status, &tr.ExecutionDate, &tr.ErrorMessage); err != nil {
			return nil, err
		}
		results = append(results, tr)
	}
	return results, nil
}

func (r *Repository) ListActiveRuns(ctx context.Context, filter Filter, paging Paging) ([]ActiveRun, error) {
	where, args := buildWhereClause(filter, &paging)
	sql := `SELECT test_run_id, run_name, plan_name, status, passed_tests, failed_tests 
            FROM v_active_test_runs` + where
	
	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var runs []ActiveRun
	for rows.Next() {
		var rn ActiveRun
		if err := rows.Scan(&rn.RunID, &rn.RunName, &rn.PlanName, &rn.Status, &rn.Status, &rn.PassedTests, &rn.FailedTests); err != nil {
			return nil, err
		}
		runs = append(runs, rn)
	}
	return runs, nil
}

func (r *Repository) ListRunSummaries(ctx context.Context, filter Filter, paging Paging) ([]RunSummary, error) {
	where, args := buildWhereClause(filter, &paging)
	// Используем v_test_execution_summary, как наиболее подходящее под "RunSummaries"
	sql := `SELECT test_run_id, test_run_name, total_cases, passed, failed, pass_rate 
            FROM v_test_execution_summary` + where
	
	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summs []RunSummary
	for rows.Next() {
		var sm RunSummary
		if err := rows.Scan(&sm.RunID, &sm.RunName, &sm.TotalCases, &sm.Passed, &sm.Failed, &sm.PassRate); err != nil {
			return nil, err
		}
		summs = append(summs, sm)
	}
	return summs, nil
}

func (r *Repository) ListTestCaseStatistics(ctx context.Context, filter Filter, paging Paging) ([]CaseStats, error) {
	where, args := buildWhereClause(filter, &paging)
	sql := `SELECT test_case_id, name, pass_rate_percent, avg_duration_minutes 
            FROM v_test_case_statistics` + where
	
	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var steps []CaseStats
	for rows.Next() {
		var st CaseStats
		if err := rows.Scan(&st.TestCaseID, &st.Name, &st.PassRatePercent, &st.AvgDurationMins); err != nil {
			return nil, err
		}
		steps = append(steps, st)
	}
	return steps, nil
}
