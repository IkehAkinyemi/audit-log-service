package mongodb

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	testDBURI = "mongodb+srv://IkehAkinyemi:kwZN74bwGnvsO4xJ@cluster0.wqrmh26.mongodb.net/test?retryWrites=true&w=majority"
)

var logRes = new(LogRepository)
var tokenRes = new(TokenRepository)

func TestMain(m *testing.M) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	serverAPIOptions := options.ServerAPI(options.ServerAPIVersion1)
	clientOptions := options.Client().
		ApplyURI(testDBURI).
		SetServerAPIOptions(serverAPIOptions)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal("Can't connect to DB", err)
	}

	logRes = NewLogRepository(client)
	tokenRes = NewTokenRepository(client)

	os.Exit(m.Run())
}
