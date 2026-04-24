//go:build integration

package integration_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAudit_InsertCreatesAuditRecord verifies that the audit trigger on test_plans
// fires on INSERT and writes a record to audit_log.
func TestAudit_InsertCreatesAuditRecord(t *testing.T) {
	ctx := context.Background()

	var before int
	require.NoError(t,
		superConn.QueryRow(ctx,
			`SELECT COUNT(*) FROM audit_log WHERE table_name = 'test_plans' AND operation = 'INSERT'`,
		).Scan(&before),
	)

	var planID int
	require.NoError(t,
		superConn.QueryRow(ctx,
			`INSERT INTO test_plans (name, priority_id, start_date, end_date, owner_user_id, status)
			 VALUES ($1, $2, CURRENT_DATE, CURRENT_DATE+1, $3, 'DRAFT') RETURNING test_plan_id`,
			fmt.Sprintf("audit-test-plan-%d", time.Now().UnixNano()),
			priorityID, testLeadUserID,
		).Scan(&planID),
	)
	t.Cleanup(func() {
		superConn.Exec(ctx, "DELETE FROM test_plans WHERE test_plan_id = $1", planID) //nolint:errcheck
	})

	var after int
	require.NoError(t,
		superConn.QueryRow(ctx,
			`SELECT COUNT(*) FROM audit_log WHERE table_name = 'test_plans' AND operation = 'INSERT'`,
		).Scan(&after),
	)

	assert.Equal(t, before+1, after,
		"audit_log should have exactly one new INSERT record for test_plans")
}

// TestAudit_TesterCannotSelectAuditLog verifies that db_tester has no direct
// SELECT privilege on audit_log (no GRANT exists; RLS adds another layer).
func TestAudit_TesterCannotSelectAuditLog(t *testing.T) {
	conn := connectAs(t, "tester_a", "testerapass")
	mustFail(t, conn, "SELECT audit_id FROM audit_log LIMIT 1")
}

// TestAudit_AdminCanSelectAuditLogView verifies that db_admin can read
// v_admin_audit_log. The view is owned by postgres (superuser), so the
// underlying audit_log access bypasses RLS when evaluated via the view owner.
func TestAudit_AdminCanSelectAuditLogView(t *testing.T) {
	conn := connectAs(t, "admin_user", "adminpass")

	rows, err := conn.Query(context.Background(), "SELECT audit_id FROM v_admin_audit_log LIMIT 1")
	require.NoError(t, err, "admin_user must be able to query v_admin_audit_log")
	rows.Close()
}

// TestAudit_TestLeadCannotSelectAdminAuditLog verifies that db_test_lead has no
// SELECT privilege on v_admin_audit_log (granted only to db_admin).
func TestAudit_TestLeadCannotSelectAdminAuditLog(t *testing.T) {
	conn := connectAs(t, "test_lead_user", "leadpass")
	mustFail(t, conn, "SELECT audit_id FROM v_admin_audit_log LIMIT 1")
}

// TestAudit_UpdateSetsUpdatedAt verifies the trg_testplans_updated_at trigger:
// after UPDATE, the updated_at column is set to a timestamp >= the pre-update time.
func TestAudit_UpdateSetsUpdatedAt(t *testing.T) {
	ctx := context.Background()

	var planID int
	require.NoError(t,
		superConn.QueryRow(ctx,
			`INSERT INTO test_plans (name, priority_id, start_date, end_date, owner_user_id, status)
			 VALUES ($1, $2, CURRENT_DATE, CURRENT_DATE+1, $3, 'DRAFT') RETURNING test_plan_id`,
			fmt.Sprintf("updated-at-plan-%d", time.Now().UnixNano()),
			priorityID, testLeadUserID,
		).Scan(&planID),
	)
	t.Cleanup(func() {
		superConn.Exec(ctx, "DELETE FROM test_plans WHERE test_plan_id = $1", planID) //nolint:errcheck
	})

	before := time.Now().UTC().Add(-time.Second) // small buffer for clock skew

	_, err := superConn.Exec(ctx,
		"UPDATE test_plans SET description = 'trigger test' WHERE test_plan_id = $1", planID,
	)
	require.NoError(t, err)

	var updatedAt *time.Time
	require.NoError(t,
		superConn.QueryRow(ctx,
			"SELECT updated_at FROM test_plans WHERE test_plan_id = $1", planID,
		).Scan(&updatedAt),
	)

	require.NotNil(t, updatedAt, "updated_at must be set after UPDATE")
	assert.True(t, updatedAt.After(before),
		"updated_at (%v) should be after pre-update time (%v)", updatedAt, before)
}
