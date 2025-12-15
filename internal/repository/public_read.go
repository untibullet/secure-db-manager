package repository

import "context"

func (r *Repository) ListPublicTestPlans(ctx context.Context, filter Filter, paging Paging) ([]TestPlan, error) {
	where, args := buildWhereClause(filter, &paging)
	sql := `SELECT test_plan_id, test_plan_name, description, start_date, end_date, status, owner, priority 
            FROM v_public_test_plans` + where

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []TestPlan
	for rows.Next() {
		var tp TestPlan
		if err := rows.Scan(&tp.ID, &tp.Name, &tp.Description, &tp.StartDate, &tp.EndDate, &tp.Status, &tp.Owner, &tp.Priority); err != nil {
			return nil, err
		}
		plans = append(plans, tp)
	}
	return plans, nil
}

func (r *Repository) ListPublicResults(ctx context.Context, filter Filter, paging Paging) ([]PublicResult, error) {
	where, args := buildWhereClause(filter, &paging)
	sql := `SELECT test_result_id, test_run_id, test_case_name, status, execution_date 
            FROM v_public_results` + where
	
	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []PublicResult
	for rows.Next() {
		var tr PublicResult
		if err := rows.Scan(&tr.ResultID, &tr.RunID, &tr.TestCaseName, &tr.Status, &tr.ExecutionDate); err != nil {
			return nil, err
		}
		results = append(results, tr)
	}
	return results, nil
}

func (r *Repository) ListSharedEnvironments(ctx context.Context, filter Filter) ([]Environment, error) {
	// Paging не требуется по сигнатуре
	where, args := buildWhereClause(filter, nil)
	sql := `SELECT env_config_id, environment_name, description, is_active, params_count 
            FROM v_shared_environments` + where

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var environments []Environment
	for rows.Next() {
		var env Environment
		if err := rows.Scan(&env.ID, &env.Name, &env.Description, &env.IsActive, &env.ParamsCount); err != nil {
			return nil, err
		}
		environments = append(environments, env)
	}
	return environments, nil
}

func (r *Repository) ListSharedReports(ctx context.Context, filter Filter, paging Paging) ([]Report, error) {
	where, args := buildWhereClause(filter, &paging)
	sql := `SELECT report_id, report_name, creation_date, template, author 
            FROM v_shared_reports` + where

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []Report
	for rows.Next() {
		var rep Report
		if err := rows.Scan(&rep.ID, &rep.Name, &rep.CreationDate, &rep.Template, &rep.Author); err != nil {
			return nil, err
		}
		reports = append(reports, rep)
	}
	return reports, nil
}
