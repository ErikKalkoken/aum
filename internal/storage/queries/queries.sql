-- name: CreateReport :exec
INSERT INTO
  reports (app_id, arch, machine_id, os, version)
VALUES
  (?, ?, ?, ?, ?);

-- name: CountApplicationByID :one
SELECT
  count(*)
FROM
  applications
WHERE
  app_id = ?;

-- name: ListApplicationsWithMetric :many
SELECT
  sqlc.embed(ap),
  COUNT(DISTINCT(rp.machine_id)) as user_count,
  MAX(rp.timestamp) as latest
FROM
  applications ap
  JOIN reports rp ON rp.app_id = ap.app_id
GROUP BY
  ap.app_id
ORDER BY
  ap.name;

-- name: ListApplicationPlatforms :many
SELECT
  platform, COUNT(platform)
FROM
  reports_platforms
WHERE
  app_id = ?
GROUP BY
  platform;

-- name: UpdateOrCreateApplication :exec
INSERT INTO
  applications (app_id, name)
VALUES
  (?1, ?2) ON CONFLICT(app_id) DO
UPDATE
SET
  name = ?2
WHERE
  app_id = ?1;