//go:build integration

package integration_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestViews_TestLead_CanInsertIntoTestPlansManage verifies that the flat
// auto-updatable view v_test_plans_manage accepts INSERT from db_test_lead.
func TestViews_TestLead_CanInsertIntoTestPlansManage(t *testing.T) {
	conn := connectAs(t, "test_lead_user", "leadpass")

	var newID int
	err := conn.QueryRow(context.Background(),
		`INSERT INTO v_test_plans_manage (name, priority_id, start_date, end_date, owner_user_id, status)
		 VALUES ($1, $2, CURRENT_DATE, CURRENT_DATE+1, $3, 'DRAFT') RETURNING test_plan_id`,
		fmt.Sprintf("views-test-plan-%d", testLeadUserID),
		priorityID,
		testLeadUserID,
	).Scan(&newID)
	require.NoError(t, err)
	assert.Greater(t, newID, 0)

	t.Cleanup(func() {
		superConn.Exec(context.Background(), //nolint:errcheck
			"DELETE FROM test_plans WHERE test_plan_id = $1", newID)
	})
}

// TestViews_TestLead_CanInsertIntoTestCasesManage verifies that the flat
// auto-updatable view v_test_cases_manage accepts INSERT from db_test_lead.
func TestViews_TestLead_CanInsertIntoTestCasesManage(t *testing.T) {
	conn := connectAs(t, "test_lead_user", "leadpass")

	var newID int
	err := conn.QueryRow(context.Background(),
		`INSERT INTO v_test_cases_manage (name, priority_id, owner_user_id)
		 VALUES ($1, $2, $3) RETURNING test_case_id`,
		fmt.Sprintf("views-test-case-%d", testLeadUserID),
		priorityID,
		testLeadUserID,
	).Scan(&newID)
	require.NoError(t, err)
	assert.Greater(t, newID, 0)

	t.Cleanup(func() {
		superConn.Exec(context.Background(), //nolint:errcheck
			"DELETE FROM test_cases WHERE test_case_id = $1", newID)
	})
}

// TestViews_Guest_CannotInsertIntoPublicTestPlans verifies that db_guest has no
// INSERT privilege on v_public_test_plans (SELECT-only view).
func TestViews_Guest_CannotInsertIntoPublicTestPlans(t *testing.T) {
	conn := connectAs(t, "guest_user", "guestpass")

	mustFail(t, conn,
		`INSERT INTO v_public_test_plans (name, priority_id, start_date, end_date, owner_user_id, status)
		 VALUES ('should-fail', $1, CURRENT_DATE, CURRENT_DATE+1, $2, 'DRAFT')`,
		priorityID, guestUserID,
	)
}

// TestViews_TestLead_CannotInsertIntoLeadAllCases verifies that v_lead_all_cases
// (a JOIN view) rejects DML, confirming it is not auto-updatable.
func TestViews_TestLead_CannotInsertIntoLeadAllCases(t *testing.T) {
	conn := connectAs(t, "test_lead_user", "leadpass")

	// db_test_lead has only SELECT on v_lead_all_cases (no INSERT grant),
	// so this fails with 42501 before even reaching the non-updatable-view check.
	mustFail(t, conn,
		`INSERT INTO v_lead_all_cases (test_case_name, priority, owner)
		 VALUES ('x', 'CRITICAL', 'someone')`,
	)
}

// TestViews_UsersManage_NoPasswordHash confirms that v_users_manage excludes the
// password_hash column, preventing accidental exposure of credentials.
func TestViews_UsersManage_NoPasswordHash(t *testing.T) {
	conn := connectAs(t, "admin_user", "adminpass")

	rows, err := conn.Query(context.Background(), "SELECT * FROM v_users_manage LIMIT 0")
	require.NoError(t, err)
	defer rows.Close()

	var cols []string
	for _, fd := range rows.FieldDescriptions() {
		cols = append(cols, string(fd.Name))
	}

	assert.NotContains(t, cols, "password_hash",
		"v_users_manage must not expose password_hash")
	// Sanity check: the view should have at least the user_id column.
	assert.Contains(t, cols, "user_id")
}

// TestViews_ActiveTestRuns_SafeColumns confirms that v_active_test_runs exposes
// the expected aggregated columns and does not leak raw FK or credential fields.
func TestViews_ActiveTestRuns_SafeColumns(t *testing.T) {
	conn := connectAs(t, "test_lead_user", "leadpass")

	rows, err := conn.Query(context.Background(), "SELECT * FROM v_active_test_runs LIMIT 0")
	require.NoError(t, err)
	defer rows.Close()

	var cols []string
	for _, fd := range rows.FieldDescriptions() {
		cols = append(cols, string(fd.Name))
	}

	expected := []string{
		"test_run_id", "run_name", "plan_name", "version_string",
		"environment", "tool_config", "status",
		"created_by_name", "total_tests", "passed_tests", "failed_tests", "blocked_tests",
	}
	for _, col := range expected {
		assert.Contains(t, cols, col, "v_active_test_runs must have column %q", col)
	}
	assert.NotContains(t, cols, "password_hash")
	// Raw FK replaced by created_by_name.
	assert.NotContains(t, cols, "created_by",
		"raw FK 'created_by' should be replaced by 'created_by_name'")
}
