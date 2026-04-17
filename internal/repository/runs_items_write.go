package repository

import (
	"context"
	"fmt"

	"github.com/untibullet/secure-db-manager/internal/domain"
)

func (r *AppRepository) CreateRun(ctx context.Context, dto domain.RunDTO) (int, error) {
	var id int
	err := r.db.QueryRow(ctx,
		`INSERT INTO v_test_runs_manage (name, test_plan_id, version_id, env_config_id, tool_config_id, created_by, start_date, status)
		 VALUES ($1, $2, $3, $4, $5, $6, NOW(), 'PLANNED') RETURNING test_run_id`,
		dto.Name, dto.PlanID, dto.VersionID, dto.EnvID, dto.ToolConfigID, dto.UserID,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("CreateRun: %w", err)
	}
	return id, nil
}

func (r *AppRepository) UpdateRunStatus(ctx context.Context, id int, status string) error {
	_, err := r.db.Exec(ctx, `UPDATE v_test_runs_manage SET status=$1 WHERE test_run_id=$2`, status, id)
	if err != nil {
		return fmt.Errorf("UpdateRunStatus: %w", err)
	}
	return nil
}

func (r *AppRepository) DeleteRun(ctx context.Context, id int) error {
	_, err := r.db.Exec(ctx, `DELETE FROM v_test_runs_manage WHERE test_run_id=$1`, id)
	if err != nil {
		return fmt.Errorf("DeleteRun: %w", err)
	}
	return nil
}

func (r *AppRepository) AddRunItem(ctx context.Context, runID, testCaseID, order int) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO v_test_run_items (test_run_id, test_case_id, execution_order) VALUES ($1, $2, $3)`,
		runID, testCaseID, order,
	)
	if err != nil {
		return fmt.Errorf("AddRunItem: %w", err)
	}
	return nil
}

func (r *AppRepository) RemoveRunItem(ctx context.Context, itemID int) error {
	_, err := r.db.Exec(ctx, `DELETE FROM v_test_run_items WHERE run_item_id=$1`, itemID)
	if err != nil {
		return fmt.Errorf("RemoveRunItem: %w", err)
	}
	return nil
}
