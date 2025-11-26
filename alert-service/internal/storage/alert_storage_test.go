package storage

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	cleanup := func() {
		db.Close()
	}

	return db, mock, cleanup
}

func TestAlertStorage_CreateAlert(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	storage := NewAlertStorage(db)
	ctx := context.Background()
	createdAt := time.Now()

	mock.ExpectExec("INSERT INTO alerts").
		WithArgs("test-source", "high", "test description", []byte(`{"key": "value"}`), "geo_location", "192.168.1.1", createdAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := storage.CreateAlert(
		ctx,
		"test-source",
		"high",
		"test description",
		[]byte(`{"key": "value"}`),
		"geo_location",
		"192.168.1.1",
		createdAt,
	)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAlertStorage_CreateAlert_Error(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	storage := NewAlertStorage(db)
	ctx := context.Background()

	mock.ExpectExec("INSERT INTO alerts").
		WillReturnError(sql.ErrConnDone)

	err := storage.CreateAlert(
		ctx,
		"test-source",
		"high",
		"test description",
		[]byte(`{}`),
		"geo_location",
		"192.168.1.1",
		time.Now(),
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error creating alert")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAlertStorage_GetAlerts(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	storage := NewAlertStorage(db)
	ctx := context.Background()
	createdAt := time.Now()

	rows := sqlmock.NewRows([]string{"id", "source", "severity", "description", "whole_event", "enrichment_type", "ip_address", "created_at"}).
		AddRow(1, "source1", "high", "desc1", []byte(`{}`), "geo_location", "10.0.0.1", createdAt).
		AddRow(2, "source2", "low", "desc2", []byte(`{}`), "threat_intel", "10.0.0.2", createdAt)

	mock.ExpectQuery("SELECT (.+) FROM alerts ORDER BY created_at DESC").
		WillReturnRows(rows)

	alerts, err := storage.GetAlerts(ctx)

	assert.NoError(t, err)
	assert.Len(t, alerts, 2)
	assert.Equal(t, "source1", alerts[0].Source)
	assert.Equal(t, "source2", alerts[1].Source)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAlertStorage_GetAlerts_Empty(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	storage := NewAlertStorage(db)
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"id", "source", "severity", "description", "whole_event", "enrichment_type", "ip_address", "created_at"})

	mock.ExpectQuery("SELECT (.+) FROM alerts ORDER BY created_at DESC").
		WillReturnRows(rows)

	alerts, err := storage.GetAlerts(ctx)

	assert.NoError(t, err)
	assert.Empty(t, alerts)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAlertStorage_GetAlertByID(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	storage := NewAlertStorage(db)
	ctx := context.Background()
	createdAt := time.Now()

	t.Run("existing alert", func(t *testing.T) {
		row := sqlmock.NewRows([]string{"id", "source", "severity", "description", "whole_event", "enrichment_type", "ip_address", "created_at"}).
			AddRow(1, "test-source", "critical", "critical alert", []byte(`{}`), "network_analysis", "172.16.0.1", createdAt)

		mock.ExpectQuery("SELECT (.+) FROM alerts WHERE id = \\$1").
			WithArgs("1").
			WillReturnRows(row)

		alert, err := storage.GetAlertByID(ctx, "1")

		assert.NoError(t, err)
		assert.NotNil(t, alert)
		assert.Equal(t, "test-source", alert.Source)
		assert.Equal(t, "critical", alert.Severity)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("non-existing alert", func(t *testing.T) {
		mock.ExpectQuery("SELECT (.+) FROM alerts WHERE id = \\$1").
			WithArgs("999").
			WillReturnError(sql.ErrNoRows)

		alert, err := storage.GetAlertByID(ctx, "999")

		assert.Error(t, err)
		assert.Nil(t, alert)
		assert.Contains(t, err.Error(), "alert not found")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestAlertStorage_GetAlertsByDays(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	storage := NewAlertStorage(db)
	ctx := context.Background()
	createdAt := time.Now()

	rows := sqlmock.NewRows([]string{"id", "source", "severity", "description", "whole_event", "enrichment_type", "ip_address", "created_at"}).
		AddRow(1, "recent-source", "low", "recent alert", []byte(`{}`), "user_context", "8.8.8.8", createdAt)

	mock.ExpectQuery("SELECT (.+) FROM alerts WHERE created_at >= NOW\\(\\) - INTERVAL").
		WithArgs(3).
		WillReturnRows(rows)

	alerts, err := storage.GetAlertsByDays(ctx, 3)

	assert.NoError(t, err)
	assert.Len(t, alerts, 1)
	assert.Equal(t, "recent-source", alerts[0].Source)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAlertStorage_GetLastSyncTime(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	storage := NewAlertStorage(db)
	ctx := context.Background()

	t.Run("no alerts returns zero time", func(t *testing.T) {
		row := sqlmock.NewRows([]string{"max"}).AddRow(nil)

		mock.ExpectQuery("SELECT MAX\\(created_at\\) FROM alerts").
			WillReturnRows(row)

		lastSync, err := storage.GetLastSyncTime(ctx)

		assert.NoError(t, err)
		assert.True(t, lastSync.IsZero())
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns most recent alert time", func(t *testing.T) {
		expectedTime := time.Now().Truncate(time.Second)
		row := sqlmock.NewRows([]string{"max"}).AddRow(expectedTime)

		mock.ExpectQuery("SELECT MAX\\(created_at\\) FROM alerts").
			WillReturnRows(row)

		lastSync, err := storage.GetLastSyncTime(ctx)

		assert.NoError(t, err)
		assert.Equal(t, expectedTime, lastSync)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
