package repository

import "context"

func (r *Repository) CreateRun(ctx context.Context, dto RunDTO) (int, error) {
	sql := `INSERT INTO v_test_runs_manage (name, test_plan_id, version_id, env_config_id, tool_config_id, created_by, start_date, status)
            VALUES ($1, $2, $3, $4, $5, $6, NOW(), 'PLANNED') RETURNING test_run_id`
	var id int
	err := r.db.QueryRow(ctx, sql, dto.Name, dto.PlanID, dto.VersionID, dto.EnvID, dto.ToolConfigID, dto.UserID).Scan(&id)
	return id, err
}

func (r *Repository) UpdateRunStatus(ctx context.Context, id int, status string) error {
	sql := `UPDATE v_test_runs_manage SET status=$1 WHERE test_run_id=$2`
	_, err := r.db.Exec(ctx, sql, status, id)
	return err
}

func (r *Repository) DeleteRun(ctx context.Context, id int) error {
	sql := `DELETE FROM v_test_runs_manage WHERE test_run_id=$1`
	_, err := r.db.Exec(ctx, sql, id)
	return err
}

func (r *Repository) AddRunItem(ctx context.Context, runID, testCaseID, order int) error {
	sql := `INSERT INTO v_test_run_items (test_run_id, test_case_id, execution_order) VALUES ($1, $2, $3)`
	_, err := r.db.Exec(ctx, sql, runID, testCaseID, order)
	return err
}

func (r *Repository) RemoveRunItem(ctx context.Context, itemID int) error {
    sql := `DELETE FROM v_test_run_items WHERE run_item_id = $1`
    _, err := r.db.Exec(ctx, sql, itemID)
    return err
}

