package storage_test

import (
	"context"
	"example/telemetry/internal/storage"
	"example/telemetry/internal/storage/queries"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStorage(t *testing.T) {
	ctx := context.Background()
	db, err := storage.InitDB(ctx, ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	q := queries.New(db)
	t.Run("should return false when app does not exist", func(t *testing.T) {
		storage.TruncateTables(db)
		got, err := storage.ApplicationExists(ctx, q, "dummy")
		if assert.NoError(t, err) {
			assert.False(t, got)
		}
	})
	t.Run("should return true when app exists", func(t *testing.T) {
		storage.TruncateTables(db)
		err := q.UpdateOrCreateApplication(ctx, queries.UpdateOrCreateApplicationParams{AppID: "dummy", Name: "dummy"})
		if err != nil {
			t.Fatal(err)
		}
		got, err := storage.ApplicationExists(ctx, q, "dummy")
		if assert.NoError(t, err) {
			assert.True(t, got)
		}
	})

}
