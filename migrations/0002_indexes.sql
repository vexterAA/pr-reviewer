CREATE INDEX IF NOT EXISTS idx_users_team_active ON users (team_name, is_active, id);
CREATE INDEX IF NOT EXISTS idx_pr_status ON pull_requests (status);
CREATE INDEX IF NOT EXISTS idx_pr_author ON pull_requests (author_id);
CREATE INDEX IF NOT EXISTS idx_pr_reviewers_reviewer ON pull_request_reviewers (reviewer_id, pull_request_id);
