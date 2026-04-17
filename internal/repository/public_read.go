package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/untibullet/secure-db-manager/internal/domain"
)

func (r *AppRepository) ListPublicTestPlans(ctx context.Context, filter domain.Filter, paging domain.Paging) ([]domain.TestPlan, error) {
	where, args, err := buildWhereClause(filter, &paging)
	if err != nil {
		return nil, fmt.Errorf("ListPublicTestPlans: %w", err)
	}
	sql := `SELECT test_plan_id, test_plan_name, description, start_date, end_date, status, owner, priority
	        FROM v_public_test_plans` + where

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("ListPublicTestPlans: %w", err)
	}
	defer rows.Close()

	var plans []domain.TestPlan
	for rows.Next() {
		var tp domain.TestPlan
		if err := rows.Scan(&tp.ID, &tp.Name, &tp.Description, &tp.StartDate, &tp.EndDate, &tp.Status, &tp.Owner, &tp.Priority); err != nil {
			return nil, fmt.Errorf("ListPublicTestPlans: %w", err)
		}
		plans = append(plans, tp)
	}
	return plans, rows.Err()
}

func (r *AppRepository) GetTestPlanByID(ctx context.Context, id int) (*domain.TestPlan, error) {
	sql := `SELECT test_plan_id, test_plan_name, description, start_date, end_date, status, owner, priority
	        FROM v_public_test_plans WHERE test_plan_id = $1`
	var tp domain.TestPlan
	err := r.db.QueryRow(ctx, sql, id).Scan(&tp.ID, &tp.Name, &tp.Description, &tp.StartDate, &tp.EndDate, &tp.Status, &tp.Owner, &tp.Priority)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("GetTestPlanByID: %w", err)
	}
	return &tp, nil
}

func (r *AppRepository) ListPublicResults(ctx context.Context, filter domain.Filter, paging domain.Paging) ([]domain.PublicResult, error) {
	where, args, err := buildWhereClause(filter, &paging)
	if err != nil {
		return nil, fmt.Errorf("ListPublicResults: %w", err)
	}
	sql := `SELECT test_result_id, test_run_id, test_case_name, status, execution_date
	        FROM v_public_results` + where

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("ListPublicResults: %w", err)
	}
	defer rows.Close()

	var results []domain.PublicResult
	for rows.Next() {
		var tr domain.PublicResult
		if err := rows.Scan(&tr.ResultID, &tr.RunID, &tr.TestCaseName, &tr.Status, &tr.ExecutionDate); err != nil {
			return nil, fmt.Errorf("ListPublicResults: %w", err)
		}
		results = append(results, tr)
	}
	return results, rows.Err()
}

func (r *AppRepository) ListSharedEnvironments(ctx context.Context, filter domain.Filter) ([]domain.Environment, error) {
	where, args, err := buildWhereClause(filter, nil)
	if err != nil {
		return nil, fmt.Errorf("ListSharedEnvironments: %w", err)
	}
	sql := `SELECT env_config_id, environment_name, description, is_active, params_count
	        FROM v_shared_environments` + where

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("ListSharedEnvironments: %w", err)
	}
	defer rows.Close()

	var environments []domain.Environment
	for rows.Next() {
		var env domain.Environment
		if err := rows.Scan(&env.ID, &env.Name, &env.Description, &env.IsActive, &env.ParamsCount); err != nil {
			return nil, fmt.Errorf("ListSharedEnvironments: %w", err)
		}
		environments = append(environments, env)
	}
	return environments, rows.Err()
}

func (r *AppRepository) ListSharedReports(ctx context.Context, filter domain.Filter, paging domain.Paging) ([]domain.Report, error) {
	where, args, err := buildWhereClause(filter, &paging)
	if err != nil {
		return nil, fmt.Errorf("ListSharedReports: %w", err)
	}
	sql := `SELECT report_id, report_name, creation_date, template, author
	        FROM v_shared_reports` + where

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("ListSharedReports: %w", err)
	}
	defer rows.Close()

	var reports []domain.Report
	for rows.Next() {
		var rep domain.Report
		if err := rows.Scan(&rep.ID, &rep.Name, &rep.CreationDate, &rep.Template, &rep.Author); err != nil {
			return nil, fmt.Errorf("ListSharedReports: %w", err)
		}
		reports = append(reports, rep)
	}
	return reports, rows.Err()
}
