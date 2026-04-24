//go:build integration

package integration_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRLS_TesterA_SeesOnlyOwnResults verifies that tester_a sees only their own
// results via v_tester_my_results (enforced by view filter + RLS on test_results).
func TestRLS_TesterA_SeesOnlyOwnResults(t *testing.T) {
	conn := connectAs(t, "tester_a", "testerapass")

	ids := collectIDs(t, conn, "SELECT test_result_id FROM v_tester_my_results")

	assert.Contains(t, ids, resultAID, "tester_a should see their own result")
	assert.Contains(t, ids, nonFinalResultID, "tester_a should see their IN_PROGRESS result")
	assert.NotContains(t, ids, resultBID, "tester_a must NOT see tester_b's result")
}

// TestRLS_TesterB_DoesNotSeeTesterAResults verifies that tester_b cannot see
// tester_a's results through v_tester_my_results.
func TestRLS_TesterB_DoesNotSeeTesterAResults(t *testing.T) {
	conn := connectAs(t, "tester_b", "testerbpass")

	ids := collectIDs(t, conn, "SELECT test_result_id FROM v_tester_my_results")

	assert.Contains(t, ids, resultBID, "tester_b should see their own result")
	assert.NotContains(t, ids, resultAID, "tester_b must NOT see tester_a's result")
}

// TestRLS_Guest_SeesOnlyFinalResults verifies that guest_user only sees final
// (is_final=true) results via v_public_results. IN_PROGRESS results are hidden.
func TestRLS_Guest_SeesOnlyFinalResults(t *testing.T) {
	conn := connectAs(t, "guest_user", "guestpass")

	ids := collectIDs(t, conn, "SELECT test_result_id FROM v_public_results")

	assert.Contains(t, ids, resultAID, "guest should see PASSED result")
	assert.Contains(t, ids, resultBID, "guest should see PASSED result")
	assert.NotContains(t, ids, nonFinalResultID, "guest must NOT see IN_PROGRESS result")
}

// TestRLS_TesterA_CannotInsertResultForAnotherUser verifies that the RLS INSERT
// policy on test_results (testresults_insert_own) blocks tester_a from inserting
// a result attributed to another user.
func TestRLS_TesterA_CannotInsertResultForAnotherUser(t *testing.T) {
	conn := connectAs(t, "tester_a", "testerapass")

	// tester_a has INSERT on v_results_manage but RLS WITH CHECK requires
	// executor_user_id to match CURRENT_USER.
	_, err := conn.Exec(context.Background(),
		`INSERT INTO v_results_manage (run_item_id, status_id, executor_user_id)
		 VALUES ($1, $2, $3)`,
		runItemID2, passedStatusID, testerBUserID,
	)
	require.Error(t, err, "RLS should block inserting result for another user")
}
