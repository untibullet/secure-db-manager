package repository

import "context"

func (r *AppRepository) CreateResult(ctx context.Context, dto ResultDTO) (int, error) {
	sql := `INSERT INTO v_results_manage (run_item_id, status_id, result_summary, executor_user_id)
	        VALUES ($1, $2, $3, $4) RETURNING test_result_id`
	var id int
	err := r.db.QueryRow(ctx, sql, dto.RunItemID, dto.StatusID, dto.Summary, dto.ExecutorID).Scan(&id)
	return id, err
}

func (r *AppRepository) UpdateResult(ctx context.Context, id int, dto ResultDTO) error {
	sql := `UPDATE v_results_manage SET status_id=$1, result_summary=$2 WHERE test_result_id=$3`
	_, err := r.db.Exec(ctx, sql, dto.StatusID, dto.Summary, id)
	return err
}

func (r *AppRepository) DeleteResult(ctx context.Context, id int) error {
	sql := `DELETE FROM v_results_manage WHERE test_result_id=$1`
	_, err := r.db.Exec(ctx, sql, id)
	return err
}

func (r *AppRepository) AddResultArtifact(ctx context.Context, resultID int, dto ArtifactDTO) (int, error) {
	sql := `INSERT INTO v_test_result_artifacts (test_result_id, kind, file_path)
            VALUES ($1, $2, $3) RETURNING artifact_id`
	var id int
	err := r.db.QueryRow(ctx, sql, resultID, dto.Kind, dto.Path).Scan(&id)
	return id, err
}
