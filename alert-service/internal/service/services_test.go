package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"censys_alert_system/external"
	"censys_alert_system/internal/models"
	"censys_alert_system/internal/service/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAlertService_GetAlerts(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mockStorage := mocks.NewAlertStorageInterface(t)
		service := NewAlertService(mockStorage, nil)

		expectedAlerts := []models.Alert{
			{ID: "some-uuid-1", Source: "source1", Severity: "high"},
			{ID: "some-uuid-2", Source: "source2", Severity: "low"},
		}

		// Use On().Return() instead of EXPECT()
		mockStorage.On("GetAlerts", ctx).Return(expectedAlerts, nil)

		alerts, err := service.GetAlerts(ctx)

		assert.NoError(t, err)
		assert.Len(t, alerts, 2)
		mockStorage.AssertExpectations(t)
	})

	t.Run("storage error", func(t *testing.T) {
		mockStorage := mocks.NewAlertStorageInterface(t)
		service := NewAlertService(mockStorage, nil)

		mockStorage.On("GetAlerts", ctx).Return(nil, errors.New("database error"))

		alerts, err := service.GetAlerts(ctx)

		assert.Error(t, err)
		assert.Nil(t, alerts)
		assert.Contains(t, err.Error(), "service: error getting alerts")
		mockStorage.AssertExpectations(t)
	})
}

func TestAlertService_GetAlertByID(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mockStorage := mocks.NewAlertStorageInterface(t)
		service := NewAlertService(mockStorage, nil)

		expectedAlert := &models.Alert{ID: "some-uuid", Source: "test", Severity: "critical"}
		mockStorage.On("GetAlertByID", ctx, "1").Return(expectedAlert, nil)

		alert, err := service.GetAlertByID(ctx, "1")

		assert.NoError(t, err)
		assert.Equal(t, "test", alert.Source)
		mockStorage.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		mockStorage := mocks.NewAlertStorageInterface(t)
		service := NewAlertService(mockStorage, nil)

		mockStorage.On("GetAlertByID", ctx, "999").Return(nil, errors.New("alert not found"))

		alert, err := service.GetAlertByID(ctx, "999")

		assert.Error(t, err)
		assert.Nil(t, alert)
		mockStorage.AssertExpectations(t)
	})
}

func TestAlertService_GetAlertsByDays(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mockStorage := mocks.NewAlertStorageInterface(t)
		service := NewAlertService(mockStorage, nil)

		expectedAlerts := []models.Alert{
			{ID: "some-uuid-1", Source: "recent", Severity: "medium"},
		}
		mockStorage.On("GetAlertsByDays", ctx, 7).Return(expectedAlerts, nil)

		alerts, err := service.GetAlertsByDays(ctx, 7)

		assert.NoError(t, err)
		assert.Len(t, alerts, 1)
		mockStorage.AssertExpectations(t)
	})

	t.Run("invalid days - zero", func(t *testing.T) {
		mockStorage := mocks.NewAlertStorageInterface(t)
		service := NewAlertService(mockStorage, nil)

		alerts, err := service.GetAlertsByDays(ctx, 0)

		assert.Error(t, err)
		assert.Nil(t, alerts)
		assert.Contains(t, err.Error(), "days must be greater than 0")
	})

	t.Run("invalid days - negative", func(t *testing.T) {
		mockStorage := mocks.NewAlertStorageInterface(t)
		service := NewAlertService(mockStorage, nil)

		alerts, err := service.GetAlertsByDays(ctx, -5)

		assert.Error(t, err)
		assert.Nil(t, alerts)
	})
}

func TestAlertService_PerformSync(t *testing.T) {
	ctx := context.Background()

	t.Run("first sync success", func(t *testing.T) {
		mockStorage := mocks.NewAlertStorageInterface(t)
		mockClient := mocks.NewAPIClientInterface(t)
		service := NewAlertService(mockStorage, mockClient)

		externalAlerts := []external.ExternalAlert{
			{Source: "ext1", Severity: "high", Description: "desc1", CreatedAt: time.Now()},
		}

		mockClient.On("CheckHealth", ctx).Return(nil)
		mockStorage.On("GetLastSyncTime", ctx).Return(time.Time{}, nil)
		mockClient.On("FetchAllAlerts", ctx).Return(externalAlerts, nil)
		mockStorage.On("CreateAlert",
			ctx,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		).Return(nil)
		mockStorage.On("UpdateLastSyncTime", ctx, mock.Anything).Return(nil)

		err := service.PerformSync(ctx)

		assert.NoError(t, err)
		mockStorage.AssertExpectations(t)
		mockClient.AssertExpectations(t)
	})

	t.Run("incremental sync success", func(t *testing.T) {
		mockStorage := mocks.NewAlertStorageInterface(t)
		mockClient := mocks.NewAPIClientInterface(t)
		service := NewAlertService(mockStorage, mockClient)

		lastSync := time.Now().Add(-24 * time.Hour)
		externalAlerts := []external.ExternalAlert{
			{Source: "ext1", Severity: "low", Description: "new alert", CreatedAt: time.Now()},
		}

		mockClient.On("CheckHealth", ctx).Return(nil)
		mockStorage.On("GetLastSyncTime", ctx).Return(lastSync, nil)
		mockClient.On("FetchAlertsSince", ctx, lastSync).Return(externalAlerts, nil)
		mockStorage.On("CreateAlert",
			ctx,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		).Return(nil)
		mockStorage.On("UpdateLastSyncTime", ctx, mock.Anything).Return(nil)

		err := service.PerformSync(ctx)

		assert.NoError(t, err)
		mockStorage.AssertExpectations(t)
		mockClient.AssertExpectations(t)
	})

	t.Run("no new alerts", func(t *testing.T) {
		mockStorage := mocks.NewAlertStorageInterface(t)
		mockClient := mocks.NewAPIClientInterface(t)
		service := NewAlertService(mockStorage, mockClient)

		mockClient.On("CheckHealth", ctx).Return(nil)
		mockStorage.On("GetLastSyncTime", ctx).Return(time.Time{}, nil)
		mockClient.On("FetchAllAlerts", ctx).Return([]external.ExternalAlert{}, nil)

		err := service.PerformSync(ctx)

		assert.NoError(t, err)
		mockStorage.AssertNotCalled(t, "CreateAlert")
	})

	t.Run("fetch error", func(t *testing.T) {
		mockStorage := mocks.NewAlertStorageInterface(t)
		mockClient := mocks.NewAPIClientInterface(t)
		service := NewAlertService(mockStorage, mockClient)

		mockClient.On("CheckHealth", ctx).Return(nil)
		mockStorage.On("GetLastSyncTime", ctx).Return(time.Time{}, nil)
		mockClient.On("FetchAllAlerts", ctx).Return(nil, errors.New("API error"))

		err := service.PerformSync(ctx)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to fetch alerts")
	})

	t.Run("context cancellation", func(t *testing.T) {
		mockStorage := mocks.NewAlertStorageInterface(t)
		mockClient := mocks.NewAPIClientInterface(t)
		service := NewAlertService(mockStorage, mockClient)

		cancelCtx, cancel := context.WithCancel(ctx)
		cancel()

		externalAlerts := []external.ExternalAlert{
			{Source: "ext1", Severity: "high", Description: "desc1", CreatedAt: time.Now()},
		}

		mockClient.On("CheckHealth", cancelCtx).Return(nil)
		mockStorage.On("GetLastSyncTime", cancelCtx).Return(time.Time{}, nil)
		mockClient.On("FetchAllAlerts", cancelCtx).Return(externalAlerts, nil)

		err := service.PerformSync(cancelCtx)

		assert.Error(t, err)
		assert.Equal(t, context.Canceled, err)
	})

	t.Run("health check failure continues", func(t *testing.T) {
		mockStorage := mocks.NewAlertStorageInterface(t)
		mockClient := mocks.NewAPIClientInterface(t)
		service := NewAlertService(mockStorage, mockClient)

		mockClient.On("CheckHealth", ctx).Return(errors.New("health check failed"))
		mockStorage.On("GetLastSyncTime", ctx).Return(time.Time{}, nil)
		mockClient.On("FetchAllAlerts", ctx).Return([]external.ExternalAlert{}, nil)

		err := service.PerformSync(ctx)

		assert.NoError(t, err)
	})
}
