package storage

import (
	"Llamacommunicator/pkg/entities"
	"context"

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
		strg.Log.Errorln("Error in ReadActionOptionEntity for ", name, err)
		return entities.Action{}, err
	}
	return action, nil
}

func (strg *StorageReader) ReadBasePrompt(name string, ctx context.Context) (entities.BasePrompt, error) {
	basePromptCollection := strg.Db.Collection("baseprompts")
	var prompt entities.BasePrompt
	err := basePromptCollection.FindOne(ctx, bson.M{"prompt_name": name}).Decode(&prompt)
	if err != nil {
		strg.Log.Errorln("Error in ReadBasePrompt", err)
		return entities.BasePrompt{}, err
	}
	return prompt, nil
}

func (strg *StorageReader) ReadSingleObject(name string, ctx context.Context) (entities.RelevantObject, error) {
	basePromptCollection := strg.Db.Collection("objects")
	var relObject entities.RelevantObject
	strg.Log.Infoln("Trying to read ", name)
	err := basePromptCollection.FindOne(ctx, bson.M{"object_name": name}).Decode(&relObject)
	if err != nil {
		strg.Log.Errorln("Error in ReadSingleObject", err)
		return entities.RelevantObject{}, err
	}
	return relObject, nil
}

func (strg *StorageReader) ReadAllObjects(ctx context.Context) ([]entities.RelevantObject, error) {
	objCollection := strg.Db.Collection("objects")
	var obs []entities.RelevantObject
	cursor, err := objCollection.Find(context.Background(), bson.M{})
	if err != nil {
		strg.Log.Errorln(err)
		return obs, err
	}
	if err = cursor.All(ctx, &obs); err != nil {
		strg.Log.Errorln(err)
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
		strg.Log.Errorln("Error in ReadPlayer", err)
		return entities.Player{}, err
	}
	return player, nil

}

func (strg *StorageReader) ReadAllLocations(ctx context.Context) ([]entities.Location, error) {
	locationCollection := strg.Db.Collection("locations")
	var locs []entities.Location
	cursor, err := locationCollection.Find(context.Background(), bson.M{})
	if err != nil {
		return locs, err
	}
	if err = cursor.All(ctx, &locs); err != nil {
		strg.Log.Errorln(err)
		return locs, err
	}
	return locs, nil
}

func (strg *StorageReader) ReadLocation(name string, ctx context.Context) (entities.Location, error) {
	locationCollection := strg.Db.Collection("locations")
	var loc entities.Location
	err := locationCollection.FindOne(ctx, bson.D{{Key: "location_name", Value: name}}).Decode(&loc)
	if err != nil {
		strg.Log.Errorln("Error in ReadLocation", err)
		return entities.Location{}, err
	}
	return loc, nil
}

func (strg *StorageReader) ReadActionToken(ctx context.Context, id primitive.ObjectID) (entities.ActionToken, error) {
	actionTokenCollection := strg.Db.Collection("actiontokens")
	var tok entities.ActionToken
	err := actionTokenCollection.FindOne(ctx, bson.D{{Key: "_id", Value: id}}).Decode(&tok)
	if err != nil {
		strg.Log.Errorln("Error in ReadActionToken", err)
		return entities.ActionToken{}, err
	}
	return tok, nil
}

func (strg *StorageReader) ReadPlayerHistory(maxlen int, username string) []string {
	pl, err := strg.ReadPlayer(username, context.Background())
	if err != nil {
		strg.Log.Errorln("Error in ReadPlayer during ReadPlayerHistory", err)
	}
	sub := 0
	length := len(pl.HistoryArray)
	arr := pl.HistoryArray
	if length < maxlen {
		strg.Log.Infoln("Player history has less than 50 entries, using all of it.")
		return arr
	} else {
		strg.Log.Infoln("Returning last 50 entries.")
		sub = maxlen
		h := arr[len(arr)-sub:]
		return h
	}

}

func (strg *StorageReader) ReadMaterials(mats []string, ac entities.AssistantContext, pc entities.PlayerContext, ctx context.Context) ([]entities.Material, error) {
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
					Type:        "locations",
					Name:        avac.LocationName,
					Description: avac.Description,
				},
			)
		}
	}
	if m["assistantAssetsInView"] {
		for _, asset := range ac.AssetsInView {
			obj, err := strg.ReadSingleObject(asset, context.Background())
			if err != nil {
				strg.Log.Errorln("Error reading single Object for assistantAssetsInView")
			}
			s = append(
				s,
				entities.Material{
					Type:        "assistantAssetsInView",
					Name:        obj.ObjectName,
					Description: obj.Description,
				},
			)
		}
	}
	if m["playerAssetsInView"] {
		for _, asset := range pc.AssetsInView {
			obj, err := strg.ReadSingleObject(asset, context.Background())
			if err != nil {
				strg.Log.Errorln("Error reading single Object for assistantAssetsInView")
			}
			s = append(
				s,
				entities.Material{
					Type:        "playerAssetsInView",
					Name:        obj.ObjectName,
					Description: obj.Description,
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
					Type:        "objects",
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
