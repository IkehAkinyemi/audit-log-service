package mongodb

import (
	"context"
	"testing"
	"time"

	"github.com/IkehAkinyemi/logaudit/internal/repository/model"
	"github.com/IkehAkinyemi/logaudit/internal/utils"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
)

func createLog(t *testing.T) *model.Log {
	log := &model.Log{
		Timestamp: time.Now(),
		Action:    "updated",
		Actor: model.Actor{
			Type:      "user",
			ID:        "12300",
			Extension: map[string]any{"userAgent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/83.0.4103.116 Safari/537.36"},
		},
		Entity: model.Entity{
			Type:      "inventory",
			Extension: map[string]any{"item_id": "f66020564728"},
		},
		Context: model.Context{
			IPAddr:    "192.168.1.1",
			Location:  "New York, NY",
			Extension: map[string]any{"inventory_section": "electronics"},
		},
		Extension: map[string]any{"notes": "Inventory successfully updated"},
	}

	db = "test"

	res, err := logRes.AddLog(log)
	require.NoError(t, err)

	// check the returned document ID
	result := res.(*mongo.InsertOneResult)
	require.Equal(t, log.ID, result.InsertedID)

	return log
}

func TestAddLog(t *testing.T) {
	createLog(t)
}

func TestGetAllLogs(t *testing.T) {
	filters := utils.Filters{
		Action:         "updated",
		ActorType:      "user",
		ActorID:        "12300",
		EntityType:     "inventory",
		StartTimestamp: time.Now(),
		SortField:      "timestamp",
		SortDescending: false,
		PageSize:       5,
		Page:           1,
	}

	n := 10
	for i := 0; i < n; i++ {
		createLog(t)
	}

	logs, metadata, err := logRes.GetAllLogs(context.Background(), filters)
	require.NoError(t, err)
	require.Len(t, logs, 5)
	require.NotEmpty(t, metadata)

	for _, log := range logs {
		require.NotEmpty(t, log)

		require.Equal(t, filters.Action, log.Action)

		require.Equal(t, filters.ActorType, log.Actor.Type)
		require.Equal(t, filters.ActorID, log.Actor.ID)
		require.NotEmpty(t, log.Actor.Extension)

		require.Equal(t, filters.EntityType, log.Entity.Type)

		require.WithinDuration(t, filters.StartTimestamp, log.Timestamp, 5*time.Second)
	}
}
