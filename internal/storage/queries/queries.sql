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

-- name: ListApplications :many
SELECT
  *
FROM
  applications
ORDER BY app_id;

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