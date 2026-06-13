-- Forum queries for sqlc code generation
-- All queries use parameterized placeholders (never string interpolation)
-- Security: T-05-06 mitigated

-- name: GetPoint :one
SELECT id, name, intent, curator, foundation_compat, commit_date,
       subscribers, avg_rating, rating_count, tags
FROM points
WHERE id = ?;

-- name: ListPoints :many
SELECT id, name, intent, curator, foundation_compat, commit_date,
       subscribers, avg_rating, rating_count, tags
FROM points
ORDER BY id
LIMIT ? OFFSET ?;

-- name: UpsertPoint :exec
INSERT INTO points (id, name, intent, curator, foundation_compat, commit_date, tags)
VALUES (?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
  name             = excluded.name,
  intent           = excluded.intent,
  curator          = excluded.curator,
  foundation_compat = excluded.foundation_compat,
  commit_date      = excluded.commit_date,
  tags             = excluded.tags;

-- name: RebuildFTS :exec
INSERT INTO points_fts(points_fts) VALUES('rebuild');

-- SearchPoints is implemented manually in sqlite.go (FTS5 virtual table not parsed by sqlc).
-- Placeholder: this comment is intentional.

-- name: AddSubscription :exec
INSERT INTO subscriptions (user_id, point_id) VALUES (?, ?)
ON CONFLICT(user_id, point_id) DO NOTHING;

-- name: RemoveSubscription :exec
DELETE FROM subscriptions WHERE user_id = ? AND point_id = ?;

-- name: GetSubscriptions :many
SELECT p.id, p.name, p.intent, p.curator, p.foundation_compat, p.commit_date,
       p.subscribers, p.avg_rating, p.rating_count, p.tags
FROM subscriptions s
JOIN points p ON p.id = s.point_id
WHERE s.user_id = ?
ORDER BY p.id;

-- name: SetRating :exec
INSERT INTO ratings (user_id, point_id, stars) VALUES (?, ?, ?)
ON CONFLICT(user_id, point_id) DO UPDATE SET stars = excluded.stars;

-- name: GetRatingSummary :one
SELECT COALESCE(AVG(CAST(stars AS REAL)), 0.0) AS avg_rating,
       CAST(COUNT(*) AS INTEGER) AS rating_count
FROM ratings
WHERE point_id = ?;

-- name: UpdatePointAggregates :exec
UPDATE points
SET avg_rating   = (SELECT COALESCE(AVG(CAST(stars AS REAL)), 0.0) FROM ratings WHERE point_id = points.id),
    rating_count = (SELECT COUNT(*) FROM ratings WHERE point_id = points.id),
    subscribers  = (SELECT COUNT(*) FROM subscriptions WHERE point_id = points.id)
WHERE id = ?;

-- name: GetConflicts :many
SELECT id, point_a, point_b, status, patch_pr_url, created_at
FROM conflict_threads
WHERE (point_a = ? AND point_b = ?) OR (point_a = ? AND point_b = ?)
ORDER BY created_at;

-- name: UpsertConflictThread :exec
INSERT INTO conflict_threads (id, point_a, point_b, status, patch_pr_url)
VALUES (?, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
  status       = excluded.status,
  patch_pr_url = excluded.patch_pr_url;

-- name: TruncateAll :exec
DELETE FROM conflict_threads;

-- name: TruncateRatings :exec
DELETE FROM ratings;

-- name: TruncateSubscriptions :exec
DELETE FROM subscriptions;

-- name: TruncatePoints :exec
DELETE FROM points;
