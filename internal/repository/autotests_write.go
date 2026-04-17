package repository

import (
	"context"
	"fmt"

	"github.com/untibullet/secure-db-manager/internal/domain"
)

func (r *AppRepository) CreateAutotest(ctx context.Context, dto domain.AutotestDTO) (int, error) {
	var id int
	err := r.db.QueryRow(ctx,
		`INSERT INTO v_autotests_manage (test_case_id, name, description, owner_user_id, is_active)
		 VALUES ($1, $2, $3, $4, $5) RETURNING autotest_id`,
		dto.TestCaseID, dto.Name, dto.Description, dto.OwnerID, dto.IsActive,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("CreateAutotest: %w", err)
	}
	return id, nil
}

func (r *AppRepository) UpdateAutotest(ctx context.Context, id int, dto domain.AutotestDTO) error {
	_, err := r.db.Exec(ctx,
		`UPDATE v_autotests_manage SET name=$1, description=$2, is_active=$3 WHERE autotest_id=$4`,
		dto.Name, dto.Description, dto.IsActive, id,
	)
	if err != nil {
		return fmt.Errorf("UpdateAutotest: %w", err)
	}
	return nil
}

func (r *AppRepository) DeleteAutotest(ctx context.Context, id int) error {
	_, err := r.db.Exec(ctx, `DELETE FROM v_autotests_manage WHERE autotest_id=$1`, id)
	if err != nil {
		return fmt.Errorf("DeleteAutotest: %w", err)
	}
	return nil
}

func (r *AppRepository) CreateAutotestVersion(ctx context.Context, autotestID int, dto domain.VersionDTO) (int, error) {
	var id int
	err := r.db.QueryRow(ctx,
		`INSERT INTO v_autotest_versions (autotest_id, version_string, commit_hash)
		 VALUES ($1, $2, $3) RETURNING version_id`,
		autotestID, dto.VersionString, dto.CommitHash,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("CreateAutotestVersion: %w", err)
	}
	return id, nil
}
