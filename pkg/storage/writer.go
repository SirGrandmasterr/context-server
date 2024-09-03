package storage

import (
	"Llamacommunicator/pkg/entities"
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

type StorageWriter struct {
	Log *zap.SugaredLogger
	Db  *mongo.Database
}

func NewStorageWriter(log *zap.SugaredLogger, db *mongo.Database) *StorageWriter {
	return &StorageWriter{
		Log: log,
		Db:  db,
	}
}

func (strg *StorageWriter) SaveActionOptionEntity(action entities.Action, ctx context.Context) error {
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

func (strg *StorageWriter) SavePlayerContext(collection string, action entities.PlayerContext, ctx context.Context) error {
	actionCollection := strg.Db.Collection("actions")
	opts := options.Update().SetUpsert(true)
	_, err := actionCollection.UpdateOne(ctx, nil, action, opts)
	if err != nil {
		return err
	}
	return nil
}

func boolHelper(val bool) *bool {
	return &val
}
