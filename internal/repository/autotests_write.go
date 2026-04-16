package repository

import "context"

func (r *Repository) CreateAutotest(ctx context.Context, dto AutotestDTO) (int, error) {
	sql := `INSERT INTO v_autotests_manage (test_case_id, name, description, owner_user_id, is_active)
            VALUES ($1, $2, $3, $4, $5) RETURNING autotest_id`
	var id int
	err := r.db.QueryRow(ctx, sql, dto.TestCaseID, dto.Name, dto.Description, dto.OwnerID, dto.IsActive).Scan(&id)
	return id, err
}

func (r *Repository) UpdateAutotest(ctx context.Context, id int, dto AutotestDTO) error {
	sql := `UPDATE v_autotests_manage SET name=$1, description=$2, is_active=$3 WHERE autotest_id=$4`
	_, err := r.db.Exec(ctx, sql, dto.Name, dto.Description, dto.IsActive, id)
	return err
}

func (r *Repository) DeleteAutotest(ctx context.Context, id int) error {
	sql := `DELETE FROM v_autotests_manage WHERE autotest_id=$1`
	_, err := r.db.Exec(ctx, sql, id)
	return err
}

func (r *Repository) CreateAutotestVersion(ctx context.Context, autotestID int, dto VersionDTO) (int, error) {
	sql := `INSERT INTO v_autotest_versions (autotest_id, version_string, commit_hash) 
            VALUES ($1, $2, $3) RETURNING version_id`
	var id int
	err := r.db.QueryRow(ctx, sql, autotestID, dto.VersionString, dto.CommitHash).Scan(&id)
	return id, err
}
