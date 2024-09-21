package storage

import (
	"Llamacommunicator/pkg/entities"
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
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

func (strg *StorageReader) ReadActionOptionEntity(name string, ctx context.Context) (entities.Action, error) {
	actionCollection := strg.Db.Collection("actions")
	var action entities.Action
	err := actionCollection.FindOne(ctx, bson.M{"actionname": name}).Decode(&action)
	if err != nil {
		strg.Log.Panicln("Error in ReadActionOptionEntity", err)
		return entities.Action{}, err
	}
	return action, nil
}

/*func (strg *StorageReader) ReadPlayerHistory(id string, ctx context.Context) (entities.PlayerContext, error) {
	re
}*/
