//go:build integration

package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// loginAs performs POST /auth/login and returns the JWT token.
func loginAs(t *testing.T, username, password string) string {
	t.Helper()
	body := fmt.Sprintf(`{"username":%q,"password":%q}`, username, password)
	resp, err := http.Post(testServerURL+"/api/auth/login", "application/json", bytes.NewBufferString(body))
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode, "loginAs: unexpected status for %s", username)

	var result struct {
		Token string `json:"token"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	require.NotEmpty(t, result.Token, "loginAs: empty token for %s", username)
	return result.Token
}

// TestAPI_Login_Returns200AndJWT verifies the happy path of POST /auth/login (E2E).
func TestAPI_Login_Returns200AndJWT(t *testing.T) {
	body := `{"username":"tester_a","password":"testerapass"}`
	resp, err := http.Post(testServerURL+"/api/auth/login", "application/json", bytes.NewBufferString(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.NotEmpty(t, result["token"], "response must contain token")
	assert.NotEmpty(t, result["expires_at"], "response must contain expires_at")
}

// TestAPI_TestPlans_NoAuth_Returns401 verifies that GET /test-plans without a JWT returns 401.
func TestAPI_TestPlans_NoAuth_Returns401(t *testing.T) {
	resp, err := http.Get(testServerURL + "/api/test-plans")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// TestAPI_Results_TesterSeesOwnOnly verifies that a tester only sees their own results
// via GET /results (RLS on v_tester_my_results enforced through tester's PG connection).
func TestAPI_Results_TesterSeesOwnOnly(t *testing.T) {
	token := loginAs(t, "tester_a", "testerapass")

	req, err := http.NewRequest(http.MethodGet, testServerURL+"/api/results", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result struct {
		Data []struct {
			ResultID int `json:"test_result_id"`
		} `json:"data"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))

	ids := make([]int, len(result.Data))
	for i, r := range result.Data {
		ids[i] = r.ResultID
	}

	assert.Contains(t, ids, resultAID, "tester_a must see their own PASSED result")
	assert.NotContains(t, ids, resultBID, "tester_a must not see tester_b's result")
}

// TestAPI_TestPlans_TesterCreate_Returns403 verifies that a tester cannot create a test
// plan (privilege denied on v_test_plans_manage → mapped to 403).
func TestAPI_TestPlans_TesterCreate_Returns403(t *testing.T) {
	token := loginAs(t, "tester_a", "testerapass")

	body := `{"name":"Forbidden Plan"}`
	req, err := http.NewRequest(http.MethodPost, testServerURL+"/api/test-plans", bytes.NewBufferString(body))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}
