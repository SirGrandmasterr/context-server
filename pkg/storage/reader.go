package storage

import (
	"Llamacommunicator/pkg/entities"
	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
	err := actionCollection.FindOne(ctx, bson.M{"action_name": name}).Decode(&action)
	if err != nil {
		strg.Log.Panicln("Error in ReadActionOptionEntity", err)
		return entities.Action{}, err
	}
	return action, nil
}

func (strg *StorageReader) ReadBasePrompt(name string, ctx context.Context) (entities.BasePrompt, error) {
	basePromptCollection := strg.Db.Collection("baseprompts")
	var prompt entities.BasePrompt
	err := basePromptCollection.FindOne(ctx, bson.M{"prompt_name": name}).Decode(&prompt)
	if err != nil {
		strg.Log.Panicln("Error in ReadBasePrompt", err)
		return entities.BasePrompt{}, err
	}
	return prompt, nil
}

func (strg *StorageReader) ReadSingleObject(name string, ctx context.Context) (entities.RelevantObject, error) {
	basePromptCollection := strg.Db.Collection("objects")
	var relObject entities.RelevantObject
	err := basePromptCollection.FindOne(ctx, bson.M{"objectName": name}).Decode(&relObject)
	if err != nil {
		strg.Log.Panicln("Error in ReadBasePrompt", err)
		return entities.RelevantObject{}, err
	}
	return relObject, nil
}

func (strg *StorageReader) ReadAllObjects(ctx context.Context) ([]entities.RelevantObject, error) {
	objCollection := strg.Db.Collection("objects")
	var obs []entities.RelevantObject
	cursor, err := objCollection.Find(context.Background(), bson.M{})
	if err = cursor.All(ctx, &obs); err != nil {
		log.Fatal(err)
		return obs, err
	}
	return obs, nil
}

func (strg *StorageReader) ReadPlayer(username string, ctx context.Context) (entities.Player, error) {
	playerCollection := strg.Db.Collection("players")
	strg.Log.Infoln("Searching for Player: ", username)
	var player entities.Player
	err := playerCollection.FindOne(ctx, bson.D{{Key: "username", Value: username}}).Decode(&player)
	if err != nil {
		strg.Log.Panicln("Error in ReadPlayer", err)
		return entities.Player{}, err
	}
	return player, nil

}

func (strg *StorageReader) ReadAllLocations(ctx context.Context) ([]entities.Location, error) {
	locationCollection := strg.Db.Collection("locations")
	var locs []entities.Location
	cursor, err := locationCollection.Find(context.Background(), bson.M{})
	if err = cursor.All(ctx, &locs); err != nil {
		log.Fatal(err)
		return locs, err
	}
	return locs, nil
}

func (strg *StorageReader) ReadLocation(name string, ctx context.Context) (entities.Location, error) {
	locationCollection := strg.Db.Collection("locations")
	var loc entities.Location
	err := locationCollection.FindOne(ctx, bson.D{{Key: "location_name", Value: name}}).Decode(&loc)
	if err != nil {
		strg.Log.Panicln("Error in ReadPlayer", err)
		return entities.Location{}, err
	}
	return loc, nil
}

func (strg *StorageReader) ReadActionToken(ctx context.Context, id primitive.ObjectID) (entities.ActionToken, error) {
	actionTokenCollection := strg.Db.Collection("actiontokens")
	var tok entities.ActionToken
	err := actionTokenCollection.FindOne(ctx, bson.D{{Key: "_id", Value: id}}).Decode(&tok)
	if err != nil {
		strg.Log.Panicln("Error in ReadPlayer", err)
		return entities.ActionToken{}, err
	}
	return tok, nil
}

func (strg *StorageReader) ReadMaterials(mats []string, ac entities.AssistantContext, ctx context.Context) ([]entities.Material, error) {
	m := make(map[string]bool)
	for _, mat := range mats {
		m[mat] = true
	}
	var s []entities.Material
	if m["focus"] {
		ob, err := strg.ReadSingleObject(ac.FocusedAsset, ctx)
		if err != nil {
			strg.Log.Errorln("Could not retrieve focused object material")
		}
		s = append(
			s,
			entities.Material{
				Type:        "focus",
				Name:        ob.ObjectName,
				Description: ob.Description,
			},
		)
	}
	if m["options"] {
		for _, avac := range ac.AvailableActions {
			action, err := strg.ReadActionOptionEntity(avac, ctx)
			if err != nil {
				strg.Log.Errorln("Could not retrieve actionOption Material")
			}
			s = append(
				s,
				entities.Material{
					Type:        "options",
					Name:        action.ActionName,
					Description: action.Description,
				},
			)
		}
	}
	if m["locations"] {
		locs, err := strg.ReadAllLocations(ctx)
		if err != nil {
			strg.Log.Errorln("Could not retrieve Locations Material")
		}
		for _, avac := range locs {
			s = append(
				s,
				entities.Material{
					Type:        "options",
					Name:        avac.LocationName,
					Description: avac.Description,
				},
			)
		}
	}
	if m["objects"] {
		obs, err := strg.ReadAllObjects(ctx)
		if err != nil {
			strg.Log.Errorln("Could not retrieve Objects Material")
		}
		for _, avac := range obs {
			s = append(
				s,
				entities.Material{
					Type:        "options",
					Name:        avac.ObjectName,
					Description: avac.Description,
				},
			)
		}

	}
	return s, nil
}

/*func (strg *StorageReader) ReadPlayerHistory(id string, ctx context.Context) (entities.PlayerContext, error) {
	re
}*/
