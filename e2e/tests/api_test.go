package tests

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"testing"
	"time"
)

var baseURL = envOr("BASE_URL", "http://api_e2e:8080")

func TestE2EHappyPathAndErrors(t *testing.T) {
	waitForReady(t)

	createTeam(t, "team-a", []user{
		{ID: "u1", Username: "Alice", IsActive: true},
		{ID: "u2", Username: "Bob", IsActive: true},
		{ID: "u3", Username: "Charlie", IsActive: true},
	})

	setActive(t, "u1", true)
	setActive(t, "u2", true)
	setActive(t, "u3", true)

	prID := "pr-1"
	pr := createPR(t, prID, "Initial", "u1")
	if len(pr.AssignedReviewers) == 0 {
		t.Fatalf("expected reviewers to be assigned")
	}
	for _, r := range pr.AssignedReviewers {
		if r == pr.AuthorID {
			t.Fatalf("author should not be reviewer")
		}
	}

	old := pr.AssignedReviewers[0]
	updated, replacedBy := reassign(t, pr.ID, old)
	if replacedBy == "" || replacedBy == old {
		t.Fatalf("expected new reviewer different from old")
	}
	if !contains(updated.AssignedReviewers, replacedBy) || contains(updated.AssignedReviewers, old) {
		t.Fatalf("expected reviewers updated")
	}

	reviewList := getReview(t, replacedBy)
	if !containsShort(reviewList.PullRequests, pr.ID) {
		t.Fatalf("expected pr in reviewer list")
	}

	merged := mergePR(t, pr.ID)
	if merged.Status != "MERGED" {
		t.Fatalf("expected merged status")
	}
	again := mergePR(t, pr.ID)
	if again.Status != "MERGED" {
		t.Fatalf("expected idempotent merge")
	}

	expectErrorCode(t, http.StatusNotFound, "NOT_FOUND", func() (*http.Response, error) {
		return doJSON("POST", "/pullRequest/reassign", map[string]any{"pull_request_id": "absent", "old_user_id": "u2"})
	})
	expectErrorCode(t, http.StatusConflict, "PR_MERGED", func() (*http.Response, error) {
		return doJSON("POST", "/pullRequest/reassign", map[string]any{"pull_request_id": pr.ID, "old_user_id": replacedBy})
	})
	expectErrorCode(t, http.StatusNotFound, "NOT_FOUND", func() (*http.Response, error) {
		return doJSON("POST", "/pullRequest/create", map[string]any{"pull_request_id": "pr-ghost", "pull_request_name": "Ghost", "author_id": "ghost"})
	})
}

type user struct {
	ID       string
	Username string
	IsActive bool
}

type pullRequest struct {
	ID                string   `json:"pull_request_id"`
	Name              string   `json:"pull_request_name"`
	AuthorID          string   `json:"author_id"`
	Status            string   `json:"status"`
	AssignedReviewers []string `json:"assigned_reviewers"`
}

type reviewList struct {
	UserID       string             `json:"user_id"`
	PullRequests []pullRequestShort `json:"pull_requests"`
}

type pullRequestShort struct {
	ID string `json:"pull_request_id"`
}

func createTeam(t *testing.T, name string, members []user) {
	body := map[string]any{
		"team_name": name,
		"members":   []map[string]any{},
	}
	for _, m := range members {
		body["members"] = append(body["members"].([]map[string]any), map[string]any{
			"user_id":   m.ID,
			"username":  m.Username,
			"is_active": m.IsActive,
		})
	}
	resp, err := doJSON("POST", "/team/add", body)
	if err != nil {
		t.Fatalf("createTeam request error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("createTeam expected 201, got %d", resp.StatusCode)
	}
}

func setActive(t *testing.T, userID string, isActive bool) {
	resp, err := doJSON("POST", "/users/setIsActive", map[string]any{
		"user_id":   userID,
		"is_active": isActive,
	})
	if err != nil {
		t.Fatalf("setActive request error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("setActive expected 200, got %d", resp.StatusCode)
	}
}

func createPR(t *testing.T, id, name, author string) pullRequest {
	resp, err := doJSON("POST", "/pullRequest/create", map[string]any{
		"pull_request_id":   id,
		"pull_request_name": name,
		"author_id":         author,
	})
	if err != nil {
		t.Fatalf("createPR request error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("createPR expected 201, got %d", resp.StatusCode)
	}
	var wrapper struct {
		PR pullRequest `json:"pr"`
	}
	decodeBody(t, resp.Body, &wrapper)
	return wrapper.PR
}

func mergePR(t *testing.T, id string) pullRequest {
	resp, err := doJSON("POST", "/pullRequest/merge", map[string]any{"pull_request_id": id})
	if err != nil {
		t.Fatalf("merge request error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("merge expected 200, got %d", resp.StatusCode)
	}
	var wrapper struct {
		PR pullRequest `json:"pr"`
	}
	decodeBody(t, resp.Body, &wrapper)
	return wrapper.PR
}

func reassign(t *testing.T, id, old string) (pullRequest, string) {
	resp, err := doJSON("POST", "/pullRequest/reassign", map[string]any{
		"pull_request_id": id,
		"old_user_id":     old,
	})
	if err != nil {
		t.Fatalf("reassign request error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("reassign expected 200, got %d", resp.StatusCode)
	}
	var wrapper struct {
		PR         pullRequest `json:"pr"`
		ReplacedBy string      `json:"replaced_by"`
	}
	decodeBody(t, resp.Body, &wrapper)
	return wrapper.PR, wrapper.ReplacedBy
}

func getReview(t *testing.T, userID string) reviewList {
	req, _ := http.NewRequest("GET", baseURL+"/users/getReview?user_id="+userID, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("getReview request error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("getReview expected 200, got %d", resp.StatusCode)
	}
	var rl reviewList
	decodeBody(t, resp.Body, &rl)
	return rl
}

func expectErrorCode(t *testing.T, status int, code string, do func() (*http.Response, error)) {
	resp, err := do()
	if err != nil {
		t.Fatalf("request error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != status {
		t.Fatalf("expected status %d, got %d", status, resp.StatusCode)
	}
	var body struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	decodeBody(t, resp.Body, &body)
	if body.Error.Code != code {
		t.Fatalf("expected code %s, got %s", code, body.Error.Code)
	}
}

func doJSON(method, path string, body any) (*http.Response, error) {
	data, _ := json.Marshal(body)
	req, _ := http.NewRequest(method, baseURL+path, bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	return http.DefaultClient.Do(req)
}

func decodeBody(t *testing.T, r io.Reader, v any) {
	t.Helper()
	if err := json.NewDecoder(r).Decode(v); err != nil {
		t.Fatalf("decode body: %v", err)
	}
}

func contains(list []string, value string) bool {
	for _, v := range list {
		if v == value {
			return true
		}
	}
	return false
}

func containsShort(list []pullRequestShort, id string) bool {
	for _, pr := range list {
		if pr.ID == id {
			return true
		}
	}
	return false
}

func waitForReady(t *testing.T) {
	t.Helper()
	timeout := time.After(30 * time.Second)
	for {
		select {
		case <-timeout:
			t.Fatalf("server did not become ready")
		default:
			resp, err := http.Get(baseURL + "/team/get?team_name=__ping__")
			if err == nil {
				resp.Body.Close()
				return
			}
			time.Sleep(500 * time.Millisecond)
		}
	}
}

func TestMain(m *testing.M) {
	if v := os.Getenv("BASE_URL"); v != "" {
		baseURL = v
	}
	os.Exit(m.Run())
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
