package command

import (
	"Llamacommunicator/pkg/entities"
	"Llamacommunicator/pkg/storage"
	"context"
	"encoding/json"
	"io"
	"os"

	"github.com/urfave/cli"
)

type MigrateCommand struct {
	BaseCommand
	storageWriter storage.StorageWriter
}

func NewMigrateCommand(baseCommand BaseCommand) *MigrateCommand {
	return &MigrateCommand{
		BaseCommand:   baseCommand,
		storageWriter: *storage.NewStorageWriter(baseCommand.Log, baseCommand.Db),
	}
}

func (cmd *MigrateCommand) Run(cliCtx *cli.Context) error {
	err := cmd.BaseCommand.Db.CreateCollection(context.Background(), "actions")
	if err != nil {
		cmd.Log.Errorln(err)
	}
	jsonFile, err := os.Open(`./pkg/storage/init_db.json`)

	if err != nil {
		cmd.Log.Errorln(err)
	}

	byteValue, _ := io.ReadAll(jsonFile)

	// we initialize our Users array
	var initJson entities.InitJson

	// we unmarshal our byteArray which contains our
	// jsonFile's content into 'users' which we defined above
	err = json.Unmarshal(byteValue, &initJson)
	if err != nil {
		cmd.Log.Errorln(err)
	}
	cmd.Log.Infoln("We have unmarshalled.")
	cmd.Log.Infoln("Uploading ActionOption Entities")
	cmd.Log.Infoln(initJson.Actions)
	for _, action := range initJson.Actions {
		err = cmd.storageWriter.SaveActionOptionEntity2(action, context.Background())
		if err != nil {
			cmd.Log.Errorln(err)
		}
	}
	cmd.Log.Infoln("Uploading Objects")
	for _, obj := range initJson.Objects {
		err = cmd.storageWriter.SaveObject(obj, context.Background())
		if err != nil {
			cmd.Log.Errorln(err)
		}
	}
	cmd.Log.Infoln("Uploading Baseprompts")
	for _, obj := range initJson.BasePrompts {
		err = cmd.storageWriter.SaveBasePrompt(obj, context.Background())
		if err != nil {
			cmd.Log.Errorln(err)
		}
	}
	cmd.Log.Infoln("Uploading Locations")
	for _, obj := range initJson.Locations {
		err = cmd.storageWriter.SaveLocations(obj, context.Background())
		if err != nil {
			cmd.Log.Errorln(err)
		}
	}
	cmd.Log.Infoln("Creating Player")
	for _, obj := range initJson.Players {
		err = cmd.storageWriter.SavePlayers(obj, context.Background())
		if err != nil {
			cmd.Log.Errorln(err)
		}
	}

	defer jsonFile.Close()
	return nil

}
