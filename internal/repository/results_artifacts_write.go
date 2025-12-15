package repository

import "context"

func (r *Repository) CreateResult(ctx context.Context, dto ResultDTO) (int, error) {
	// Тестировщик обычно пишет в v_tester_my_results, но лид может писать в v_lead_all_results.
	// Используем v_lead_all_results как более общее, при условии наличия прав.
	sql := `INSERT INTO v_lead_all_results (run_item_id, status_id, result_summary, executor_user_id) 
            VALUES ($1, $2, $3, $4) RETURNING test_result_id`
	var id int
	err := r.db.QueryRow(ctx, sql, dto.RunItemID, dto.StatusID, dto.Summary, dto.ExecutorID).Scan(&id)
	return id, err
}

func (r *Repository) UpdateResult(ctx context.Context, id int, dto ResultDTO) error {
    sql := `
        UPDATE v_lead_all_results
        SET status_id = $1, 
            result_summary = $2
        WHERE test_result_id = $3
    `
    _, err := r.db.Exec(ctx, sql, dto.StatusID, dto.Summary, id)
    return err
}

func (r *Repository) DeleteResult(ctx context.Context, id int) error {
    sql := `DELETE FROM v_lead_all_results WHERE test_result_id = $1`
    _, err := r.db.Exec(ctx, sql, id)
    return err
}

func (r *Repository) AddResultArtifact(ctx context.Context, resultID int, dto ArtifactDTO) (int, error) {
	sql := `INSERT INTO v_test_result_artifacts (test_result_id, file_path, file_name) 
            VALUES ($1, $2, $3) RETURNING artifact_id`
	var id int
	err := r.db.QueryRow(ctx, sql, resultID, dto.Path, dto.Name).Scan(&id)
	return id, err
}
