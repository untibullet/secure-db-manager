package repository

import (
	"context"
	"fmt"

	"github.com/untibullet/secure-db-manager/internal/domain"
)

func (r *AppRepository) CreateResult(ctx context.Context, dto domain.ResultDTO) (int, error) {
	var id int
	err := r.db.QueryRow(ctx,
		`INSERT INTO v_results_manage (run_item_id, status_id, result_summary, executor_user_id)
		 VALUES ($1, $2, $3, $4) RETURNING test_result_id`,
		dto.RunItemID, dto.StatusID, dto.Summary, dto.ExecutorID,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("CreateResult: %w", err)
	}
	return id, nil
}

func (r *AppRepository) UpdateResult(ctx context.Context, id int, dto domain.ResultDTO) error {
	_, err := r.db.Exec(ctx,
		`UPDATE v_results_manage SET status_id=$1, result_summary=$2 WHERE test_result_id=$3`,
		dto.StatusID, dto.Summary, id,
	)
	if err != nil {
		return fmt.Errorf("UpdateResult: %w", err)
	}
	return nil
}

func (r *AppRepository) DeleteResult(ctx context.Context, id int) error {
	_, err := r.db.Exec(ctx, `DELETE FROM v_results_manage WHERE test_result_id=$1`, id)
	if err != nil {
		return fmt.Errorf("DeleteResult: %w", err)
	}
	return nil
}

func (r *AppRepository) AddResultArtifact(ctx context.Context, resultID int, dto domain.ArtifactDTO) (int, error) {
	var id int
	err := r.db.QueryRow(ctx,
		`INSERT INTO v_test_result_artifacts (test_result_id, kind, file_path)
		 VALUES ($1, $2, $3) RETURNING artifact_id`,
		resultID, dto.Kind, dto.Path,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("AddResultArtifact: %w", err)
	}
	return id, nil
}
