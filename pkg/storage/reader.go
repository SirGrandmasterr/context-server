package storage

import (
	"Llamacommunicator/pkg/entities"
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

type StorageReader struct {
	Log *zap.SugaredLogger
	Db  *mongo.Database
}

func NewStorageReader(log *zap.SugaredLogger, db *mongo.Database) *StorageReader {
	return &StorageReader{
		Log: log,
		Db:  db,
	}
}

func (strg *StorageReader) ReadActionOptionEntity(action entities.Action, ctx context.Context) error {
	actionCollection := strg.Db.Collection("actions")
	opts := options.Update().SetUpsert(true)
	filter := bson.D{{Key: `action_name`, Value: action.ActionName}}
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "action_name", Value: action.ActionName}, {Key: "description", Value: action.Description}}}}
	_, err := actionCollection.UpdateOne(context.Background(), filter, update, opts)
	if err != nil {
		strg.Log.Errorln("Error inserting Action Option", err)
	}
	return nil
}
