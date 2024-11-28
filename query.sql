-- name: CreateRecord :exec
INSERT INTO records (
  uid, data
) VALUES (
  ?, ?
);
