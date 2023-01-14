package main

import (
	"context"
	"expvar"
	"flag"
	"log"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/IkehAkinyemi/logaudit/internal/jsonlog"
	"github.com/IkehAkinyemi/logaudit/internal/repository/mongodb"
	"github.com/IkehAkinyemi/logaudit/internal/utils"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Holds the application logic and dependencies
type service struct {
	logger *jsonlog.Logger
	config utils.Config
	db     *mongodb.Repository
	wg     sync.WaitGroup
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

func main() {
	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	// Parse flag argument
	var cfgFile string
	flag.StringVar(&cfgFile, "cfg-file", "", "Directory path to configuration file")
	flag.Parse()

	config, err := utils.GetConfig(cfgFile)
	if err != nil {
		logger.PrintFatal(err, nil)
		return
	}

	client, err := connectDB(config.MongoDB.ConnURI)
	if err != nil {
		logger.PrintFatal(err, nil)
		return
	}
	logger.PrintInfo("database connection established", nil)
	defer closeDB(client)

	// Configuring metrics using expvar
	expvar.Publish("goroutines", expvar.Func(func() any {
		return runtime.NumGoroutine()
	}))
	expvar.Publish("timestamp", expvar.Func(func() any {
		// records the current Unix timestamp when metrics was taken
		return time.Now().Unix()
	}))

	service := &service{
		logger: logger,
		config: *config,
		db:     mongodb.New(client),
	}

	err = service.serve()

	if err != nil {
		logger.PrintFatal(err, nil)
	}
}
