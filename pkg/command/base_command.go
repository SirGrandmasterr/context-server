package command

import (
	"Llamacommunicator/pkg/config"

	"github.com/gofiber/storage/mongodb/v2"
	"go.mongodb.org/mongo-driver/mongo"

	"go.uber.org/zap"
)

// BaseCommand holds common command properties
type BaseCommand struct {
	Config *config.Specification
	Log    *zap.SugaredLogger
	Db     *mongo.Database
}

// NewBaseCommand creates a structure with common shared properties of the commands
func NewBaseCommand(config *config.Specification, logger *zap.SugaredLogger) BaseCommand {
	bc := BaseCommand{
		Config: config,
		Log:    logger,
		Db:     &mongo.Database{},
	}
	bc.Db = bc.NewDatabaseConnection()
	return bc
}

func (cmd *BaseCommand) NewDatabaseConnection() *mongo.Database {

	store := mongodb.New(mongodb.Config{
		ConnectionURI: cmd.Config.DBConnLink,
		Database:      "llamadrama",
		Collection:    "llama_storage",
		Reset:         false,
	})

	return store.Conn()

}
