// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package queries

import (
	"database/sql"
)

type Report struct {
	ID        int64
	AppID     string
	Arch      string
	MachineID string
	Os        string
	Timestamp sql.NullTime
	Version   string
}