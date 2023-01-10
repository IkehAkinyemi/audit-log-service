package mongodb

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"errors"

	"github.com/IkehAkinyemi/logaudit/internal/repository/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	tokensCollection = "tokens"
)

// NewAPIToken cryptographically secure random value
// for authentication token.
func generateAPIToken(serviceID string) (*model.Token, error) {
	token := &model.Token{
		ServiceID: model.ServiceID(serviceID),
	}

	randomByte := make([]byte, 16)

	_, err := rand.Read(randomByte)
	if err != nil {
		return nil, err
	}

	token.Plaintext = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomByte)

	hash := sha256.Sum256([]byte(token.Plaintext))
	token.Hash = hash[:]

	return token, nil
}

// NewAPIToken generates a new API token, stores to the DB,
// and returns it to the caller.
func (r *Repository) NewAPIKey(ctx context.Context, serviceID string) (*model.Token, error) {
	token, err := generateAPIToken(serviceID)
	if err != nil {
		return nil, err
	}

	err = r.AddNewToken(ctx, token)
	return token, err
}

// AddNewToken adds a new token record to the tokens collection.
func (r *Repository) AddNewToken(ctx context.Context, token *model.Token) error {
	collection := r.client.Database(db).Collection(tokensCollection)
	record := bson.D{
		{Key: "Hash", Value: string(token.Hash)},
		{Key: "ServiceID", Value: token.ServiceID},
	}
	_, err := collection.InsertOne(ctx, record)
	return err
}

// UpdateToken updates a token; perceived as resetting a token.
func (r *Repository) UpdateToken(ctx context.Context, serviceID model.ServiceID) (*model.Token, error) {
	token, err := generateAPIToken(string(serviceID))
	if err != nil {
		return nil, err
	}

	collection := r.client.Database(db).Collection(tokensCollection)
	result, err := collection.UpdateOne(
		ctx,
		bson.M{"ServiceID": serviceID},
		bson.D{
			{Key: "$set", Value: bson.D{
				{Key: "Hash", Value: string(token.Hash)},
			}},
			{Key: "$currentDate", Value: bson.D{
				{Key: "lastModified", Value: true},
			}},
		},
	)

	switch {
	case result.MatchedCount == 0:
		return nil, model.ErrRecordNotFound
	}

	return token, err
}

// GetServiceIDTokenBy retrieves a token by its API key.
func (r *Repository) GetTokenByAPIKey(ctx context.Context, tokenPlaintext string) (*model.ServiceID, error) {
	tokenHash := sha256.Sum256([]byte(tokenPlaintext))
	var token model.Token

	collection := r.client.Database(db).Collection(tokensCollection)
	filter := bson.M{"Hash": string(tokenHash[:])}
	projection := bson.D{
		{Key: "ServiceID", Value: 1},
	}
	err := collection.FindOne(ctx, filter, options.FindOne().SetProjection(projection)).Decode(&token)

	if err != nil {
		switch {
		case errors.Is(err, mongo.ErrNoDocuments):
			return nil, model.ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &token.ServiceID, nil
}
