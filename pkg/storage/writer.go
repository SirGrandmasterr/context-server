package storage

import (
	"Llamacommunicator/pkg/entities"
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

func (strg *StorageWriter) SaveActionOptionEntity2(action entities.Action, ctx context.Context) error {
	actionCollection := strg.Db.Collection("actions")
	_, err := actionCollection.InsertOne(ctx, action)
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

func (strg *StorageWriter) SaveObject(object entities.RelevantObject, ctx context.Context) error {
	objectCollection := strg.Db.Collection("objects")
	opts := options.Update().SetUpsert(true)

	filter := bson.D{{Key: `object_name`, Value: object.ObjectName}}
	update := bson.D{{Key: "$set", Value: bson.D{
		{Key: "object_name", Value: object.ObjectName},
		{Key: "object_type", Value: object.ObjectType},
		{Key: "object_location", Value: object.ObjectLocation},
		{Key: "description", Value: object.Description},
		{Key: "actions", Value: object.Artist}}}}
	_, err := objectCollection.UpdateOne(context.Background(), filter, update, opts)
	if err != nil {
		strg.Log.Errorln("Error inserting Object", err)
	}
	return nil
}

func (strg *StorageWriter) SaveBasePrompt(prompt entities.BasePrompt, ctx context.Context) error {
	BasePromptCollection := strg.Db.Collection("baseprompts")
	opts := options.Update().SetUpsert(true)

	filter := bson.D{{Key: `prompt_name`, Value: prompt.PromptName}}
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "prompt_name", Value: prompt.PromptName}, {Key: "prompt", Value: prompt.Prompt}}}}
	_, err := BasePromptCollection.UpdateOne(context.Background(), filter, update, opts)
	if err != nil {
		strg.Log.Errorln("Error inserting Object", err)
	}
	return nil
}

func (strg *StorageWriter) SaveLocations(loc entities.Location, ctx context.Context) error {
	LocationCollection := strg.Db.Collection("locations")
	opts := options.Update().SetUpsert(true)

	filter := bson.D{{Key: `location_name`, Value: loc.LocationName}}
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "location_name", Value: loc.LocationName}, {Key: "description", Value: loc.Description}}}}
	_, err := LocationCollection.UpdateOne(context.Background(), filter, update, opts)
	if err != nil {
		strg.Log.Errorln("Error inserting Object", err)
	}
	return nil
}

func (strg *StorageWriter) SaveActionToken(loc entities.ActionToken, ctx context.Context) (primitive.ObjectID, error) {
	strg.Log.Infoln("About to create Collection?")
	ActionTokenCollection := strg.Db.Collection("actiontokens")
	strg.Log.Infoln("About to insert ", loc.ID)
	insert, err := ActionTokenCollection.InsertOne(context.Background(), loc)
	if err != nil {
		strg.Log.Errorln("Error inserting Object", err)
	}
	strg.Log.Infoln("Inserted Actiontoken ", loc.ID)
	id := insert.InsertedID.(primitive.ObjectID)
	strg.Log.Infoln("returning id: ", id)
	return id, err
}

func (strg *StorageWriter) DeleteActionToken(id primitive.ObjectID, ctx context.Context) error {
	actionTokenCollection := strg.Db.Collection("actiontokens")
	actionTokenCollection.DeleteOne(ctx, bson.D{{Key: "_id", Value: id}})
	return nil
}

func (strg *StorageWriter) UpdateActionTokenStage(id primitive.ObjectID, currentStage int, ctx context.Context) error {
	actionTokenCollection := strg.Db.Collection("actiontokens")

	filter := bson.D{{Key: `_id`, Value: id}}
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "currentStage", Value: currentStage}}}}
	_, err := actionTokenCollection.UpdateOne(context.Background(), filter, update, nil)
	if err != nil {
		strg.Log.Errorln("Error inserting Object", err)
	}
	return nil
}

func (strg *StorageWriter) SavePlayers(pl entities.Player, ctx context.Context) error {
	LocationCollection := strg.Db.Collection("players")
	opts := options.Update().SetUpsert(true)

	filter := bson.D{{Key: `username`, Value: pl.Username}}
	update := bson.D{{Key: "$set", Value: bson.D{
		{Key: "password", Value: pl.Password},
		{Key: "username", Value: pl.Username},
		{Key: "_id", Value: pl.ID},
		{Key: "historyarray", Value: pl.HistoryArray},
	}}}
	_, err := LocationCollection.UpdateOne(context.Background(), filter, update, opts)
	if err != nil {
		strg.Log.Errorln("Error inserting Object", err)
	}
	return nil
}

func (strg *StorageWriter) UpdatePlayerHistory(username string, history string) error {

	playerCollection := strg.Db.Collection("players")
	filter := bson.D{{Key: `username`, Value: username}}
	update := bson.D{{Key: "$set", Value: bson.D{
		{Key: "history", Value: history},
	}}}
	_, err := playerCollection.UpdateOne(context.Background(), filter, update, nil)
	if err != nil {
		strg.Log.Errorln("Error updating Player History", err)
	}
	return nil
}

func (strg *StorageWriter) PushPlayerHistoryElement(username string, historyElement string) error {
	strg.Log.Infoln("Inserted ", username, "'s history entry ", historyElement)
	playerCollection := strg.Db.Collection("players")
	filter := bson.D{{Key: `username`, Value: username}}
	update := bson.D{{Key: "$push", Value: bson.D{
		{Key: "historyarray", Value: historyElement},
	}}}
	_, err := playerCollection.UpdateOne(context.Background(), filter, update, nil)
	if err != nil {
		strg.Log.Errorln("Error updating Player History", err)
	}
	return nil
}

func (strg *StorageWriter) ResetPlayerHistory(username string) error {

	playerCollection := strg.Db.Collection("players")
	newHistory := []string{""}
	filter := bson.D{{Key: `username`, Value: username}}
	update := bson.D{{Key: "$set", Value: bson.D{
		{Key: "history", Value: ""},
		{Key: "historyarray", Value: newHistory},
	}}}
	_, err := playerCollection.UpdateOne(context.Background(), filter, update, nil)
	if err != nil {
		strg.Log.Errorln("Error wiping Player History", err)
	}
	return nil
}

func (strg *StorageWriter) BoolHelper(val bool) *bool {
	return &val
}
