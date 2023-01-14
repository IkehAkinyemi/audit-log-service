package mongodb

import (
	"context"

	"github.com/IkehAkinyemi/logaudit/internal/repository/model"
	"github.com/IkehAkinyemi/logaudit/internal/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	db                 = "audit-log"
	eventLogCollection = "logs"
)

// Repository defines a Mongodb-based log repository.
type Repository struct {
	client *mongo.Client
}

// New instantiates a new Mongodb-based log repository.
func New(client *mongo.Client) *Repository {
	return &Repository{client}
}

// AddLog adds a log record to the event_log collection.
func (r *Repository) AddLog(ctx context.Context, eventLog *model.AuditEvent) (any, error) {
	collection := r.client.Database(db).Collection(eventLogCollection)

	if eventLog.ID == primitive.NilObjectID {
		eventLog.ID = primitive.NewObjectID()
	}

	return collection.InsertOne(ctx, eventLog)
}

// GetAggregatedLogs returns all the log records that's matched
// by the query_string.
func (r *Repository) GetAllLogs(ctx context.Context, filter utils.Filters) ([]*model.AuditEvent, utils.Metadata, error) {
	collection := r.client.Database(db).Collection(eventLogCollection)

	// Set up the pipeline to perform the filtering and pagination.
	pipeline := []bson.M{}

	// Filter the results by the specified criteria.
	if filter.Action != "" {
		pipeline = append(pipeline, bson.M{"$match": bson.M{"action": filter.Action}})
	}
	if filter.ActorType != "" {
		pipeline = append(pipeline, bson.M{"$match": bson.M{"actor.type": filter.ActorType}})
	}
	if filter.ActorID != "" {
		pipeline = append(pipeline, bson.M{"$match": bson.M{"actor.id": filter.ActorID}})
	}
	if filter.EntityType != "" {
		pipeline = append(pipeline, bson.M{"$match": bson.M{"entity.type": filter.EntityType}})
	}
	if !filter.StartTimestamp.IsZero() {
		pipeline = append(pipeline, bson.M{"$match": bson.M{"timestamp": bson.M{"$gte": filter.StartTimestamp}}})
	}
	if !filter.EndTimestamp.IsZero() {
		pipeline = append(pipeline, bson.M{"$match": bson.M{"timestamp": bson.M{"$lte": filter.EndTimestamp}}})
	}

	// Sort the results by the specified field.
	if filter.SortField != "" {
		sortDirection := 1
		if filter.SortDescending {
			sortDirection = -1
		}
		pipeline = append(pipeline, bson.M{"$sort": bson.M{filter.SortField: sortDirection}})
	}

	// Paginate the results.
	if filter.PageSize > 0 {
		pipeline = append(pipeline, bson.M{"$skip": filter.PageSize * (filter.Page - 1)})
		pipeline = append(pipeline, bson.M{"$limit": filter.PageSize})
	}

	// Execute the pipeline and retrieve the results.
	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, utils.Metadata{}, err
	}

	var events []*model.AuditEvent
	err = cursor.All(ctx, &events)
	if err != nil {
		return nil, utils.Metadata{}, err
	}

	// Retrieve the total number of documents that match the filter criteria.
	count, err := collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, utils.Metadata{}, err
	}

	// Return the results along with metadata about the pagination.
	metadata := utils.Metadata{
		TotalRecords: int(count),
		CurrentPage:  filter.Page,
		PageSize:     filter.PageSize,
	}
	return events, metadata, nil
}
