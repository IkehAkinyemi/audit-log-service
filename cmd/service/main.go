package main

import (
	"context"
	"log"
	"os"
	"sync"
	"time"

	"github.com/IkehAkinyemi/logaudit/internal/jsonlog"
	"github.com/IkehAkinyemi/logaudit/internal/repository/mongodb"
	"github.com/IkehAkinyemi/logaudit/internal/utils"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Holds the application logic and dependencies
type service struct {
	logger    *jsonlog.Logger
	config    utils.Config
	logs      *mongodb.LogRepository
	tokens    *mongodb.TokenRepository
	msgBroker *msgBroker
	wg        sync.WaitGroup
}

func main() {
	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	config, err := utils.ParseConfig()
	if err != nil {
		logger.PrintFatal(err, nil)
		return
	}

	// connect message broker
	conn, err := amqp.Dial(config.AMQP_CONN_URI)
	if err != nil {
		logger.PrintFatal(err, nil)
		return
	}
	logger.PrintInfo("connection to message broker established", nil)
	defer conn.Close()

	// connect DB
	client, err := connectDB(config.DBConnURI)
	if err != nil {
		logger.PrintFatal(err, nil)
		return
	}
	logger.PrintInfo("database connection established", nil)
	defer closeDB(client)

	msgBroker, err := newMsgBroker(conn, "logs")
	if err != nil {
		logger.PrintFatal(err, nil)
		return
	}

	service := &service{
		logger:    logger,
		config:    *config,
		logs:      mongodb.NewLogRepository(client),
		tokens:    mongodb.NewTokenRepository(client),
		msgBroker: msgBroker,
	}

	go service.processLogs()

	err = service.serve()
	if err != nil {
		logger.PrintFatal(err, nil)
	}
}

// connectDB establishes connection to MongoDB
func connectDB(connURI string) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	serverAPIOptions := options.ServerAPI(options.ServerAPIVersion1)
	clientOptions := options.Client().
		ApplyURI(connURI).
		SetServerAPIOptions(serverAPIOptions)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	// Check the connection
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatal(err)
	}

	return client, err
}

// closeDB close database connection.
func closeDB(client *mongo.Client) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client.Disconnect(ctx)
}
