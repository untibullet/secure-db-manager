package repository

import "context"

func (r *Repository) CreateTestCase(ctx context.Context, dto TestCaseDTO) (int, error) {
	sql := `INSERT INTO v_lead_all_cases (name, description, priority_id, owner_user_id, is_automated) 
            VALUES ($1, $2, $3, $4, $5) RETURNING test_case_id`
	var id int
	err := r.db.QueryRow(ctx, sql, dto.Name, dto.Description, dto.PriorityID, dto.OwnerID, dto.IsAutomated).Scan(&id)
	return id, err
}

func (r *Repository) UpdateTestCase(ctx context.Context, id int, dto TestCaseDTO) error {
	sql := `UPDATE v_lead_all_cases SET name=$1, description=$2, priority_id=$3 WHERE test_case_id=$4`
	_, err := r.db.Exec(ctx, sql, dto.Name, dto.Description, dto.PriorityID, id)
	return err
}

func (r *Repository) DeleteTestCase(ctx context.Context, id int) error {
	sql := `DELETE FROM v_lead_all_cases WHERE test_case_id=$1`
	_, err := r.db.Exec(ctx, sql, id)
	return err
}

func (r *Repository) CreateTestCaseStep(ctx context.Context, tcID int, dto StepDTO) (int, error) {
	sql := `INSERT INTO v_test_case_steps (test_case_id, step_order, action_text, expected_result)
            VALUES ($1, $2, $3, $4) RETURNING step_id`
	var id int
	err := r.db.QueryRow(ctx, sql, tcID, dto.Order, dto.Action, dto.Expected).Scan(&id)
	return id, err
}

func (r *Repository) UpdateTestCaseStep(ctx context.Context, id int, dto StepDTO) error {
    sql := `
        UPDATE v_test_case_steps 
        SET step_order = $1, 
            action_text = $2, 
            expected_result = $3 
        WHERE step_id = $4
    `
    // Команда Exec возвращает tag, который можно проверить на RowsAffected, если нужно
    _, err := r.db.Exec(ctx, sql, dto.Order, dto.Action, dto.Expected, id)
    return err
}

func (r *Repository) DeleteTestCaseStep(ctx context.Context, id int) error {
    sql := `DELETE FROM v_test_case_steps WHERE step_id = $1`
    _, err := r.db.Exec(ctx, sql, id)
    return err
}

