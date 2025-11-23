CREATE VIEW IF NOT EXISTS pr_status_counts AS
SELECT status, COUNT(*) AS total
FROM pull_requests
GROUP BY status;
