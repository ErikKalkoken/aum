-- name: CreateReport :exec
INSERT INTO
  reports (app_id, arch, machine_id, os, version)
VALUES
  (?, ?, ?, ?, ?);