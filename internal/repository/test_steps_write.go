package repository

import (
	"context"
	"fmt"

	"github.com/untibullet/secure-db-manager/internal/domain"
)

func (r *AppRepository) CreateTestCase(ctx context.Context, dto domain.TestCaseDTO) (int, error) {
	var id int
	err := r.db.QueryRow(ctx,
		`INSERT INTO v_test_cases_manage (name, description, priority_id, owner_user_id, is_automated)
		 VALUES ($1, $2, $3, $4, $5) RETURNING test_case_id`,
		dto.Name, dto.Description, dto.PriorityID, dto.OwnerID, dto.IsAutomated,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("CreateTestCase: %w", err)
	}
	return id, nil
}

func (r *AppRepository) UpdateTestCase(ctx context.Context, id int, dto domain.TestCaseDTO) error {
	_, err := r.db.Exec(ctx,
		`UPDATE v_test_cases_manage SET name=$1, description=$2, priority_id=$3 WHERE test_case_id=$4`,
		dto.Name, dto.Description, dto.PriorityID, id,
	)
	if err != nil {
		return fmt.Errorf("UpdateTestCase: %w", err)
	}
	return nil
}

func (r *AppRepository) DeleteTestCase(ctx context.Context, id int) error {
	_, err := r.db.Exec(ctx, `DELETE FROM v_test_cases_manage WHERE test_case_id=$1`, id)
	if err != nil {
		return fmt.Errorf("DeleteTestCase: %w", err)
	}
	return nil
}

func (r *AppRepository) CreateTestCaseStep(ctx context.Context, tcID int, dto domain.StepDTO) (int, error) {
	var id int
	err := r.db.QueryRow(ctx,
		`INSERT INTO v_test_case_steps (test_case_id, step_order, action_text, expected_result)
		 VALUES ($1, $2, $3, $4) RETURNING step_id`,
		tcID, dto.Order, dto.Action, dto.Expected,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("CreateTestCaseStep: %w", err)
	}
	return id, nil
}

func (r *AppRepository) UpdateTestCaseStep(ctx context.Context, id int, dto domain.StepDTO) error {
	_, err := r.db.Exec(ctx,
		`UPDATE v_test_case_steps SET step_order=$1, action_text=$2, expected_result=$3 WHERE step_id=$4`,
		dto.Order, dto.Action, dto.Expected, id,
	)
	if err != nil {
		return fmt.Errorf("UpdateTestCaseStep: %w", err)
	}
	return nil
}

func (r *AppRepository) DeleteTestCaseStep(ctx context.Context, id int) error {
	_, err := r.db.Exec(ctx, `DELETE FROM v_test_case_steps WHERE step_id=$1`, id)
	if err != nil {
		return fmt.Errorf("DeleteTestCaseStep: %w", err)
	}
	return nil
}
