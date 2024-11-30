CREATE TABLE IF NOT EXISTS applications (
  app_id TEXT NOT NULL PRIMARY KEY,
  name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS reports (
  id INTEGER PRIMARY KEY,
  app_id TEXT NOT NULL,
  arch TEXT NOT NULL,
  machine_id TEXT NOT NULL,
  os TEXT NOT NULL,
  timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
  version TEXT NOT NULL,
  FOREIGN KEY (app_id) REFERENCES applications(app_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS reports_idx_1 ON reports (app_id);

CREATE INDEX IF NOT EXISTS reports_idx_2 ON reports (arch);

CREATE INDEX IF NOT EXISTS reports_idx_3 ON reports (os);

CREATE INDEX IF NOT EXISTS reports_idx_4 ON reports (timestamp);

CREATE VIEW IF NOT EXISTS reports_platforms AS
SELECT
  ap.app_id,
  ap.name,
  concat(os, "/", arch) platform
FROM
  reports rp
  JOIN applications ap ON ap.app_id = rp.app_id
WHERE
  os <> ""
  AND arch <> "";