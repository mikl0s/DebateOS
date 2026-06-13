-- Forum database schema
-- FTS5 is confirmed functional in modernc.org/sqlite v1.46.1

CREATE TABLE IF NOT EXISTS points (
  id               TEXT PRIMARY KEY,
  name             TEXT NOT NULL,
  intent           TEXT NOT NULL DEFAULT '',
  curator          TEXT NOT NULL DEFAULT '',
  foundation_compat TEXT NOT NULL DEFAULT '[]',  -- JSON array of foundation IDs
  commit_date      TEXT NOT NULL DEFAULT '',       -- ISO 8601
  subscribers      INTEGER NOT NULL DEFAULT 0,
  avg_rating       REAL NOT NULL DEFAULT 0.0,
  rating_count     INTEGER NOT NULL DEFAULT 0,
  tags             TEXT NOT NULL DEFAULT '[]'      -- JSON array of tags
);

-- FTS5 virtual table for full-text search over points
-- VERIFIED: modernc.org/sqlite v1.46.1 has FTS5 compiled in (tested empirically)
-- content='points' makes this an external-content table backed by the points table.
-- After UpsertPoint, run: INSERT INTO points_fts(points_fts) VALUES('rebuild') to sync.
CREATE VIRTUAL TABLE IF NOT EXISTS points_fts USING fts5(
  name, intent, curator, id UNINDEXED,
  content='points', content_rowid='rowid'
);

CREATE TABLE IF NOT EXISTS subscriptions (
  user_id   TEXT NOT NULL,
  point_id  TEXT NOT NULL REFERENCES points(id),
  PRIMARY KEY (user_id, point_id)
);

CREATE TABLE IF NOT EXISTS ratings (
  user_id   TEXT NOT NULL,
  point_id  TEXT NOT NULL REFERENCES points(id),
  stars     INTEGER NOT NULL CHECK(stars BETWEEN 1 AND 5),
  PRIMARY KEY (user_id, point_id)
);

CREATE TABLE IF NOT EXISTS conflict_threads (
  id           TEXT PRIMARY KEY,
  point_a      TEXT NOT NULL,
  point_b      TEXT NOT NULL,
  status       TEXT NOT NULL DEFAULT 'open',
  patch_pr_url TEXT NOT NULL DEFAULT '',
  created_at   TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);
