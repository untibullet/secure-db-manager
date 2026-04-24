package repository

import (
	"context"
	"fmt"

	"github.com/untibullet/secure-db-manager/internal/domain"
)

func (r *AppRepository) CreateTestPlan(ctx context.Context, dto domain.TestPlanDTO) (int, error) {
	var id int
	err := r.db.QueryRow(ctx,
		`INSERT INTO v_test_plans_manage (name, description, priority_id, start_date, end_date, acceptance_criteria, owner_user_id, status)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING test_plan_id`,
		dto.Name, dto.Description, dto.PriorityID, dto.StartDate, dto.EndDate,
		dto.AcceptanceCriteria, dto.OwnerID, dto.Status,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("CreateTestPlan: %w", mapRepoError(err))
	}
	return id, nil
}

func (r *AppRepository) UpdateTestPlan(ctx context.Context, id int, dto domain.TestPlanDTO) error {
	_, err := r.db.Exec(ctx,
		`UPDATE v_test_plans_manage SET name=$1, description=$2, priority_id=$3, start_date=$4, end_date=$5,
		 acceptance_criteria=$6, status=$7 WHERE test_plan_id=$8`,
		dto.Name, dto.Description, dto.PriorityID, dto.StartDate, dto.EndDate,
		dto.AcceptanceCriteria, dto.Status, id,
	)
	if err != nil {
		return fmt.Errorf("UpdateTestPlan: %w", mapRepoError(err))
	}
	return nil
}

func (r *AppRepository) DeleteTestPlan(ctx context.Context, id int) error {
	_, err := r.db.Exec(ctx, `DELETE FROM v_test_plans_manage WHERE test_plan_id=$1`, id)
	if err != nil {
		return fmt.Errorf("DeleteTestPlan: %w", mapRepoError(err))
	}
	return nil
}
